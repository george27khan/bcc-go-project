package http_loader

import (
	"bcc-go-project/internal/domain/entity"
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

func (l *HttpLoader) Load(ctx context.Context, url entity.Url) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, string(url), nil)

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
