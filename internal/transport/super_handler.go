package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
)

type SuperHandler struct {
	service *services.SuperService
}

func NewSuperHandler(s *services.SuperService) *SuperHandler {
	return &SuperHandler{service: s}
}

func (h *SuperHandler) GetGlobalStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetGlobalStats(r.Context())
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *SuperHandler) GetAdminsHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}

	admins, total, err := h.service.GetAdmins(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"admins": admins,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *SuperHandler) GetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *SuperHandler) SetSettingHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if err := h.service.SetSetting(r.Context(), body.Key, body.Value); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SuperHandler) GetAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	analytics, err := h.service.GetAnalytics(r.Context(), dateFrom, dateTo)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func (h *SuperHandler) GetGlobalLogsHandler(w http.ResponseWriter, r *http.Request) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	movType := r.URL.Query().Get("type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}

	logs, total, err := h.service.GetGlobalLogs(r.Context(), dateFrom, dateTo, movType, limit, offset)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *SuperHandler) SuperRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/stats", h.GetGlobalStatsHandler)
	r.Get("/analytics", h.GetAnalyticsHandler)
	r.Get("/logs", h.GetGlobalLogsHandler)
	r.Get("/admins", h.GetAdminsHandler)
	r.Get("/settings", h.GetSettingsHandler)
	r.Post("/settings", h.SetSettingHandler)
	return r
}
