package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type ProductService struct {
	productStore  store.ProductStore
	categoryStore store.CategoryStore
	shopStore     store.ShopStore
	stockStore    store.StockStore
}

func NewProductService(s store.ProductStore, c store.CategoryStore, sh store.ShopStore, st store.StockStore) *ProductService {
	return &ProductService{productStore: s, categoryStore: c, shopStore: sh, stockStore: st}
}

func (p *ProductService) GetAllProducts(ctx context.Context, ownerID, limit, offset int) ([]*models.Product, int, error) {
	return p.productStore.GetAll(ctx, ownerID, limit, offset)
}

func (p *ProductService) GetProductByUUID(ctx context.Context, uuid uuid.UUID, ownerID int) (*models.Product, error) {
	return p.productStore.GetByUUID(ctx, uuid, ownerID)
}

func (p *ProductService) CreateProduct(ctx context.Context, model *models.Product) (*models.Product, error) {
	cat, err := p.categoryStore.GetByID(ctx, model.ProductCategory.CategoryID, model.OwnerID)
	if err != nil {
		return nil, errors.New("catégorie introuvable")
	}
	if cat.OwnerID != model.OwnerID {
		return nil, errors.New("catégorie non autorisée pour cet utilisateur")
	}
	if model.ProductUUID == uuid.Nil {
		model.ProductUUID = uuid.New()
	}
	if model.UnitPrice < 0 {
		return nil, errors.New("le prix ne peut pas être négatif")
	}

	created, err := p.productStore.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	// Initialise le stock à 0 pour toutes les boutiques actives de l'owner
	shops, _, err := p.shopStore.GetAll(ctx, model.OwnerID, 1000, 0)
	if err == nil {
		for _, sh := range shops {
			_ = p.stockStore.InitStock(ctx, created.ProductID, sh.ShopID, model.OwnerID)
		}
	}

	return created, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, uuid uuid.UUID, ownerID int, product *models.Product) (*models.Product, error) {
	currentProduct, err := p.productStore.GetByUUID(ctx, uuid, ownerID)
	if err != nil {
		return nil, err
	}

	if product.UnitPrice != 0 {
		if product.UnitPrice < 0 {
			return nil, errors.New("le prix ne peut pas être négatif")
		}
		currentProduct.UnitPrice = product.UnitPrice
	}
	if product.ProductCategory != nil && product.ProductCategory.CategoryID > 0 {
		currentProduct.ProductCategory = product.ProductCategory
	}
	if product.ProductName != "" {
		currentProduct.ProductName = product.ProductName
	}

	return p.productStore.Update(ctx, uuid, currentProduct)
}

func (p *ProductService) DeleteProduct(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	return p.productStore.Delete(ctx, uuid, ownerID)
}

func (p *ProductService) RestoreProduct(ctx context.Context, uuid uuid.UUID, ownerID int) error {
	return p.productStore.Restore(ctx, uuid, ownerID)
}
