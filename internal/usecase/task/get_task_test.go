package task

import (
	"bcc-go-project/internal/domain/entity"
	"bcc-go-project/internal/infrastructure/repository/errors_repo"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type mockGetTask struct {
	repo *MockGetTaskRepository
}

func TestGetTask(t *testing.T) {
	type TestCase struct {
		name        string
		prepare     func(tt *TestCase, m *mockGetTask)
		ctx         context.Context
		idTask      entity.IdTask
		expected    *entity.Task
		expectedErr error
	}
	TestCases := []*TestCase{
		&TestCase{
			name: "success",
			prepare: func(tt *TestCase, m *mockGetTask) {
				m.repo.EXPECT().Get(gomock.Any(), tt.idTask).
					Return(&entity.Task{
						Id:      0,
						Timeout: 60,
						Status:  entity.TaskStatusProcess,
						Files:   nil,
					}, nil)
			},
			ctx:    context.Background(),
			idTask: entity.IdTask(0),
			expected: &entity.Task{
				Id:      0,
				Timeout: 60,
				Status:  entity.TaskStatusProcess,
				Files:   nil,
			},
		},
		&TestCase{
			name: "context repo timeout",
			prepare: func(tt *TestCase, m *mockGetTask) {
				//var cancel context.CancelFunc
				tt.ctx, _ = context.WithTimeout(tt.ctx, 100*time.Millisecond)
				//defer cancel()
				m.repo.EXPECT().Get(gomock.Any(), tt.idTask).DoAndReturn(
					func(ctx context.Context, idTask entity.IdTask) (*entity.Task, error) {
						<-ctx.Done()
						return nil, ctx.Err()
					})
			},
			ctx:         context.Background(),
			idTask:      entity.IdTask(0),
			expected:    nil,
			expectedErr: context.DeadlineExceeded,
		},
		&TestCase{
			name: "task not found",
			prepare: func(tt *TestCase, m *mockGetTask) {
				m.repo.EXPECT().Get(gomock.Any(), tt.idTask).
					Return(nil, errors_repo.ErrTaskNotExist)
			},
			ctx:         context.Background(),
			idTask:      entity.IdTask(0),
			expected:    nil,
			expectedErr: errors_repo.ErrTaskNotExist,
		},
	}

	for _, tt := range TestCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repoMock := NewMockGetTaskRepository(ctrl)
			m := &mockGetTask{
				repo: repoMock,
			}

			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			tf := NewGetTaskUseCase(repoMock)
			got, err := tf.GetTask(tt.ctx, tt.idTask)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, got, tt.expected)
		})
	}
}
