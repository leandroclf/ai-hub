// Package objectstore encapsula o acesso a objetos grandes em S3
// (DAD-05). Entradas/saidas volumosas ficam no S3; metadados e
// resultados pequenos ficam no PostgreSQL (fora deste pacote). Na
// referencia local/dev, aponta para o LocalStack.
package objectstore

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client encapsula o cliente S3 e o bucket padrao do dominio dono do
// objeto (cada dominio referencia apenas seus proprios objetos).
type Client struct {
	s3     *s3.Client
	Bucket string
}

// New conecta ao endpoint informado (LocalStack em local/dev).
func New(ctx context.Context, endpoint, region, bucket string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		return nil, fmt.Errorf("objectstore: carregar config aws: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
	return &Client{s3: client, Bucket: bucket}, nil
}

// EnsureBucket cria o bucket se ainda nao existir (bootstrap local/dev).
func (c *Client) EnsureBucket(ctx context.Context) error {
	_, err := c.s3.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(c.Bucket)})
	if err != nil {
		// Ja existente e aceitavel em bootstrap idempotente local/dev.
		return nil
	}
	return nil
}

// Put grava um objeto imutavel sob a chave key (DAD-05: versao
// imutavel ou promocao a objeto definitivo sem sobrescrita — esta
// referencia local nao reforca imutabilidade fisica, apenas a
// convencao de nomear por checksum/versao no chamador).
func (c *Client) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("objectstore: gravar objeto %s: %w", key, err)
	}
	return nil
}

// Get le um objeto pela chave.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("objectstore: ler objeto %s: %w", key, err)
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
