package biz

import (
	"context"
	"strings"
	"time"

	v1 "crow/api/admin/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	ErrRoleNotFound         = errors.NotFound(v1.ErrorReason_ROLE_NOT_FOUND.String(), "role not found")
	ErrRoleInvalidArgument  = errors.BadRequest(v1.ErrorReason_ROLE_INVALID_ARGUMENT.String(), "invalid role argument")
	ErrRoleNameConflict     = errors.Conflict(v1.ErrorReason_ROLE_NAME_CONFLICT.String(), "role name already exists")
)

type Role struct {
	ID         int64
	RoleName   string
	CreateTime time.Time
	UpdateTime time.Time
}

type RoleRepo interface {
	FindByID(context.Context, int64) (*Role, error)
	ListRoles(context.Context, ...RoleListOption) ([]*Role, error)
	CreateRole(context.Context, *Role) (*Role, error)
	UpdateRole(context.Context, *Role) (*Role, error)
	DeleteRole(context.Context, int64) error
}

type RoleListOption func(*RoleListOptions)

type RoleListOptions struct {
	Offset int
	Limit  int
}

func RoleListOffset(offset int) RoleListOption {
	return func(o *RoleListOptions) {
		o.Offset = offset
	}
}

func RoleListLimit(limit int) RoleListOption {
	return func(o *RoleListOptions) {
		o.Limit = limit
	}
}

type RoleUsecase struct {
	repo RoleRepo
}

func NewRoleUsecase(repo RoleRepo) *RoleUsecase {
	return &RoleUsecase{repo: repo}
}

func (uc *RoleUsecase) GetRole(ctx context.Context, id int64) (*Role, error) {
	if id <= 0 {
		return nil, ErrRoleInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *RoleUsecase) ListRoles(ctx context.Context, opts ...RoleListOption) ([]*Role, error) {
	return uc.repo.ListRoles(ctx, opts...)
}

func (uc *RoleUsecase) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	if err := validateRole(role); err != nil {
		return nil, err
	}
	return uc.repo.CreateRole(ctx, role)
}

func (uc *RoleUsecase) UpdateRole(ctx context.Context, role *Role) (*Role, error) {
	if role == nil || role.ID <= 0 {
		return nil, ErrRoleInvalidArgument
	}
	if err := validateRole(role); err != nil {
		return nil, err
	}
	return uc.repo.UpdateRole(ctx, role)
}

func (uc *RoleUsecase) DeleteRole(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrRoleInvalidArgument
	}
	return uc.repo.DeleteRole(ctx, id)
}

func validateRole(role *Role) error {
	if role == nil {
		return ErrRoleInvalidArgument
	}
	role.RoleName = strings.TrimSpace(role.RoleName)
	if role.RoleName == "" {
		return ErrRoleInvalidArgument
	}
	return nil
}
