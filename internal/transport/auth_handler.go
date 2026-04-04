package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/middlewares"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/ctxkeys"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/responses"
	"github.com/huguescodeur/zoro_rest_api_go/internal/services"
)

type loginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type resetPasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3" example:"jean_vendeur"`
	Email     string `json:"email" validate:"required,email" example:"jean@zoro.com"`
	Password  string `json:"password" validate:"required,min=6" example:"pass1234"`
	Firstname string `json:"firstname" validate:"required" example:"Jean"`
	Lastname  string `json:"lastname" validate:"required" example:"Dupont"`
	Phone     string `json:"phone" validate:"required,len=10" example:"0708091011"`
	Role      string `json:"role" validate:"required,oneof=admin vendeur" example:"vendeur"`

	ParentID *int `json:"parentId,omitempty"`
}

type AuthHandler struct {
	service  *services.AuthService
	validate *validator.Validate
}

func NewAuthHandler(s *services.AuthService) *AuthHandler {
	return &AuthHandler{
		service:  s,
		validate: validator.New()}
}

// RegisterHandler godoc
// @Summary      Inscription d'un utilisateur
// @Description  Crée un nouveau compte utilisateur et retourne un token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterRequest  true  "Données d'inscription"
// @Success      201  {object}  map[string]interface{} "Ex: {message: '...', user: {}, token: '...'}"
// @Failure      400  {string}  string "Payload ou validation invalide"
// @Router       /auth/register [post]
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	newUser := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Phone:     req.Phone,
		Role:      req.Role,
		ParentID:  req.ParentID,
	}

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Payload invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(newUser); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	user, token, err := h.service.Register(newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Inscription réussie",
		"user":    user,
		"token":   token,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// LoginHandler godoc
// @Summary      Connexion utilisateur
// @Description  Authentifie l'utilisateur et retourne un token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      loginRequest  true  "Identifiant et mot de passe"
// @Success      200  {object}  map[string]interface{} "Ex: {user: {}, token: '...'}"
// @Failure      401  {string}  string "Identifiants incorrects"
// @Router       /auth/login [post]
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	user, token, err := h.service.Login(req.Identifier, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"user":  user,
		"token": token,
	}); err != nil {
		return
	}
}

// LogoutHandler godoc
// @Summary      Déconnexion
// @Description  Invalide la session (côté client). Retourne un message de confirmation.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string "message: Déconnexion réussie"
// @Router       /auth/logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Déconnexion réussie. Veuillez supprimer le token côté client."}); err != nil {
		return
	}
}

// ResetPasswordHandler godoc
// @Summary      Changement de mot de passe
// @Description  Permet à l'utilisateur connecté de modifier son mot de passe
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        passwords  body      resetPasswordRequest  true  "Ancien et nouveau mot de passe"
// @Success      200  {object}  map[string]string "message: Mot de passe mis à jour"
// @Failure      403  {string}  string "Ancien mot de passe incorrect"
// @Router       /auth/reset-password [patch]
func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := r.Context().Value(ctxkeys.UserIDKey).(int)

	currentUserUUIDStr, ok := r.Context().Value(ctxkeys.UserUUIDKey).(string)
	if !ok {
		http.Error(w, "Identifiant UUID introuvable dans la session", http.StatusUnauthorized)
		return
	}

	currentUserUUID, err := uuid.Parse(currentUserUUIDStr)

	if err != nil {
		http.Error(w, "Format d'identifiant de session corrompu", http.StatusInternalServerError)
		return
	}

	var req resetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	err = h.service.ResetPassword(currentUserID, currentUserUUID, req.OldPassword, req.NewPassword)
	if err != nil {
		http.Error(w, "l'ancien mot de passe est incorrect", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"message": "Votre mot de passe a été mis à jour avec succès"}`))
}

func (h *AuthHandler) AuthRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.RegisterHandler)
	r.Post("/login", h.LoginHandler)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware)
		r.Post("/logout", h.LogoutHandler)
		r.Patch("/reset-password", h.ResetPasswordHandler)

	})
	return r
}
