package services

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/pkg/utils"
	"github.com/huguescodeur/zoro_rest_api_go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authStore store.AuthStore
	userStore store.UserStore
	secret    string
}

func NewAuthService(s store.AuthStore, u store.UserStore) *AuthService {
	return &AuthService{authStore: s, userStore: u, secret: os.Getenv("JWT_SECRET")}
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

func (s *AuthService) Register(u *models.User) (*models.User, string, error) {
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

	if _, err := s.authStore.CreateUser(u); err != nil {
		return nil, "", err
	}

	userInDB, err := s.authStore.GetByEmailOrUsername(u.Email)
	if err != nil {
		return nil, "", err
	}

	token, err := s.generateJWT(userInDB)
	if err != nil {
		return nil, "", err
	}

	return userInDB, token, nil
}

func (s *AuthService) Login(identifier, password string) (*models.User, string, error) {
	user, err := s.authStore.GetByEmailOrUsername(identifier)
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

func (s *AuthService) ResetPassword(requesterID int, targetUUID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userStore.GetByUUID(targetUUID)
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

	return s.authStore.UpdatePassword(user.ID, string(newHash))
}
