package store

import (
	"context"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryStore interface {
	GetAll(ctx context.Context, ownerID int) ([]*models.Category, error)
	GetByID(ctx context.Context, id, ownerID int) (*models.Category, error)
	Create(ctx context.Context, c *models.Category) (*models.Category, error)
	Update(ctx context.Context, id, ownerID int, name string) (*models.Category, error)
	Delete(ctx context.Context, id, ownerID int) error
	Restore(ctx context.Context, id, ownerID int) error
}

type categoryStore struct {
	db *pgxpool.Pool
}

func NewCategoryStore(db *pgxpool.Pool) CategoryStore {
	return &categoryStore{db: db}
}

func (s *categoryStore) GetAll(ctx context.Context, ownerID int) ([]*models.Category, error) {
	rows, err := s.db.Query(ctx,
		`SELECT category_id, uuid, category_name, user_id FROM categories
         WHERE user_id = $1 AND deleted_at IS NULL ORDER BY category_name`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*models.Category
	for rows.Next() {
		c := &models.Category{}
		if err := rows.Scan(&c.CategoryID, &c.CategoryUUID, &c.CategoryName, &c.OwnerID); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (s *categoryStore) GetByID(ctx context.Context, id, ownerID int) (*models.Category, error) {
	c := &models.Category{}
	err := s.db.QueryRow(ctx,
		`SELECT category_id, uuid, category_name, user_id FROM categories
         WHERE category_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, ownerID,
	).Scan(&c.CategoryID, &c.CategoryUUID, &c.CategoryName, &c.OwnerID)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	return c, nil
}

func (s *categoryStore) Create(ctx context.Context, c *models.Category) (*models.Category, error) {
	var id int
	err := s.db.QueryRow(ctx,
		`INSERT INTO categories (uuid, category_name, user_id) VALUES ($1, $2, $3) RETURNING category_id`,
		c.CategoryUUID, c.CategoryName, c.OwnerID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	c.CategoryID = id
	return c, nil
}

func (s *categoryStore) Update(ctx context.Context, id, ownerID int, name string) (*models.Category, error) {
	tag, err := s.db.Exec(ctx,
		`UPDATE categories SET category_name = $1 WHERE category_id = $2 AND user_id = $3 AND deleted_at IS NULL`,
		name, id, ownerID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errs.ErrNotFound
	}
	return s.GetByID(ctx, id, ownerID)
}

func (s *categoryStore) Delete(ctx context.Context, id, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE categories SET deleted_at = NOW() WHERE category_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, ownerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *categoryStore) Restore(ctx context.Context, id, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE categories SET deleted_at = NULL WHERE category_id = $1 AND user_id = $2 AND deleted_at IS NOT NULL`,
		id, ownerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}
