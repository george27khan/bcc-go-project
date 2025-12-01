package http_loader

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

type HttpLoader struct {
}

func NewHttpLoader() *HttpLoader {
	return &HttpLoader{}
}

func (l *HttpLoader) Load(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("HttpLoader.Load error: %w", err)
	}
	log.Printf("контекст: %v", ctx.Err())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HttpLoader.Load error: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
