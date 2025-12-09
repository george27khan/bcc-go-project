package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

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
		&TestCase{
			name: "success",
			prepare: func(tt *TestCase, m *mockCreateTask) {
				m.repo.EXPECT().Create(gomock.Any(), tt.Task).
					Return(entity.IdTask(0), nil)
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
		//&TestCase{
		//	name: "context canceled",
		//	prepare: func(tt *TestCase, m *mockCreateTask) {
		//		var cancel context.CancelFunc
		//		tt.ctx, cancel = context.WithCancel(tt.ctx)
		//		cancel()
		//	},
		//	ctx:            context.Background(),
		//	Task:           entity.Task{},
		//	expectedidTask: entity.IdTask(0),
		//	expectedStatus: entity.TaskStatusProcess,
		//	expectedErr:    context.Canceled,
		//},
		//&TestCase{
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
		//	expectedidTask: entity.IdTask(0),
		//	expectedStatus: entity.TaskStatusProcess,
		//	expectedErr:    context.DeadlineExceeded,
		//},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rep := NewMockCreateTaskRepository(ctrl)
			loader := NewMockHttpLoader(ctrl)
			m := &mockCreateTask{rep}

			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			tf := NewCreateTaskUseCase(rep, loader)
			idTask, status, err := tf.CreateTask(tt.ctx, tt.Task)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, idTask, tt.expectedIdTask)
			require.Equal(t, status, tt.expectedStatus)
		})
	}
}
