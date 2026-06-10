package transport

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/ctxkeys"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/responses"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
)

type orderItemRequest struct {
	ProductID int `json:"product_id" validate:"required,gt=0"`
	Quantity  int `json:"quantity" validate:"required,gt=0"`
}

type createOrderRequest struct {
	ShopID            int                `json:"shop_id" validate:"required,gt=0"`
	Items             []orderItemRequest `json:"items" validate:"required,min=1,dive"`
	CustomerName      string             `json:"customer_name"`
	CustomerPhone     string             `json:"customer_phone"`
	PaymentMethod     string             `json:"payment_method" validate:"omitempty,oneof=ESPECES ORANGE_MONEY MTN_MOMO WAVE CREDIT AUTRE"`
	Notes             string             `json:"notes"`
	DeliveryAddress   string             `json:"delivery_address"`
	ImmediateDelivery bool               `json:"immediate_delivery"`
}

type OrderHandler struct {
	orderService *services.OrderService
	validate     *validator.Validate
}

func NewOrderHandler(s *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: s, validate: validator.New()}
}

// AllOrdersHandler godoc
// @Summary      Liste paginée des commandes
// @Tags         Orders
// @Param        page   query  int  false  "Numéro de la page (défaut: 1)"
// @Param        limit  query  int  false  "Nombre d'éléments (défaut: 20)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Router       /orders [get]
func (h *OrderHandler) AllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	shopID, _ := strconv.Atoi(r.URL.Query().Get("shop_id"))
	dateFrom := validateDateParam(r.URL.Query().Get("date_from"))
	dateTo := validateDateParam(r.URL.Query().Get("date_to"))

	orders, total, err := h.orderService.GetAllOrders(ctx, ownerID, shopID, dateFrom, dateTo, limit, offset)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"orders":      orders,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": math.Ceil(float64(total) / float64(limit)),
	})
}

// GetOrderByUUIDHandler godoc
// @Summary      Détails d'une commande
// @Tags         Orders
// @Param        uuid  path  string  true  "UUID de la commande"
// @Security     BearerAuth
// @Success      200  {object}  models.Order
// @Router       /orders/{uuid} [get]
func (h *OrderHandler) GetOrderByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	orderUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderByUUID(ctx, orderUUID, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// CreateOrderHandler godoc
// @Summary      Créer une commande
// @Description  Crée une commande et décrémente le stock atomiquement
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        order  body  createOrderRequest  true  "Données de la commande"
// @Security     BearerAuth
// @Success      201  {object}  models.Order
// @Router       /orders [post]
func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
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
		items[i] = models.OrderItem{ProductID: it.ProductID, Quantity: it.Quantity}
	}

	order, err := h.orderService.CreateOrder(ctx, req.ShopID, ownerID, items, req.CustomerName, req.CustomerPhone, req.PaymentMethod, req.Notes, req.DeliveryAddress, req.ImmediateDelivery)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// CancelOrderHandler godoc
// @Summary      Annuler une commande
// @Description  Annule une commande et restaure le stock
// @Tags         Orders
// @Param        uuid  path  string  true  "UUID de la commande"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /orders/{uuid} [delete]
func (h *OrderHandler) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	orderUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := h.orderService.CancelOrder(ctx, orderUUID, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Commande annulée avec succès"})
}

// ConfirmOrderHandler godoc
// @Summary      Confirmer une commande (PENDING → CONFIRMED)
// @Tags         Orders
// @Param        uuid  path  string  true  "UUID de la commande"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /orders/{uuid}/confirm [patch]
func (h *OrderHandler) ConfirmOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}
	orderUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}
	if err := h.orderService.ConfirmOrder(ctx, orderUUID, ownerID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Commande confirmée"})
}

// DeliverOrderHandler godoc
// @Summary      Marquer une commande comme livrée (CONFIRMED → DELIVERED)
// @Tags         Orders
// @Param        uuid  path  string  true  "UUID de la commande"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /orders/{uuid}/deliver [patch]
func (h *OrderHandler) DeliverOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}
	orderUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}
	if err := h.orderService.DeliverOrder(ctx, orderUUID, ownerID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Commande livrée"})
}

func (h *OrderHandler) OrderRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.AllOrdersHandler)
	r.Post("/", h.CreateOrderHandler)

	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.GetOrderByUUIDHandler)
		r.Delete("/", h.CancelOrderHandler)
		r.Patch("/confirm", h.ConfirmOrderHandler)
		r.Patch("/deliver", h.DeliverOrderHandler)
	})

	return r
}
