package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShopStore interface {
	DoShopsBelongToOwner(ctx context.Context, shopIDs []int, ownerID int) (bool, error)
	IsChildOf(ctx context.Context, vendeurID int, adminID int) (bool, error)
	GetAll(ctx context.Context, ownerID, limit, offset int) ([]*models.Shop, int, error)
	GetAllShopVendeur(ctx context.Context, vendeurID int) ([]*models.Shop, error)
	GetByUUID(ctx context.Context, uuid uuid.UUID, ownerID int) (*models.Shop, error)
	Create(ctx context.Context, shop *models.Shop) (*models.Shop, error)
	AssignVendeurToShops(ctx context.Context, vendeurID int, shopIDs []int) error
	Update(ctx context.Context, shop *models.Shop) (*models.Shop, error)
	Delete(ctx context.Context, uuid uuid.UUID, ownerID int) error
	Restore(ctx context.Context, uuid uuid.UUID, ownerID int) error
}

type shopStore struct {
	db *pgxpool.Pool
}

func NewShopStore(db *pgxpool.Pool) ShopStore {
	return &shopStore{db: db}
}

func (s *shopStore) DoShopsBelongToOwner(ctx context.Context, shopIDs []int, ownerID int) (bool, error) {
	if len(shopIDs) == 0 {
		return true, nil
	}

	placeholders := make([]string, len(shopIDs))
	args := make([]any, len(shopIDs)+1)

	for i, id := range shopIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(shopIDs)] = ownerID

	q := fmt.Sprintf(
		`SELECT COUNT(*) FROM shops WHERE shop_id IN (%s) AND user_id = $%d`,
		strings.Join(placeholders, ","), len(shopIDs)+1,
	)

	var count int
	if err := s.db.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return false, err
	}
	return count == len(shopIDs), nil
}

func (s *shopStore) IsChildOf(ctx context.Context, vendeurID int, adminID int) (bool, error) {
	var count int
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE id = $1 AND parent_id = $2`, vendeurID, adminID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *shopStore) GetAll(ctx context.Context, ownerID, limit, offset int) ([]*models.Shop, int, error) {
	var total int
	_ = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM shops WHERE user_id = $1 AND deleted_at IS NULL`, ownerID,
	).Scan(&total)

	q := `SELECT shop_id, uuid, shop_name, shop_address, shop_phone, shop_mail, created_at, updated_at, deleted_at, user_id
          FROM shops
          WHERE user_id = $1 AND deleted_at IS NULL
          ORDER BY created_at DESC
          LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var listShops []*models.Shop
	for rows.Next() {
		shop := &models.Shop{}
		if err = rows.Scan(
			&shop.ShopID, &shop.ShopUUID, &shop.ShopName, &shop.ShopAddress,
			&shop.ShopPhone, &shop.ShopMail, &shop.CreatedAt, &shop.UpdatedAt,
			&shop.DeletedAt, &shop.OwnerID,
		); err != nil {
			return nil, 0, err
		}
		listShops = append(listShops, shop)
	}
	return listShops, total, rows.Err()
}

func (s *shopStore) GetAllShopVendeur(ctx context.Context, vendeurID int) ([]*models.Shop, error) {
	q := `SELECT s.shop_id, s.uuid, s.shop_name, s.shop_address, s.created_at, s.updated_at, us.assigned_at
          FROM shops s
          INNER JOIN user_shops us ON s.shop_id = us.shop_id
          WHERE us.user_id = $1 AND s.deleted_at IS NULL`

	rows, err := s.db.Query(ctx, q, vendeurID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listShopUser []*models.Shop
	for rows.Next() {
		shop := &models.Shop{}
		if err := rows.Scan(
			&shop.ShopID, &shop.ShopUUID, &shop.ShopName, &shop.ShopAddress,
			&shop.CreatedAt, &shop.UpdatedAt, &shop.AssignedAt,
		); err != nil {
			return nil, err
		}
		listShopUser = append(listShopUser, shop)
	}
	return listShopUser, rows.Err()
}

func (s *shopStore) GetByUUID(ctx context.Context, uuid uuid.UUID, ownerID int) (*models.Shop, error) {
	q := `SELECT shop_id, uuid, shop_name, shop_address, shop_phone, shop_mail, created_at, updated_at, deleted_at, user_id
          FROM shops
          WHERE uuid = $1 AND user_id = $2 AND deleted_at IS NULL`

	shop := &models.Shop{}
	if err := s.db.QueryRow(ctx, q, uuid, ownerID).Scan(
		&shop.ShopID, &shop.ShopUUID, &shop.ShopName, &shop.ShopAddress,
		&shop.ShopPhone, &shop.ShopMail, &shop.CreatedAt, &shop.UpdatedAt,
		&shop.DeletedAt, &shop.OwnerID,
	); err != nil {
		return nil, err
	}
	return shop, nil
}

func (s *shopStore) Create(ctx context.Context, shop *models.Shop) (*models.Shop, error) {
	timeNow := time.Now().UTC()
	q := `INSERT INTO shops (uuid, shop_name, shop_address, shop_phone, shop_mail, created_at, updated_at, user_id)
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(ctx, q,
		shop.ShopUUID, shop.ShopName, shop.ShopAddress,
		shop.ShopPhone, shop.ShopMail, timeNow, timeNow, shop.OwnerID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetByUUID(ctx, shop.ShopUUID, shop.OwnerID)
}

func (s *shopStore) AssignVendeurToShops(ctx context.Context, vendeurID int, shopIDs []int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := `INSERT INTO user_shops (user_id, shop_id, assigned_at)
          VALUES ($1, $2, NOW())
          ON CONFLICT (user_id, shop_id) DO UPDATE SET assigned_at = NOW()`

	for _, shopID := range shopIDs {
		if _, err := tx.Exec(ctx, q, vendeurID, shopID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *shopStore) Update(ctx context.Context, shop *models.Shop) (*models.Shop, error) {
	updatedAt := time.Now().UTC()
	q := `UPDATE shops
          SET shop_name = $1, shop_address = $2, shop_phone = $3, shop_mail = $4, updated_at = $5
          WHERE uuid = $6 AND user_id = $7 AND deleted_at IS NULL`

	tag, err := s.db.Exec(ctx, q,
		shop.ShopName, shop.ShopAddress, shop.ShopPhone, shop.ShopMail,
		updatedAt, shop.ShopUUID, shop.OwnerID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("Shop n'existe pas! Impossible d'update")
	}
	return s.GetByUUID(ctx, shop.ShopUUID, shop.OwnerID)
}

func (s *shopStore) Delete(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE shops SET deleted_at = NOW() WHERE uuid = $1 AND user_id = $2 AND deleted_at IS NULL`,
		uuid, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *shopStore) Restore(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE shops SET deleted_at = NULL, updated_at = NOW() WHERE uuid = $1 AND user_id = $2 AND deleted_at IS NOT NULL`,
		uuid, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}
