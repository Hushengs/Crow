package biz

import (
	"context"
	"time"

	v1 "crow/api/admin/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	ErrAdminRoleNotFound        = errors.NotFound(v1.ErrorReason_ADMIN_ROLE_NOT_FOUND.String(), "admin role not found")
	ErrAdminRoleInvalidArgument = errors.BadRequest(v1.ErrorReason_ADMIN_ROLE_INVALID_ARGUMENT.String(), "invalid admin role argument")
	ErrAdminRoleConflict        = errors.Conflict(v1.ErrorReason_ADMIN_ROLE_CONFLICT.String(), "admin role relation already exists")
)

type AdminRole struct {
	ID         int64
	AdminID    int64
	RoleID     int64
	CreateTime time.Time
	UpdateTime time.Time
}

type AdminRoleRepo interface {
	FindByID(context.Context, int64) (*AdminRole, error)
	ListAdminRoles(context.Context, ...AdminRoleListOption) ([]*AdminRole, error)
	CreateAdminRole(context.Context, *AdminRole) (*AdminRole, error)
	UpdateAdminRole(context.Context, *AdminRole) (*AdminRole, error)
	DeleteAdminRole(context.Context, int64) error
}

type AdminRoleListOption func(*AdminRoleListOptions)

type AdminRoleListOptions struct {
	Offset int
	Limit  int
}

func AdminRoleListOffset(offset int) AdminRoleListOption {
	return func(o *AdminRoleListOptions) {
		o.Offset = offset
	}
}

func AdminRoleListLimit(limit int) AdminRoleListOption {
	return func(o *AdminRoleListOptions) {
		o.Limit = limit
	}
}

type AdminRoleUsecase struct {
	repo AdminRoleRepo
}

func NewAdminRoleUsecase(repo AdminRoleRepo) *AdminRoleUsecase {
	return &AdminRoleUsecase{repo: repo}
}

func (uc *AdminRoleUsecase) GetAdminRole(ctx context.Context, id int64) (*AdminRole, error) {
	if id <= 0 {
		return nil, ErrAdminRoleInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *AdminRoleUsecase) ListAdminRoles(ctx context.Context, opts ...AdminRoleListOption) ([]*AdminRole, error) {
	return uc.repo.ListAdminRoles(ctx, opts...)
}

func (uc *AdminRoleUsecase) CreateAdminRole(ctx context.Context, relation *AdminRole) (*AdminRole, error) {
	if err := validateAdminRole(relation); err != nil {
		return nil, err
	}
	return uc.repo.CreateAdminRole(ctx, relation)
}

func (uc *AdminRoleUsecase) UpdateAdminRole(ctx context.Context, relation *AdminRole) (*AdminRole, error) {
	if relation == nil || relation.ID <= 0 {
		return nil, ErrAdminRoleInvalidArgument
	}
	if err := validateAdminRole(relation); err != nil {
		return nil, err
	}
	return uc.repo.UpdateAdminRole(ctx, relation)
}

func (uc *AdminRoleUsecase) DeleteAdminRole(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrAdminRoleInvalidArgument
	}
	return uc.repo.DeleteAdminRole(ctx, id)
}

func validateAdminRole(relation *AdminRole) error {
	if relation == nil || relation.AdminID <= 0 || relation.RoleID <= 0 {
		return ErrAdminRoleInvalidArgument
	}
	return nil
}
