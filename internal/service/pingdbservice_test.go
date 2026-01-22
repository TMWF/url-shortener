package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	// Укажите правильный путь до вашего репозитория
)

// --- Мок DBPinger (если он не определен в другом месте для тестов) ---
// Эта структура должна быть в том же пакете, что и тесты, или импортирована.
type mockDBPinger struct {
	mock.Mock
}

// PingDB implements repository.DBPinger.
func (m *mockDBPinger) PingDB() error {
	args := m.Called()
	return args.Error(0)
}

// --- Тесты ---

func TestPingDBService_PingDB(t *testing.T) {
	tests := []struct {
		name             string
		mockPingError    error
		expectedError    error
		expectPingCalled bool
	}{
		{
			name:             "Successful Ping",
			mockPingError:    nil,
			expectedError:    nil,
			expectPingCalled: true,
		},
		{
			name:             "Ping Returns Error",
			mockPingError:    errors.New("database is not available"),
			expectedError:    errors.New("database is not available"),
			expectPingCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockDBPinger)

			if tt.expectPingCalled {
				mockRepo.On("PingDB").Return(tt.mockPingError).Once()
			} else {
				// Если метод не должен быть вызван, мы можем не настраивать On,
				// а потом проверить с помощью AssertNotCalled.
			}

			pingService := NewPingDBService(mockRepo)

			err := pingService.PingDB()
			assert.Equal(t, tt.expectedError, err, "Expected error mismatch")

			if tt.expectPingCalled {
				mockRepo.AssertNumberOfCalls(t, "PingDB", 1)
			} else {
				mockRepo.AssertNotCalled(t, "PingDB")
			}
		})
	}
}
