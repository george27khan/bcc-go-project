package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

// SyncRunner реализация синхронного запуска загрузки для тестов
type SyncRunner struct{}

func (sr *SyncRunner) GoTask(ctx context.Context, object entity.Task, f func(context.Context, *sync.WaitGroup, entity.Task)) {
	f(ctx, &sync.WaitGroup{}, object)
}
func (sr *SyncRunner) GoFile(ctx context.Context, wg *sync.WaitGroup, idTask entity.IdTask, file entity.File, f func(context.Context, *sync.WaitGroup, entity.IdTask, entity.File)) {
	f(ctx, wg, idTask, file)
}

type mockCreateTask struct {
	repo *MockCreateTaskRepository
}

func TestCreateTask(t *testing.T) {
	type TestCase struct {
		name           string
		prepare        func(tt *TestCase, m *mockCreateTask)
		ctx            context.Context
		Task           entity.Task
		expectedIdTask entity.IdTask
		expectedStatus entity.Status
		expectedErr    error
	}
	testCases := []*TestCase{
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockCreateTask) {
				m.repo.EXPECT().Create(gomock.Any(), tt.Task).
					Return(entity.IdTask(0), nil)
				m.repo.EXPECT().UpdateFileData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				m.repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
			prepare: func(tt *TestCase, m *mockCreateTask) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx:            context.Background(),
			Task:           entity.Task{},
			expectedIdTask: entity.IdTask(0),
			expectedStatus: entity.Status(""),
			expectedErr:    context.Canceled,
		},
		//{
		//	name: "context repo timeout",
		//	prepare: func(tt *TestCase, m *mockCreateTask) {
		//		//var cancel context.CancelFunc
		//		tt.ctx, _ = context.WithTimeout(tt.ctx, 100*time.Millisecond)
		//		//defer cancel()
		//		m.repo.EXPECT().Create(gomock.Any(), tt.Task).DoAndReturn(
		//			func(ctx context.Context, Task entity.Task) (entity.IdTask, error) {
		//				<-ctx.Done()
		//				return entity.IdTask(0), ctx.Err()
		//			})
		//	},
		//	ctx:            context.Background(),
		//	Task:           entity.Task{},
		//	expectedIdTask: entity.IdTask(0),
		//	expectedStatus: entity.Status(""),
		//	expectedErr:    context.DeadlineExceeded,
		//},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rep := NewMockCreateTaskRepository(ctrl)
			loader := NewMockHttpLoader(ctrl)
			runner := &SyncRunner{}
			m := &mockCreateTask{rep}

			if tt.prepare != nil {
				tt.prepare(tt, m)
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
