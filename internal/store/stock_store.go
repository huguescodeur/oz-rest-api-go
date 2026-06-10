package store

import (
	"context"
	"fmt"
	"time"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockStore interface {
	GetAllStocks(ctx context.Context, ownerID int) ([]*models.Stock, error)
	GetLowStocks(ctx context.Context, ownerID, shopID int) ([]*models.Stock, error)
	GetStockByProductAndShop(ctx context.Context, productID, shopID, ownerID int) (*models.Stock, error)
	InitStock(ctx context.Context, productID, shopID, ownerID int) error
	SetMinStock(ctx context.Context, productID, shopID, ownerID, minStock int) error
	AdjustStock(ctx context.Context, productID, shopID, ownerID, quantityChange int, movementType, comment string) error
	GetMovements(ctx context.Context, ownerID int) ([]*models.StockMovement, error)
	GetMovementsByProduct(ctx context.Context, productID, ownerID int) ([]*models.StockMovement, error)
	GetMovementsByShop(ctx context.Context, shopID, ownerID int) ([]*models.StockMovement, error)
	GetLogs(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string, limit, offset int) ([]*models.StockMovement, int, error)
}

type stockStore struct {
	db *pgxpool.Pool
}

func NewStockStore(db *pgxpool.Pool) StockStore {
	return &stockStore{db: db}
}

const stockSelectBase = `SELECT
    st.product_id, p.product_name, st.shop_id, sh.shop_name,
    st.quantity, p.unit_price, st.min_stock, (st.quantity <= st.min_stock) AS low_stock, st.updated_at
  FROM stocks st
  INNER JOIN products p ON st.product_id = p.product_id
  INNER JOIN shops sh ON st.shop_id = sh.shop_id`

func (s *stockStore) scanStocks(ctx context.Context, q string, args ...any) ([]*models.Stock, error) {
	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []*models.Stock
	for rows.Next() {
		st := &models.Stock{}
		if err := rows.Scan(
			&st.ProductID, &st.ProductName,
			&st.ShopID, &st.ShopName,
			&st.Quantity, &st.UnitPrice, &st.MinStock, &st.LowStock, &st.UpdatedAt,
		); err != nil {
			return nil, err
		}
		stocks = append(stocks, st)
	}
	return stocks, rows.Err()
}

func (s *stockStore) GetAllStocks(ctx context.Context, ownerID int) ([]*models.Stock, error) {
	q := stockSelectBase + `
  WHERE p.user_id = $1 AND p.deleted_at IS NULL AND sh.deleted_at IS NULL
  ORDER BY st.updated_at DESC`
	return s.scanStocks(ctx, q, ownerID)
}

func (s *stockStore) GetLowStocks(ctx context.Context, ownerID, shopID int) ([]*models.Stock, error) {
	shopCond := ""
	if shopID > 0 {
		shopCond = fmt.Sprintf(" AND st.shop_id = %d", shopID)
	}
	q := stockSelectBase + fmt.Sprintf(`
  WHERE p.user_id = $1 AND p.deleted_at IS NULL AND sh.deleted_at IS NULL
    AND st.quantity <= st.min_stock%s
  ORDER BY st.quantity ASC`, shopCond)
	return s.scanStocks(ctx, q, ownerID)
}

func (s *stockStore) SetMinStock(ctx context.Context, productID, shopID, ownerID, minStock int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE stocks SET min_stock = $1, updated_at = NOW()
         WHERE product_id = $2 AND shop_id = $3`,
		minStock, productID, shopID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *stockStore) GetStockByProductAndShop(ctx context.Context, productID, shopID, ownerID int) (*models.Stock, error) {
	q := stockSelectBase + `
  WHERE st.product_id = $1 AND st.shop_id = $2 AND p.user_id = $3
    AND p.deleted_at IS NULL AND sh.deleted_at IS NULL`

	stock := &models.Stock{}
	if err := s.db.QueryRow(ctx, q, productID, shopID, ownerID).Scan(
		&stock.ProductID, &stock.ProductName,
		&stock.ShopID, &stock.ShopName,
		&stock.Quantity, &stock.UnitPrice, &stock.MinStock, &stock.LowStock, &stock.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return stock, nil
}

func (s *stockStore) InitStock(ctx context.Context, productID, shopID, ownerID int) error {
	q := `INSERT INTO stocks (product_id, shop_id, quantity, updated_at)
          VALUES ($1, $2, 0, $3)
          ON CONFLICT (product_id, shop_id) DO UPDATE SET updated_at = EXCLUDED.updated_at`

	_, err := s.db.Exec(ctx, q, productID, shopID, time.Now().UTC())
	return err
}

func (s *stockStore) AdjustStock(ctx context.Context, productID, shopID, ownerID, quantityChange int, movementType, comment string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()

	tag, err := tx.Exec(ctx,
		`UPDATE stocks SET quantity = quantity + $1, updated_at = $2 WHERE product_id = $3 AND shop_id = $4`,
		quantityChange, now, productID, shopID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO stock_movements (product_id, shop_id, user_id, quantity_change, movement_type, comment, created_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		productID, shopID, ownerID, quantityChange, movementType, comment, now,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *stockStore) GetMovements(ctx context.Context, ownerID int) ([]*models.StockMovement, error) {
	q := `SELECT sm.movement_id, sm.product_id, sm.shop_id, sm.user_id,
                 sm.quantity_change, sm.movement_type, sm.comment, sm.created_at
          FROM stock_movements sm
          INNER JOIN products p ON sm.product_id = p.product_id
          WHERE sm.user_id = $1
          ORDER BY sm.created_at DESC`

	return s.scanMovements(ctx, q, ownerID)
}

func (s *stockStore) GetMovementsByProduct(ctx context.Context, productID, ownerID int) ([]*models.StockMovement, error) {
	q := `SELECT sm.movement_id, sm.product_id, sm.shop_id, sm.user_id,
                 sm.quantity_change, sm.movement_type, sm.comment, sm.created_at
          FROM stock_movements sm
          INNER JOIN products p ON sm.product_id = p.product_id
          WHERE sm.product_id = $1 AND sm.user_id = $2
          ORDER BY sm.created_at DESC`

	return s.scanMovements(ctx, q, productID, ownerID)
}

func (s *stockStore) GetMovementsByShop(ctx context.Context, shopID, ownerID int) ([]*models.StockMovement, error) {
	q := `SELECT sm.movement_id, sm.product_id, sm.shop_id, sm.user_id,
                 sm.quantity_change, sm.movement_type, sm.comment, sm.created_at
          FROM stock_movements sm
          INNER JOIN shops sh ON sm.shop_id = sh.shop_id
          WHERE sm.shop_id = $1 AND sm.user_id = $2
          ORDER BY sm.created_at DESC`

	return s.scanMovements(ctx, q, shopID, ownerID)
}

func (s *stockStore) GetLogs(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string, limit, offset int) ([]*models.StockMovement, int, error) {
	cond := ""
	if shopID > 0 {
		cond = fmt.Sprintf(" AND sm.shop_id = %d", shopID)
	}
	if dateFrom != "" {
		cond += fmt.Sprintf(" AND DATE(sm.created_at) >= '%s'", dateFrom)
	}
	if dateTo != "" {
		cond += fmt.Sprintf(" AND DATE(sm.created_at) <= '%s'", dateTo)
	}

	var total int
	_ = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM stock_movements sm
		INNER JOIN products p ON sm.product_id = p.product_id
		WHERE p.user_id = $1%s`, cond), ownerID).Scan(&total)

	q := fmt.Sprintf(`
		SELECT sm.movement_id,
		       sm.product_id, p.product_name,
		       sm.shop_id,    sh.shop_name,
		       sm.user_id,    u.username,
		       sm.quantity_change, sm.movement_type, COALESCE(sm.comment,''), sm.created_at
		FROM stock_movements sm
		INNER JOIN products p  ON sm.product_id = p.product_id
		INNER JOIN shops    sh ON sm.shop_id    = sh.shop_id
		LEFT  JOIN users    u  ON sm.user_id    = u.id
		WHERE p.user_id = $1%s
		ORDER BY sm.created_at DESC
		LIMIT $2 OFFSET $3`, cond)

	rows, err := s.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movements []*models.StockMovement
	for rows.Next() {
		m := &models.StockMovement{}
		if err := rows.Scan(
			&m.MovementID,
			&m.ProductID, &m.ProductName,
			&m.ShopID, &m.ShopName,
			&m.UserID, &m.Username,
			&m.QuantityChange, &m.MovementType, &m.Comment, &m.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return movements, total, nil
}

func (s *stockStore) scanMovements(ctx context.Context, q string, args ...any) ([]*models.StockMovement, error) {
	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*models.StockMovement
	for rows.Next() {
		m := &models.StockMovement{}
		if err := rows.Scan(
			&m.MovementID, &m.ProductID, &m.ShopID, &m.UserID,
			&m.QuantityChange, &m.MovementType, &m.Comment, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}
	return movements, rows.Err()
}
