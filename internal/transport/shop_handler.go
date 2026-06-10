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

type shopsResponse struct {
	Count int            `json:"count"`
	Shops []*models.Shop `json:"shops"`
}

type shopRequest struct {
	ShopName    string `json:"shopName" validate:"required,min=3"`
	ShopAddress string `json:"shopAddress" validate:"required"`
	ShopPhone   string `json:"shopPhone" validate:"required"`
	ShopMail    string `json:"shopMail" validate:"required"`
}

type AssignShopsRequest struct {
	ShopIDs   []int `json:"shopIDs" validate:"required"`
	VendeurID int   `json:"vendeurID" validate:"required"`
}

type ShopHandler struct {
	shopService *services.ShopService
	validate    *validator.Validate
}

func NewShopHandler(s *services.ShopService) *ShopHandler {
	return &ShopHandler{shopService: s, validate: validator.New()}
}

// AssignVendeurToShopsHandler godoc
// @Summary      Assigner un vendeur à des magasins
// @Tags         shops
// @Accept       json
// @Produce      json
// @Param        payload  body  AssignShopsRequest  true  "IDs du vendeur et des magasins"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /shops/assign-vendeur [post]
func (h *ShopHandler) AssignVendeurToShopsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req AssignShopsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := ctx.Value(ctxkeys.UserRoleKey).(string)

	if err := h.shopService.AssignVendeurToShops(ctx, req.VendeurID, req.ShopIDs, ownerID, ownerRole); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vendeur assigné avec succès"})
}

// AllShopHandler godoc
// @Summary      Liste paginée des magasins
// @Tags         Shops
// @Param        page   query  int  false  "Numéro de la page (défaut: 1)"
// @Param        limit  query  int  false  "Nombre d'éléments (défaut: 20)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Router       /shops [get]
func (h *ShopHandler) AllShopHandler(w http.ResponseWriter, r *http.Request) {
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

	shops, total, err := h.shopService.GetAllShops(ctx, ownerID, limit, offset)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"shops":       shops,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": math.Ceil(float64(total) / float64(limit)),
	})
}

// AllShopVendeurHandler godoc
// @Summary      Liste les magasins d'un vendeur
// @Tags         Shops
// @Security     BearerAuth
// @Success      200  {object}  shopsResponse
// @Router       /shops/vendeur [get]
func (h *ShopHandler) AllShopVendeurHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vendeurID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shops, err := h.shopService.GetAllShopsVendeur(ctx, vendeurID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shopsResponse{Count: len(shops), Shops: shops})
}

// GetShopByUUIDHandler godoc
// @Summary      Détails d'un magasin
// @Tags         Shops
// @Param        uuid  path  string  true  "UUID du magasin"
// @Security     BearerAuth
// @Success      200  {object}  models.Shop
// @Router       /shops/{uuid} [get]
func (h *ShopHandler) GetShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shop, err := h.shopService.GetShopByUUID(ctx, shopUUID, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shop)
}

// CreateShopHandler godoc
// @Summary      Créer un magasin
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Param        shop  body  shopRequest  true  "Données du magasin"
// @Security     BearerAuth
// @Success      201  {object}  models.Shop
// @Router       /shops [post]
func (h *ShopHandler) CreateShopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req shopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json Invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shopCreate := &models.Shop{
		ShopUUID:    uuid.New(),
		ShopName:    req.ShopName,
		ShopAddress: req.ShopAddress,
		ShopPhone:   req.ShopPhone,
		ShopMail:    req.ShopMail,
		OwnerID:     ownerID,
	}

	shop, err := h.shopService.CreateShop(ctx, shopCreate)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shop)
}

// UpdateShopByUUIDHandler godoc
// @Summary      Modifier un magasin
// @Tags         Shops
// @Param        uuid  path  string      true  "UUID du magasin"
// @Param        shop  body  shopRequest  true  "Nouvelles données"
// @Security     BearerAuth
// @Success      200  {object}  models.Shop
// @Router       /shops/{uuid} [put]
func (h *ShopHandler) UpdateShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req shopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json invalide", http.StatusBadRequest)
		return
	}

	shopUpdate := &models.Shop{
		ShopName:    req.ShopName,
		ShopAddress: req.ShopAddress,
		ShopPhone:   req.ShopPhone,
		ShopMail:    req.ShopMail,
		OwnerID:     ownerID,
	}

	shop, err := h.shopService.UpdateShop(ctx, shopUUID, ownerID, shopUpdate)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shop)
}

// DeleteShopByUUIDHandler godoc
// @Summary      Supprimer un magasin (soft delete)
// @Tags         Shops
// @Param        uuid  path  string  true  "UUID du magasin"
// @Security     BearerAuth
// @Success      204  "No Content"
// @Router       /shops/{uuid} [delete]
func (h *ShopHandler) DeleteShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	if err := h.shopService.DeleteShop(ctx, shopUUID, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreShopByUUIDHandler godoc
// @Summary      Restaurer un magasin
// @Tags         Shops
// @Param        uuid  path  string  true  "UUID du magasin"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /shops/{uuid} [patch]
func (h *ShopHandler) RestoreShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	if err := h.shopService.RestoreShop(ctx, shopUUID, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Magasin restauré avec succès"})
}

func (h *ShopHandler) ShopRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.AllShopHandler)
	r.Get("/vendeur", h.AllShopVendeurHandler)
	r.Post("/assign-vendeur", h.AssignVendeurToShopsHandler)
	r.Post("/", h.CreateShopHandler)

	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.GetShopByUUIDHandler)
		r.Put("/", h.UpdateShopByUUIDHandler)
		r.Delete("/", h.DeleteShopByUUIDHandler)
		r.Patch("/", h.RestoreShopByUUIDHandler)
	})

	return r
}
