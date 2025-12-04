package http_loader

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"fmt"
	"io"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpLoader struct {
	client HttpClient
}

func NewHttpLoader(client HttpClient) *HttpLoader {
	return &HttpLoader{client: client}
}

func (l *HttpLoader) Load(ctx context.Context, url entity.Url) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, string(url), nil)
	if err != nil {
		return nil, fmt.Errorf("HttpLoader.Load error: %w", err)
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HttpLoader.Load error: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HttpLoader.Load error: %w", err)
	}
	return data, nil
}
