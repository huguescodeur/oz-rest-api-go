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

type ShopHandler struct {
	shopService *services.ShopService
	validate    *validator.Validate
}

type AssignShopsRequest struct {
	ShopIDs   []int `json:"shopIDs" validate:"required" example:"1,2,3"`
	VendeurID int   `json:"vendeurID" validate:"required" example:"5"`
}

func NewShopHandler(s *services.ShopService) *ShopHandler {
	return &ShopHandler{
		shopService: s,
		validate:    validator.New(),
	}
}

// AssignVendeurToShopsHandler godoc
// @Summary      Assigner un vendeur à des magasins
// @Description  Lier un vendeur à des magasin.
// @Description  Note: Un Admin ne peut assigner que ses propres vendeurs (parent_id doit correspondre).
// @Tags         shops
// @Accept       json
// @Produce      json
// @Param        payload  body      AssignShopsRequest  true  "IDs du vendeur et des magasins"
// @Success      200      {object}  map[string]string   "message: Assignation réussie"
// @Failure      403      {string}  string              "Vous n'êtes pas autorisé (Vendeur non lié à cet Admin)"
// @Router       /shops/assign-vendeur [post]
// @Security     BearerAuth
func (h *ShopHandler) AssignVendeurToShopsHandler(w http.ResponseWriter, r *http.Request) {
	var req AssignShopsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

	err := h.shopService.AssignVendeurToShops(req.VendeurID, req.ShopIDs, ownerID, ownerRole)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur d'assignation: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vendeur assigné avec succès"})
}

// AllShopHandler godoc
// @Summary      Liste tous les magasins
// @Description  Récupère la liste des magasins de l'utilisateur
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  shopsResponse
// @Failure      500  {string}  string "Erreur de récupération"
// @Router       /shops [get]
func (h *ShopHandler) AllShopHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shops, err := h.shopService.GetAllShops(ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		http.Error(w, "Erreur de récupération: "+err.Error(), code)
		return
	}

	response := shopsResponse{
		Count: len(shops),
		Shops: shops,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Json invalide", http.StatusInternalServerError)
		return
	}

}

// AllShopVendeurHandler godoc
// @Summary      Liste les magasins d'un vendeur
// @Description  Récupère tous les magasins associés au vendeur connecté (via son ID d'owner)
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  shopsResponse "Liste des magasins avec le compte"
// @Failure      401  {string}  string "Utilisateur non identifié"
// @Failure      404  {string}  string "Aucun magasin trouvé"
// @Failure      500  {string}  string "Erreur interne ou JSON invalide"
// @Router       /shops/vendeur [get]
func (h *ShopHandler) AllShopVendeurHandler(w http.ResponseWriter, r *http.Request) {
	vendeurID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shops, err := h.shopService.GetAllShopsVendeur(vendeurID)
	if err != nil {
		code := errs.MapHTTPError(err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		http.Error(w, "Erreur de récupération: "+err.Error(), code)
		return
	}

	response := shopsResponse{
		Count: len(shops),
		Shops: shops,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Json invalide", http.StatusInternalServerError)
		return
	}

}

// GetShopByUUIDHandler godoc
// @Summary      Détails d'un magasin
// @Description  Récupère un magasin spécifique par son UUID
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Param        uuid   path      string  true  "UUID du magasin"
// @Security     BearerAuth
// @Success      200  {object}  models.Shop
// @Failure      400  {string}  string "Format d'identifiant invalide"
// @Failure      404  {string}  string "Aucun magasin trouvé"
// @Router       /shops/{uuid} [get]
func (h *ShopHandler) GetShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	uuidStr := chi.URLParam(r, "uuid")
	shopUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shop, err := h.shopService.GetShopByUUID(shopUUID, ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(
			w,
			"Aucun magasin trouvé:"+err.Error(),
			// err.Error(),
			code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shop)
}

// CreateShopHandler godoc
// @Summary      Créer un magasin
// @Description  Ajoute un nouveau magasin
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Param        shop  body      shopRequest  true  "Données du magasin"
// @Security     BearerAuth
// @Success      201  {object}  models.Shop
// @Failure      400  {string}  string "Json Invalide"
// @Router       /shops [post]
func (h *ShopHandler) CreateShopHandler(w http.ResponseWriter, r *http.Request) {
	var req shopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json Invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shopCreate := &models.Shop{
		ShopName:    req.ShopName,
		ShopAddress: req.ShopAddress,
		ShopPhone:   req.ShopPhone,
		ShopMail:    req.ShopMail,
		OwnerID:     ownerID,
	}

	shopCreate.ShopUUID = uuid.New()

	shop, err := h.shopService.CreateShop(shopCreate)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur lors de la création: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shop)
}

// UpdateShopByUUIDHandler godoc
// @Summary      Modifier un magasin
// @Description  Met à jour les informations d'un magasin existant
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Param        uuid  path      string       true  "UUID du magasin"
// @Param        shop  body      shopRequest  true  "Nouvelles données"
// @Security     BearerAuth
// @Success      200  {object}  models.Shop
// @Failure      400  {string}  string "Données invalides"
// @Failure      404  {string}  string "Mise à jour impossible"
// @Router       /shops/{uuid} [put]
func (h *ShopHandler) UpdateShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	var req shopRequest
	uuidStr := chi.URLParam(r, "uuid")
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shopUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

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

	shop, err := h.shopService.UpdateShop(shopUUID, ownerID, shopUpdate)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Mise à jour impossible: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(shop)
}

// DeleteShopByUUIDHandler godoc
// @Summary      Supprimer un magasin
// @Description  Suppression logique (soft delete) d'un magasin
// @Tags         Shops
// @Param        uuid   path      string  true  "UUID du magasin"
// @Security     BearerAuth
// @Success      204  "No Content"
// @Failure      404  {string}  string "Echec de suppression"
// @Router       /shops/{uuid} [delete]
func (h *ShopHandler) DeleteShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	uuidStr := chi.URLParam(r, "uuid")
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	shopUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := h.shopService.DeleteShop(shopUUID, ownerID); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Echec de suppression: "+err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreShopByUUIDHandler godoc
// @Summary      Restaurer un magasin
// @Description  Réactive un magasin supprimé
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Param        uuid   path      string  true  "UUID du magasin"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string "message: Magasin restauré avec succès"
// @Failure      404  {string}  string "Erreur lors du restore"
// @Router       /shops/{uuid} [patch]
func (h *ShopHandler) RestoreShopByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	uuidStr := chi.URLParam(r, "uuid")
	shopUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	if err := h.shopService.RestoreShop(shopUUID, ownerID); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur lors du restore: "+err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusOK)
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
