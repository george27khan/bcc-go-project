package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

type mockCreateTask struct {
	repo *MockCreateTaskRepository
}

type mockLoader struct {
	loader *MockHttpLoader
}

type mockRunner struct {
	runner *MockBackgroundRunner
}

func TestCreateTask(t *testing.T) {
	type TestCase struct {
		name           string
		prepare        func(tt *TestCase, m *mockCreateTask, r *mockRunner)
		ctx            context.Context
		Task           entity.Task
		expectedIdTask entity.IdTask
		expectedStatus entity.Status
		expectedErr    error
	}
	testCases := []*TestCase{
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockCreateTask, r *mockRunner) {
				m.repo.EXPECT().Create(gomock.Any(), tt.Task).Return(entity.IdTask(0), nil)
				r.runner.EXPECT().GoTask(gomock.Any(), gomock.Any(), gomock.Any())
			},
			ctx: context.Background(),
			Task: entity.Task{
				Id:     entity.IdTask(0),
				Status: entity.TaskStatusProcess,
			},
			expectedIdTask: entity.IdTask(0),
			expectedStatus: entity.TaskStatusProcess,
			expectedErr:    nil,
		},
		{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockCreateTask, r *mockRunner) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx:         context.Background(),
			expectedErr: context.Canceled,
		},
		{
			name: "context repo timeout",
			prepare: func(tt *TestCase, m *mockCreateTask, r *mockRunner) {
				//var cancel context.CancelFunc
				tt.ctx, _ = context.WithTimeout(tt.ctx, 100*time.Millisecond)
				m.repo.EXPECT().Create(tt.ctx, gomock.Any()).DoAndReturn(
					func(ctx context.Context, Task entity.Task) (entity.IdTask, error) {
						<-ctx.Done()
						return entity.IdTask(0), ctx.Err()
					})
			},
			ctx:         context.Background(),
			expectedErr: context.DeadlineExceeded,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rep := NewMockCreateTaskRepository(ctrl)
			loader := NewMockHttpLoader(ctrl)
			runner := NewMockBackgroundRunner(ctrl)

			r := &mockRunner{runner}
			m := &mockCreateTask{rep}

			if tt.prepare != nil {
				tt.prepare(tt, m, r)
			}

			tf := NewCreateTaskUseCase(rep, loader, runner, context.Background())
			idTask, status, err := tf.CreateTask(tt.ctx, tt.Task)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedIdTask, idTask)
			require.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestRunDownload(t *testing.T) {
	type TestCase struct {
		name        string
		prepare     func(tt *TestCase, m *mockCreateTask, r *mockRunner)
		ctx         context.Context
		Task        entity.Task
		wg          *sync.WaitGroup
		expectedErr error
	}
	testCases := []*TestCase{
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockCreateTask, r *mockRunner) {
				r.runner.EXPECT().GoFile(gomock.Any(), tt.wg, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File) error) {
						wg.Done()
					},
				)
				m.repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			ctx: context.Background(),
			wg:  &sync.WaitGroup{},
			Task: entity.Task{
				Id:      entity.IdTask(0),
				Status:  entity.TaskStatusProcess,
				Timeout: 60 * time.Second,
				Files:   []entity.File{{Url: "https://google.com"}},
			},
			expectedErr: nil,
		},
		{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockCreateTask, r *mockRunner) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx: context.Background(),
			wg:  &sync.WaitGroup{},
			Task: entity.Task{
				Id:      entity.IdTask(0),
				Status:  entity.TaskStatusProcess,
				Timeout: 60 * time.Second,
				Files:   []entity.File{{Url: "https://google.com"}},
			},
			expectedErr: context.Canceled,
		},
		{
			name: "context timeout",
			prepare: func(tt *TestCase, m *mockCreateTask, r *mockRunner) {
				r.runner.EXPECT().GoFile(gomock.Any(), tt.wg, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File) error) {
						wg.Done()
					},
				)
				m.repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
			},
			ctx: context.Background(),
			wg:  &sync.WaitGroup{},
			Task: entity.Task{
				Id:      entity.IdTask(0),
				Status:  entity.TaskStatusProcess,
				Timeout: 60 * time.Second,
				Files:   []entity.File{{Url: "https://google.com"}},
			},
			expectedErr: context.DeadlineExceeded,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rep := NewMockCreateTaskRepository(ctrl)
			loader := NewMockHttpLoader(ctrl)
			runner := NewMockBackgroundRunner(ctrl)
			r := &mockRunner{runner}
			m := &mockCreateTask{rep}
			if tt.prepare != nil {
				tt.prepare(tt, m, r)
			}

			tf := NewCreateTaskUseCase(rep, loader, runner, context.Background())
			err := tf.RunDownload(tt.ctx, tt.wg, tt.Task)
			tt.wg.Wait()
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDownloadFile(t *testing.T) {
	type TestCase struct {
		name        string
		prepare     func(tt *TestCase, m *mockCreateTask, l *mockLoader)
		ctx         context.Context
		idTask      entity.IdTask
		file        entity.File
		wg          *sync.WaitGroup
		expectedErr error
	}
	testCases := []*TestCase{
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockCreateTask, l *mockLoader) {
				tt.wg.Add(1)
				l.loader.EXPECT().Load(tt.ctx, tt.file.Url).Return([]byte("test"), nil)
				m.repo.EXPECT().UpdateFileData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			ctx:         context.Background(),
			wg:          &sync.WaitGroup{},
			idTask:      1,
			file:        entity.File{Url: "https://google.com"},
			expectedErr: nil,
		},
		{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockCreateTask, l *mockLoader) {
				tt.wg.Add(1)
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx:         context.Background(),
			wg:          &sync.WaitGroup{},
			idTask:      1,
			file:        entity.File{Url: "https://google.com"},
			expectedErr: context.Canceled,
		},
		{
			name: "UpdateFileData error",
			prepare: func(tt *TestCase, m *mockCreateTask, l *mockLoader) {
				tt.wg.Add(1)
				l.loader.EXPECT().Load(tt.ctx, tt.file.Url).Return([]byte("test"), nil)
				m.repo.EXPECT().UpdateFileData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
			},
			ctx:         context.Background(),
			wg:          &sync.WaitGroup{},
			idTask:      1,
			file:        entity.File{Url: "https://google.com"},
			expectedErr: context.DeadlineExceeded,
		},
		{
			name: "HttpLoader.Load error",
			prepare: func(tt *TestCase, m *mockCreateTask, l *mockLoader) {
				tt.wg.Add(1)
				l.loader.EXPECT().Load(tt.ctx, tt.file.Url).Return(nil, nil)
			},
			ctx:         context.Background(),
			wg:          &sync.WaitGroup{},
			idTask:      1,
			file:        entity.File{Url: "https://google.com"},
			expectedErr: context.DeadlineExceeded,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rep := NewMockCreateTaskRepository(ctrl)
			loader := NewMockHttpLoader(ctrl)
			runner := NewMockBackgroundRunner(ctrl)
			m := &mockCreateTask{rep}
			l := &mockLoader{loader}
			if tt.prepare != nil {
				tt.prepare(tt, m, l)
			}

			tf := NewCreateTaskUseCase(rep, loader, runner, context.Background())
			err := tf.DownloadFile(tt.ctx, tt.wg, tt.idTask, tt.file)
			tt.wg.Wait()
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
