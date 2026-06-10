package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthStore interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmailOrUsername(ctx context.Context, identifier string) (*models.User, error)
	UpdatePassword(ctx context.Context, userID int, newHash string) error
	CreateResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetValidResetToken(ctx context.Context, token string) (*models.ResetToken, error)
	MarkTokenUsed(ctx context.Context, token string) error
}

type authStore struct {
	db *pgxpool.Pool
}

func NewAuthStore(db *pgxpool.Pool) AuthStore {
	return &authStore{db: db}
}

func (s *authStore) CreateUser(ctx context.Context, u *models.User) (*models.User, error) {
	q := `INSERT INTO users (uuid, username, firstname, lastname, email, phone, password_hash, role, parent_id, created_at, updated_at)
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())`
	_, err := s.db.Exec(ctx, q, u.UUID, u.Username, u.Firstname, u.Lastname, u.Email, u.Phone, u.PasswordHash, u.Role, u.ParentID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *authStore) GetByEmailOrUsername(ctx context.Context, identifier string) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, uuid, username, firstname, lastname, email, phone, password_hash, role, parent_id, created_at, updated_at
          FROM users
          WHERE (email = $1 OR username = $1) AND deleted_at IS NULL LIMIT 1`

	err := s.db.QueryRow(ctx, q, identifier).Scan(
		&u.ID,
		&u.UUID,
		&u.Username,
		&u.Firstname,
		&u.Lastname,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&u.Role,
		&u.ParentID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("identifiants invalides ou compte supprimé")
		}
		return nil, err
	}
	return u, nil
}

func (s *authStore) UpdatePassword(ctx context.Context, userID int, newHash string) error {
	q := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	tag, err := s.db.Exec(ctx, q, newHash, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}
	return nil
}

func (s *authStore) CreateResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	// Invalider les anciens tokens non utilisés pour cet utilisateur
	_, _ = s.db.Exec(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`, userID)
	_, err := s.db.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, expiresAt)
	return err
}

func (s *authStore) GetValidResetToken(ctx context.Context, token string) (*models.ResetToken, error) {
	rt := &models.ResetToken{}
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, token, expires_at, used_at, created_at
		 FROM password_reset_tokens
		 WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()`,
		token,
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.UsedAt, &rt.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("lien invalide ou expiré")
		}
		return nil, err
	}
	return rt, nil
}

func (s *authStore) MarkTokenUsed(ctx context.Context, token string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE token = $1`,
		token)
	return err
}
