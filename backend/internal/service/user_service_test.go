package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	roleOwner   = &domain.Role{ID: 1, Code: domain.RoleOwner}
	roleAdmin   = &domain.Role{ID: 2, Code: domain.RoleAdmin}
	roleCashier = &domain.Role{ID: 3, Code: domain.RoleCashier}
	admin       = Actor{UserID: 2, Role: domain.RoleAdmin}
)

func newUserSvc() (*UserService, *mockUserRepo, *mockRoleRepo, *mockTokenRepo) {
	users, roles, tokens := &mockUserRepo{}, &mockRoleRepo{}, &mockTokenRepo{}
	roles.On("FindByID", mock.Anything, int64(1)).Return(roleOwner, nil).Maybe()
	roles.On("FindByID", mock.Anything, int64(2)).Return(roleAdmin, nil).Maybe()
	roles.On("FindByID", mock.Anything, int64(3)).Return(roleCashier, nil).Maybe()
	roles.On("FindByID", mock.Anything, int64(9)).Return(nil, domain.ErrNotFound).Maybe()
	return NewUserService(users, roles, tokens, &fakeTx{}, fakeHasher{}, fixedClock(time.Now())), users, roles, tokens
}

func TestCreateUser_NormalisesEmailAndHashesPassword(t *testing.T) {
	svc, users, _, _ := newUserSvc()
	users.On("EmailExists", mock.Anything, "kasir2@toko.id", int64(0)).Return(false, nil)
	users.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "kasir2@toko.id" && u.PasswordHash == "hash:rahasia123" && u.IsActive && u.RoleID == 3
	})).Run(func(a mock.Arguments) { a.Get(1).(*domain.User).ID = 10 }).Return(nil)
	users.On("FindByID", mock.Anything, int64(10)).Return(&domain.User{ID: 10}, nil)

	_, err := svc.Create(context.Background(), admin, dto.CreateUserRequest{
		Name: "Kasir Dua", Email: " Kasir2@Toko.id ", Password: "rahasia123", RoleID: 3})
	require.NoError(t, err)
	users.AssertExpectations(t)
}

func TestCreateUser_AdminCannotCreateOwner(t *testing.T) {
	svc, users, _, _ := newUserSvc()
	_, err := svc.Create(context.Background(), admin, dto.CreateUserRequest{Name: "X", Email: "x@x.id", Password: "12345678", RoleID: 1})
	assertStatus(t, err, http.StatusForbidden)
	users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateUser_DuplicateEmailAndUnknownRole(t *testing.T) {
	svc, users, _, _ := newUserSvc()
	users.On("EmailExists", mock.Anything, "dup@toko.id", int64(0)).Return(true, nil)
	_, err := svc.Create(context.Background(), owner, dto.CreateUserRequest{Name: "X", Email: "dup@toko.id", Password: "12345678", RoleID: 3})
	ae := assertStatus(t, err, http.StatusUnprocessableEntity)
	assert.Equal(t, "email", ae.Errors[0].Field)

	_, err = svc.Create(context.Background(), owner, dto.CreateUserRequest{Name: "X", Email: "a@toko.id", Password: "12345678", RoleID: 9})
	ae = assertStatus(t, err, http.StatusUnprocessableEntity)
	assert.Equal(t, "role_id", ae.Errors[0].Field)
}

func TestUpdateUser_AdminCannotEditOwner(t *testing.T) {
	svc, users, _, _ := newUserSvc()
	users.On("FindByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, RoleCode: domain.RoleOwner}, nil)
	_, err := svc.Update(context.Background(), admin, 1, dto.UpdateUserRequest{Name: "X", Email: "x@x.id", RoleID: 1, IsActive: true})
	assertStatus(t, err, http.StatusForbidden)
}

func TestUpdateUser_CannotDeactivateSelf(t *testing.T) {
	svc, users, _, _ := newUserSvc()
	users.On("FindByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, RoleCode: domain.RoleOwner, RoleID: 1}, nil)
	_, err := svc.Update(context.Background(), owner, 1, dto.UpdateUserRequest{Name: "X", Email: "x@x.id", RoleID: 1, IsActive: false})
	assertStatus(t, err, http.StatusBadRequest)
}

func TestUpdateUser_KeepsAtLeastOneOwner(t *testing.T) {
	svc, users, _, _ := newUserSvc()
	users.On("FindByID", mock.Anything, int64(5)).Return(&domain.User{ID: 5, RoleCode: domain.RoleOwner, RoleID: 1}, nil)
	users.On("CountActiveByRole", mock.Anything, domain.RoleOwner).Return(int64(1), nil)
	_, err := svc.Update(context.Background(), owner, 5, dto.UpdateUserRequest{Name: "X", Email: "x@x.id", RoleID: 3, IsActive: true})
	assertStatus(t, err, http.StatusBadRequest)
}

func TestUpdateUser_PasswordResetRevokesSessions(t *testing.T) {
	svc, users, _, tokens := newUserSvc()
	users.On("FindByID", mock.Anything, int64(7)).Return(&domain.User{ID: 7, RoleCode: domain.RoleCashier, RoleID: 3}, nil)
	users.On("EmailExists", mock.Anything, "k@toko.id", int64(7)).Return(false, nil)
	users.On("Update", mock.Anything, mock.Anything).Return(nil)
	users.On("UpdatePassword", mock.Anything, int64(7), "hash:barubaru1").Return(nil)
	tokens.On("RevokeAllByUser", mock.Anything, int64(7), mock.Anything).Return(nil)

	_, err := svc.Update(context.Background(), admin, 7, dto.UpdateUserRequest{
		Name: "Kasir", Email: "k@toko.id", RoleID: 3, IsActive: true, Password: "barubaru1"})
	require.NoError(t, err)
	tokens.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	t.Run("cannot delete self", func(t *testing.T) {
		svc, _, _, _ := newUserSvc()
		assertStatus(t, svc.Delete(context.Background(), owner, owner.UserID), http.StatusBadRequest)
	})
	t.Run("soft deletes and revokes", func(t *testing.T) {
		svc, users, _, tokens := newUserSvc()
		users.On("FindByID", mock.Anything, int64(7)).Return(&domain.User{ID: 7, RoleCode: domain.RoleCashier}, nil)
		users.On("SoftDelete", mock.Anything, int64(7)).Return(nil)
		tokens.On("RevokeAllByUser", mock.Anything, int64(7), mock.Anything).Return(nil)
		require.NoError(t, svc.Delete(context.Background(), owner, 7))
		users.AssertExpectations(t)
	})
}
