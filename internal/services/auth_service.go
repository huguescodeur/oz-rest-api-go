package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/mailer"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/utils"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authStore   store.AuthStore
	userStore   store.UserStore
	secret      string
	mailer      *mailer.Mailer
	frontendURL string
}

func NewAuthService(s store.AuthStore, u store.UserStore) *AuthService {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	return &AuthService{
		authStore:   s,
		userStore:   u,
		secret:      os.Getenv("JWT_SECRET"),
		mailer:      mailer.New(),
		frontendURL: frontendURL,
	}
}

func (s *AuthService) generateJWT(user *models.User) (string, error) {
	secret := []byte(s.secret)

	ownerID := user.ID
	if user.ParentID != nil {
		ownerID = *user.ParentID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   user.ID,
		"userUUID": user.UUID.String(),
		"ownerID":  ownerID,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(secret)
}

func (s *AuthService) Register(ctx context.Context, u *models.User) (*models.User, string, error) {
	u.Role = "admin"
	u.ParentID = nil

	cleanUsername, ok := utils.SanitizeUsername(u.Username)
	if !ok {
		return nil, "", errors.New("format de username invalide (lettres, chiffres, underscore, doit commencer par une lettre)")
	}
	u.Username = cleanUsername

	cleanPhone, ok := utils.SanitizePhone(u.Phone)
	if !ok {
		return nil, "", errors.New("numéro de téléphone invalide (doit être 10 chiffres commençant par 01, 05 ou 07)")
	}
	u.Phone = cleanPhone

	u.UUID = uuid.New()
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	u.PasswordHash = string(hashed)

	if _, err := s.authStore.CreateUser(ctx, u); err != nil {
		return nil, "", err
	}

	userInDB, err := s.authStore.GetByEmailOrUsername(ctx, u.Email)
	if err != nil {
		return nil, "", err
	}

	token, err := s.generateJWT(userInDB)
	if err != nil {
		return nil, "", err
	}

	return userInDB, token, nil
}

func (s *AuthService) Login(ctx context.Context, identifier, password string) (*models.User, string, error) {
	user, err := s.authStore.GetByEmailOrUsername(ctx, identifier)
	if err != nil {
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("identifiants invalides")
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) UpdateMe(ctx context.Context, userID int, firstname, lastname, username, email, phone string) error {
	cleanUsername, ok := utils.SanitizeUsername(username)
	if !ok {
		return errors.New("format de username invalide")
	}
	cleanPhone, ok := utils.SanitizePhone(phone)
	if !ok {
		return errors.New("numéro de téléphone invalide")
	}
	return s.userStore.UpdateMe(ctx, userID, firstname, lastname, cleanUsername, email, cleanPhone)
}

func (s *AuthService) ForgotPassword(ctx context.Context, identifier string) error {
	user, err := s.authStore.GetByEmailOrUsername(ctx, identifier)
	if err != nil {
		// On retourne toujours succès pour ne pas révéler si l'email existe
		return nil
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Errorf("génération token: %w", err)
	}
	token := hex.EncodeToString(raw)
	expiresAt := time.Now().Add(15 * time.Minute)

	fmt.Printf("[AUTH] ForgotPassword: user trouvé id=%d email=%s\n", user.ID, user.Email)
	if err := s.authStore.CreateResetToken(ctx, user.ID, token, expiresAt); err != nil {
		fmt.Printf("[AUTH] CreateResetToken erreur: %v\n", err)
		return fmt.Errorf("création token: %w", err)
	}
	fmt.Printf("[AUTH] Token créé, lancement goroutine envoi email\n")

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)
	fullName := user.Firstname + " " + user.Lastname
	go func() {
		fmt.Printf("[AUTH] Goroutine démarrée, envoi à %s\n", user.Email)
		if err := s.mailer.SendPasswordReset(user.Email, fullName, resetLink); err != nil {
			fmt.Printf("[AUTH] Erreur envoi email reset à %s: %v\n", user.Email, err)
		} else {
			fmt.Printf("[AUTH] Email reset envoyé avec succès à %s\n", user.Email)
		}
	}()
	return nil
}

func (s *AuthService) ValidateResetToken(ctx context.Context, token string) error {
	_, err := s.authStore.GetValidResetToken(ctx, token)
	return err
}

func (s *AuthService) ResetPasswordByToken(ctx context.Context, token, newPassword string) error {
	rt, err := s.authStore.GetValidResetToken(ctx, token)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.authStore.UpdatePassword(ctx, rt.UserID, string(hash)); err != nil {
		return err
	}
	return s.authStore.MarkTokenUsed(ctx, token)
}

func (s *AuthService) ResetPassword(ctx context.Context, requesterID int, targetUUID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userStore.GetByUUID(ctx, targetUUID)
	if err != nil {
		return errors.New("utilisateur non trouvé")
	}

	if requesterID != user.ID {
		return errors.New("interdit : vous ne pouvez modifier que votre propre mot de passe")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return err
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)

	return s.authStore.UpdatePassword(ctx, user.ID, string(newHash))
}
