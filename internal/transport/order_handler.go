package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/ctxkeys"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/errs"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/responses"
	"github.com/huguescodeur/zoro_rest_api_go/internal/services"
)

type orderItemRequest struct {
	ProductID int `json:"product_id" validate:"required,gt=0"`
	Quantity  int `json:"quantity" validate:"required,gt=0"`
}

type createOrderRequest struct {
	ShopID int                `json:"shop_id" validate:"required,gt=0"`
	Items  []orderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type OrderHandler struct {
	orderService *services.OrderService
	validate     *validator.Validate
}

func NewOrderHandler(s *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: s,
		validate:     validator.New(),
	}
}

// AllOrdersHandler godoc
// @Summary      Liste toutes les commandes
// @Description  Récupère toutes les commandes de l'utilisateur connecté avec leurs items
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {string}  string "Utilisateur non identifié"
// @Router       /orders [get]
func (h *OrderHandler) AllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetAllOrders(ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur de récupération: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(orders),
		"orders": orders,
	})
}

// GetOrderByUUIDHandler godoc
// @Summary      Détails d'une commande
// @Description  Récupère une commande par son UUID avec tous ses items
// @Tags         Orders
// @Produce      json
// @Param        uuid  path  string  true  "UUID de la commande"
// @Security     BearerAuth
// @Success      200  {object}  models.Order
// @Failure      400  {string}  string "Format d'identifiant invalide"
// @Failure      404  {string}  string "Commande introuvable"
// @Router       /orders/{uuid} [get]
func (h *OrderHandler) GetOrderByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	uuidStr := chi.URLParam(r, "uuid")
	orderUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderByUUID(orderUUID, ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Commande introuvable: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// CreateOrderHandler godoc
// @Summary      Créer une commande
// @Description  Crée une commande, décrémente le stock et enregistre les mouvements SALE
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        order  body  createOrderRequest  true  "Données de la commande"
// @Security     BearerAuth
// @Success      201  {object}  models.Order
// @Failure      400  {string}  string "Données invalides ou stock insuffisant"
// @Router       /orders [post]
func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	items := make([]models.OrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = models.OrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		}
	}

	order, err := h.orderService.CreateOrder(req.ShopID, ownerID, items)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur lors de la création: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// CancelOrderHandler godoc
// @Summary      Annuler une commande
// @Description  Annule une commande et restaure le stock de chaque item
// @Tags         Orders
// @Param        uuid  path  string  true  "UUID de la commande"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Failure      404  {string}  string "Commande introuvable"
// @Router       /orders/{uuid} [delete]
func (h *OrderHandler) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	uuidStr := chi.URLParam(r, "uuid")
	orderUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := h.orderService.CancelOrder(orderUUID, ownerID); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Annulation impossible: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Commande annulée avec succès"})
}

func (h *OrderHandler) OrderRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.AllOrdersHandler)
	r.Post("/", h.CreateOrderHandler)

	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.GetOrderByUUIDHandler)
		r.Delete("/", h.CancelOrderHandler)
	})

	return r
}
