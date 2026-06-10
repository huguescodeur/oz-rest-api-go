package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type ShopService struct {
	shopStore store.ShopStore
}

func NewShopService(s store.ShopStore) *ShopService {
	return &ShopService{shopStore: s}
}

func (s *ShopService) GetAllShops(ctx context.Context, ownerID, limit, offset int) ([]*models.Shop, int, error) {
	return s.shopStore.GetAll(ctx, ownerID, limit, offset)
}

func (s *ShopService) GetAllShopsVendeur(ctx context.Context, vendeurID int) ([]*models.Shop, error) {
	return s.shopStore.GetAllShopVendeur(ctx, vendeurID)
}

func (s *ShopService) GetShopByUUID(ctx context.Context, uuid uuid.UUID, ownerID int) (*models.Shop, error) {
	return s.shopStore.GetByUUID(ctx, uuid, ownerID)
}

func (s *ShopService) CreateShop(ctx context.Context, model *models.Shop) (*models.Shop, error) {
	if model.ShopUUID == uuid.Nil {
		model.ShopUUID = uuid.New()
	}
	return s.shopStore.Create(ctx, model)
}

func (s *ShopService) AssignVendeurToShops(ctx context.Context, vendeurID int, shopIDs []int, ownerID int, ownerRole string) error {
	if ownerRole != "super" {
		isChild, err := s.shopStore.IsChildOf(ctx, vendeurID, ownerID)
		if err != nil {
			return err
		}
		if !isChild {
			return errs.ErrUnauthorized
		}

		belongs, err := s.shopStore.DoShopsBelongToOwner(ctx, shopIDs, ownerID)
		if err != nil {
			return err
		}
		if !belongs {
			return errs.ErrUnauthorized
		}
	}
	return s.shopStore.AssignVendeurToShops(ctx, vendeurID, shopIDs)
}

func (s *ShopService) UpdateShop(ctx context.Context, uuid uuid.UUID, ownerID int, shop *models.Shop) (*models.Shop, error) {
	currentShop, err := s.shopStore.GetByUUID(ctx, uuid, ownerID)
	if err != nil {
		return nil, err
	}

	if shop.ShopName != "" {
		currentShop.ShopName = shop.ShopName
	}
	if shop.ShopAddress != "" {
		currentShop.ShopAddress = shop.ShopAddress
	}
	if shop.ShopPhone != "" {
		currentShop.ShopPhone = shop.ShopPhone
	}
	if shop.ShopMail != "" {
		currentShop.ShopMail = shop.ShopMail
	}

	return s.shopStore.Update(ctx, currentShop)
}

func (s *ShopService) DeleteShop(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	return s.shopStore.Delete(ctx, uuid, ownerID)
}

func (s *ShopService) RestoreShop(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	return s.shopStore.Restore(ctx, uuid, ownerID)
}
