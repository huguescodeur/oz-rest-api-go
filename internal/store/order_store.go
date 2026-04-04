package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
)

type OrderStore interface {
	GetAll(ownerID int) ([]*models.Order, error)
	GetByUUID(orderUUID uuid.UUID, ownerID int) (*models.Order, error)
	Create(order *models.Order) (*models.Order, error)
	Cancel(orderUUID uuid.UUID, ownerID int) error
}

type orderStore struct {
	db *sql.DB
}

func NewOrderStore(db *sql.DB) OrderStore {
	return &orderStore{db: db}
}

func (s *orderStore) GetAll(ownerID int) ([]*models.Order, error) {
	q := `SELECT 
            o.order_id, o.uuid, o.shop_id, o.user_id, o.total_amount, o.created_at
          FROM orders o
          WHERE o.user_id = ?
          ORDER BY o.created_at DESC`

	rows, err := s.db.Query(q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(
			&order.OrderID,
			&order.OrderUUID,
			&order.ShopID,
			&order.UserID,
			&order.TotalAmount,
			&order.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	for _, order := range orders {
		items, err := s.getItemsByOrderID(order.OrderID)
		if err != nil {
			return nil, err
		}
		order.Items = items
	}

	return orders, nil
}

func (s *orderStore) GetByUUID(orderUUID uuid.UUID, ownerID int) (*models.Order, error) {
	q := `SELECT 
            o.order_id, o.uuid, o.shop_id, o.user_id, o.total_amount, o.created_at
          FROM orders o
          WHERE o.uuid = ? AND o.user_id = ?`

	row := s.db.QueryRow(q, orderUUID, ownerID)

	order := &models.Order{}
	if err := row.Scan(
		&order.OrderID,
		&order.OrderUUID,
		&order.ShopID,
		&order.UserID,
		&order.TotalAmount,
		&order.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	items, err := s.getItemsByOrderID(order.OrderID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (s *orderStore) getItemsByOrderID(orderID int) ([]models.OrderItem, error) {
	q := `SELECT item_id, order_id, product_id, quantity, unit_price
          FROM order_items
          WHERE order_id = ?`

	rows, err := s.db.Query(q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		item := models.OrderItem{}
		if err := rows.Scan(
			&item.ItemID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *orderStore) Create(order *models.Order) (*models.Order, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	orderQ := `INSERT INTO orders (uuid, shop_id, user_id, total_amount, created_at)
               VALUES (?, ?, ?, ?, ?)`

	res, err := tx.Exec(orderQ, order.OrderUUID, order.ShopID, order.UserID, order.TotalAmount, now)
	if err != nil {
		return nil, err
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	order.OrderID = int(orderID)

	itemQ := `INSERT INTO order_items (order_id, product_id, quantity, unit_price)
              VALUES (?, ?, ?, ?)`

	stockQ := `UPDATE stocks 
               SET quantity = quantity - ?, updated_at = ?
               WHERE product_id = ? AND shop_id = ? AND quantity >= ?`

	movementQ := `INSERT INTO stock_movements 
                  (product_id, shop_id, user_id, quantity_change, movement_type, comment, created_at)
                  VALUES (?, ?, ?, ?, 'SALE', ?, ?)`

	for i := range order.Items {
		item := &order.Items[i]
		item.OrderID = order.OrderID

		itemRes, err := tx.Exec(itemQ, order.OrderID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			return nil, err
		}
		itemID, _ := itemRes.LastInsertId()
		item.ItemID = int(itemID)

		stockRes, err := tx.Exec(stockQ, item.Quantity, now, item.ProductID, order.ShopID, item.Quantity)
		if err != nil {
			return nil, err
		}
		count, _ := stockRes.RowsAffected()
		if count == 0 {
			return nil, errors.New("stock insuffisant pour effectuer cette opération")
		}

		comment := "Vente via commande"
		_, err = tx.Exec(movementQ, item.ProductID, order.ShopID, order.UserID, -item.Quantity, comment, now)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetByUUID(order.OrderUUID, order.UserID)
}

func (s *orderStore) Cancel(orderUUID uuid.UUID, ownerID int) error {
	order, err := s.GetByUUID(orderUUID, ownerID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	stockQ := `UPDATE stocks SET quantity = quantity + ?, updated_at = ?
               WHERE product_id = ? AND shop_id = ?`

	movementQ := `INSERT INTO stock_movements 
                  (product_id, shop_id, user_id, quantity_change, movement_type, comment, created_at)
                  VALUES (?, ?, ?, ?, 'ADJUSTMENT', ?, ?)`

	for _, item := range order.Items {
		_, err = tx.Exec(stockQ, item.Quantity, now, item.ProductID, order.ShopID)
		if err != nil {
			return err
		}

		comment := "Annulation commande " + orderUUID.String()
		_, err = tx.Exec(movementQ, item.ProductID, order.ShopID, ownerID, item.Quantity, comment, now)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`DELETE FROM order_items WHERE order_id = ?`, order.OrderID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM orders WHERE order_id = ?`, order.OrderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
