package services

import (
	"context"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type SuperService struct {
	store store.SuperStore
}

func NewSuperService(s store.SuperStore) *SuperService {
	return &SuperService{store: s}
}

func (s *SuperService) GetGlobalStats(ctx context.Context) (*models.GlobalStats, error) {
	return s.store.GetGlobalStats(ctx)
}

func (s *SuperService) GetAdmins(ctx context.Context, limit, offset int) ([]*models.AdminSummary, int, error) {
	return s.store.GetAdmins(ctx, limit, offset)
}

func (s *SuperService) GetSettings(ctx context.Context) (map[string]string, error) {
	return s.store.GetSettings(ctx)
}

func (s *SuperService) SetSetting(ctx context.Context, key, value string) error {
	return s.store.SetSetting(ctx, key, value)
}

func (s *SuperService) GetAnalytics(ctx context.Context, dateFrom, dateTo string) (*models.SuperAnalytics, error) {
	return s.store.GetAnalytics(ctx, dateFrom, dateTo)
}

func (s *SuperService) GetGlobalLogs(ctx context.Context, dateFrom, dateTo, movType string, limit, offset int) ([]*models.GlobalActivity, int, error) {
	return s.store.GetGlobalLogs(ctx, dateFrom, dateTo, movType, limit, offset)
}
