package biz

import (
	"context"
	"strings"
	"time"

	v1 "crow/api/admin/v1"

	"github.com/go-kratos/kratos/v3/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	AdminStatusDisabled uint32 = 0
	AdminStatusNormal   uint32 = 1
	AdminStatusLocked   uint32 = 2
)

var (
	ErrAdminNotFound         = errors.NotFound(v1.ErrorReason_ADMIN_NOT_FOUND.String(), "admin not found")
	ErrAdminInvalidArgument  = errors.BadRequest(v1.ErrorReason_ADMIN_INVALID_ARGUMENT.String(), "invalid admin argument")
	ErrAdminUsernameConflict = errors.Conflict(v1.ErrorReason_ADMIN_USERNAME_CONFLICT.String(), "admin username already exists")
)

// Admin is the domain model for a backend administrator.
type Admin struct {
	ID                int64
	Username          string
	Password          string
	PasswordHash      string
	RealName          string
	RoleID            int64
	Status            uint32
	LastLoginTime     *time.Time
	PasswordUpdatedAt *time.Time
	Remark            string
	CreateTime        time.Time
	UpdateTime        time.Time
}

// AdminRepo is the persistence boundary for administrators.
type AdminRepo interface {
	FindByID(context.Context, int64) (*Admin, error)
	ListAdmins(context.Context, ...AdminListOption) ([]*Admin, error)
	CreateAdmin(context.Context, *Admin) (*Admin, error)
	UpdateAdmin(context.Context, *Admin) (*Admin, error)
	DeleteAdmin(context.Context, int64) error
}

// AdminListOption configures admin list queries.
type AdminListOption func(*AdminListOptions)

// AdminListOptions are admin list query options.
type AdminListOptions struct {
	Offset int
	Limit  int
}

// AdminListOffset sets an offset for admin list queries.
func AdminListOffset(offset int) AdminListOption {
	return func(o *AdminListOptions) {
		o.Offset = offset
	}
}

// AdminListLimit sets a limit for admin list queries.
func AdminListLimit(limit int) AdminListOption {
	return func(o *AdminListOptions) {
		o.Limit = limit
	}
}

// AdminUsecase is the admin usecase.
type AdminUsecase struct {
	repo AdminRepo
}

// NewAdminUsecase creates a new AdminUsecase.
func NewAdminUsecase(repo AdminRepo) *AdminUsecase {
	return &AdminUsecase{repo: repo}
}

// GetAdmin returns an admin by ID.
func (uc *AdminUsecase) GetAdmin(ctx context.Context, id int64) (*Admin, error) {
	if id <= 0 {
		return nil, ErrAdminInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

// ListAdmins returns a page of admins.
func (uc *AdminUsecase) ListAdmins(ctx context.Context, opts ...AdminListOption) ([]*Admin, error) {
	return uc.repo.ListAdmins(ctx, opts...)
}

// CreateAdmin creates a new admin.
func (uc *AdminUsecase) CreateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	if err := normalizeAndValidateAdmin(admin, true); err != nil {
		return nil, err
	}
	hash, err := hashPassword(admin.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	admin.PasswordHash = hash
	admin.Password = ""
	admin.PasswordUpdatedAt = &now
	return uc.repo.CreateAdmin(ctx, admin)
}

// UpdateAdmin updates an existing admin.
func (uc *AdminUsecase) UpdateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	if admin == nil || admin.ID <= 0 {
		return nil, ErrAdminInvalidArgument
	}
	if err := normalizeAndValidateAdmin(admin, false); err != nil {
		return nil, err
	}
	if admin.Password != "" {
		hash, err := hashPassword(admin.Password)
		if err != nil {
			return nil, err
		}
		now := time.Now()
		admin.PasswordHash = hash
		admin.PasswordUpdatedAt = &now
	}
	admin.Password = ""
	return uc.repo.UpdateAdmin(ctx, admin)
}

// DeleteAdmin deletes an admin by ID.
func (uc *AdminUsecase) DeleteAdmin(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrAdminInvalidArgument
	}
	return uc.repo.DeleteAdmin(ctx, id)
}

func normalizeAndValidateAdmin(admin *Admin, requirePassword bool) error {
	if admin == nil {
		return ErrAdminInvalidArgument
	}
	admin.Username = strings.TrimSpace(admin.Username)
	admin.RealName = strings.TrimSpace(admin.RealName)
	admin.Remark = strings.TrimSpace(admin.Remark)
	admin.Password = strings.TrimSpace(admin.Password)

	if admin.Username == "" || admin.RoleID < 0 || !isValidAdminStatus(admin.Status) {
		return ErrAdminInvalidArgument
	}
	if requirePassword && admin.Password == "" {
		return ErrAdminInvalidArgument
	}
	if admin.Password != "" && len(admin.Password) < 6 {
		return ErrAdminInvalidArgument
	}
	if !requirePassword && admin.Password == "" && admin.PasswordHash == "" {
		return ErrAdminInvalidArgument
	}
	return nil
}

func isValidAdminStatus(status uint32) bool {
	switch status {
	case AdminStatusDisabled, AdminStatusNormal, AdminStatusLocked:
		return true
	default:
		return false
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func cloneAdmin(admin *Admin) *Admin {
	if admin == nil {
		return nil
	}
	cloned := *admin
	if admin.LastLoginTime != nil {
		t := *admin.LastLoginTime
		cloned.LastLoginTime = &t
	}
	if admin.PasswordUpdatedAt != nil {
		t := *admin.PasswordUpdatedAt
		cloned.PasswordUpdatedAt = &t
	}
	return &cloned
}
