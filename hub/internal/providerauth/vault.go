package providerauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

type Secret struct {
	Value   string
	Version string
}
type SecretResolver interface {
	Resolve(context.Context, string, string) (Secret, error)
}
type AWSVault struct{}

// Resolve uses the AWS JSON API with the SDK's credential chain and SigV4
// signer. References are never substituted for decrypted secret values.
// Contract: https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
func (AWSVault) Resolve(ctx context.Context, ref, version string) (Secret, error) {
	unavailable := errors.New("provider secret unavailable")
	if ref == "" || len(ref) > 2048 {
		return Secret{}, unavailable
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil || cfg.Region == "" {
		return Secret{}, unavailable
	}
	endpoint := os.Getenv("SECRETS_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://secretsmanager." + cfg.Region + ".amazonaws.com"
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && os.Getenv("ENVIRONMENT") != "local") {
		return Secret{}, unavailable
	}
	payload := map[string]string{"SecretId": ref}
	if version != "" {
		payload["VersionId"] = version
	}
	body, _ := json.Marshal(payload)
	sum := sha256.Sum256(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Secret{}, unavailable
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "secretsmanager.GetSecretValue")
	cred, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return Secret{}, unavailable
	}
	if err = v4.NewSigner().SignHTTP(ctx, cred, req, hex.EncodeToString(sum[:]), "secretsmanager", cfg.Region, time.Now()); err != nil {
		return Secret{}, unavailable
	}
	client := http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return Secret{}, unavailable
	}
	defer resp.Body.Close()
	var result struct {
		SecretString string
		VersionId    string
	}
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 96*1024)).Decode(&result) != nil || result.SecretString == "" || result.VersionId == "" || (version != "" && version != result.VersionId) {
		return Secret{}, unavailable
	}
	return Secret{Value: result.SecretString, Version: result.VersionId}, nil
}
