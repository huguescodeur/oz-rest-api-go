package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ── Mock AuthStore ──────────────────────────────────────────────────────────

type mockAuthStore struct{ mock.Mock }

func (m *mockAuthStore) CreateUser(ctx context.Context, u *models.User) (*models.User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockAuthStore) GetByEmailOrUsername(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockAuthStore) UpdatePassword(ctx context.Context, userID int, hash string) error {
	return m.Called(ctx, userID, hash).Error(0)
}
func (m *mockAuthStore) CreateResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	return m.Called(ctx, userID, token, expiresAt).Error(0)
}
func (m *mockAuthStore) GetValidResetToken(ctx context.Context, token string) (*models.ResetToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ResetToken), args.Error(1)
}
func (m *mockAuthStore) MarkTokenUsed(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

// ── Mock UserStore ──────────────────────────────────────────────────────────

type mockUserStore struct{ mock.Mock }

func (m *mockUserStore) CreateUser(ctx context.Context, u *models.User) (*models.User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserStore) GetAll(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}
func (m *mockUserStore) GetAllByOwner(ctx context.Context, ownerID, limit, offset int) ([]*models.User, int, error) {
	args := m.Called(ctx, ownerID, limit, offset)
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}
func (m *mockUserStore) GetArchivedByOwner(ctx context.Context, ownerID, limit, offset int) ([]*models.User, int, error) {
	args := m.Called(ctx, ownerID, limit, offset)
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}
func (m *mockUserStore) GetByUUID(ctx context.Context, userUUID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserStore) GetByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserStore) Update(ctx context.Context, u *models.User) (*models.User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserStore) GetByUUIDWithParent(ctx context.Context, userUUID uuid.UUID, ownerID int) (*models.User, error) {
	args := m.Called(ctx, userUUID, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserStore) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserStore) Restore(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserStore) DeleteByOwner(ctx context.Context, vendeurID uuid.UUID, ownerID int) error {
	return m.Called(ctx, vendeurID, ownerID).Error(0)
}
func (m *mockUserStore) RestoreByOwner(ctx context.Context, id uuid.UUID, ownerID int) error {
	return m.Called(ctx, id, ownerID).Error(0)
}
func (m *mockUserStore) AssignShop(ctx context.Context, userID, shopID int) error {
	return m.Called(ctx, userID, shopID).Error(0)
}
func (m *mockUserStore) AssignShops(ctx context.Context, userID int, shopIDs []int) error {
	return m.Called(ctx, userID, shopIDs).Error(0)
}
func (m *mockUserStore) UpdateMe(ctx context.Context, userID int, firstname, lastname, username, email, phone string) error {
	return m.Called(ctx, userID, firstname, lastname, username, email, phone).Error(0)
}
func (m *mockUserStore) GetMyShops(ctx context.Context, userID int) ([]*models.Shop, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Shop), args.Error(1)
}

// ── Tests ───────────────────────────────────────────────────────────────────

func newAuthService(a *mockAuthStore, u *mockUserStore) *services.AuthService {
	return services.NewAuthService(a, u)
}

func makeHash(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}

func TestLogin_Success(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	u := &models.User{
		ID:           1,
		UUID:         uuid.New(),
		Email:        "test@oz.com",
		PasswordHash: makeHash(t, "secret123"),
		Role:         "admin",
	}
	as.On("GetByEmailOrUsername", mock.Anything, "test@oz.com").Return(u, nil)

	svc := newAuthService(as, us)
	_, _, err := svc.Login(context.Background(), "test@oz.com", "secret123")
	assert.NoError(t, err)
	as.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	u := &models.User{ID: 1, UUID: uuid.New(), PasswordHash: makeHash(t, "secret123"), Role: "admin"}
	as.On("GetByEmailOrUsername", mock.Anything, "test@oz.com").Return(u, nil)

	svc := newAuthService(as, us)
	_, _, err := svc.Login(context.Background(), "test@oz.com", "wrongpassword")
	assert.Error(t, err)
}

func TestLogin_UserNotFound(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	as.On("GetByEmailOrUsername", mock.Anything, "ghost@oz.com").Return(nil, errors.New("not found"))

	svc := newAuthService(as, us)
	_, _, err := svc.Login(context.Background(), "ghost@oz.com", "anypassword")
	assert.Error(t, err)
}

func TestForgotPassword_UnknownIdentifier_ReturnsNil(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	// Store returns error (user doesn't exist) → service must still return nil
	as.On("GetByEmailOrUsername", mock.Anything, "nobody@oz.com").Return(nil, errors.New("not found"))

	svc := newAuthService(as, us)
	err := svc.ForgotPassword(context.Background(), "nobody@oz.com")
	assert.NoError(t, err, "anti-enumeration: must return nil even when user is not found")
}

func TestValidateResetToken_Valid(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	rt := &models.ResetToken{ID: 1, UserID: 42, Token: "abc", ExpiresAt: time.Now().Add(10 * time.Minute)}
	as.On("GetValidResetToken", mock.Anything, "abc").Return(rt, nil)

	svc := newAuthService(as, us)
	err := svc.ValidateResetToken(context.Background(), "abc")
	assert.NoError(t, err)
}

func TestValidateResetToken_Invalid(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	as.On("GetValidResetToken", mock.Anything, "expired-token").Return(nil, errors.New("lien invalide ou expiré"))

	svc := newAuthService(as, us)
	err := svc.ValidateResetToken(context.Background(), "expired-token")
	assert.Error(t, err)
}

func TestResetPasswordByToken_Success(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	rt := &models.ResetToken{ID: 1, UserID: 42, Token: "valid-token", ExpiresAt: time.Now().Add(10 * time.Minute)}
	as.On("GetValidResetToken", mock.Anything, "valid-token").Return(rt, nil)
	as.On("UpdatePassword", mock.Anything, 42, mock.AnythingOfType("string")).Return(nil)
	as.On("MarkTokenUsed", mock.Anything, "valid-token").Return(nil)

	svc := newAuthService(as, us)
	err := svc.ResetPasswordByToken(context.Background(), "valid-token", "newpass123")
	assert.NoError(t, err)
	as.AssertExpectations(t)
}

func TestResetPasswordByToken_InvalidToken(t *testing.T) {
	as := &mockAuthStore{}
	us := &mockUserStore{}

	as.On("GetValidResetToken", mock.Anything, "bad-token").Return(nil, errors.New("lien invalide ou expiré"))

	svc := newAuthService(as, us)
	err := svc.ResetPasswordByToken(context.Background(), "bad-token", "newpass123")
	assert.Error(t, err)
}
