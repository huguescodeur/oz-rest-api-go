package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type ProductService struct {
	productStore store.ProductStore
}

func NewProductService(s store.ProductStore) *ProductService {
	return &ProductService{productStore: s}
}

func (p *ProductService) GetAllProducts(ctx context.Context, ownerID, limit, offset int) ([]*models.Product, int, error) {
	return p.productStore.GetAll(ctx, ownerID, limit, offset)
}

func (p *ProductService) GetProductByUUID(ctx context.Context, uuid uuid.UUID, ownerID int) (*models.Product, error) {
	return p.productStore.GetByUUID(ctx, uuid, ownerID)
}

func (p *ProductService) CreateProduct(ctx context.Context, model *models.Product) (*models.Product, error) {
	if model.ProductUUID == uuid.Nil {
		model.ProductUUID = uuid.New()
	}
	if model.UnitPrice < 0 {
		return nil, errors.New("le prix ne peut pas être négatif")
	}
	return p.productStore.Create(ctx, model)
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
