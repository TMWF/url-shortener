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
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

// GetUsersURLs implements [repository.DBStorage].
func (m *mockDBPinger) GetUsersURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.GetUserURLsResponseModel), args.Error(1)
}

// SaveBatchURL implements [repository.DBStorage].
func (m *mockDBPinger) SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]string, error) {
	args := m.Called(ctx, urlBatch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// SaveURL implements [repository.DBStorage].
func (m *mockDBPinger) SaveURL(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

// SaveUser implements [repository.DBStorage].
func (m *mockDBPinger) SaveUser(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// DeleteUserURLs implements [repository.DBStorage].
func (m *mockDBPinger) DeleteUserURLs(jobs []model.DeleteUserURLsJobModel) error {
	args := m.Called(jobs)
	return args.Error(0)
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
