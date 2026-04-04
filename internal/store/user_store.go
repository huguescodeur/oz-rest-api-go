package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/errs"
)

type UserStore interface {
	CreateUser(u *models.User) (*models.User, error)
	GetAll(limit, offset int) ([]*models.User, int, error)
	GetAllByOwner(ownerID, limit, offset int) ([]*models.User, int, error)
	GetByUUID(userUUID uuid.UUID) (*models.User, error)
	Update(u *models.User) (*models.User, error)
	GetByUUIDWithParent(userUUID uuid.UUID, ownerID int) (*models.User, error)
	Delete(uuid uuid.UUID) error
	Restore(uuid uuid.UUID) error
	DeleteByOwner(vendeurID uuid.UUID, ownerID int) error
	RestoreByOwner(uuid uuid.UUID, ownerID int) error
}

type userStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) UserStore {
	return &userStore{db: db}
}

func (s *userStore) CreateUser(u *models.User) (*models.User, error) {
	q := `INSERT INTO users (
            uuid, username, firstname, lastname, email, 
            phone, password_hash, role, parent_id, 
            created_at, updated_at
          ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	_, err := s.db.Exec(
		q,
		u.UUID,
		u.Username,
		u.Firstname,
		u.Lastname,
		u.Email,
		u.Phone,
		u.PasswordHash,
		u.Role,
		u.ParentID,
	)

	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userStore) GetAll(limit, offset int) ([]*models.User, int, error) {
	var users []*models.User
	var total int

	s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)

	q := `SELECT id, uuid, username, firstname, lastname, email, phone, role, parent_id, created_at, deleted_at 
          FROM users 
          ORDER BY created_at DESC 
          LIMIT ? OFFSET ?`

	rows, err := s.db.Query(q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &models.User{}
		rows.Scan(&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname, &u.Email, &u.Phone, &u.Role, &u.ParentID, &u.CreatedAt, &u.DeletedAt)
		users = append(users, u)
	}
	return users, total, nil
}

func (s *userStore) GetAllByOwner(ownerID, limit, offset int) ([]*models.User, int, error) {
	var users []*models.User
	var total int

	s.db.QueryRow("SELECT COUNT(*) FROM users WHERE parent_id = ?", ownerID).Scan(&total)

	q := `SELECT id, uuid, username, firstname, lastname, email, phone, role, parent_id, created_at, deleted_at 
          FROM users 
          WHERE parent_id = ? 
          ORDER BY username ASC 
          LIMIT ? OFFSET ?`

	rows, err := s.db.Query(q, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &models.User{}
		rows.Scan(&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname, &u.Email, &u.Phone, &u.Role, &u.ParentID, &u.CreatedAt, &u.DeletedAt)
		users = append(users, u)
	}
	return users, total, nil
}

func (s *userStore) GetByUUID(userUUID uuid.UUID) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, uuid, username, firstname, lastname, email, phone, password_hash, role, parent_id, created_at 
          FROM users 
          WHERE uuid = ? AND deleted_at IS NULL`

	err := s.db.QueryRow(q, userUUID).Scan(
		&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname,
		&u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.ParentID, &u.CreatedAt,
	)
	return u, err
}

func (s *userStore) GetByUUIDWithParent(userUUID uuid.UUID, ownerID int) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, uuid, username, firstname, lastname, email, phone, role, parent_id, created_at 
          FROM users 
          WHERE uuid = ? AND (id = ? OR parent_id = ?) AND deleted_at IS NULL`

	err := s.db.QueryRow(q, userUUID, ownerID, ownerID).Scan(
		&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname,
		&u.Email, &u.Phone, &u.Role, &u.ParentID, &u.CreatedAt,
	)
	return u, err
}

func (s *userStore) Update(u *models.User) (*models.User, error) {
	updatedAt := time.Now().UTC()
	q := `UPDATE users SET 
            username = ?, 
            firstname = ?, 
            lastname = ?, 
            email = ?, 
            phone = ?, 
            role = ?, 
            parent_id = ?,
            updated_at = ?
          WHERE uuid = ? AND deleted_at IS NULL`

	raw, err := s.db.Exec(q,
		u.Username,
		u.Firstname,
		u.Lastname,
		u.Email,
		u.Phone,
		u.Role,
		u.ParentID,
		updatedAt,
		u.UUID,
	)
	if err != nil {
		return nil, err
	}
	count, err := raw.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("User n'existe pas! Impossible d'update")
	}

	return s.GetByUUID(u.UUID)
}

func (s *userStore) Delete(uuid uuid.UUID) error {
	q := `UPDATE users SET deleted_at = NOW() WHERE uuid = ? AND deleted_at IS NULL`
	raw, err := s.db.Exec(q, uuid)
	if err != nil {
		return err
	}

	count, err := raw.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *userStore) Restore(uuid uuid.UUID) error {
	q := `UPDATE users 
          SET deleted_at = NULL, updated_at = NOW() 
          WHERE uuid = ? AND deleted_at IS NOT NULL`

	raw, err := s.db.Exec(q, uuid)
	if err != nil {
		return err
	}

	count, err := raw.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errs.ErrNotFound
	}

	return nil
}

func (s *userStore) DeleteByOwner(vendeurUUID uuid.UUID, ownerID int) error {
	q := `UPDATE users SET deleted_at = NOW() WHERE uuid = ? AND parent_id = ? AND deleted_at IS NULL`
	raw, err := s.db.Exec(q, vendeurUUID, ownerID)
	if err != nil {
		return err
	}

	count, err := raw.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *userStore) RestoreByOwner(uuid uuid.UUID, ownerID int) error {
	q := `UPDATE users 
          SET deleted_at = NULL, updated_at = NOW() 
          WHERE uuid = ? AND parent_id = ? AND deleted_at IS NOT NULL`

	raw, err := s.db.Exec(q, uuid, ownerID)
	if err != nil {
		return err
	}

	count, err := raw.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errs.ErrNotFound
	}

	return nil
}
