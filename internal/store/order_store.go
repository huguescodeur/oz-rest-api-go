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

type OrderStore interface {
	GetAll(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string, limit, offset int) ([]*models.Order, int, error)
	GetByUUID(ctx context.Context, orderUUID uuid.UUID, ownerID int) (*models.Order, error)
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	UpdateStatus(ctx context.Context, orderUUID uuid.UUID, ownerID int, status string) error
	Cancel(ctx context.Context, orderUUID uuid.UUID, ownerID int) error
}

type orderStore struct {
	db *pgxpool.Pool
}

func NewOrderStore(db *pgxpool.Pool) OrderStore {
	return &orderStore{db: db}
}

func (s *orderStore) GetAll(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string, limit, offset int) ([]*models.Order, int, error) {
	cond := ""
	if shopID > 0 {
		cond = fmt.Sprintf(" AND o.shop_id = %d", shopID)
	}
	if dateFrom != "" {
		cond += fmt.Sprintf(" AND DATE(o.created_at) >= '%s'", dateFrom)
	}
	if dateTo != "" {
		cond += fmt.Sprintf(" AND DATE(o.created_at) <= '%s'", dateTo)
	}

	var total int
	_ = s.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM orders o WHERE o.user_id = $1%s`, cond), ownerID).Scan(&total)

	q := fmt.Sprintf(`SELECT o.order_id, o.uuid, o.shop_id, COALESCE(s.shop_name,'') AS shop_name,
	             o.user_id, o.total_amount, o.status,
	             o.customer_name, o.customer_phone, o.payment_method, o.notes,
	             o.delivery_address, o.created_at, o.updated_at
	      FROM orders o
	      LEFT JOIN shops s ON s.shop_id = o.shop_id
	      WHERE o.user_id = $1%s
	      ORDER BY o.created_at DESC
	      LIMIT $2 OFFSET $3`, cond)

	rows, err := s.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{}
		var custName, custPhone, payMethod, notes, delivAddr *string
		if err := rows.Scan(
			&o.OrderID, &o.OrderUUID, &o.ShopID, &o.ShopName,
			&o.UserID, &o.TotalAmount,
			&o.Status, &custName, &custPhone, &payMethod, &notes,
			&delivAddr, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if custName != nil  { o.CustomerName    = *custName  }
		if custPhone != nil { o.CustomerPhone   = *custPhone }
		if payMethod != nil { o.PaymentMethod   = *payMethod }
		if notes != nil     { o.Notes           = *notes     }
		if delivAddr != nil { o.DeliveryAddress = *delivAddr }
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for _, o := range orders {
		items, err := s.getItemsByOrderID(ctx, o.OrderID)
		if err != nil {
			return nil, 0, err
		}
		o.Items = items
	}
	return orders, total, nil
}

func (s *orderStore) GetByUUID(ctx context.Context, orderUUID uuid.UUID, ownerID int) (*models.Order, error) {
	q := `SELECT o.order_id, o.uuid, o.shop_id, COALESCE(s.shop_name,'') AS shop_name,
	             o.user_id, o.total_amount, o.status,
	             o.customer_name, o.customer_phone, o.payment_method, o.notes,
	             o.delivery_address, o.created_at, o.updated_at
	      FROM orders o
	      LEFT JOIN shops s ON s.shop_id = o.shop_id
	      WHERE o.uuid = $1 AND o.user_id = $2`

	o := &models.Order{}
	var custName, custPhone, payMethod, notes, delivAddr *string
	if err := s.db.QueryRow(ctx, q, orderUUID, ownerID).Scan(
		&o.OrderID, &o.OrderUUID, &o.ShopID, &o.ShopName,
		&o.UserID, &o.TotalAmount,
		&o.Status, &custName, &custPhone, &payMethod, &notes,
		&delivAddr, &o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	if custName != nil  { o.CustomerName    = *custName  }
	if custPhone != nil { o.CustomerPhone   = *custPhone }
	if payMethod != nil { o.PaymentMethod   = *payMethod }
	if notes != nil     { o.Notes           = *notes     }
	if delivAddr != nil { o.DeliveryAddress = *delivAddr }

	items, err := s.getItemsByOrderID(ctx, o.OrderID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

func (s *orderStore) getItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	rows, err := s.db.Query(ctx,
		`SELECT item_id, order_id, product_id, quantity, unit_price FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		item := models.OrderItem{}
		if err := rows.Scan(&item.ItemID, &item.OrderID, &item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *orderStore) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()

	status := order.Status
	if status == "" {
		status = "PENDING"
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (uuid, shop_id, user_id, total_amount, status,
		                    customer_name, customer_phone, payment_method, notes,
		                    delivery_address, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		 RETURNING order_id`,
		order.OrderUUID, order.ShopID, order.UserID, order.TotalAmount, status,
		nullStr(order.CustomerName), nullStr(order.CustomerPhone),
		nullStr(order.PaymentMethod), nullStr(order.Notes),
		nullStr(order.DeliveryAddress), now,
	).Scan(&order.OrderID)
	if err != nil {
		return nil, err
	}

	for i := range order.Items {
		item := &order.Items[i]
		item.OrderID = order.OrderID

		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, unit_price)
             VALUES ($1, $2, $3, $4) RETURNING item_id`,
			order.OrderID, item.ProductID, item.Quantity, item.UnitPrice,
		).Scan(&item.ItemID)
		if err != nil {
			return nil, err
		}

		tag, err := tx.Exec(ctx,
			`UPDATE stocks SET quantity = quantity - $1, updated_at = $2
             WHERE product_id = $3 AND shop_id = $4 AND quantity >= $1`,
			item.Quantity, now, item.ProductID, order.ShopID,
		)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, errors.New("stock insuffisant pour ce produit")
		}

		movComment := fmt.Sprintf("Commande #%s", order.OrderUUID.String()[:8])
		if order.Notes != "" {
			movComment += " — " + order.Notes
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, shop_id, user_id, quantity_change, movement_type, comment, created_at)
             VALUES ($1, $2, $3, $4, 'SALE', $5, $6)`,
			item.ProductID, order.ShopID, order.UserID, -item.Quantity, movComment, now,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetByUUID(ctx, order.OrderUUID, order.UserID)
}

func (s *orderStore) UpdateStatus(ctx context.Context, orderUUID uuid.UUID, ownerID int, status string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE uuid = $2 AND user_id = $3`,
		status, orderUUID, ownerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *orderStore) Cancel(ctx context.Context, orderUUID uuid.UUID, ownerID int) error {
	order, err := s.GetByUUID(ctx, orderUUID, ownerID)
	if err != nil {
		return err
	}
	if order.Status == "CANCELLED" {
		return errors.New("commande déjà annulée")
	}
	if order.Status == "DELIVERED" {
		return errors.New("impossible d'annuler une commande livrée")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()

	for _, item := range order.Items {
		if _, err = tx.Exec(ctx,
			`UPDATE stocks SET quantity = quantity + $1, updated_at = $2 WHERE product_id = $3 AND shop_id = $4`,
			item.Quantity, now, item.ProductID, order.ShopID,
		); err != nil {
			return err
		}

		if _, err = tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, shop_id, user_id, quantity_change, movement_type, comment, created_at)
             VALUES ($1, $2, $3, $4, 'ADJUSTMENT', $5, $6)`,
			item.ProductID, order.ShopID, ownerID, item.Quantity,
			"Annulation commande "+orderUUID.String(), now,
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec(ctx,
		`UPDATE orders SET status = 'CANCELLED', updated_at = $1 WHERE uuid = $2`,
		now, orderUUID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
