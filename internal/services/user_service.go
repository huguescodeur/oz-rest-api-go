package services

import (
	"context"
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

func (u *UserService) CreateUser(ctx context.Context, user *models.User, ownerID int, ownerRole string) (*models.User, error) {
	if ownerRole != "admin" && ownerRole != "super" {
		return nil, errors.New("accès refusé : privilèges insuffisants")
	}

	shopID := user.ShopID

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

	created, err := u.userStore.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	if shopID != nil && *shopID > 0 {
		if err := u.userStore.AssignShop(ctx, created.ID, *shopID); err != nil {
			return nil, err
		}
	}

	return u.authStore.GetByEmailOrUsername(ctx, user.Email)
}

func (u *UserService) GetAllUser(ctx context.Context, ownerID int, ownerRole string, limit, offset int, showArchived bool) ([]*models.User, int, error) {
	if ownerRole == "super" {
		return u.userStore.GetAll(ctx, limit, offset)
	}
	if ownerRole == "vendeur" {
		return nil, 0, errors.New("accès interdit")
	}
	if showArchived {
		return u.userStore.GetArchivedByOwner(ctx, ownerID, limit, offset)
	}
	return u.userStore.GetAllByOwner(ctx, ownerID, limit, offset)
}

func (u *UserService) GetUserByUUID(ctx context.Context, userUUID uuid.UUID, ownerID int, ownerRole string) (*models.User, error) {
	if ownerRole == "super" {
		return u.userStore.GetByUUID(ctx, userUUID)
	}
	return u.userStore.GetByUUIDWithParent(ctx, userUUID, ownerID)
}

func (u *UserService) UpdateUser(ctx context.Context, targetUUID uuid.UUID, input *models.User, ownerID int, ownerRole string) (*models.User, error) {
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

	existingUser, err := u.GetUserByUUID(ctx, targetUUID, ownerID, ownerRole)
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

	result, err := u.userStore.Update(ctx, existingUser)
	if err != nil {
		return nil, err
	}

	if input.ShopID != nil {
		if err := u.userStore.AssignShop(ctx, existingUser.ID, *input.ShopID); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (u *UserService) DeleteUser(ctx context.Context, userUUID uuid.UUID, ownerID int, ownerRole string) error {
	if ownerRole == "super" {
		return u.userStore.Delete(ctx, userUUID)
	}
	return u.userStore.DeleteByOwner(ctx, userUUID, ownerID)
}

func (u *UserService) GetMyShops(ctx context.Context, userID int) ([]*models.Shop, error) {
	return u.userStore.GetMyShops(ctx, userID)
}

func (u *UserService) AssignShops(ctx context.Context, targetUUID uuid.UUID, ownerID int, shopIDs []int) error {
	target, err := u.userStore.GetByUUIDWithParent(ctx, targetUUID, ownerID)
	if err != nil {
		return err
	}
	return u.userStore.AssignShops(ctx, target.ID, shopIDs)
}

func (u *UserService) GetUserShops(ctx context.Context, targetUUID uuid.UUID, ownerID int) ([]*models.Shop, error) {
	target, err := u.userStore.GetByUUIDWithParent(ctx, targetUUID, ownerID)
	if err != nil {
		return nil, err
	}
	return u.userStore.GetMyShops(ctx, target.ID)
}

func (u *UserService) UpdateMe(ctx context.Context, userID int, firstname, lastname, username, email, phone string) error {
	cleanUsername, ok := utils.SanitizeUsername(username)
	if !ok {
		return errors.New("format de username invalide")
	}
	cleanPhone, ok := utils.SanitizePhone(phone)
	if !ok {
		return errors.New("numéro de téléphone invalide")
	}
	return u.userStore.UpdateMe(ctx, userID, firstname, lastname, cleanUsername, email, cleanPhone)
}

func (u *UserService) RestoreUser(ctx context.Context, userUUID uuid.UUID, ownerID int, ownerRole string) error {
	if ownerRole == "super" {
		return u.userStore.Restore(ctx, userUUID)
	}
	return u.userStore.RestoreByOwner(ctx, userUUID, ownerID)
}
