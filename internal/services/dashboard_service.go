package services

import (
	"context"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type DashboardService struct {
	store store.DashboardStore
}

func NewDashboardService(s store.DashboardStore) *DashboardService {
	return &DashboardService{store: s}
}

func (s *DashboardService) GetStats(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string) (*models.DashboardStats, error) {
	return s.store.GetStats(ctx, ownerID, shopID, dateFrom, dateTo)
}

func (s *DashboardService) GetNotifications(ctx context.Context, ownerID, shopID int) (*models.NotificationSummary, error) {
	return s.store.GetNotifications(ctx, ownerID, shopID)
}

func (s *DashboardService) GetOnboardingStatus(ctx context.Context, ownerID int) (*models.OnboardingStatus, error) {
	return s.store.GetOnboardingStatus(ctx, ownerID)
}
