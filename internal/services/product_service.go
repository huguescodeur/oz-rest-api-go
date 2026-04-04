package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/store"
)

type ProductService struct {
	productStore store.ProductStore
}

func NewProductService(s store.ProductStore) *ProductService {
	return &ProductService{
		productStore: s,
	}
}

func (p *ProductService) GetAllProducts(ownerID int) ([]*models.Product, error) {
	products, err := p.productStore.GetAll(ownerID)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (p *ProductService) GetProductByUUID(uuid uuid.UUID, ownerID int) (*models.Product, error) {
	product, err := p.productStore.GetByUUID(uuid, ownerID)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (p *ProductService) CreateProduct(model *models.Product) (*models.Product, error) {
	if model.ProductUUID == uuid.Nil {
		model.ProductUUID = uuid.New()
	}
	if model.UnitPrice < 0 {
		return nil, errors.New("le prix ne peut pas être négatif")
	}

	createdProduct, err := p.productStore.Create(model)
	if err != nil {
		return nil, err
	}

	return createdProduct, nil
}

func (p *ProductService) UpdateProduct(uuid uuid.UUID, ownerID int, product *models.Product) (*models.Product, error) {
	currentProduct, err := p.productStore.GetByUUID(uuid, ownerID)
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

	productUpdated, err := p.productStore.Update(uuid, currentProduct)
	if err != nil {
		return nil, err
	}

	return productUpdated, nil
}
func (p *ProductService) DeleteProduct(uuid uuid.UUID, ownerID int) error {
	if err := p.productStore.Delete(uuid, ownerID); err != nil {
		return err
	}

	return nil
}
func (p *ProductService) RestoreProduct(uuid uuid.UUID, ownerID int) error {
	if err := p.productStore.Restore(uuid, ownerID); err != nil {
		return err
	}

	return nil
}
