package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dpbrackin/ready-set-go/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthRepository struct {
	mock.Mock
}

type mockClock struct {
	mock.Mock
}

// Now implements auth.Clock.
func (m *mockClock) Now() time.Time {
	args := m.Called()

	return args.Get(0).(time.Time)
}

// AddUser implements auth.AuthRepository.
func (m *mockAuthRepository) AddUser(ctx context.Context, params auth.UserWithPassword) error {
	args := m.Called(ctx, params)

	return args.Error(0)
}

// GetSession implements auth.AuthRepository.
func (m *mockAuthRepository) GetSession(ctx context.Context, sessionID string) (auth.Session, error) {
	args := m.Called(ctx, sessionID)
	// Check if the returned value is of type auth.Session
	session, ok := args.Get(0).(auth.Session)
	if !ok {
		return auth.Session{}, fmt.Errorf("expected auth.Session, but got %T", args.Get(0))
	}

	return session, args.Error(1)
}

// GetUserByUsername implements auth.AuthRepository.
func (m *mockAuthRepository) GetUserByUsername(ctx context.Context, username string) (auth.UserWithPassword, error) {
	args := m.Called(ctx, username)
	// Check if the returned value is of type auth.Session
	user, ok := args.Get(0).(auth.UserWithPassword)
	if !ok {
		return auth.UserWithPassword{}, fmt.Errorf("expected auth.Session, but got %T", args.Get(0))
	}

	return user, args.Error(1)
}

// CreateSession implements auth.AuthRepository.
func (m *mockAuthRepository) CreateSession(ctx context.Context, session auth.Session) error {
	args := m.Called(ctx, session)

	return args.Error(0)
}

func TestValidAuthenticateSession(t *testing.T) {
	validSession := auth.Session{
		ID: "session1",
		User: auth.User{
			Username: "user1",
			ID:       1,
		},
		CreatedAt:    time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC),
		RevokedAt:    time.Time{},
		ExpiresAt:    time.Date(2025, 12, 12, 0, 0, 0, 0, time.UTC),
		LastActiveAt: time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC),
		IsRevoked:    false,
	}

	repository := new(mockAuthRepository)
	repository.On("GetSession", mock.Anything, "session1").Return(validSession, nil)

	fakeClock := new(mockClock)
	fakeClock.On("Now").Return(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	service := auth.NewAuthService(auth.NewAuthServiceParams{
		Repository: repository,
		Clock:      fakeClock,
	})

	ctx := context.Background()
	res, err := service.AuthenticateSession(ctx, "session1")

	assert.Nil(t, err)
	assert.Equal(t, validSession.User, res)
}

func TestExpiredSession(t *testing.T) {
	expiredSession := auth.Session{
		ID: "session1",
		User: auth.User{
			Username: "user1",
			ID:       1,
		},
		CreatedAt:    time.Date(2023, 12, 12, 0, 0, 0, 0, time.UTC),
		RevokedAt:    time.Time{},
		ExpiresAt:    time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC),
		LastActiveAt: time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC),
		IsRevoked:    false,
	}

	repository := new(mockAuthRepository)
	repository.On("GetSession", mock.Anything, "session1").Return(expiredSession, nil)

	fakeClock := new(mockClock)
	fakeClock.On("Now").Return(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	service := auth.NewAuthService(auth.NewAuthServiceParams{
		Repository: repository,
		Clock:      fakeClock,
	})

	ctx := context.Background()
	_, err := service.AuthenticateSession(ctx, "session1")

	assert.NotNil(t, err)
}

func TestRevokedSession(t *testing.T) {
	revokedSession := auth.Session{
		ID: "session1",
		User: auth.User{
			Username: "user1",
			ID:       1,
		},
		CreatedAt:    time.Date(2022, 12, 12, 0, 0, 0, 0, time.UTC),
		RevokedAt:    time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2025, 12, 12, 0, 0, 0, 0, time.UTC),
		LastActiveAt: time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC),
		IsRevoked:    true,
	}

	repository := new(mockAuthRepository)
	repository.On("GetSession", mock.Anything, "session1").Return(revokedSession, nil)

	fakeClock := new(mockClock)
	fakeClock.On("Now").Return(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	service := auth.NewAuthService(auth.NewAuthServiceParams{
		Repository: repository,
		Clock:      fakeClock,
	})

	ctx := context.Background()
	_, err := service.AuthenticateSession(ctx, "session1")

	assert.NotNil(t, err)
}
