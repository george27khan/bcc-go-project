package http_loader

import (
	"fmt"
	"io"
	"net/http"
)

type HttpLoader struct {
}

func NewHttpLoader() *HttpLoader {
	return &HttpLoader{}
}

func (l *HttpLoader) Load(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HttpLoader.Load error: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
