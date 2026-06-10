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

type UserHandler struct {
	userService *services.UserService
	validate    *validator.Validate
}

func NewUserHandler(s *services.UserService) *UserHandler {
	return &UserHandler{userService: s, validate: validator.New()}
}

type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty"`
	Firstname *string `json:"firstname,omitempty"`
	Lastname  *string `json:"lastname,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty"`
	Role      *string `json:"role" validate:"required,oneof=admin vendeur"`
	ParentID  *int    `json:"parentId,omitempty"`
	ShopID    *int    `json:"shop_id,omitempty"`
}

// GetAllUsersHandler godoc
// @Summary      Liste paginée des utilisateurs
// @Param        page   query  int  false  "Numéro de la page (défaut: 1)"
// @Param        limit  query  int  false  "Nombre d'éléments (défaut: 10)"
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := ctx.Value(ctxkeys.UserRoleKey).(string)
	showArchived := r.URL.Query().Get("archived") == "true"

	users, total, err := h.userService.GetAllUser(ctx, ownerID, ownerRole, limit, offset, showArchived)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"users":       users,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": math.Ceil(float64(total) / float64(limit)),
	})
}

// AddUserHandler godoc
// @Summary      Créer un nouvel utilisateur
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      models.User  true  "Données de l'utilisateur"
// @Success      201   {object}  map[string]interface{}
// @Router       /users [post]
func (h *UserHandler) AddUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := ctx.Value(ctxkeys.UserRoleKey).(string)

	if err := h.validate.Struct(newUser); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	user, err := h.userService.CreateUser(ctx, &newUser, ownerID, ownerRole)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Utilisateur créé avec succès", "user": user})
}

// GetUserByUUIDHandler godoc
// @Summary      Récupère un utilisateur
// @Tags         users
// @Param        uuid  path  string  true  "User UUID"
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Router       /users/{uuid} [get]
func (h *UserHandler) GetUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := ctx.Value(ctxkeys.UserRoleKey).(string)

	user, err := h.userService.GetUserByUUID(ctx, userUUID, ownerID, ownerRole)
	if err != nil {
		http.Error(w, "Aucun utilisateur trouvé: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUserByUUIDHandler godoc
// @Summary      Met à jour un utilisateur
// @Tags         users
// @Param        uuid  path  string            true  "User UUID"
// @Param        user  body  UpdateUserRequest  true  "Update User"
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Router       /users/{uuid} [put]
func (h *UserHandler) UpdateUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	updateUser := &models.User{}
	if req.Username != nil {
		updateUser.Username = *req.Username
	}
	if req.Email != nil {
		updateUser.Email = *req.Email
	}
	if req.Firstname != nil {
		updateUser.Firstname = *req.Firstname
	}
	if req.Lastname != nil {
		updateUser.Lastname = *req.Lastname
	}
	if req.Phone != nil {
		updateUser.Phone = *req.Phone
	}
	if req.Role != nil {
		updateUser.Role = *req.Role
	}
	updateUser.ParentID = req.ParentID
	updateUser.ShopID = req.ShopID

	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := ctx.Value(ctxkeys.UserRoleKey).(string)

	user, err := h.userService.UpdateUser(ctx, userUUID, updateUser, ownerID, ownerRole)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour : "+err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUserByUUIDHandler godoc
// @Summary      Supprime un utilisateur
// @Tags         users
// @Param        uuid  path  string  true  "User UUID"
// @Security     BearerAuth
// @Success      204  "No Content"
// @Router       /users/{uuid} [delete]
func (h *UserHandler) DeleteUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := ctx.Value(ctxkeys.UserRoleKey).(string)

	if err := h.userService.DeleteUser(ctx, userUUID, ownerID, ownerRole); err != nil {
		http.Error(w, "Echec de suppression: "+err.Error(), errs.MapHTTPError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreUserByUUIDHandler godoc
// @Summary      Restaure un utilisateur
// @Tags         users
// @Param        uuid  path  string  true  "User UUID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /users/{uuid} [patch]
func (h *UserHandler) RestoreUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := ctx.Value(ctxkeys.UserRoleKey).(string)

	if err := h.userService.RestoreUser(ctx, userUUID, ownerID, ownerRole); err != nil {
		http.Error(w, "Erreur lors du restore: "+err.Error(), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Utilisateur restauré avec succès"})
}

func (h *UserHandler) GetMyShopsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(ctxkeys.UserIDKey).(int)

	shops, err := h.userService.GetMyShops(ctx, userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des boutiques", http.StatusInternalServerError)
		return
	}
	if shops == nil {
		shops = []*models.Shop{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"shops": shops})
}

func (h *UserHandler) GetUserShopsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)

	userUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "UUID invalide", http.StatusBadRequest)
		return
	}

	shops, err := h.userService.GetUserShops(ctx, userUUID, ownerID)
	if err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}
	if shops == nil {
		shops = []*models.Shop{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"shops": shops})
}

func (h *UserHandler) AssignShopsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerID, _ := ctx.Value(ctxkeys.OwnerIDKey).(int)

	userUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		http.Error(w, "UUID invalide", http.StatusBadRequest)
		return
	}

	var body struct {
		ShopIDs []int `json:"shop_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	if err := h.userService.AssignShops(ctx, userUUID, ownerID, body.ShopIDs); err != nil {
		http.Error(w, errs.SafeMessage(err, errs.MapHTTPError(err)), errs.MapHTTPError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Boutiques mises à jour"})
}

func (h *UserHandler) UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/me/shops", h.GetMyShopsHandler)
	r.Get("/", h.GetAllUsersHandler)
	r.Post("/", h.AddUserHandler)

	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.GetUserByUUIDHandler)
		r.Put("/", h.UpdateUserByUUIDHandler)
		r.Delete("/", h.DeleteUserByUUIDHandler)
		r.Patch("/", h.RestoreUserByUUIDHandler)
		r.Get("/shops", h.GetUserShopsHandler)
		r.Put("/shops", h.AssignShopsHandler)
	})

	return r
}
