package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
)

type AuthStore interface {
	CreateUser(user *models.User) (*models.User, error)
	GetByEmailOrUsername(identifier string) (*models.User, error)
	UpdatePassword(userID int, newHash string) error
}

type authStore struct {
	db *sql.DB
}

func NewAuthStore(db *sql.DB) AuthStore {
	return &authStore{db: db}
}

func (s *authStore) CreateUser(u *models.User) (*models.User, error) {
	q := `INSERT INTO users (uuid, username, firstname, lastname, email, phone, password_hash, role, parent_id, created_at, updated_at) 
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW() )`
	_, err := s.db.Exec(q, u.UUID, u.Username, u.Firstname, u.Lastname, u.Email, u.Phone, u.PasswordHash, u.Role, u.ParentID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *authStore) GetByEmailOrUsername(identifier string) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, uuid, username, firstname, lastname, email, phone, password_hash, role, parent_id, created_at, updated_at 
          FROM users 
          WHERE (email = ? OR username = ?) AND deleted_at IS NULL LIMIT 1`

	err := s.db.QueryRow(q, identifier, identifier).Scan(
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
		if err == sql.ErrNoRows {
			return nil, errors.New("identifiants invalides ou compte supprimé")
		}
		return nil, err
	}
	return u, nil
}

func (s *authStore) UpdatePassword(userID int, newHash string) error {
	q := `UPDATE users SET password_hash = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL`
	res, err := s.db.Exec(q, newHash, userID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	return nil
}
