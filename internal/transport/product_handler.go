package transport

import (
	"encoding/json"
	"fmt"
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

type productsResponse struct {
	Count    int               `json:"count"`
	Products []*models.Product `json:"products"`
}

type productRequest struct {
	Name       string `json:"productName" validate:"required"`
	CategoryID int    `json:"categoryID" validate:"required"`
	UnitPrice  int    `json:"unitPrice" validate:"required,gte=0"`
	// OwnerID     int       `json:"ownerID" validate:"required"`
}

type ProductHandler struct {
	productService *services.ProductService
	validate       *validator.Validate
}

func NewProductHandler(s *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: s,
		validate:       validator.New(),
	}
}

// AllProductHandler godoc
// @Summary      Liste tous les produits
// @Description  Récupère la liste des produits appartenant à l'utilisateur connecté
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  productsResponse
// @Failure      401  {string}  string "Utilisateur non identifié"
// @Failure      500  {string}  string "Erreur interne"
// @Router       /products [get]
func (h *ProductHandler) AllProductHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	// fmt.Printf("Owner ID: %+v", ownerID)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	products, err := h.productService.GetAllProducts(ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		http.Error(w, "Erreur de récupération: "+err.Error(), code)
		return
	}

	response := productsResponse{
		Count:    len(products),
		Products: products,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Json invalide", http.StatusInternalServerError)
		return
	}

}

// GetProductByUUIDHandler godoc
// @Summary      Détails d'un produit
// @Description  Récupère un produit spécifique par son UUID (doit appartenir à l'owner)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        uuid   path      string  true  "UUID du produit"
// @Security     BearerAuth
// @Success      200  {object}  models.Product
// @Failure      400  {string}  string "Format d'identifiant invalide"
// @Failure      404  {string}  string "Aucun produit trouvé"
// @Router       /products/{uuid} [get]
func (h *ProductHandler) GetProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	// fmt.Printf("Owner ID: %+v\n", ownerID)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	uuidStr := chi.URLParam(r, "uuid")
	productUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	product, err := h.productService.GetProductByUUID(productUUID, ownerID)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(
			w,
			"Aucun produit trouvé:"+err.Error(),
			// err.Error(),
			code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "Json invalide", http.StatusInternalServerError)
		return
	}

}

// CreateProductHandler godoc
// @Summary      Créer un produit
// @Description  Ajoute un nouveau produit dans la base de données
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        product  body      productRequest  true  "Données du produit"
// @Security     BearerAuth
// @Success      201  {object}  productRequest
// @Failure      400  {string}  string "Json ou validation invalide"
// @Router       /products [post]
func (h *ProductHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req productRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json Invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	productModel := &models.Product{
		ProductUUID: uuid.New(),
		ProductName: req.Name,
		UnitPrice:   req.UnitPrice,
		OwnerID:     ownerID,
	}

	if req.CategoryID > 0 {
		productModel.ProductCategory = &models.Category{
			CategoryID: req.CategoryID,
		}
	}

	product, err := h.productService.CreateProduct(productModel)
	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur lors de la création: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// UpdateProductByUUIDHandler godoc
// @Summary      Modifier un produit
// @Description  Met à jour les informations d'un produit existant
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        uuid     path      string          true  "UUID du produit"
// @Param        product  body      productRequest  true  "Nouvelles données"
// @Security     BearerAuth
// @Success      200  {object}  models.Product
// @Failure      400  {string}  string "Données invalides"
// @Failure      404  {string}  string "Produit introuvable"
// @Router       /products/{uuid} [put]
func (h *ProductHandler) UpdateProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req productRequest
	uuidStr := chi.URLParam(r, "uuid")

	productUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json invalide", http.StatusBadRequest)
		return
	}
	fmt.Printf("Requête reçue : %+v\n", req)

	productModel := &models.Product{
		ProductName: req.Name,
		UnitPrice:   req.UnitPrice,
		OwnerID:     ownerID,
	}

	if req.CategoryID > 0 {
		productModel.ProductCategory = &models.Category{
			CategoryID: req.CategoryID,
		}
	}

	product, err := h.productService.UpdateProduct(productUUID, ownerID, productModel)

	if err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Mise à jour impossible: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

// DeleteProductByUUIDHandler godoc
// @Summary      Supprimer un produit
// @Description  Suppression logique d'un produit (soft delete)
// @Tags         Products
// @Param        uuid   path      string  true  "UUID du produit"
// @Security     BearerAuth
// @Success      204  "No Content"
// @Failure      404  {string}  string "Echec de suppression"
// @Router       /products/{uuid} [delete]
func (h *ProductHandler) DeleteProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	uuidStr := chi.URLParam(r, "uuid")
	productUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := h.productService.DeleteProduct(productUUID, ownerID); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Echec de suppression: "+err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]string{"message": "Produit supprimé avec succès"})
}

// RestoreProductByUUIDHandler godoc
// @Summary      Restaurer un produit
// @Description  Réactive un produit qui a été supprimé
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        uuid   path      string  true  "UUID du produit"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string "message: Produit restauré avec succès"
// @Failure      404  {string}  string "Erreur lors du restore"
// @Router       /products/{uuid} [patch]
func (h *ProductHandler) RestoreProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	uuidStr := chi.URLParam(r, "uuid")
	productUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := h.productService.RestoreProduct(productUUID, ownerID); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur lors du restore: "+err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Produit restauré avec succès"})
}

func (h *ProductHandler) ProductRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.AllProductHandler)
	r.Post("/", h.CreateProductHandler)

	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.GetProductByUUIDHandler)
		r.Put("/", h.UpdateProductByUUIDHandler)
		r.Delete("/", h.DeleteProductByUUIDHandler)
		r.Patch("/", h.RestoreProductByUUIDHandler)
	})

	return r
}
