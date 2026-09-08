// Package objectstore streams immutable, versioned S3 content. File authorization
// and retention belong to the database catalog; object keys are never authority.
package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

const InlineLimit int64 = 1 << 20
const ObjectLimit int64 = 1 << 30
const PartSize int64 = 8 << 20

var ErrIntegrity = errors.New("objectstore: object integrity or content type mismatch")
var ErrTooLarge = errors.New("objectstore: content exceeds declared limit")

type Client struct {
	s3      *s3.Client
	presign *s3.PresignClient
	Bucket  string
}

func New(ctx context.Context, endpoint, region, bucket string) (*Client, error) {
	cfg, e := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if e != nil {
		return nil, e
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})
	publicClient := client
	if endpoint := os.Getenv("S3_PUBLIC_ENDPOINT"); endpoint != "" {
		publicClient = s3.NewFromConfig(cfg, func(o *s3.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	}
	return &Client{s3: client, presign: s3.NewPresignClient(publicClient), Bucket: bucket}, nil
}
func (c *Client) EnsureBucket(ctx context.Context) error {
	_, e := c.s3.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(c.Bucket)})
	if e != nil {
		var ae smithy.APIError
		if !errors.As(e, &ae) || (ae.ErrorCode() != "BucketAlreadyOwnedByYou" && ae.ErrorCode() != "BucketAlreadyExists") {
			return e
		}
		if _, e = c.s3.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(c.Bucket)}); e != nil {
			return e
		}
	}
	_, e = c.s3.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{Bucket: aws.String(c.Bucket), VersioningConfiguration: &types.VersioningConfiguration{Status: types.BucketVersioningStatusEnabled}})
	return e
}
func (c *Client) Put(ctx context.Context, key string, data []byte, contentType string) error {
	if int64(len(data)) > InlineLimit {
		return ErrTooLarge
	}
	sum := sha256.Sum256(data)
	_, e := c.s3.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), Body: bytes.NewReader(data), ContentType: aws.String(contentType), ContentLength: aws.Int64(int64(len(data))), ChecksumSHA256: aws.String(base64.StdEncoding.EncodeToString(sum[:])), IfNoneMatch: aws.String("*")})
	return e
}
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	out, e := c.s3.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key)})
	if e != nil {
		return nil, e
	}
	defer out.Body.Close()
	if aws.ToInt64(out.ContentLength) > InlineLimit {
		return nil, ErrTooLarge
	}
	body, e := io.ReadAll(io.LimitReader(out.Body, InlineLimit+1))
	if int64(len(body)) > InlineLimit {
		return nil, ErrTooLarge
	}
	return body, e
}

type Metadata struct {
	Key         string
	Version     string
	Size        int64
	SHA256      string
	ContentType string
}

func (c *Client) Head(ctx context.Context, key, version string) (Metadata, error) {
	in := &s3.HeadObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key)}
	if version != "" {
		in.VersionId = aws.String(version)
	}
	out, e := c.s3.HeadObject(ctx, in)
	if e != nil {
		return Metadata{}, e
	}
	return Metadata{Key: key, Version: aws.ToString(out.VersionId), Size: aws.ToInt64(out.ContentLength), ContentType: aws.ToString(out.ContentType)}, nil
}
func contentMatches(declared string, sample []byte) bool {
	if declared == "application/octet-stream" {
		return true
	}
	if declared == "application/json" {
		s := bytes.TrimSpace(sample)
		return len(s) > 0 && (s[0] == '{' || s[0] == '[')
	}
	detected := http.DetectContentType(sample)
	if declared == "text/plain" {
		return len(detected) >= 10 && detected[:10] == "text/plain"
	}
	return detected == declared
}

// Verify streams the exact version and independently hashes bytes. HEAD metadata
// supplied by uploaders is insufficient evidence of checksum or content type.
func (c *Client) Verify(ctx context.Context, key, version string, size int64, checksum, contentType string) (Metadata, error) {
	m, e := c.Head(ctx, key, version)
	if e != nil {
		return m, e
	}
	if m.Size != size || m.Size > ObjectLimit || m.Version == "" || m.Version == "null" || m.ContentType != contentType {
		return m, ErrIntegrity
	}
	out, e := c.s3.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), VersionId: aws.String(m.Version)})
	if e != nil {
		return m, e
	}
	defer out.Body.Close()
	hash := sha256.New()
	reader := io.LimitReader(out.Body, size+1)
	sample := make([]byte, 512)
	n, e := io.ReadFull(reader, sample)
	if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
		return m, e
	}
	sample = sample[:n]
	hash.Write(sample)
	copied, e := io.CopyBuffer(hash, reader, make([]byte, 64*1024))
	if e != nil {
		return m, e
	}
	m.SHA256 = hex.EncodeToString(hash.Sum(nil))
	if int64(n)+copied != size || (checksum != "" && m.SHA256 != checksum) || !contentMatches(contentType, sample) {
		return m, ErrIntegrity
	}
	return m, nil
}

type UploadURL struct {
	PartNumber int32       `json:"part_number,omitempty"`
	URL        string      `json:"url"`
	Headers    http.Header `json:"headers"`
	Size       int64       `json:"size"`
}
type CompletedPart struct {
	PartNumber int32  `json:"part_number"`
	ETag       string `json:"etag"`
}

