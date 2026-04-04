package services

import (
	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/errs"
	"github.com/huguescodeur/zoro_rest_api_go/internal/store"
)

type ShopService struct {
	shopStore store.ShopStore
}

func NewShopService(s store.ShopStore) *ShopService {
	return &ShopService{
		shopStore: s,
	}
}

func (s *ShopService) GetAllShops(ownerID int) ([]*models.Shop, error) {
	shops, err := s.shopStore.GetAll(ownerID)
	if err != nil {
		return nil, err
	}

	return shops, nil
}
func (s *ShopService) GetAllShopsVendeur(vendeurID int) ([]*models.Shop, error) {
	shops, err := s.shopStore.GetAll(vendeurID)
	if err != nil {
		return nil, err
	}

	return shops, nil
}

func (s *ShopService) GetShopByUUID(uuid uuid.UUID, ownerID int) (*models.Shop, error) {
	shop, err := s.shopStore.GetByUUID(uuid, ownerID)
	if err != nil {
		return nil, err
	}

	return shop, nil
}
func (s *ShopService) CreateShop(model *models.Shop) (*models.Shop, error) {
	if model.ShopUUID == uuid.Nil {
		model.ShopUUID = uuid.New()
	}

	shop, err := s.shopStore.Create(model)
	if err != nil {
		return nil, err
	}

	return shop, nil
}

func (s *ShopService) AssignVendeurToShops(vendeurID int, shopIDs []int, ownerID int, ownerRole string) error {
	if ownerRole != "super" {
		isChild, err := s.shopStore.IsChildOf(vendeurID, ownerID)
		if err != nil {
			return err
		}
		if !isChild {

			return errs.ErrUnauthorized
		}

		belongs, err := s.shopStore.DoShopsBelongToOwner(shopIDs, ownerID)
		if err != nil {
			return err
		}
		if !belongs {

			return errs.ErrUnauthorized
		}
	}
	return s.shopStore.AssignVendeurToShops(vendeurID, shopIDs)
}

func (s *ShopService) UpdateShop(uuid uuid.UUID, ownerID int, shop *models.Shop) (*models.Shop, error) {
	currentShop, err := s.shopStore.GetByUUID(uuid, ownerID)
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

	updatedShop, err := s.shopStore.Update(currentShop)
	if err != nil {
		return nil, err
	}

	return updatedShop, nil
}
func (s *ShopService) DeleteShop(uuid uuid.UUID, ownerID int) error {
	if err := s.shopStore.Delete(uuid, ownerID); err != nil {
		return err
	}

	return nil
}

func (s *ShopService) RestoreShop(uuid uuid.UUID, ownerID int) error {
	if err := s.shopStore.Restore(uuid, ownerID); err != nil {
		return err
	}

	return nil
}
