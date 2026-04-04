package services

import (
	"errors"
	"os"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/utils"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userStore store.UserStore
	authStore store.AuthStore
	secret    string
}

func NewUserService(s store.UserStore, a store.AuthStore) *UserService {
	return &UserService{userStore: s, authStore: a, secret: os.Getenv("JWT_SECRET")}
}

func (u *UserService) CreateUser(user *models.User, ownerID int, ownerRole string) (*models.User, error) {
	if ownerRole != "admin" && ownerRole != "super" {
		return nil, errors.New("accès refusé : privilèges insuffisants")
	}

	switch ownerRole {
	case "admin":
		user.Role = "vendeur"
		user.ParentID = &ownerID
	case "super":
		if user.Role == "" {
			user.Role = "admin"
		}
		user.ParentID = nil
	}

	cleanUsername, ok := utils.SanitizeUsername(user.Username)
	if !ok {
		return nil, errors.New("format de username invalide")
	}
	user.Username = cleanUsername

	cleanPhone, ok := utils.SanitizePhone(user.Phone)
	if !ok {
		return nil, errors.New("numéro de téléphone invalide")
	}
	user.Phone = cleanPhone

	user.UUID = uuid.New()
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hashed)

	if _, err := u.userStore.CreateUser(user); err != nil {
		return nil, err
	}

	userInDB, err := u.authStore.GetByEmailOrUsername(user.Email)
	if err != nil {
		return nil, err
	}

	return userInDB, nil
}

func (u *UserService) GetAllUser(ownerID int, ownerRole string, limit, offset int) ([]*models.User, int, error) {
	if ownerRole == "super" {
		return u.userStore.GetAll(limit, offset)
	}
	if ownerRole == "vendeur" {
		return nil, 0, errors.New("accès interdit")
	}
	return u.userStore.GetAllByOwner(ownerID, limit, offset)
}

func (u *UserService) GetUserByUUID(userUUID uuid.UUID, ownerID int, ownerRole string) (*models.User, error) {
	if ownerRole == "super" {
		return u.userStore.GetByUUID(userUUID)
	}

	return u.userStore.GetByUUIDWithParent(userUUID, ownerID)
}

func (u *UserService) UpdateUser(targetUUID uuid.UUID, input *models.User, ownerID int, ownerRole string) (*models.User, error) {
	cleanUsername, ok := utils.SanitizeUsername(input.Username)
	if !ok {
		return nil, errors.New("format de username invalide (lettres, chiffres, underscore, doit commencer par une lettre)")
	}
	input.Username = cleanUsername

	cleanPhone, ok := utils.SanitizePhone(input.Phone)
	if !ok {
		return nil, errors.New("numéro de téléphone invalide (doit être 10 chiffres commençant par 01, 05 ou 07)")
	}
	input.Phone = cleanPhone

	// fmt.Printf("Clean Username: %+v\n", cleanUsername)
	// fmt.Printf("Clean Phone: %+v\n", cleanPhone)

	existingUser, err := u.GetUserByUUID(targetUUID, ownerID, ownerRole)
	if err != nil {
		return nil, err
	}

	if input.Username != "" {
		existingUser.Username = input.Username
	}
	if input.Firstname != "" {
		existingUser.Firstname = input.Firstname
	}
	if input.Lastname != "" {
		existingUser.Lastname = input.Lastname
	}
	if input.Phone != "" {
		existingUser.Phone = input.Phone
	}

	if (ownerID == existingUser.ID || ownerRole == "super") && input.Email != "" {
		existingUser.Email = input.Email
	}

	if ownerRole == "super" {
		if input.Role != "" {
			existingUser.Role = input.Role
		}
		if input.ParentID != nil {
			existingUser.ParentID = input.ParentID
		}
	}

	updatedUser, err := u.userStore.Update(existingUser)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil

}

func (u *UserService) DeleteUser(userUUID uuid.UUID, ownerID int, ownerRole string) error {
	if ownerRole == "super" {
		return u.userStore.Delete(userUUID)
	}

	return u.userStore.DeleteByOwner(userUUID, ownerID)
}

func (u *UserService) RestoreUser(userUUID uuid.UUID, ownerID int, ownerRole string) error {
	if ownerRole == "super" {
		return u.userStore.Restore(userUUID)
	}

	return u.userStore.RestoreByOwner(userUUID, ownerID)
}
