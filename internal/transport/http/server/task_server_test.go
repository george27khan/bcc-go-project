package server

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockGetTask struct {
	UseCase *MockTaskGetUseCase
}

func TestGetDownloadsId(t *testing.T) {
	type TestCase struct {
		name            string
		prepare         func(tt *TestCase, m *mockGetTask)
		ctx             context.Context
		req             GetDownloadsIdRequestObject
		expectType      any
		expectedUrlFile UrlFile
		expectedUrlErr  UrlErr
		expectId        IdTask
		expectStatus    TaskStatus
		expectedErr     error
	}
	TestCases := []*TestCase{
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockGetTask) {
				m.UseCase.EXPECT().GetTask(gomock.Any(), entity.IdTask(tt.req.Id)).Return(
					&entity.Task{
						Id:     entity.IdTask(0),
						Status: entity.TaskStatusDone,
						Files: []entity.File{
							{
								Id:  0,
								Url: entity.Url("https://google.com"),
							},
							{
								Error: errors.New("TIMEOUT"),
								Url:   entity.Url("https://google.com"),
							},
						},
					},
					nil,
				)
			},
			ctx:        context.Background(),
			req:        GetDownloadsIdRequestObject{Id: 0},
			expectType: GetDownloadsId200JSONResponse{},
			expectedUrlFile: UrlFile{
				FileId: 0,
				Url:    "https://google.com",
			},
			expectedUrlErr: UrlErr{
				Error: struct {
					Code string `json:"code"`
				}{
					Code: "TIMEOUT",
				},
				Url: "https://google.com",
			},
			expectId:     IdTask(0),
			expectStatus: DONE,
			expectedErr:  nil,
		},
		{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockGetTask) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx:         context.Background(),
			req:         GetDownloadsIdRequestObject{},
			expectType:  GetDownloadsId500JSONResponse{},
			expectedErr: nil,
		},
	}
	for _, tt := range TestCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mGet := NewMockTaskGetUseCase(ctrl)
			mFile := NewMockTaskFileUseCase(ctrl)
			mCreate := NewMockTaskCreateUseCase(ctrl)

			ts := NewTaskServer(mCreate, mGet, mFile)

			m := &mockGetTask{mGet}
			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			resp, err := ts.GetDownloadsId(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			switch respType := resp.(type) {
			case GetDownloadsId200JSONResponse:
				_ = respType
				val, ok := resp.(GetDownloadsId200JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, val.Status, tt.expectStatus)
				require.Equal(t, val.Id, tt.expectId)

				urlFile, err := val.Files[0].AsUrlFile()
				require.NoError(t, err)
				require.Equal(t, urlFile, tt.expectedUrlFile)

				urlErr, err := val.Files[1].AsUrlErr()
				require.NoError(t, err)
				require.Equal(t, urlErr, tt.expectedUrlErr)

				t.Log("GetDownloadsId200JSONResponse", val)
			case GetDownloadsId400JSONResponse:
			default:
				require.Fail(t, "unexpected response type")
			}

		})
	}
}
