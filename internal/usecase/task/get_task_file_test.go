package task

import (
	"bcc-go-project/internal/domain/entity"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mocks struct {
	mockRepo *MockTaskFileRepository
}

func TestGetTaskFile(t *testing.T) {
	type TestCase struct {
		name        string
		prepare     func(tt *TestCase, m *mocks)
		idTask      entity.IdTask
		idFile      entity.IdFile
		expected    []byte
		expectedErr string
	}
	testCases := []*TestCase{
		&TestCase{
			name: "success",
			prepare: func(tt *TestCase, m *mocks) {
				m.mockRepo.EXPECT().GetTaskFile(gomock.Any(), tt.idTask, tt.idFile).
					Return([]byte("Hello World"), nil)
			},
			idTask:   entity.IdTask(0),
			idFile:   entity.IdFile(0),
			expected: []byte("Hello World"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rep := NewMockTaskFileRepository(ctrl)
			m := &mocks{rep}

			if tt.prepare != nil {
				tt.prepare(tt, m)
			}

			tf := NewTaskFileUseCase(rep)
			got, err := tf.GetTaskFile(context.Background(), tt.idTask, tt.idFile)
			if tt.expectedErr != "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, got, tt.expected)
		})
	}
}
