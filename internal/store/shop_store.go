package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/errs"
)

type ShopStore interface {
	DoShopsBelongToOwner(shopIDs []int, ownerID int) (bool, error)
	IsChildOf(vendeurID int, adminID int) (bool, error)
	GetAll(ownerID int) ([]*models.Shop, error)
	GetAllShopVendeur(vendeurID int) ([]*models.Shop, error)
	GetByUUID(uuid uuid.UUID, ownerID int) (*models.Shop, error)
	Create(shop *models.Shop) (*models.Shop, error)
	AssignVendeurToShops(vendeurID int, shopIDs []int) error
	Update(shop *models.Shop) (*models.Shop, error)
	Delete(uuid uuid.UUID, ownerID int) error
	Restore(uuid uuid.UUID, ownerID int) error
}

type shopStore struct {
	db *sql.DB
}

func NewShopStore(db *sql.DB) ShopStore {
	return &shopStore{db: db}
}

func (s *shopStore) DoShopsBelongToOwner(shopIDs []int, ownerID int) (bool, error) {
	if len(shopIDs) == 0 {
		return true, nil
	}

	placeholders := make([]string, len(shopIDs))
	args := make([]interface{}, len(shopIDs)+1)

	for i, id := range shopIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(shopIDs)] = ownerID

	q := `SELECT COUNT(*) FROM shops WHERE shop_id IN (` +
		strings.Join(placeholders, ",") + `) AND user_id = ?`

	var count int
	err := s.db.QueryRow(q, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == len(shopIDs), nil
}

func (s *shopStore) IsChildOf(vendeurID int, adminID int) (bool, error) {
	var count int
	q := `SELECT COUNT(*) FROM users WHERE id = ? AND parent_id = ?`
	err := s.db.QueryRow(q, vendeurID, adminID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *shopStore) GetAll(ownerID int) ([]*models.Shop, error) {
	q := `SELECT * FROM shops WHERE user_id = ? AND deleted_at IS NULL
		  ORDER BY created_at DESC`

	raws, err := s.db.Query(q, ownerID)
	if err != nil {
		return nil, err
	}
	defer raws.Close()

	var listShops []*models.Shop
	for raws.Next() {
		shop := &models.Shop{}
		if err = raws.Scan(
			&shop.ShopID,
			&shop.ShopUUID,
			&shop.ShopName,
			&shop.ShopAddress,
			&shop.ShopPhone,
			&shop.ShopMail,
			&shop.CreatedAt,
			&shop.UpdatedAt,
			&shop.DeletedAt,
			&shop.OwnerID); err != nil {
			return nil, err
		}

		listShops = append(listShops, shop)
	}

	return listShops, nil
}

func (s *shopStore) GetAllShopVendeur(vendeurID int) ([]*models.Shop, error) {
	q := `SELECT s.shop_id, s.shop_name, s.shop_address, s.uuid, s.created_at, s.updated_at, us.assigned_at
		  FROM shops s
		  INNER JOIN user_shops us ON s.shop_id = us.shop_id
		  WHERE us.user_id = ? AND s.deleted_at IS NULL`

	rows, err := s.db.Query(q, vendeurID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var listShopUser []*models.Shop
	for rows.Next() {
		shop := &models.Shop{}
		if err := rows.Scan(&shop.ShopID, &shop.ShopName, &shop.ShopAddress, shop.ShopUUID, &shop.CreatedAt, &shop.UpdatedAt, &shop.AssignedAt); err != nil {
			return nil, err
		}

		listShopUser = append(listShopUser, shop)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return listShopUser, nil
}

func (s *shopStore) GetByUUID(uuid uuid.UUID, ownerID int) (*models.Shop, error) {
	q := `SELECT * FROM shops WHERE uuid = ? AND user_id = ? AND deleted_at IS NULL`

	raw := s.db.QueryRow(q, uuid, ownerID)

	shop := &models.Shop{}
	if err := raw.Scan(
		&shop.ShopID,
		&shop.ShopUUID,
		&shop.ShopName,
		&shop.ShopAddress,
		&shop.ShopPhone,
		&shop.ShopMail,
		&shop.CreatedAt,
		&shop.UpdatedAt,
		&shop.DeletedAt,
		&shop.OwnerID); err != nil {
		return nil, err
	}

	return shop, nil
}

func (s *shopStore) Create(shop *models.Shop) (*models.Shop, error) {
	timeNow := time.Now().UTC()

	q := `INSERT INTO shops (uuid, shop_name, shop_address, shop_phone, shop_mail, created_at, updated_at, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(q,
		shop.ShopUUID,
		shop.ShopName,
		shop.ShopAddress,
		shop.ShopPhone,
		shop.ShopMail,
		timeNow,
		timeNow, shop.OwnerID)
	if err != nil {
		return nil, err
	}

	return s.GetByUUID(shop.ShopUUID, shop.OwnerID)
}

func (s *shopStore) AssignVendeurToShops(vendeurID int, shopIDs []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	q := `INSERT INTO user_shops (user_id, shop_id, assigned_at) 
          VALUES (?, ?, NOW())
          ON DUPLICATE KEY UPDATE assigned_at = NOW()`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, shopID := range shopIDs {
		_, err := stmt.Exec(vendeurID, shopID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *shopStore) Update(shop *models.Shop) (*models.Shop, error) {
	updatedAt := time.Now().UTC()

	q := `UPDATE shops 
          SET shop_name = ?, shop_address = ?, shop_phone = ?, shop_mail = ?, updated_at = ? 
          WHERE uuid = ? AND user_id = ? AND deleted_at IS NULL`

	raw, err := s.db.Exec(
		q,
		shop.ShopName,
		shop.ShopAddress,
		shop.ShopPhone,
		shop.ShopMail,
		updatedAt,
		shop.ShopUUID, shop.OwnerID)
	if err != nil {
		return nil, err
	}
	count, err := raw.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("Shop n'existe pas! Impossible d'update")
	}

	return s.GetByUUID(shop.ShopUUID, shop.OwnerID)
}

func (s *shopStore) Delete(uuid uuid.UUID, ownerID int) error {
	q := `UPDATE shops SET deleted_at = NOW() WHERE uuid = ? AND user_id = ? AND deleted_at IS NULL`

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

func (s *shopStore) Restore(uuid uuid.UUID, ownerID int) error {
	q := `UPDATE shops 
          SET deleted_at = NULL, updated_at = NOW() 
          WHERE uuid = ? AND user_id = ? AND deleted_at IS NOT NULL`

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
