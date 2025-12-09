package task

import (
	"bcc-go-project/internal/domain/entity"
	errors_repo "bcc-go-project/internal/infrastructure/repository/errors_repo"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type mockGetTaskFile struct {
	repo *MockTaskFileRepository
}

func TestGetTaskFile(t *testing.T) {
	type TestCase struct {
		name        string
		prepare     func(tt *TestCase, m *mockGetTaskFile)
		ctx         context.Context
		idTask      entity.IdTask
		idFile      entity.IdFile
		expected    []byte
		expectedErr error
	}
	testCases := []*TestCase{
		&TestCase{
			name: "success",
			prepare: func(tt *TestCase, m *mockGetTaskFile) {
				m.repo.EXPECT().GetTaskFile(gomock.Any(), tt.idTask, tt.idFile).
					Return([]byte("Hello World"), nil)
			},
			ctx:      context.Background(),
			idTask:   entity.IdTask(0),
			idFile:   entity.IdFile(0),
			expected: []byte("Hello World"),
		},
		&TestCase{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockGetTaskFile) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()

			},
			ctx:         context.Background(),
			idTask:      entity.IdTask(0),
			idFile:      entity.IdFile(0),
			expected:    nil,
			expectedErr: context.Canceled,
		},
		&TestCase{
			name: "context repo timeout",
			prepare: func(tt *TestCase, m *mockGetTaskFile) {
				//var cancel context.CancelFunc
				tt.ctx, _ = context.WithTimeout(tt.ctx, 100*time.Millisecond)
				//defer cancel()
				m.repo.EXPECT().GetTaskFile(gomock.Any(), tt.idTask, tt.idFile).DoAndReturn(
					func(ctx context.Context, idTask entity.IdTask, idFile entity.IdFile) ([]byte, error) {
						<-ctx.Done()
						return nil, ctx.Err()
					})
			},
			ctx:         context.Background(),
			idTask:      entity.IdTask(0),
			idFile:      entity.IdFile(0),
			expected:    nil,
			expectedErr: context.DeadlineExceeded,
		},
		&TestCase{
			name: "task not found",
			prepare: func(tt *TestCase, m *mockGetTaskFile) {
				m.repo.EXPECT().GetTaskFile(gomock.Any(), tt.idTask, tt.idFile).
					Return(nil, errors_repo.ErrTaskNotExist)
			},
			ctx:         context.Background(),
			idTask:      entity.IdTask(0),
			idFile:      entity.IdFile(0),
			expected:    nil,
			expectedErr: errors_repo.ErrTaskNotExist,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rep := NewMockTaskFileRepository(ctrl)
			m := &mockGetTaskFile{rep}

			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			tf := NewTaskFileUseCase(rep)
			got, err := tf.GetTaskFile(tt.ctx, tt.idTask, tt.idFile)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, got, tt.expected)
		})
	}
}
