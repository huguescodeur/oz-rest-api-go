package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
)

type ProductStore interface {
	GetAll(ownerID int) ([]*models.Product, error)
	GetByUUID(uuid uuid.UUID, ownerID int) (*models.Product, error)
	GetByID(productID, ownerID int) (*models.Product, error)
	Create(product *models.Product) (*models.Product, error)
	Update(uuid uuid.UUID, product *models.Product) (*models.Product, error)
	Delete(uuid uuid.UUID, ownerID int) error
	Restore(uuid uuid.UUID, ownerID int) error
}

type productStore struct {
	db *sql.DB
}

func NewProductSotre(db *sql.DB) ProductStore {
	return &productStore{db: db}
}

func (s *productStore) GetAll(ownerID int) ([]*models.Product, error) {
	q := `SELECT 
            p.product_id, p.uuid, p.product_name, p.unit_price, p.created_at, p.updated_at, p.user_id,
            c.category_id, c.uuid, c.category_name, c.user_id
        FROM products p
        INNER JOIN categories c ON p.category_id = c.category_id
        WHERE p.user_id = ? AND p.deleted_at IS NULL
        ORDER BY p.created_at DESC, p.product_id DESC`

	raws, err := s.db.Query(q, ownerID)
	if err != nil {
		return nil, err
	}
	defer raws.Close()

	var listProducts []*models.Product

	for raws.Next() {
		product := &models.Product{}
		product.ProductCategory = &models.Category{}

		if err = raws.Scan(
			&product.ProductID,
			&product.ProductUUID,
			&product.ProductName,
			&product.UnitPrice,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.OwnerID,
			&product.ProductCategory.CategoryID,
			&product.ProductCategory.CategoryUUID,
			&product.ProductCategory.CategoryName,
			&product.ProductCategory.OwnerID,
		); err != nil {
			return nil, err
		}

		listProducts = append(listProducts, product)
	}

	return listProducts, nil
}

func (s *productStore) GetByUUID(uuid uuid.UUID, ownerID int) (*models.Product, error) {
	q := `SELECT
	        p.product_id, p.uuid, p.product_name, p.unit_price, p.created_at, p.updated_at, p.user_id,
	        c.category_id, c.uuid, c.category_name, c.user_id
	    FROM products p
	    INNER JOIN categories c ON p.category_id = c.category_id
	    WHERE p.uuid = ? AND p.user_id = ? AND c.user_id = ? AND p.deleted_at IS NULL`

	raw := s.db.QueryRow(q, uuid, ownerID, ownerID)

	product := &models.Product{}
	category := &models.Category{}
	if err := raw.Scan(
		&product.ProductID,
		&product.ProductUUID,
		&product.ProductName,
		&product.UnitPrice,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.OwnerID,
		&category.CategoryID,
		&category.CategoryUUID,
		&category.CategoryName,
		&category.OwnerID,
	); err != nil {
		return nil, err
	}

	product.ProductCategory = category

	return product, nil
}

func (s *productStore) GetByID(productID, ownerID int) (*models.Product, error) {
	q := `SELECT
            p.product_id, p.uuid, p.product_name, p.unit_price, p.created_at, p.updated_at, p.user_id,
            c.category_id, c.uuid, c.category_name, c.user_id
        FROM products p
        INNER JOIN categories c ON p.category_id = c.category_id
        WHERE p.product_id = ? AND p.user_id = ? AND c.user_id = ? AND p.deleted_at IS NULL`

	raw := s.db.QueryRow(q, productID, ownerID, ownerID)

	product := &models.Product{}
	category := &models.Category{}
	if err := raw.Scan(
		&product.ProductID,
		&product.ProductUUID,
		&product.ProductName,
		&product.UnitPrice,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.OwnerID,
		&category.CategoryID,
		&category.CategoryUUID,
		&category.CategoryName,
		&category.OwnerID,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	product.ProductCategory = category
	return product, nil
}

func (s *productStore) Create(product *models.Product) (*models.Product, error) {
	timeNow := time.Now().UTC()

	q := `INSERT INTO products (uuid, product_name, category_id, unit_price, created_at, updated_at, user_id) 
          VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(q,
		product.ProductUUID,
		product.ProductName,
		product.ProductCategory.CategoryID,
		product.UnitPrice,
		timeNow,
		timeNow, product.OwnerID)
	if err != nil {
		return nil, err
	}

	return s.GetByUUID(product.ProductUUID, product.OwnerID)
}

func (s *productStore) Update(uuid uuid.UUID, product *models.Product) (*models.Product, error) {
	updatedAt := time.Now().UTC()

	q := `UPDATE products 
          SET product_name = ?, category_id = ?, unit_price = ?, updated_at = ? 
          WHERE uuid = ? AND user_id = ? AND deleted_at IS NULL`

	raw, err := s.db.Exec(
		q,
		product.ProductName,
		product.ProductCategory.CategoryID,
		product.UnitPrice,
		updatedAt,
		uuid, product.OwnerID)
	if err != nil {
		return nil, err
	}
	count, err := raw.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("Ce produit n'existe pas! Impossible d'update")
	}

	return s.GetByUUID(product.ProductUUID, product.OwnerID)
}

func (s *productStore) Delete(uuid uuid.UUID, ownerID int) error {
	// q := `DELETE FROM products WHERE product_id = ?`
	q := `UPDATE products SET deleted_at = NOW() WHERE uuid = ? AND user_id = ? AND deleted_at IS NULL`

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

func (s *productStore) Restore(u uuid.UUID, ownerID int) error {
	q := `UPDATE products 
          SET deleted_at = NULL, updated_at = NOW() 
          WHERE uuid = ? AND user_id = ? AND deleted_at IS NOT NULL`

	raw, err := s.db.Exec(q, u, ownerID)
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
