package transport

import (
	"encoding/json"
	"log"
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

type productsResponse struct {
	Count    int               `json:"count"`
	Products []*models.Product `json:"products"`
}

type productRequest struct {
	Name        string `json:"product_name" validate:"required"`
	CategoryID  int    `json:"category_id" validate:"required"`
	UnitPrice   int    `json:"unit_price" validate:"required,gte=0"`
	Description string `json:"description" `
}

type ProductHandler struct {
	productService *services.ProductService
	validate       *validator.Validate
}

func NewProductHandler(s *services.ProductService) *ProductHandler {
	return &ProductHandler{productService: s, validate: validator.New()}
}

// AllProductHandler godoc
// @Summary      Liste paginée des produits
// @Tags         Products
// @Param        page   query  int  false  "Numéro de la page (défaut: 1)"
// @Param        limit  query  int  false  "Nombre d'éléments (défaut: 20)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Router       /products [get]
func (h *ProductHandler) AllProductHandler(w http.ResponseWriter, r *http.Request) {
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

	products, total, err := h.productService.GetAllProducts(ctx, ownerID, limit, offset)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"products":    products,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": math.Ceil(float64(total) / float64(limit)),
	})
}

// GetProductByUUIDHandler godoc
// @Summary      Détails d'un produit
// @Tags         Products
// @Param        uuid  path  string  true  "UUID du produit"
// @Security     BearerAuth
// @Success      200  {object}  models.Product
// @Router       /products/{uuid} [get]
func (h *ProductHandler) GetProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	productUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	product, err := h.productService.GetProductByUUID(ctx, productUUID, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// CreateProductHandler godoc
// @Summary      Créer un produit
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        product  body  productRequest  true  "Données du produit"
// @Security     BearerAuth
// @Success      201  {object}  models.Product
// @Router       /products [post]
func (h *ProductHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req productRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// fmt.Println("Erreur Json:", err.Error())
		http.Error(w, "Erreur Json:", http.StatusBadRequest)
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
		Description: req.Description,
	}
	if req.CategoryID > 0 {
		productModel.ProductCategory = &models.Category{CategoryID: req.CategoryID}
	}

	product, err := h.productService.CreateProduct(ctx, productModel)
	if err != nil {
		if err.Error() == "catégorie introuvable" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// fmt.Println("Erreur Create:", err.Error())
		// errs.SafeMessage(err, errs.MapHTTPError(err))
		http.Error(w, err.Error(), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// UpdateProductByUUIDHandler godoc
// @Summary      Modifier un produit
// @Tags         Products
// @Param        uuid     path  string          true  "UUID du produit"
// @Param        product  body  productRequest  true  "Nouvelles données"
// @Security     BearerAuth
// @Success      200  {object}  models.Product
// @Router       /products/{uuid} [put]
func (h *ProductHandler) UpdateProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	productUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		log.Println("Product UUID:", productUUID)
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	var req productRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Json invalide", http.StatusBadRequest)
		return
	}

	productModel := &models.Product{ProductName: req.Name, UnitPrice: req.UnitPrice, OwnerID: ownerID}
	if req.CategoryID > 0 {
		productModel.ProductCategory = &models.Category{CategoryID: req.CategoryID}
	}

	product, err := h.productService.UpdateProduct(ctx, productUUID, ownerID, productModel)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// DeleteProductByUUIDHandler godoc
// @Summary      Supprimer un produit (soft delete)
// @Tags         Products
// @Param        uuid  path  string  true  "UUID du produit"
// @Security     BearerAuth
// @Success      204  "No Content"
// @Router       /products/{uuid} [delete]
func (h *ProductHandler) DeleteProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	productUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := h.productService.DeleteProduct(ctx, productUUID, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreProductByUUIDHandler godoc
// @Summary      Restaurer un produit
// @Tags         Products
// @Param        uuid  path  string  true  "UUID du produit"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /products/{uuid} [patch]
func (h *ProductHandler) RestoreProductByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	productUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	if err := h.productService.RestoreProduct(ctx, productUUID, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
