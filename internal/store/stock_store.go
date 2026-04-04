package store

import (
	"database/sql"
	"time"

	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/errs"
)

type StockStore interface {
	GetAllStocks(ownerID int) ([]*models.Stock, error)
	GetStockByProductAndShop(productID, shopID, ownerID int) (*models.Stock, error)
	InitStock(productID, shopID, ownerID int) error
	AdjustStock(productID, shopID, ownerID, quantityChange int, movementType, comment string) error
	GetMovements(ownerID int) ([]*models.StockMovement, error)
	GetMovementsByProduct(productID, ownerID int) ([]*models.StockMovement, error)
	GetMovementsByShop(shopID, ownerID int) ([]*models.StockMovement, error)
}

type stockStore struct {
	db *sql.DB
}

func NewStockStore(db *sql.DB) StockStore {
	return &stockStore{db: db}
}

func (s *stockStore) GetAllStocks(ownerID int) ([]*models.Stock, error) {
	q := `SELECT 
            st.product_id, p.product_name, st.shop_id, sh.shop_name, st.quantity, st.updated_at
          FROM stocks st
          INNER JOIN products p ON st.product_id = p.product_id
          INNER JOIN shops sh ON st.shop_id = sh.shop_id
          WHERE p.user_id = ? AND p.deleted_at IS NULL AND sh.deleted_at IS NULL
          ORDER BY st.updated_at DESC`

	rows, err := s.db.Query(q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []*models.Stock
	for rows.Next() {
		stock := &models.Stock{}
		if err := rows.Scan(
			&stock.ProductID,
			&stock.ProductName,
			&stock.ShopID,
			&stock.ShopName,
			&stock.Quantity,
			&stock.UpdatedAt,
		); err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (s *stockStore) GetStockByProductAndShop(productID, shopID, ownerID int) (*models.Stock, error) {
	q := `SELECT 
            st.product_id, p.product_name, st.shop_id, sh.shop_name, st.quantity, st.updated_at
          FROM stocks st
          INNER JOIN products p ON st.product_id = p.product_id
          INNER JOIN shops sh ON st.shop_id = sh.shop_id
          WHERE st.product_id = ? AND st.shop_id = ? AND p.user_id = ?
            AND p.deleted_at IS NULL AND sh.deleted_at IS NULL`

	row := s.db.QueryRow(q, productID, shopID, ownerID)

	stock := &models.Stock{}
	if err := row.Scan(
		&stock.ProductID,
		&stock.ProductName,
		&stock.ShopID,
		&stock.ShopName,
		&stock.Quantity,
		&stock.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return stock, nil
}

func (s *stockStore) InitStock(productID, shopID, ownerID int) error {
	q := `INSERT INTO stocks (product_id, shop_id, quantity, updated_at)
          VALUES (?, ?, 0, ?)
          ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at)`

	_, err := s.db.Exec(q, productID, shopID, time.Now().UTC())
	return err
}

func (s *stockStore) AdjustStock(productID, shopID, ownerID, quantityChange int, movementType, comment string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateQ := `UPDATE stocks 
                SET quantity = quantity + ?, updated_at = ?
                WHERE product_id = ? AND shop_id = ?`

	res, err := tx.Exec(updateQ, quantityChange, time.Now().UTC(), productID, shopID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errs.ErrNotFound
	}

	movementQ := `INSERT INTO stock_movements 
                  (product_id, shop_id, user_id, quantity_change, movement_type, comment, created_at)
                  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(movementQ, productID, shopID, ownerID, quantityChange, movementType, comment, time.Now().UTC())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *stockStore) GetMovements(ownerID int) ([]*models.StockMovement, error) {
	q := `SELECT 
            sm.movement_id, sm.product_id, sm.shop_id, sm.user_id,
            sm.quantity_change, sm.movement_type, sm.comment, sm.created_at
          FROM stock_movements sm
          INNER JOIN products p ON sm.product_id = p.product_id
          WHERE sm.user_id = ?
          ORDER BY sm.created_at DESC`

	rows, err := s.db.Query(q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*models.StockMovement
	for rows.Next() {
		m := &models.StockMovement{}
		if err := rows.Scan(
			&m.MovementID,
			&m.ProductID,
			&m.ShopID,
			&m.UserID,
			&m.QuantityChange,
			&m.MovementType,
			&m.Comment,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}

	return movements, nil
}

func (s *stockStore) GetMovementsByProduct(productID, ownerID int) ([]*models.StockMovement, error) {
	q := `SELECT 
            sm.movement_id, sm.product_id, sm.shop_id, sm.user_id,
            sm.quantity_change, sm.movement_type, sm.comment, sm.created_at
          FROM stock_movements sm
          INNER JOIN products p ON sm.product_id = p.product_id
          WHERE sm.product_id = ? AND sm.user_id = ?
          ORDER BY sm.created_at DESC`

	rows, err := s.db.Query(q, productID, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*models.StockMovement
	for rows.Next() {
		m := &models.StockMovement{}
		if err := rows.Scan(
			&m.MovementID,
			&m.ProductID,
			&m.ShopID,
			&m.UserID,
			&m.QuantityChange,
			&m.MovementType,
			&m.Comment,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}

	return movements, nil
}

func (s *stockStore) GetMovementsByShop(shopID, ownerID int) ([]*models.StockMovement, error) {
	q := `SELECT 
            sm.movement_id, sm.product_id, sm.shop_id, sm.user_id,
            sm.quantity_change, sm.movement_type, sm.comment, sm.created_at
          FROM stock_movements sm
          INNER JOIN shops sh ON sm.shop_id = sh.shop_id
          WHERE sm.shop_id = ? AND sm.user_id = ?
          ORDER BY sm.created_at DESC`

	rows, err := s.db.Query(q, shopID, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*models.StockMovement
	for rows.Next() {
		m := &models.StockMovement{}
		if err := rows.Scan(
			&m.MovementID,
			&m.ProductID,
			&m.ShopID,
			&m.UserID,
			&m.QuantityChange,
			&m.MovementType,
			&m.Comment,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}

	return movements, nil
}
