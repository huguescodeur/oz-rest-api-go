package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/middlewares"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/ctxkeys"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/responses"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
)

type loginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type resetPasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type updateMeRequest struct {
	Firstname string `json:"firstname" validate:"required"`
	Lastname  string `json:"lastname" validate:"required"`
	Username  string `json:"username" validate:"required,min=3"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,len=10"`
}

type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3" example:"hugues_vendeur"`
	Email     string `json:"email" validate:"required,email" example:"hugues@zoro.com"`
	Password  string `json:"password" validate:"required,min=6" example:"pass1234"`
	Firstname string `json:"firstname" validate:"required" example:"Hugues"`
	Lastname  string `json:"lastname" validate:"required" example:"Dupont"`
	Phone     string `json:"phone" validate:"required,len=10" example:"0708091011"`
	Role      string `json:"role" validate:"required,oneof=admin vendeur" example:"vendeur"`
	ParentID  *int   `json:"parentId,omitempty"`
}

type AuthHandler struct {
	service  *services.AuthService
	validate *validator.Validate
}

func NewAuthHandler(s *services.AuthService) *AuthHandler {
	return &AuthHandler{service: s, validate: validator.New()}
}

// RegisterHandler godoc
// @Summary      Inscription d'un utilisateur
// @Description  Crée un nouveau compte utilisateur et retourne un token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterRequest  true  "Données d'inscription"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {string}  string "Payload ou validation invalide"
// @Router       /auth/register [post]
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

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

	user, token, err := h.service.Register(r.Context(), newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Inscription réussie",
		"user":    user,
		"token":   token,
	})
}

// LoginHandler godoc
// @Summary      Connexion utilisateur
// @Description  Authentifie l'utilisateur et retourne un token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      loginRequest  true  "Identifiant et mot de passe"
// @Success      200  {object}  map[string]interface{}
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

	user, token, err := h.service.Login(r.Context(), req.Identifier, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"user": user, "token": token})
}

// LogoutHandler godoc
// @Summary      Déconnexion
// @Description  Invalide la session (côté client).
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /auth/logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Déconnexion réussie. Veuillez supprimer le token côté client."})
}

// ResetPasswordHandler godoc
// @Summary      Changement de mot de passe
// @Description  Permet à l'utilisateur connecté de modifier son mot de passe
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        passwords  body      resetPasswordRequest  true  "Ancien et nouveau mot de passe"
// @Success      200  {object}  map[string]string
// @Failure      403  {string}  string "Ancien mot de passe incorrect"
// @Router       /auth/reset-password [patch]
func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUserID, _ := ctx.Value(ctxkeys.UserIDKey).(int)

	currentUserUUIDStr, ok := ctx.Value(ctxkeys.UserUUIDKey).(string)
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

	if err := h.service.ResetPassword(ctx, currentUserID, currentUserUUID, req.OldPassword, req.NewPassword); err != nil {
		http.Error(w, "l'ancien mot de passe est incorrect", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Votre mot de passe a été mis à jour avec succès"}`))
}

func (h *AuthHandler) UpdateMeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := ctx.Value(ctxkeys.UserIDKey).(int)

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	if err := h.service.UpdateMe(ctx, userID, req.Firstname, req.Lastname, req.Username, req.Email, req.Phone); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profil mis à jour avec succès"})
}

func (h *AuthHandler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Identifier == "" {
		http.Error(w, "Email ou username requis", http.StatusBadRequest)
		return
	}
	// Toujours répondre succès (évite l'énumération d'emails)
	_ = h.service.ForgotPassword(r.Context(), body.Identifier)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Si ce compte existe, un email de réinitialisation a été envoyé."})
}

func (h *AuthHandler) ValidateResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token manquant", http.StatusBadRequest)
		return
	}
	if err := h.service.ValidateResetToken(r.Context(), token); err != nil {
		http.Error(w, "Lien invalide ou expiré", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"valid": true})
}

func (h *AuthHandler) ResetPasswordByTokenHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}
	if body.Token == "" || len(body.NewPassword) < 6 {
		http.Error(w, "Token et mot de passe (min 6 caractères) requis", http.StatusBadRequest)
		return
	}
	if err := h.service.ResetPasswordByToken(r.Context(), body.Token, body.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Mot de passe réinitialisé avec succès."})
}

func (h *AuthHandler) AuthRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.RegisterHandler)
	r.Post("/login", h.LoginHandler)
	r.Post("/forgot-password", h.ForgotPasswordHandler)
	r.Get("/validate-reset-token", h.ValidateResetTokenHandler)
	r.Post("/reset-password-token", h.ResetPasswordByTokenHandler)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware)
		r.Post("/logout", h.LogoutHandler)
		r.Patch("/reset-password", h.ResetPasswordHandler)
		r.Patch("/me", h.UpdateMeHandler)
	})
	return r
}
