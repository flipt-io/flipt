package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.flipt.io/flipt/internal/server/authz/engine/rego/source"
)

type CloudPolicySource struct {
	host   string
	apiKey string
}

func PolicySourceFromCloud(host, apiKey string) *CloudPolicySource {
	return &CloudPolicySource{host, apiKey}
}

func (p *CloudPolicySource) Get(ctx context.Context, seen source.Hash) ([]byte, source.Hash, error) {
	body, etag, err := fetch(ctx, seen, fmt.Sprintf("https://%s/api/auth/policies", p.host), p.apiKey)
	if err != nil {
		return nil, nil, err
	}

	return body, []byte(etag), nil
}

type CloudDataSource struct {
	host   string
	apiKey string
}

func DataSourceFromCloud(host, apiKey string) *CloudDataSource {
	return &CloudDataSource{host, apiKey}
}

func (p *CloudDataSource) Get(ctx context.Context, seen source.Hash) (map[string]any, source.Hash, error) {
	body, etag, err := fetch(ctx, seen, fmt.Sprintf("https://%s/api/auth/roles", p.host), p.apiKey)
	if err != nil {
		return nil, nil, err
	}

	data := map[string]any{}
	return data, []byte(etag), json.Unmarshal(body, &data)
}

func fetch(ctx context.Context, seen source.Hash, url string, apiKey string) ([]byte, source.Hash, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("If-None-Match", string(seen))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode == http.StatusNotModified {
		return nil, nil, source.ErrNotModified
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// get etag header
	etag := resp.Header.Get("ETag")
	if etag == string(seen) {
		return nil, nil, source.ErrNotModified
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, nil, fmt.Errorf("reading response body: %w", err)
	}

	return body, []byte(etag), nil
}
