package repositories

import (
	"context"
	"fmt"

	"github.com/dpbrackin/ready-set-go/auth"
	"github.com/dpbrackin/ready-set-go/db/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type PGAuthRepository struct {
	queries *generated.Queries
}

func NewPGAuthRepository(queries *generated.Queries) *PGAuthRepository {
	return &PGAuthRepository{
		queries: queries,
	}
}

// AddUser implements auth.AuthRepository.
func (p *PGAuthRepository) AddUser(ctx context.Context, params auth.UserWithPassword) error {
	err := p.queries.AddUser(ctx, generated.AddUserParams{
		Username: params.Username,
		Password: params.Password,
	})

	return err
}

// GetSession implements auth.AuthRepository.
func (p *PGAuthRepository) GetSession(ctx context.Context, sessionID string) (auth.Session, error) {
	session, err := p.queries.GetSession(ctx, sessionID)

	if err != nil {
		return auth.Session{}, fmt.Errorf("Failed to get session: %w", err)
	}

	return auth.Session{
		ID:           session.ID,
		User:         auth.User{Username: session.Username, ID: session.UserID.Int32},
		CreatedAt:    session.CreatedAt.Time,
		RevokedAt:    session.RevokedAt.Time,
		ExpiresAt:    session.ExpiresAt.Time,
		LastActiveAt: session.LastActiveAt.Time,
		IsRevoked:    session.RevokedAt.Valid,
	}, nil
}

// GetUserByUsername implements auth.AuthRepository.
func (p *PGAuthRepository) GetUserByUsername(ctx context.Context, username string) (auth.UserWithPassword, error) {
	user, err := p.queries.GetUserByUsername(ctx, username)

	if err != nil {
		return auth.UserWithPassword{}, fmt.Errorf("Failed to get user: %w", err)
	}

	return auth.UserWithPassword{
		User: auth.User{
			Username: user.Username,
			ID:       user.ID,
		},
		Password: user.Password,
	}, nil
}

func (p *PGAuthRepository) CreateSession(ctx context.Context, session auth.Session) error {
	err := p.queries.CreateSession(ctx, generated.CreateSessionParams{
		ID: session.ID,
		UserID: pgtype.Int4{
			Int32: int32(session.User.ID),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  session.ExpiresAt,
			Valid: true,
		},
	})

	return err
}
