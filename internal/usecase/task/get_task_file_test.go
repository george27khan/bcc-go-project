package task

import (
	"bcc-go-project/internal/domain/entity"
	errorsRepo "bcc-go-project/internal/infrastructure/repository/errors_repo"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
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
		{
			name: "success",
			prepare: func(tt *TestCase, m *mockGetTaskFile) {
				m.repo.EXPECT().GetTaskFile(gomock.Any(), tt.idTask, tt.idFile).
					Return([]byte("Hello World"), nil)
			},
			ctx:      context.Background(),
			idTask:   entity.IdTask(1),
			idFile:   entity.IdFile(1),
			expected: []byte("Hello World"),
		},
		{
			name: "context canceled",
			prepare: func(tt *TestCase, m *mockGetTaskFile) {
				var cancel context.CancelFunc
				tt.ctx, cancel = context.WithCancel(tt.ctx)
				cancel()
			},
			ctx:         context.Background(),
			expected:    nil,
			expectedErr: context.Canceled,
		},
		{
			name: "task not found",
			prepare: func(tt *TestCase, m *mockGetTaskFile) {
				m.repo.EXPECT().GetTaskFile(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errorsRepo.ErrTaskNotExist)
			},
			ctx:         context.Background(),
			idTask:      entity.IdTask(1),
			idFile:      entity.IdFile(1),
			expected:    nil,
			expectedErr: errorsRepo.ErrTaskNotExist,
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
			require.Equal(t, tt.expected, got)
		})
	}
}
