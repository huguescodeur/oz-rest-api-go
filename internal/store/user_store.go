package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore interface {
	CreateUser(ctx context.Context, u *models.User) (*models.User, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.User, int, error)
	GetAllByOwner(ctx context.Context, ownerID, limit, offset int) ([]*models.User, int, error)
	GetArchivedByOwner(ctx context.Context, ownerID, limit, offset int) ([]*models.User, int, error)
	GetByUUID(ctx context.Context, userUUID uuid.UUID) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	Update(ctx context.Context, u *models.User) (*models.User, error)
	GetByUUIDWithParent(ctx context.Context, userUUID uuid.UUID, ownerID int) (*models.User, error)
	Delete(ctx context.Context, uuid uuid.UUID) error
	Restore(ctx context.Context, uuid uuid.UUID) error
	DeleteByOwner(ctx context.Context, vendeurID uuid.UUID, ownerID int) error
	RestoreByOwner(ctx context.Context, uuid uuid.UUID, ownerID int) error
	AssignShop(ctx context.Context, userID, shopID int) error
	AssignShops(ctx context.Context, userID int, shopIDs []int) error
	UpdateMe(ctx context.Context, userID int, firstname, lastname, username, email, phone string) error
	GetMyShops(ctx context.Context, userID int) ([]*models.Shop, error)
}

type userStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) UserStore {
	return &userStore{db: db}
}

const shopSubqueries = `
	(SELECT us.shop_id FROM user_shops us WHERE us.user_id = u.id ORDER BY us.assigned_at LIMIT 1) AS shop_id,
	(SELECT STRING_AGG(s.shop_name, ', ' ORDER BY us.assigned_at) FROM user_shops us JOIN shops s ON s.shop_id = us.shop_id WHERE us.user_id = u.id) AS shop_name`

func scanUser(row interface {
	Scan(dest ...any) error
}) (*models.User, error) {
	u := &models.User{}
	var shopID *int
	var shopName *string
	err := row.Scan(
		&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname,
		&u.Email, &u.Phone, &u.Role, &u.ParentID, &u.CreatedAt, &u.DeletedAt,
		&shopID, &shopName,
	)
	if err != nil {
		return nil, err
	}
	if shopID != nil   { u.ShopID   = shopID    }
	if shopName != nil { u.ShopName = *shopName  }
	return u, nil
}

