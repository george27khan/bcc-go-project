package server

import (
	"bcc-go-project/internal/domain/entity"
	rep_err "bcc-go-project/internal/infrastructure/repository/errors_repo"
	"bufio"
	"context"
	"errors"
	"fmt"
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
		expectFailResp  ErrorResponse
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
			ctx:        context.Background(),
			req:        GetDownloadsIdRequestObject{},
			expectType: GetDownloadsId500JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    INTERNALSERVERERROR,
				Message: "GetDownloadsId: context canceled",
			},
		},
		{
			name: "GetTask error",
			prepare: func(tt *TestCase, m *mockGetTask) {
				m.UseCase.EXPECT().GetTask(gomock.Any(), entity.IdTask(tt.req.Id)).Return(
					nil,
					rep_err.ErrTaskNotExist,
				)
			},
			ctx:        context.Background(),
			req:        GetDownloadsIdRequestObject{},
			expectType: GetDownloadsId404JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    NOTFOUND,
				Message: fmt.Errorf("GetDownloadsId: %w", rep_err.ErrTaskNotExist).Error(),
			},
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
			case GetDownloadsId400JSONResponse:
				val, ok := resp.(GetDownloadsId400JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, val.Code, tt.expectFailResp.Code)
				require.Equal(t, val.Message, tt.expectFailResp.Message)
			case GetDownloadsId404JSONResponse:
				val, ok := resp.(GetDownloadsId404JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, val.Code, tt.expectFailResp.Code)
				require.Equal(t, val.Message, tt.expectFailResp.Message)
			case GetDownloadsId500JSONResponse:
				val, ok := resp.(GetDownloadsId500JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, val.Code, tt.expectFailResp.Code)
				require.Equal(t, val.Message, tt.expectFailResp.Message)
			default:
				require.Fail(t, "unexpected response type")
			}

		})
	}
}

type mockTaskFileUseCase struct {
	UseCase *MockTaskFileUseCase
}

func TestGetDownloadsIdFilesFileId(t *testing.T) {
	type TestCase struct {
		name           string
		prepare        func(tt *TestCase, m *mockTaskFileUseCase)
		ctx            context.Context
		req            GetDownloadsIdFilesFileIdRequestObject
		expectType     any
		expectedData   []byte
		expectFailResp ErrorResponse
		expectedErr    error
	}
	TestCases := []*TestCase{
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockTaskFileUseCase) {
				m.UseCase.EXPECT().GetTaskFile(gomock.Any(), entity.IdTask(tt.req.Id), entity.IdFile(tt.req.FileId)).Return(
					[]byte("Mock"), nil)
			},
			ctx:          context.Background(),
			req:          GetDownloadsIdFilesFileIdRequestObject{Id: 1, FileId: 1},
			expectType:   GetDownloadsIdFilesFileId200ApplicationoctetStreamResponse{},
			expectedData: []byte("Mock"),
			expectedErr:  nil,
		},
		//{
		//	name: "context canceled",
		//	prepare: func(tt *TestCase, m *mockTaskFileUseCase) {
		//		var cancel context.CancelFunc
		//		tt.ctx, cancel = context.WithCancel(tt.ctx)
		//		cancel()
		//	},
		//	ctx:        context.Background(),
		//	req:        GetDownloadsIdRequestObject{},
		//	expectType: GetDownloadsId500JSONResponse{},
		//	expectFailResp: ErrorResponse{
		//		Code:    INTERNALSERVERERROR,
		//		Message: "GetDownloadsId: context canceled",
		//	},
		//},
		//{
		//	name: "GetTask error",
		//	prepare: func(tt *TestCase, m *mockTaskFileUseCase) {
		//		m.UseCase.EXPECT().GetTask(gomock.Any(), entity.IdTask(tt.req.Id)).Return(
		//			nil,
		//			rep_err.ErrTaskNotExist,
		//		)
		//	},
		//	ctx:        context.Background(),
		//	req:        GetDownloadsIdRequestObject{},
		//	expectType: GetDownloadsId404JSONResponse{},
		//	expectFailResp: ErrorResponse{
		//		Code:    NOTFOUND,
		//		Message: fmt.Errorf("GetDownloadsId: %w", rep_err.ErrTaskNotExist).Error(),
		//	},
		//},
	}
	for _, tt := range TestCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mGet := NewMockTaskGetUseCase(ctrl)
			mFile := NewMockTaskFileUseCase(ctrl)
			mCreate := NewMockTaskCreateUseCase(ctrl)

			ts := NewTaskServer(mCreate, mGet, mFile)

			m := &mockTaskFileUseCase{mFile}
			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			resp, err := ts.GetDownloadsIdFilesFileId(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			switch respType := resp.(type) {
			case GetDownloadsIdFilesFileId200ApplicationoctetStreamResponse:
				_ = respType
				val, ok := resp.(GetDownloadsIdFilesFileId200ApplicationoctetStreamResponse)
				require.Equal(t, ok, true)
				r := bufio.NewScanner(val.Body)
				r.Scan()
				require.Equal(t, tt.expectedData, r.Bytes())
			case GetDownloadsIdFilesFileId400JSONResponse:
				val, ok := resp.(GetDownloadsIdFilesFileId400JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, val.Code, tt.expectFailResp.Code)
				require.Equal(t, val.Message, tt.expectFailResp.Message)
			case GetDownloadsIdFilesFileId404JSONResponse:
				val, ok := resp.(GetDownloadsIdFilesFileId404JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, val.Code, tt.expectFailResp.Code)
				require.Equal(t, val.Message, tt.expectFailResp.Message)
			case GetDownloadsIdFilesFileId500JSONResponse:
				val, ok := resp.(GetDownloadsIdFilesFileId500JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, val.Code, tt.expectFailResp.Code)
				require.Equal(t, val.Message, tt.expectFailResp.Message)
			default:
				require.Fail(t, "unexpected response type")
			}

		})
	}
}
