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

type categoryRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

type CategoryHandler struct {
	service  *services.CategoryService
	validate *validator.Validate
}

func NewCategoryHandler(s *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: s, validate: validator.New()}
}

// AllCategoriesHandler godoc
// @Summary      Liste toutes les catégories
// @Tags         Categories
// @Security     BearerAuth
// @Success      200  {array}  models.Category
// @Router       /categories [get]
func (h *CategoryHandler) AllCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	cats, err := h.service.GetAll(ctx, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"count": len(cats), "categories": cats})
}

// CreateCategoryHandler godoc
// @Summary      Créer une catégorie
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        category  body  categoryRequest  true  "Nom de la catégorie"
// @Security     BearerAuth
// @Success      201  {object}  models.Category
// @Router       /categories [post]
func (h *CategoryHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	cat, err := h.service.Create(ctx, req.Name, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cat)
}

// UpdateCategoryHandler godoc
// @Summary      Modifier une catégorie
// @Tags         Categories
// @Param        id        path  int             true  "ID de la catégorie"
// @Param        category  body  categoryRequest  true  "Nouveau nom"
// @Security     BearerAuth
// @Success      200  {object}  models.Category
// @Router       /categories/{id} [put]
func (h *CategoryHandler) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	cat, err := h.service.Update(ctx, id, ownerID, req.Name)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cat)
}

// DeleteCategoryHandler godoc
// @Summary      Supprimer une catégorie (soft delete)
// @Tags         Categories
// @Param        id  path  int  true  "ID de la catégorie"
// @Security     BearerAuth
// @Success      204
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RestoreCategoryHandler godoc
// @Summary      Restaurer une catégorie
// @Tags         Categories
// @Param        id  path  int  true  "ID de la catégorie"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /categories/{id} [patch]
func (h *CategoryHandler) RestoreCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, ok := ctx.Value(ctxkeys.OwnerIDKey).(int)
	if !ok {
		http.Error(w, "Utilisateur non identifié", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	if err := h.service.Restore(ctx, id, ownerID); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Catégorie restaurée"})
}

func (h *CategoryHandler) CategoryRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.AllCategoriesHandler)
	r.Post("/", h.CreateCategoryHandler)
	r.Route("/{id}", func(r chi.Router) {
		r.Put("/", h.UpdateCategoryHandler)
		r.Delete("/", h.DeleteCategoryHandler)
		r.Patch("/", h.RestoreCategoryHandler)
	})
	return r
}