func (s *userStore) CreateUser(ctx context.Context, u *models.User) (*models.User, error) {
	q := `INSERT INTO users (
            uuid, username, firstname, lastname, email,
            phone, password_hash, role, parent_id,
            created_at, updated_at
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
          RETURNING id`

	err := s.db.QueryRow(ctx, q,
		u.UUID, u.Username, u.Firstname, u.Lastname,
		u.Email, u.Phone, u.PasswordHash, u.Role, u.ParentID,
	).Scan(&u.ID)
	if err != nil {
		return nil, err
	}

	if u.ShopID != nil && *u.ShopID > 0 {
		_, err = s.db.Exec(ctx,
			`INSERT INTO user_shops (user_id, shop_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			u.ID, *u.ShopID,
		)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func (s *userStore) GetAll(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	var total int
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)

	q := `SELECT u.id, u.uuid, u.username, u.firstname, u.lastname, u.email, u.phone, u.role, u.parent_id, u.created_at, u.deleted_at,` +
		shopSubqueries + `
          FROM users u
          ORDER BY u.created_at DESC
          LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (s *userStore) GetAllByOwner(ctx context.Context, ownerID, limit, offset int) ([]*models.User, int, error) {
	var total int
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE parent_id = $1 AND deleted_at IS NULL", ownerID).Scan(&total)

	q := `SELECT u.id, u.uuid, u.username, u.firstname, u.lastname, u.email, u.phone, u.role, u.parent_id, u.created_at, u.deleted_at,` +
		shopSubqueries + `
          FROM users u
          WHERE u.parent_id = $1 AND u.deleted_at IS NULL
          ORDER BY u.username ASC
          LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (s *userStore) GetArchivedByOwner(ctx context.Context, ownerID, limit, offset int) ([]*models.User, int, error) {
	var total int
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE parent_id = $1 AND deleted_at IS NOT NULL", ownerID).Scan(&total)

	q := `SELECT u.id, u.uuid, u.username, u.firstname, u.lastname, u.email, u.phone, u.role, u.parent_id, u.created_at, u.deleted_at,` +
		shopSubqueries + `
          FROM users u
          WHERE u.parent_id = $1 AND u.deleted_at IS NOT NULL
          ORDER BY u.deleted_at DESC
          LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (s *userStore) GetByUUID(ctx context.Context, userUUID uuid.UUID) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, uuid, username, firstname, lastname, email, phone, password_hash, role, parent_id, created_at
          FROM users
          WHERE uuid = $1 AND deleted_at IS NULL`

	err := s.db.QueryRow(ctx, q, userUUID).Scan(
		&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname,
		&u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.ParentID, &u.CreatedAt,
	)
	return u, err
}

func (s *userStore) GetByID(ctx context.Context, id int) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, uuid, username, firstname, lastname, email, phone, password_hash, role, parent_id, created_at
	      FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := s.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname,
		&u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.ParentID, &u.CreatedAt,
	)
	return u, err
}

func (s *userStore) GetByUUIDWithParent(ctx context.Context, userUUID uuid.UUID, ownerID int) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, uuid, username, firstname, lastname, email, phone, role, parent_id, created_at
          FROM users
          WHERE uuid = $1 AND (id = $2 OR parent_id = $2) AND deleted_at IS NULL`

	err := s.db.QueryRow(ctx, q, userUUID, ownerID).Scan(
		&u.ID, &u.UUID, &u.Username, &u.Firstname, &u.Lastname,
		&u.Email, &u.Phone, &u.Role, &u.ParentID, &u.CreatedAt,
	)
	return u, err
}

func (s *userStore) Update(ctx context.Context, u *models.User) (*models.User, error) {
	updatedAt := time.Now().UTC()
	q := `UPDATE users SET
            username = $1, firstname = $2, lastname = $3, email = $4,
            phone = $5, role = $6, parent_id = $7, updated_at = $8
          WHERE uuid = $9 AND deleted_at IS NULL`

	tag, err := s.db.Exec(ctx, q,
		u.Username, u.Firstname, u.Lastname, u.Email,
		u.Phone, u.Role, u.ParentID, updatedAt, u.UUID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("User n'existe pas! Impossible d'update")
	}

	return s.GetByUUID(ctx, u.UUID)
}

func (s *userStore) AssignShop(ctx context.Context, userID, shopID int) error {
	_, err := s.db.Exec(ctx, `DELETE FROM user_shops WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	if shopID > 0 {
		_, err = s.db.Exec(ctx, `INSERT INTO user_shops (user_id, shop_id) VALUES ($1, $2)`, userID, shopID)
	}
	return err
}

func (s *userStore) AssignShops(ctx context.Context, userID int, shopIDs []int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `DELETE FROM user_shops WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, sid := range shopIDs {
		if sid > 0 {
			if _, err = tx.Exec(ctx, `INSERT INTO user_shops (user_id, shop_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, sid); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (s *userStore) UpdateMe(ctx context.Context, userID int, firstname, lastname, username, email, phone string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE users SET firstname=$1, lastname=$2, username=$3, email=$4, phone=$5, updated_at=NOW()
		 WHERE id=$6 AND deleted_at IS NULL`,
		firstname, lastname, username, email, phone, userID,
	)
	return err
}

func (s *userStore) Delete(ctx context.Context, uuid uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET deleted_at = NOW() WHERE uuid = $1 AND deleted_at IS NULL`, uuid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *userStore) Restore(ctx context.Context, uuid uuid.UUID) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users SET deleted_at = NULL, updated_at = NOW() WHERE uuid = $1 AND deleted_at IS NOT NULL`, uuid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *userStore) DeleteByOwner(ctx context.Context, vendeurUUID uuid.UUID, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users SET deleted_at = NOW() WHERE uuid = $1 AND parent_id = $2 AND deleted_at IS NULL`,
		vendeurUUID, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *userStore) GetMyShops(ctx context.Context, userID int) ([]*models.Shop, error) {
	rows, err := s.db.Query(ctx, `
		SELECT s.shop_id, s.uuid, s.shop_name, s.shop_address,
		       COALESCE(s.shop_phone,''), COALESCE(s.shop_mail,''),
		       s.created_at, s.updated_at, s.user_id, us.assigned_at
		FROM user_shops us
		JOIN shops s ON s.shop_id = us.shop_id
		WHERE us.user_id = $1 AND s.deleted_at IS NULL
		ORDER BY us.assigned_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shops []*models.Shop
	for rows.Next() {
		sh := &models.Shop{}
		if err := rows.Scan(
			&sh.ShopID, &sh.ShopUUID, &sh.ShopName, &sh.ShopAddress,
			&sh.ShopPhone, &sh.ShopMail,
			&sh.CreatedAt, &sh.UpdatedAt, &sh.OwnerID, &sh.AssignedAt,
		); err != nil {
			return nil, err
		}
		shops = append(shops, sh)
	}
	return shops, rows.Err()
}

func (s *userStore) RestoreByOwner(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users SET deleted_at = NULL, updated_at = NOW() WHERE uuid = $1 AND parent_id = $2 AND deleted_at IS NOT NULL`,
		uuid, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}
