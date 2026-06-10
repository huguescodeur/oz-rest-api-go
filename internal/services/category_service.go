package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type CategoryService struct {
	store store.CategoryStore
}

func NewCategoryService(s store.CategoryStore) *CategoryService {
	return &CategoryService{store: s}
}

func (s *CategoryService) GetAll(ctx context.Context, ownerID int) ([]*models.Category, error) {
	return s.store.GetAll(ctx, ownerID)
}

func (s *CategoryService) Create(ctx context.Context, name string, ownerID int) (*models.Category, error) {
	if name == "" {
		return nil, errors.New("le nom de la catégorie est requis")
	}
	c := &models.Category{
		CategoryUUID: uuid.New(),
		CategoryName: name,
		OwnerID:      ownerID,
	}
	return s.store.Create(ctx, c)
}

func (s *CategoryService) Update(ctx context.Context, id, ownerID int, name string) (*models.Category, error) {
	if name == "" {
		return nil, errors.New("le nom de la catégorie est requis")
	}
	return s.store.Update(ctx, id, ownerID, name)
}

func (s *CategoryService) Delete(ctx context.Context, id, ownerID int) error {
	return s.store.Delete(ctx, id, ownerID)
}

func (s *CategoryService) Restore(ctx context.Context, id, ownerID int) error {
	return s.store.Restore(ctx, id, ownerID)
}
