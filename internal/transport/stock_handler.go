package transport

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/ctxkeys"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/responses"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
)

type adjustStockRequest struct {
	ProductID      int    `json:"product_id" validate:"required,gt=0"`
	ShopID         int    `json:"shop_id" validate:"required,gt=0"`
	QuantityChange int    `json:"quantity_change" validate:"required"`
	MovementType   string `json:"movement_type" validate:"required,oneof=SALE STOCK_IN GIFT LOSS ADJUSTMENT"`
	Comment        string `json:"comment"`
}

type stockInRequest struct {
	ProductID int    `json:"product_id" validate:"required,gt=0"`
	ShopID    int    `json:"shop_id" validate:"required,gt=0"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
	Comment   string `json:"comment"`
}

type initStockRequest struct {
	ProductID int `json:"product_id" validate:"required,gt=0"`
	ShopID    int `json:"shop_id" validate:"required,gt=0"`
}

type StockHandler struct {
	stockService *services.StockService
	validate     *validator.Validate
}

func NewStockHandler(s *services.StockService) *StockHandler {
	return &StockHandler{stockService: s, validate: validator.New()}
}

// AllStocksHandler godoc
// @Summary      Liste tous les stocks
// @Tags         Stocks
// @Security     BearerAuth
// @Success      200  {array}  models.Stock
// @Router       /stocks [get]
func (h *StockHandler) AllStocksHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	stocks, err := h.stockService.GetAllStocks(ctx, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	// Filtre optionnel par boutique
	if shopIDStr := r.URL.Query().Get("shop_id"); shopIDStr != "" {
		shopID, err := strconv.Atoi(shopIDStr)
		if err == nil && shopID > 0 {
			filtered := stocks[:0]
			for _, s := range stocks {
				if s.ShopID == shopID {
					filtered = append(filtered, s)
				}
			}
			stocks = filtered
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"count": len(stocks), "stocks": stocks})
}

// GetStockByProductAndShopHandler godoc
// @Summary      Stock d'un produit dans une boutique
// @Tags         Stocks
// @Param        product_id  query  int  true  "ID du produit"
// @Param        shop_id     query  int  true  "ID de la boutique"
// @Security     BearerAuth
// @Success      200  {object}  models.Stock
// @Router       /stocks/detail [get]
func (h *StockHandler) GetStockByProductAndShopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	productID, err := strconv.Atoi(r.URL.Query().Get("product_id"))
	if err != nil || productID <= 0 {
		http.Error(w, "product_id invalide", http.StatusBadRequest)
		return
	}

	shopID, err := strconv.Atoi(r.URL.Query().Get("shop_id"))
	if err != nil || shopID <= 0 {
		http.Error(w, "shop_id invalide", http.StatusBadRequest)
		return
	}

	stock, err := h.stockService.GetStockByProductAndShop(ctx, productID, shopID, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stock)
}

// InitStockHandler godoc
// @Summary      Initialiser un stock
// @Tags         Stocks
// @Accept       json
// @Produce      json
// @Param        stock  body  initStockRequest  true  "Données d'initialisation"
// @Security     BearerAuth
// @Success      201  {object}  map[string]string
// @Router       /stocks/init [post]
func (h *StockHandler) InitStockHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req initStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	if err := h.stockService.InitStock(ctx, req.ProductID, req.ShopID, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Stock initialisé avec succès"})
}

// StockInHandler godoc
// @Summary      Entrée de stock
// @Tags         Stocks
// @Accept       json
// @Produce      json
// @Param        stock  body  stockInRequest  true  "Données d'entrée"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /stocks/in [post]
func (h *StockHandler) StockInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req stockInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	if err := h.stockService.StockIn(ctx, req.ProductID, req.ShopID, ownerID, req.Quantity, req.Comment); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Stock mis à jour avec succès"})
}

// AdjustStockHandler godoc
// @Summary      Ajustement manuel de stock
// @Tags         Stocks
// @Accept       json
// @Produce      json
// @Param        stock  body  adjustStockRequest  true  "Données du mouvement"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /stocks/adjust [post]
func (h *StockHandler) AdjustStockHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req adjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	if err := h.stockService.AdjustStock(ctx, req.ProductID, req.ShopID, ownerID, req.QuantityChange, req.MovementType, req.Comment); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Ajustement effectué avec succès"})
}

// AllMovementsHandler godoc
// @Summary      Liste tous les mouvements de stock
// @Tags         Stocks
// @Security     BearerAuth
// @Success      200  {array}  models.StockMovement
// @Router       /stocks/movements [get]
func (h *StockHandler) AllMovementsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	movements, err := h.stockService.GetMovements(ctx, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"count": len(movements), "movements": movements})
}

// MovementsByProductHandler godoc
// @Summary      Mouvements d'un produit
// @Tags         Stocks
// @Param        product_id  path  int  true  "ID du produit"
// @Security     BearerAuth
// @Success      200  {array}  models.StockMovement
// @Router       /stocks/movements/product/{product_id} [get]
func (h *StockHandler) MovementsByProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	productID, err := strconv.Atoi(chi.URLParam(r, "product_id"))
	if err != nil || productID <= 0 {
		http.Error(w, "product_id invalide", http.StatusBadRequest)
		return
	}

	movements, err := h.stockService.GetMovementsByProduct(ctx, productID, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"count": len(movements), "movements": movements})
}

// MovementsByShopHandler godoc
// @Summary      Mouvements d'une boutique
// @Tags         Stocks
// @Param        shop_id  path  int  true  "ID de la boutique"
// @Security     BearerAuth
// @Success      200  {array}  models.StockMovement
// @Router       /stocks/movements/shop/{shop_id} [get]
func (h *StockHandler) MovementsByShopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shopID, err := strconv.Atoi(chi.URLParam(r, "shop_id"))
	if err != nil || shopID <= 0 {
		http.Error(w, "shop_id invalide", http.StatusBadRequest)
		return
	}

	movements, err := h.stockService.GetMovementsByShop(ctx, shopID, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"count": len(movements), "movements": movements})
}

type setMinStockRequest struct {
	ProductID int `json:"product_id" validate:"required,gt=0"`
	ShopID    int `json:"shop_id" validate:"required,gt=0"`
	MinStock  int `json:"min_stock" validate:"gte=0"`
}

// LowStockAlertsHandler godoc
// @Summary      Produits en stock faible (quantité ≤ seuil)
// @Tags         Stocks
// @Security     BearerAuth
// @Success      200  {array}  models.Stock
// @Router       /stocks/alerts [get]
func (h *StockHandler) LowStockAlertsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}
	shopID, _ := strconv.Atoi(r.URL.Query().Get("shop_id"))
	stocks, err := h.stockService.GetLowStocks(ctx, ownerID, shopID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"count": len(stocks), "alerts": stocks})
}

// SetMinStockHandler godoc
// @Summary      Définir le seuil d'alerte stock
// @Tags         Stocks
// @Accept       json
// @Produce      json
// @Param        body  body  setMinStockRequest  true  "Seuil minimum"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /stocks/min [put]
func (h *StockHandler) SetMinStockHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}
	var req setMinStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}
	if err := h.stockService.SetMinStock(ctx, req.ProductID, req.ShopID, ownerID, req.MinStock); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Seuil d'alerte mis à jour"})
}

func (h *StockHandler) GetLogsHandler(w http.ResponseWriter, r *http.Request) {
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
		limit = 30
	}
	offset := (page - 1) * limit
	shopID, _ := strconv.Atoi(r.URL.Query().Get("shop_id"))
	dateFrom := validateDateParam(r.URL.Query().Get("date_from"))
	dateTo := validateDateParam(r.URL.Query().Get("date_to"))

	movements, total, err := h.stockService.GetLogs(ctx, ownerID, shopID, dateFrom, dateTo, limit, offset)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des logs", http.StatusInternalServerError)
		return
	}
	if movements == nil {
		movements = []*models.StockMovement{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"logs":        movements,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": math.Ceil(float64(total) / float64(limit)),
	})
}

func (h *StockHandler) StockRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.AllStocksHandler)
	r.Get("/detail", h.GetStockByProductAndShopHandler)
	r.Get("/alerts", h.LowStockAlertsHandler)

	r.Post("/init", h.InitStockHandler)
	r.Post("/in", h.StockInHandler)
	r.Post("/adjust", h.AdjustStockHandler)
	r.Put("/min", h.SetMinStockHandler)

	r.Route("/movements", func(r chi.Router) {
		r.Get("/", h.AllMovementsHandler)
		r.Get("/product/{product_id}", h.MovementsByProductHandler)
		r.Get("/shop/{shop_id}", h.MovementsByShopHandler)
	})

	return r
}
