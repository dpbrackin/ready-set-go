package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Clock interface {
	Now() time.Time
}

type AuthService struct {
	repository AuthRepository
	clock      Clock
}

type NewAuthServiceParams struct {
	Repository AuthRepository
	Clock      Clock
}

func NewAuthService(params NewAuthServiceParams) *AuthService {
	return &AuthService{
		repository: params.Repository,
		clock:      params.Clock,
	}
}

func (srv *AuthService) AuthenticateWithPassword(ctx context.Context, creds PasswordCredentials) (User, error) {
	var user User

	dbUser, err := srv.repository.GetUserByUsername(ctx, creds.Username)

	if err != nil {
		return user, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(creds.Password))

	if err != nil {
		return user, err
	}

	user = User{
		Username: dbUser.Username,
		ID:       dbUser.ID,
	}

	return user, nil
}

func (srv *AuthService) Register(ctx context.Context, creds PasswordCredentials) (User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)

	if err != nil {
		return User{}, err
	}

	err = srv.repository.AddUser(ctx, UserWithPassword{
		User: User{
			Username: creds.Username,
		},
		Password: string(hashedPassword),
	})

	if err != nil {
		return User{}, err
	}

	user := User{
		Username: creds.Username,
	}

	return user, nil
}

func (srv *AuthService) AuthenticateSession(ctx context.Context, sessionID string) (User, error) {
	session, err := srv.repository.GetSession(ctx, sessionID)

	if err != nil {
		return User{}, fmt.Errorf("Failed to get session: %w", err)
	}

	now := srv.clock.Now()

	isRevoked := session.IsRevoked && now.After(session.RevokedAt)
	isExpired := now.After(session.ExpiresAt)

	if isRevoked || isExpired {
		return User{}, fmt.Errorf("Session expired")
	}

	return session.User, nil
}

func (srv *AuthService) CreateSession(ctx context.Context, user User) (*Session, error) {
	session, err := NewSession(user)

	if err != nil {
		return nil, err
	}

	err = srv.repository.CreateSession(ctx, *session)

	if err != nil {
		return nil, err
	}

	createdSession, err := srv.repository.GetSession(ctx, session.ID)

	return &createdSession, err
}
