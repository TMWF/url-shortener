package service

import (
	"context"
	"errors"
	"testing"

	"github.com/TMWF/url-shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDBPinger struct {
	mock.Mock
}

// GetURL implements [repository.DBStorage].
func (m *mockDBPinger) GetURL(ctx context.Context, id string) (string, error) {
	panic("unimplemented")
}

// GetUsersURLs implements [repository.DBStorage].
func (m *mockDBPinger) GetUsersURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	panic("unimplemented")
}

// SaveBatchURL implements [repository.DBStorage].
func (m *mockDBPinger) SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]string, error) {
	panic("unimplemented")
}

// SaveURL implements [repository.DBStorage].
func (m *mockDBPinger) SaveURL(ctx context.Context, url string) (string, error) {
	panic("unimplemented")
}

// SaveUser implements [repository.DBStorage].
func (m *mockDBPinger) SaveUser(ctx context.Context) (int, error) {
	panic("unimplemented")
}

// DeleteUserURLs implements [repository.DBStorage].
func (m *mockDBPinger) DeleteUserURLs([]model.DeleteUserURLsJobModel) error {
	panic("unimplemented")
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
