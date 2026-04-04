package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
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
	return &StockHandler{
		stockService: s,
		validate:     validator.New(),
	}
}

// AllStocksHandler godoc
// @Summary      Liste tous les stocks
// @Description  Récupère tous les stocks des produits de l'utilisateur connecté
// @Tags         Stocks
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Stock
// @Failure      401  {string}  string "Utilisateur non identifié"
// @Failure      500  {string}  string "Erreur interne"
// @Router       /stocks [get]
func (h *StockHandler) AllStocksHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	stocks, err := h.stockService.GetAllStocks(ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur de récupération: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(stocks),
		"stocks": stocks,
	})
}

// GetStockByProductAndShopHandler godoc
// @Summary      Stock d'un produit dans une boutique
// @Description  Récupère le niveau de stock d'un produit dans une boutique spécifique
// @Tags         Stocks
// @Produce      json
// @Param        product_id  query  int  true  "ID du produit"
// @Param        shop_id     query  int  true  "ID de la boutique"
// @Security     BearerAuth
// @Success      200  {object}  models.Stock
// @Failure      400  {string}  string "Paramètres invalides"
// @Failure      404  {string}  string "Stock introuvable"
// @Router       /stocks/detail [get]
func (h *StockHandler) GetStockByProductAndShopHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
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

	stock, err := h.stockService.GetStockByProductAndShop(productID, shopID, ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Stock introuvable: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stock)
}

// InitStockHandler godoc
// @Summary      Initialiser un stock
// @Description  Crée une entrée stock à 0 pour un produit dans une boutique
// @Tags         Stocks
// @Accept       json
// @Produce      json
// @Param        stock  body  initStockRequest  true  "Données d'initialisation"
// @Security     BearerAuth
// @Success      201  {object}  map[string]string
// @Failure      400  {string}  string "Données invalides"
// @Router       /stocks/init [post]
func (h *StockHandler) InitStockHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
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

	if err := h.stockService.InitStock(req.ProductID, req.ShopID, ownerID); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur d'initialisation: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Stock initialisé avec succès"})
}

// StockInHandler godoc
// @Summary      Entrée de stock
// @Description  Ajoute de la quantité en stock (réception de marchandise)
// @Tags         Stocks
// @Accept       json
// @Produce      json
// @Param        stock  body  stockInRequest  true  "Données d'entrée"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Failure      400  {string}  string "Données invalides"
// @Router       /stocks/in [post]
func (h *StockHandler) StockInHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
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

	if err := h.stockService.StockIn(req.ProductID, req.ShopID, ownerID, req.Quantity, req.Comment); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur d'entrée stock: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Stock mis à jour avec succès"})
}

// AdjustStockHandler godoc
// @Summary      Ajustement manuel de stock
// @Description  Applique un mouvement de stock (SALE, GIFT, LOSS, ADJUSTMENT)
// @Tags         Stocks
// @Accept       json
// @Produce      json
// @Param        stock  body  adjustStockRequest  true  "Données du mouvement"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Failure      400  {string}  string "Données invalides ou stock insuffisant"
// @Router       /stocks/adjust [post]
func (h *StockHandler) AdjustStockHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
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

	if err := h.stockService.AdjustStock(req.ProductID, req.ShopID, ownerID, req.QuantityChange, req.MovementType, req.Comment); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Ajustement impossible: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ajustement effectué avec succès"})
}

// AllMovementsHandler godoc
// @Summary      Liste tous les mouvements de stock
// @Description  Récupère l'historique complet des mouvements de l'owner
// @Tags         Stocks
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.StockMovement
// @Failure      401  {string}  string "Utilisateur non identifié"
// @Router       /stocks/movements [get]
func (h *StockHandler) AllMovementsHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	movements, err := h.stockService.GetMovements(ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur de récupération: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(movements),
		"movements": movements,
	})
}

// MovementsByProductHandler godoc
// @Summary      Mouvements d'un produit
// @Description  Historique des mouvements pour un produit donné
// @Tags         Stocks
// @Produce      json
// @Param        product_id  path  int  true  "ID du produit"
// @Security     BearerAuth
// @Success      200  {array}   models.StockMovement
// @Failure      400  {string}  string "product_id invalide"
// @Router       /stocks/movements/product/{product_id} [get]
func (h *StockHandler) MovementsByProductHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	productID, err := strconv.Atoi(chi.URLParam(r, "product_id"))
	if err != nil || productID <= 0 {
		http.Error(w, "product_id invalide", http.StatusBadRequest)
		return
	}

	movements, err := h.stockService.GetMovementsByProduct(productID, ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(movements),
		"movements": movements,
	})
}

// MovementsByShopHandler godoc
// @Summary      Mouvements d'une boutique
// @Description  Historique des mouvements pour une boutique donnée
// @Tags         Stocks
// @Produce      json
// @Param        shop_id  path  int  true  "ID de la boutique"
// @Security     BearerAuth
// @Success      200  {array}   models.StockMovement
// @Failure      400  {string}  string "shop_id invalide"
// @Router       /stocks/movements/shop/{shop_id} [get]
func (h *StockHandler) MovementsByShopHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shopID, err := strconv.Atoi(chi.URLParam(r, "shop_id"))
	if err != nil || shopID <= 0 {
		http.Error(w, "shop_id invalide", http.StatusBadRequest)
		return
	}

	movements, err := h.stockService.GetMovementsByShop(shopID, ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(movements),
		"movements": movements,
	})
}

func (h *StockHandler) StockRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.AllStocksHandler)
	r.Get("/detail", h.GetStockByProductAndShopHandler)

	r.Post("/init", h.InitStockHandler)
	r.Post("/in", h.StockInHandler)
	r.Post("/adjust", h.AdjustStockHandler)

	r.Route("/movements", func(r chi.Router) {
		r.Get("/", h.AllMovementsHandler)
		r.Get("/product/{product_id}", h.MovementsByProductHandler)
		r.Get("/shop/{shop_id}", h.MovementsByShopHandler)
	})

	return r
}
