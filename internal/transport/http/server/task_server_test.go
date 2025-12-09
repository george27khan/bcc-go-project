package server

import (
	"github.com/golang/mock/gomock"
	"testing"
)

func TestGetDownloadsId(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockTaskGetUseCase(ctrl)

	NewTaskServer()
}