func (c *Client) UploadURLs(ctx context.Context, key, contentType, checksum string, size int64, ttl time.Duration) (string, []UploadURL, error) {
	presign := c.presign
	if size <= PartSize {
		sum, e := hex.DecodeString(checksum)
		if e != nil {
			return "", nil, e
		}
		out, e := presign.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), ContentType: aws.String(contentType), ContentLength: aws.Int64(size), ChecksumSHA256: aws.String(base64.StdEncoding.EncodeToString(sum)), IfNoneMatch: aws.String("*")}, s3.WithPresignExpires(ttl))
		if e != nil {
			return "", nil, e
		}
		return "", []UploadURL{{URL: out.URL, Headers: out.SignedHeader, Size: size}}, nil
	}
	start, e := c.s3.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), ContentType: aws.String(contentType)})
	if e != nil {
		return "", nil, e
	}
	id := aws.ToString(start.UploadId)
	urls := []UploadURL{}
	for offset, n := int64(0), int32(1); offset < size; offset, n = offset+PartSize, n+1 {
		length := min(PartSize, size-offset)
		out, e := presign.PresignUploadPart(ctx, &s3.UploadPartInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), UploadId: aws.String(id), PartNumber: aws.Int32(n), ContentLength: aws.Int64(length)}, s3.WithPresignExpires(ttl))
		if e != nil {
			return id, nil, e
		}
		urls = append(urls, UploadURL{PartNumber: n, URL: out.URL, Headers: out.SignedHeader, Size: length})
	}
	return id, urls, nil
}
func (c *Client) CompleteMultipart(ctx context.Context, key, id string, parts []CompletedPart, expected int) error {
	if len(parts) != expected || expected < 2 {
		return ErrIntegrity
	}
	in := []types.CompletedPart{}
	for i, p := range parts {
		if p.PartNumber != int32(i+1) || p.ETag == "" {
			return ErrIntegrity
		}
		in = append(in, types.CompletedPart{PartNumber: aws.Int32(p.PartNumber), ETag: aws.String(p.ETag)})
	}
	_, e := c.s3.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), UploadId: aws.String(id), MultipartUpload: &types.CompletedMultipartUpload{Parts: in}, IfNoneMatch: aws.String("*")})
	return e
}
func (c *Client) DownloadURL(ctx context.Context, key, version string, ttl time.Duration) (string, error) {
	if ttl <= 0 || ttl > 5*time.Minute {
		return "", errors.New("objectstore: invalid URL lifetime")
	}
	out, e := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), VersionId: aws.String(version)}, s3.WithPresignExpires(ttl))
	if e != nil {
		return "", e
	}
	return out.URL, nil
}
func (c *Client) DeleteVersion(ctx context.Context, key, version string) error {
	if version == "" {
		return errors.New("objectstore: version required")
	}
	_, e := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), VersionId: aws.String(version)})
	return e
}

// PutStream uses one bounded multipart buffer and never buffers an entire result.
func (c *Client) PutStream(ctx context.Context, key, contentType string, reader io.Reader, maxBytes int64) (Metadata, error) {
	if maxBytes < 1 || maxBytes > ObjectLimit {
		return Metadata{}, ErrTooLarge
	}
	start, e := c.s3.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), ContentType: aws.String(contentType)})
	if e != nil {
		return Metadata{}, e
	}
	id := aws.ToString(start.UploadId)
	completed := false
	defer func() {
		if !completed {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = c.s3.AbortMultipartUpload(cleanup, &s3.AbortMultipartUploadInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), UploadId: aws.String(id)})
		}
	}()
	hash := sha256.New()
	buffer := make([]byte, PartSize)
	parts := []types.CompletedPart{}
	total := int64(0)
	for n := int32(1); ; n++ {
		size, readErr := io.ReadFull(reader, buffer)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return Metadata{}, readErr
		}
		if size == 0 {
			break
		}
		total += int64(size)
		if total > maxBytes {
			return Metadata{}, ErrTooLarge
		}
		if n == 1 && !contentMatches(contentType, buffer[:min(size, 512)]) {
			return Metadata{}, ErrIntegrity
		}
		hash.Write(buffer[:size])
		part, e := c.s3.UploadPart(ctx, &s3.UploadPartInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), UploadId: aws.String(id), PartNumber: aws.Int32(n), Body: bytes.NewReader(buffer[:size]), ContentLength: aws.Int64(int64(size))})
		if e != nil {
			return Metadata{}, e
		}
		parts = append(parts, types.CompletedPart{PartNumber: aws.Int32(n), ETag: part.ETag})
		if readErr != nil {
			break
		}
	}
	if total == 0 {
		return Metadata{}, ErrIntegrity
	}
	out, e := c.s3.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{Bucket: aws.String(c.Bucket), Key: aws.String(key), UploadId: aws.String(id), MultipartUpload: &types.CompletedMultipartUpload{Parts: parts}, IfNoneMatch: aws.String("*")})
	if e != nil {
		return Metadata{}, e
	}
	completed = true
	checksum := hex.EncodeToString(hash.Sum(nil))
	m, e := c.Verify(ctx, key, aws.ToString(out.VersionId), total, checksum, contentType)
	if e != nil {
		return m, fmt.Errorf("objectstore: result custody verification: %w", e)
	}
	return m, nil
}
