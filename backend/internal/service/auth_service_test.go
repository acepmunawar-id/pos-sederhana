package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var authNow = time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)

func newAuth() (*AuthService, *mockUserRepo, *mockTokenRepo) {
	users, tokens := &mockUserRepo{}, &mockTokenRepo{}
	svc := NewAuthService(users, tokens, &fakeTx{}, fakeHasher{}, fakeIssuer{}, 7*24*time.Hour, fixedClock(authNow))
	return svc, users, tokens
}

func activeUser() *domain.User {
	return &domain.User{ID: 1, Name: "Budi", Email: "owner@toko.id", RoleCode: domain.RoleOwner,
		PasswordHash: "hash:owner123", IsActive: true}
}

func TestLogin_Success(t *testing.T) {
	svc, users, tokens := newAuth()
	users.On("FindByEmail", mock.Anything, "owner@toko.id").Return(activeUser(), nil)
	users.On("UpdateLastLogin", mock.Anything, int64(1), authNow).Return(nil)
	var stored *domain.RefreshToken
	tokens.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		stored = a.Get(1).(*domain.RefreshToken)
	}).Return(nil)

	res, err := svc.Login(context.Background(), dto.LoginRequest{Email: " owner@toko.id ", Password: "owner123"}, ClientInfo{})
	require.NoError(t, err)
	assert.Equal(t, "access-token", res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	assert.Equal(t, hash.SHA256(res.RefreshToken), stored.TokenHash, "only the hash is persisted")
	assert.Equal(t, authNow.Add(7*24*time.Hour), stored.ExpiresAt)
	assert.Equal(t, authNow, *res.User.LastLoginAt)
}

func TestLogin_WrongPasswordAndUnknownEmailLookTheSame(t *testing.T) {
	svc, users, _ := newAuth()
	users.On("FindByEmail", mock.Anything, "owner@toko.id").Return(activeUser(), nil)
	users.On("FindByEmail", mock.Anything, "nobody@toko.id").Return(nil, domain.ErrNotFound)

	_, err1 := svc.Login(context.Background(), dto.LoginRequest{Email: "owner@toko.id", Password: "wrong"}, ClientInfo{})
	_, err2 := svc.Login(context.Background(), dto.LoginRequest{Email: "nobody@toko.id", Password: "x"}, ClientInfo{})
	a1, a2 := assertStatus(t, err1, http.StatusUnauthorized), assertStatus(t, err2, http.StatusUnauthorized)
	assert.Equal(t, a1.Message, a2.Message)
}

func TestLogin_InactiveUser(t *testing.T) {
	svc, users, _ := newAuth()
	u := activeUser()
	u.IsActive = false
	users.On("FindByEmail", mock.Anything, "owner@toko.id").Return(u, nil)
	_, err := svc.Login(context.Background(), dto.LoginRequest{Email: "owner@toko.id", Password: "owner123"}, ClientInfo{})
	assertStatus(t, err, http.StatusForbidden)
}

func TestRefresh_RotatesToken(t *testing.T) {
	svc, users, tokens := newAuth()
	tokens.On("FindByHash", mock.Anything, hash.SHA256("old")).
		Return(&domain.RefreshToken{ID: 5, UserID: 1, ExpiresAt: authNow.Add(time.Hour)}, nil)
	users.On("FindByID", mock.Anything, int64(1)).Return(activeUser(), nil)
	tokens.On("Revoke", mock.Anything, int64(5), authNow).Return(nil)
	tokens.On("Create", mock.Anything, mock.Anything).Return(nil)

	res, err := svc.Refresh(context.Background(), "old", ClientInfo{})
	require.NoError(t, err)
	assert.NotEqual(t, "old", res.RefreshToken)
	tokens.AssertCalled(t, "Revoke", mock.Anything, int64(5), authNow)
}

func TestRefresh_ReusedTokenRevokesAllSessions(t *testing.T) {
	svc, _, tokens := newAuth()
	revoked := authNow.Add(-time.Minute)
	tokens.On("FindByHash", mock.Anything, hash.SHA256("stolen")).
		Return(&domain.RefreshToken{ID: 5, UserID: 1, ExpiresAt: authNow.Add(time.Hour), RevokedAt: &revoked}, nil)
	tokens.On("RevokeAllByUser", mock.Anything, int64(1), authNow).Return(nil)

	_, err := svc.Refresh(context.Background(), "stolen", ClientInfo{})
	assertStatus(t, err, http.StatusUnauthorized)
	tokens.AssertExpectations(t)
}

func TestRefresh_Expired(t *testing.T) {
	svc, _, tokens := newAuth()
	tokens.On("FindByHash", mock.Anything, hash.SHA256("exp")).
		Return(&domain.RefreshToken{ID: 5, UserID: 1, ExpiresAt: authNow}, nil)
	_, err := svc.Refresh(context.Background(), "exp", ClientInfo{})
	assertStatus(t, err, http.StatusUnauthorized)
}

func TestChangePassword(t *testing.T) {
	t.Run("wrong old password", func(t *testing.T) {
		svc, users, _ := newAuth()
		users.On("FindByID", mock.Anything, int64(1)).Return(activeUser(), nil)
		err := svc.ChangePassword(context.Background(), 1, dto.ChangePasswordRequest{OldPassword: "nope", NewPassword: "newpass123"})
		ae := assertStatus(t, err, http.StatusUnprocessableEntity)
		assert.Equal(t, "old_password", ae.Errors[0].Field)
		users.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("success revokes sessions", func(t *testing.T) {
		svc, users, tokens := newAuth()
		users.On("FindByID", mock.Anything, int64(1)).Return(activeUser(), nil)
		users.On("UpdatePassword", mock.Anything, int64(1), "hash:newpass123").Return(nil)
		tokens.On("RevokeAllByUser", mock.Anything, int64(1), authNow).Return(nil)
		require.NoError(t, svc.ChangePassword(context.Background(), 1,
			dto.ChangePasswordRequest{OldPassword: "owner123", NewPassword: "newpass123"}))
		tokens.AssertExpectations(t)
	})
}

func TestLogout_IgnoresOtherUsersToken(t *testing.T) {
	svc, _, tokens := newAuth()
	tokens.On("FindByHash", mock.Anything, hash.SHA256("t")).Return(&domain.RefreshToken{ID: 5, UserID: 2}, nil)
	require.NoError(t, svc.Logout(context.Background(), 1, "t"))
	tokens.AssertNotCalled(t, "Revoke", mock.Anything, mock.Anything, mock.Anything)
}
