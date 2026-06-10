package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductStore interface {
	GetAll(ctx context.Context, ownerID, limit, offset int) ([]*models.Product, int, error)
	GetByUUID(ctx context.Context, uuid uuid.UUID, ownerID int) (*models.Product, error)
	GetByID(ctx context.Context, productID, ownerID int) (*models.Product, error)
	Create(ctx context.Context, product *models.Product) (*models.Product, error)
	Update(ctx context.Context, uuid uuid.UUID, product *models.Product) (*models.Product, error)
	Delete(ctx context.Context, uuid uuid.UUID, ownerID int) error
	Restore(ctx context.Context, uuid uuid.UUID, ownerID int) error
}

type productStore struct {
	db *pgxpool.Pool
}

func NewProductStore(db *pgxpool.Pool) ProductStore {
	return &productStore{db: db}
}

func (s *productStore) GetAll(ctx context.Context, ownerID, limit, offset int) ([]*models.Product, int, error) {
	var total int
	_ = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM products WHERE user_id = $1 AND deleted_at IS NULL`, ownerID,
	).Scan(&total)

	q := `SELECT
            p.product_id, p.uuid, p.product_name, p.unit_price, p.created_at, p.updated_at, p.user_id,
            c.category_id, c.uuid, c.category_name, c.user_id
        FROM products p
        INNER JOIN categories c ON p.category_id = c.category_id
        WHERE p.user_id = $1 AND p.deleted_at IS NULL
        ORDER BY p.created_at DESC, p.product_id DESC
        LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var listProducts []*models.Product
	for rows.Next() {
		product := &models.Product{}
		product.ProductCategory = &models.Category{}

		if err = rows.Scan(
			&product.ProductID, &product.ProductUUID, &product.ProductName,
			&product.UnitPrice, &product.CreatedAt, &product.UpdatedAt, &product.OwnerID,
			&product.ProductCategory.CategoryID, &product.ProductCategory.CategoryUUID,
			&product.ProductCategory.CategoryName, &product.ProductCategory.OwnerID,
		); err != nil {
			return nil, 0, err
		}
		listProducts = append(listProducts, product)
	}
	return listProducts, total, rows.Err()
}

func (s *productStore) GetByUUID(ctx context.Context, uuid uuid.UUID, ownerID int) (*models.Product, error) {
	q := `SELECT
            p.product_id, p.uuid, p.product_name, p.unit_price, p.created_at, p.updated_at, p.user_id,
            c.category_id, c.uuid, c.category_name, c.user_id
        FROM products p
        INNER JOIN categories c ON p.category_id = c.category_id
        WHERE p.uuid = $1 AND p.user_id = $2 AND c.user_id = $2 AND p.deleted_at IS NULL`

	product := &models.Product{}
	category := &models.Category{}
	if err := s.db.QueryRow(ctx, q, uuid, ownerID).Scan(
		&product.ProductID, &product.ProductUUID, &product.ProductName,
		&product.UnitPrice, &product.CreatedAt, &product.UpdatedAt, &product.OwnerID,
		&category.CategoryID, &category.CategoryUUID, &category.CategoryName, &category.OwnerID,
	); err != nil {
		return nil, err
	}
	product.ProductCategory = category
	return product, nil
}

func (s *productStore) GetByID(ctx context.Context, productID, ownerID int) (*models.Product, error) {
	q := `SELECT
            p.product_id, p.uuid, p.product_name, p.unit_price, p.created_at, p.updated_at, p.user_id,
            c.category_id, c.uuid, c.category_name, c.user_id
        FROM products p
        INNER JOIN categories c ON p.category_id = c.category_id
        WHERE p.product_id = $1 AND p.user_id = $2 AND c.user_id = $2 AND p.deleted_at IS NULL`

	product := &models.Product{}
	category := &models.Category{}
	if err := s.db.QueryRow(ctx, q, productID, ownerID).Scan(
		&product.ProductID, &product.ProductUUID, &product.ProductName,
		&product.UnitPrice, &product.CreatedAt, &product.UpdatedAt, &product.OwnerID,
		&category.CategoryID, &category.CategoryUUID, &category.CategoryName, &category.OwnerID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	product.ProductCategory = category
	return product, nil
}

func (s *productStore) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	timeNow := time.Now().UTC()
	q := `INSERT INTO products (uuid, product_name, category_id, unit_price, created_at, updated_at, user_id)
          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.Exec(ctx, q,
		product.ProductUUID, product.ProductName,
		product.ProductCategory.CategoryID, product.UnitPrice,
		timeNow, timeNow, product.OwnerID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetByUUID(ctx, product.ProductUUID, product.OwnerID)
}

func (s *productStore) Update(ctx context.Context, uuid uuid.UUID, product *models.Product) (*models.Product, error) {
	updatedAt := time.Now().UTC()
	q := `UPDATE products
          SET product_name = $1, category_id = $2, unit_price = $3, updated_at = $4
          WHERE uuid = $5 AND user_id = $6 AND deleted_at IS NULL`

	tag, err := s.db.Exec(ctx, q,
		product.ProductName, product.ProductCategory.CategoryID,
		product.UnitPrice, updatedAt, uuid, product.OwnerID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("Ce produit n'existe pas! Impossible d'update")
	}
	return s.GetByUUID(ctx, product.ProductUUID, product.OwnerID)
}

func (s *productStore) Delete(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE products SET deleted_at = NOW() WHERE uuid = $1 AND user_id = $2 AND deleted_at IS NULL`,
		uuid, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *productStore) Restore(ctx context.Context, u uuid.UUID, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE products SET deleted_at = NULL, updated_at = NOW() WHERE uuid = $1 AND user_id = $2 AND deleted_at IS NOT NULL`,
		u, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}
