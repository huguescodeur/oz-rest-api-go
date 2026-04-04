package transport

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/ctxkeys"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/errs"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/responses"
	"github.com/huguescodeur/zoro_rest_api_go/internal/services"
)

type UserHandler struct {
	userService *services.UserService
	validate    *validator.Validate
}

func NewUserHandler(s *services.UserService) *UserHandler {
	return &UserHandler{
		userService: s,
		validate:    validator.New(),
	}
}

type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty" example:"nouveau_pseudo"`
	Firstname *string `json:"firstname,omitempty" example:"Prince"`
	Lastname  *string `json:"lastname,omitempty" example:"Konan"`
	Phone     *string `json:"phone,omitempty" example:"0102030405"`
	Email     *string `json:"email,omitempty" example:"nouveau@email.com"`
	Role      *string `json:"role" validate:"required,oneof=admin vendeur" example:"vendeur"`
	ParentID  *int    `json:"parentId,omitempty" `
}

// GetAllUsersHandler godoc
// @Summary      Liste paginée des utilisateurs
// @Param        page   query      int  false  "Numéro de la page (défaut: 1)"
// @Param        limit  query      int  false  "Nombre d'éléments (défaut: 10)"
// @Router       /users [get]
func (h *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Récupération des paramètres avec valeurs par défaut
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	// Calcul de l'offset (ex: page 2, limit 10 -> on saute les 10 premiers)
	offset := (page - 1) * limit

	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

	users, total, err := h.userService.GetAllUser(ownerID, ownerRole, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// 2. Réponse enrichie pour le Front-end
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
// @Summary      Créer un nouvel utilisateur (Admin/Vendeur)
// @Description  Permet à un Admin ou Super Admin de créer un utilisateur.
// @Description  Si l'auteur est Admin, le rôle est forcé à 'vendeur' et lié à son ID.
// @Description  Si l'auteur est Super Admin, il peut définir le rôle.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      models.User  true  "Données de l'utilisateur à créer (email, username, password, firstname, lastname, phone)"
// @Success      201   {object}  map[string]interface{} "Ex: {message: '...', user: {models.User}}"
// @Failure      400   {string}  string "Données JSON invalides ou erreur de validation"
// @Failure      403   {string}  string "Accès refusé : privilèges insuffisants"
// @Failure      500   {string}  string "Erreur interne du serveur"
// @Router       /users [post]
func (h *UserHandler) AddUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser models.User

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

	if err := h.validate.Struct(newUser); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	user, err := h.userService.CreateUser(&newUser, ownerID, ownerRole)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Utilisateur créé avec succès",
		"user":    user,
	})
}

// ? Start base Get All
// GetAllUsersHandler godoc
// @Summary      Liste tous les utilisateurs
// @Description  Retourne la liste des utilisateurs. Si le rôle est 'super', tout est listé. Sinon, seulement les enfants de l'owner.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "Ex: {count: 1, users: []}"
// @Failure      403  {string}  string "Interdit"
// @Router       /users [get]
// func (h *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
// 	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
// 	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

// 	users, err := h.userService.GetAllUser(ownerID, ownerRole)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusForbidden)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]any{
// 		"count": len(users),
// 		"users": users,
// 	})
// }
// ? End base Get All

// GetUserByUUIDHandler godoc
// @Summary      Récupère un utilisateur
// @Description  Récupère les détails d'un utilisateur par son UUID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        uuid   path      string  true  "User UUID"
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      400  {string}  string "Format d'identifiant invalide"
// @Failure      404  {string}  string "Aucun utilisateur trouvé"
// @Router       /users/{uuid} [get]
func (h *UserHandler) GetUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	uuidStr := chi.URLParam(r, "uuid")

	userUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

	user, err := h.userService.GetUserByUUID(userUUID, ownerID, ownerRole)
	if err != nil {
		http.Error(w, "Aucun utilisateur trouvé: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Json invalide", http.StatusInternalServerError)
		return
	}
}

// UpdateUserByUUIDHandler godoc
// @Summary      Met à jour un utilisateur
// @Description  Met à jour les informations. Seul 'super' peut changer le rôle ou le parent.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        uuid   path      string       true  "User UUID"
// @Param        user   body      UpdateUserRequest  true  "Update User"
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      400  {string}  string "Données invalides"
// @Failure      403  {string}  string "Erreur de permission"
// @Router       /users/{uuid} [put]
func (h *UserHandler) UpdateUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserRequest
	uuidStr := chi.URLParam(r, "uuid")
	userUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	// fmt.Printf("Data Req: %+v\n", req)
	// fmt.Printf("Data Req: %+v", req)

	updateUser := &models.User{
		Username:  *req.Username,
		Email:     *req.Email,
		Firstname: *req.Firstname,
		Lastname:  *req.Lastname,
		Phone:     *req.Phone,
		Role:      *req.Role,
		ParentID:  req.ParentID,
	}

	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

	user, err := h.userService.UpdateUser(userUUID, updateUser, ownerID, ownerRole)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour : "+err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// DeleteUserByUUIDHandler godoc
// @Summary      Supprime un utilisateur
// @Tags         users
// @Param        uuid   path      string  true  "User UUID"
// @Security     BearerAuth
// @Success      204  "No Content"
// @Failure      404  {string}  string "Not Found"
// @Router       /users/{uuid} [delete]
func (h *UserHandler) DeleteUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	uuidStr := chi.URLParam(r, "uuid")
	userUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

	if err := h.userService.DeleteUser(userUUID, ownerID, ownerRole); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Echec de suppression: "+err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreUserByUUIDHandler godoc
// @Summary      Restaure un utilisateur
// @Tags         users
// @Param        uuid   path      string  true  "User UUID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string "message: Utilisateur restauré"
// @Router       /users/{uuid} [patch]
func (h *UserHandler) RestoreUserByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	uuidStr := chi.URLParam(r, "uuid")
	userUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		http.Error(w, "Format d'identifiant invalide", http.StatusBadRequest)
		return
	}

	ownerID, _ := r.Context().Value(ctxkeys.OwnerIDKey).(int)
	ownerRole, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)

	if err := h.userService.RestoreUser(userUUID, ownerID, ownerRole); err != nil {
		code := errs.MapHTTPError(err)
		http.Error(w, "Erreur lors du restore: "+err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Utilisateur restauré avec succès"})
}

func (h *UserHandler) UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetAllUsersHandler)
	r.Post("/", h.AddUserHandler)

	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.GetUserByUUIDHandler)
		r.Put("/", h.UpdateUserByUUIDHandler)
		r.Delete("/", h.DeleteUserByUUIDHandler)
		r.Patch("/", h.RestoreUserByUUIDHandler)
	})

	return r
}
