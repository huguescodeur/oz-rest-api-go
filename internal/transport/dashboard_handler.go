package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/ctxkeys"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler(s *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: s}
}

// StatsHandler godoc
// @Summary      Tableau de bord — statistiques globales
// @Description  Revenus, commandes, top produits, valeur stock, alertes stock faible
// @Tags         Dashboard
// @Security     BearerAuth
// @Success      200  {object}  models.DashboardStats
// @Router       /dashboard [get]
func (h *DashboardHandler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shopID, _ := strconv.Atoi(r.URL.Query().Get("shop_id"))
	dateFrom := validateDateParam(r.URL.Query().Get("date_from"))
	dateTo := validateDateParam(r.URL.Query().Get("date_to"))

	stats, err := h.service.GetStats(ctx, ownerID, shopID, dateFrom, dateTo)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *DashboardHandler) NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}
	shopID, _ := strconv.Atoi(r.URL.Query().Get("shop_id"))
	summary, err := h.service.GetNotifications(ctx, ownerID, shopID)
	if err != nil {
		http.Error(w, "Erreur notifications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *DashboardHandler) OnboardingStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}
	status, err := h.service.GetOnboardingStatus(ctx, ownerID)
	if err != nil {
		http.Error(w, "Erreur onboarding", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *DashboardHandler) DashboardRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.StatsHandler)
	r.Get("/notifications", h.NotificationsHandler)
	r.Get("/onboarding", h.OnboardingStatusHandler)
	return r
}
