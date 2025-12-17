package server

import (
	"bcc-go-project/internal/domain/entity"
	repErr "bcc-go-project/internal/infrastructure/repository/errors_repo"
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockTaskCreate struct {
	UseCase *MockTaskCreateUseCase
}

func TestPostDownloads(t *testing.T) {
	type TestCase struct {
		name           string
		prepare        func(tt *TestCase, m *mockTaskCreate)
		ctx            context.Context
		req            PostDownloadsRequestObject
		expectedType   any
		expectId       IdTask
		expectStatus   TaskStatus
		expectFailResp ErrorResponse
		expectedErr    error
	}
	TestCases := []*TestCase{
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockTaskCreate) {
				m.UseCase.EXPECT().CreateTask(gomock.Any(), gomock.Any()).Return(
					entity.IdTask(1),
					entity.TaskStatusProcess,
					nil,
				)
			},
			ctx: context.Background(),
			req: PostDownloadsRequestObject{&PostDownloadsJSONRequestBody{
				Files: []Url{
					{
						"https://google.com",
					},
				},
				Timeout: "60s",
			}},
			expectedType: PostDownloads201JSONResponse{},
			expectId:     1,
			expectStatus: PROCESS,
			expectedErr:  nil,
		},
		{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockTaskCreate) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx:          context.Background(),
			req:          PostDownloadsRequestObject{},
			expectedType: PostDownloads500JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    ErrorCodeINTERNALSERVERERROR,
				Message: "PostDownloads: context canceled",
			},
		},
		{
			name: "GetTask error",
			prepare: func(tt *TestCase, m *mockTaskCreate) {
				m.UseCase.EXPECT().CreateTask(gomock.Any(), gomock.Any()).Return(
					entity.IdTask(0),
					entity.Status(""),
					errors.New("CreateTask error"),
				)
			},
			ctx: context.Background(),
			req: PostDownloadsRequestObject{&PostDownloadsJSONRequestBody{
				Files: []Url{
					{
						"https://google.com",
					},
				},
				Timeout: "60s",
			}},
			expectedType: PostDownloads500JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    ErrorCodeINTERNALSERVERERROR,
				Message: fmt.Errorf("PostDownloads: %w", errors.New("CreateTask error")).Error(),
			},
		},
		{
			name: "validate error",
			ctx:  context.Background(),
			req: PostDownloadsRequestObject{&PostDownloadsJSONRequestBody{
				Files: []Url{
					{
						"https:google.com",
					},
				},
				Timeout: "60s",
			}},
			expectedType: PostDownloads400JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    ErrorCodeBADREQUEST,
				Message: "PostDownloads: ошибка валидации параметров: validate error: URL должен содержать схему и хост: https:google.com",
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

			m := &mockTaskCreate{mCreate}
			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			resp, err := ts.PostDownloads(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.ErrorIs(t, tt.expectedErr, err)
				return
			}
			require.NoError(t, err)
			switch respType := resp.(type) {
			case PostDownloads201JSONResponse:
				_ = respType
				val, ok := resp.(PostDownloads201JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectStatus, val.Status)
				require.Equal(t, tt.expectId, val.Id)
			case PostDownloads400JSONResponse:
				val, ok := resp.(PostDownloads400JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
			case PostDownloads500JSONResponse:
				val, ok := resp.(PostDownloads500JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
			default:
				require.Fail(t, "unexpected response type")
			}
		})
	}
}

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
						Id:     entity.IdTask(1),
						Status: entity.TaskStatusDone,
						Files: []entity.File{
							{
								Id:  1,
								Url: "https://google.com",
							},
							{
								Error: entity.FileErrTimeout,
								Url:   "https://google.com",
							},
						},
					},
					nil,
				)
			},
			ctx:        context.Background(),
			req:        GetDownloadsIdRequestObject{Id: 1},
			expectType: GetDownloadsId200JSONResponse{},
			expectedUrlFile: UrlFile{
				FileId: 1,
				Url:    "https://google.com",
			},
			expectedUrlErr: UrlErr{
				Error: struct {
					Code UrlErrErrorCode `json:"code"`
				}{
					Code: UrlErrErrorCodeTIMEOUT,
				},
				Url: "https://google.com",
			},
			expectId:     1,
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
				Code:    ErrorCodeINTERNALSERVERERROR,
				Message: "GetDownloadsId: context canceled",
			},
		},
		{
			name: "GetTask error",
			prepare: func(tt *TestCase, m *mockGetTask) {
				m.UseCase.EXPECT().GetTask(gomock.Any(), gomock.Any()).Return(
					nil,
					repErr.ErrTaskNotExist,
				)
			},
			ctx:        context.Background(),
			req:        GetDownloadsIdRequestObject{},
			expectType: GetDownloadsId404JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    ErrorCodeNOTFOUND,
				Message: fmt.Errorf("GetDownloadsId: %w", repErr.ErrTaskNotExist).Error(),
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
				require.ErrorIs(t, tt.expectedErr, err)
				return
			}
			require.NoError(t, err)
			switch respType := resp.(type) {
			case GetDownloadsId200JSONResponse:
				_ = respType
				val, ok := resp.(GetDownloadsId200JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectStatus, val.Status)
				require.Equal(t, tt.expectId, val.Id)

				urlFile, err := val.Files[0].AsUrlFile()
				require.NoError(t, err)
				require.Equal(t, urlFile, tt.expectedUrlFile)

				urlErr, err := val.Files[1].AsUrlErr()
				require.NoError(t, err)
				require.Equal(t, tt.expectedUrlErr, urlErr)
			case GetDownloadsId400JSONResponse:
				val, ok := resp.(GetDownloadsId400JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
			case GetDownloadsId404JSONResponse:
				val, ok := resp.(GetDownloadsId404JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
			case GetDownloadsId500JSONResponse:
				val, ok := resp.(GetDownloadsId500JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
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
				m.UseCase.EXPECT().GetTaskFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					[]byte("test"), nil)
			},
			ctx:          context.Background(),
			req:          GetDownloadsIdFilesFileIdRequestObject{Id: 1, FileId: 1},
			expectType:   GetDownloadsIdFilesFileId200ApplicationoctetStreamResponse{},
			expectedData: []byte("test"),
			expectedErr:  nil,
		},
		{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockTaskFileUseCase) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx:        context.Background(),
			req:        GetDownloadsIdFilesFileIdRequestObject{Id: 1, FileId: 1},
			expectType: GetDownloadsIdFilesFileId500JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    ErrorCodeINTERNALSERVERERROR,
				Message: "GetDownloadsIdFilesFileId: context canceled",
			},
		},
		{
			name: "GetTask error",
			prepare: func(tt *TestCase, m *mockTaskFileUseCase) {
				m.UseCase.EXPECT().GetTaskFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					repErr.ErrFileNotExist,
				)
			},
			ctx:        context.Background(),
			req:        GetDownloadsIdFilesFileIdRequestObject{Id: 1, FileId: 1},
			expectType: GetDownloadsId404JSONResponse{},
			expectFailResp: ErrorResponse{
				Code:    ErrorCodeNOTFOUND,
				Message: fmt.Errorf("GetDownloadsIdFilesFileId: %w", repErr.ErrFileNotExist).Error(),
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

			m := &mockTaskFileUseCase{mFile}
			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			resp, err := ts.GetDownloadsIdFilesFileId(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.ErrorIs(t, tt.expectedErr, err)
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
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
			case GetDownloadsIdFilesFileId404JSONResponse:
				val, ok := resp.(GetDownloadsIdFilesFileId404JSONResponse)
				require.Equal(t, true, ok)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
			case GetDownloadsIdFilesFileId500JSONResponse:
				val, ok := resp.(GetDownloadsIdFilesFileId500JSONResponse)
				require.Equal(t, ok, true)
				require.Equal(t, tt.expectFailResp.Code, val.Code)
				require.Equal(t, tt.expectFailResp.Message, val.Message)
			default:
				require.Fail(t, "unexpected response type")
			}

		})
	}
}
