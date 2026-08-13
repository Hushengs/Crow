package biz

import (
	"context"
	"strings"
	"time"

	v1 "crow/api/admin/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	ErrPermissionNotFound        = errors.NotFound(v1.ErrorReason_PERMISSION_NOT_FOUND.String(), "permission not found")
	ErrPermissionInvalidArgument = errors.BadRequest(v1.ErrorReason_PERMISSION_INVALID_ARGUMENT.String(), "invalid permission argument")
	ErrPermissionHandleConflict  = errors.Conflict(v1.ErrorReason_PERMISSION_HANDLE_CONFLICT.String(), "permission handle already exists")
)

type Permission struct {
	ID         int64
	ParentID   int64
	Title      string
	Handle     string
	Weight     int32
	CreateTime time.Time
	UpdateTime time.Time
}

type PermissionRepo interface {
	FindByID(context.Context, int64) (*Permission, error)
	ListPermissions(context.Context, ...PermissionListOption) ([]*Permission, error)
	CreatePermission(context.Context, *Permission) (*Permission, error)
	UpdatePermission(context.Context, *Permission) (*Permission, error)
	DeletePermission(context.Context, int64) error
}

type PermissionListOption func(*PermissionListOptions)

type PermissionListOptions struct {
	Offset int
	Limit  int
}

func PermissionListOffset(offset int) PermissionListOption {
	return func(o *PermissionListOptions) {
		o.Offset = offset
	}
}

func PermissionListLimit(limit int) PermissionListOption {
	return func(o *PermissionListOptions) {
		o.Limit = limit
	}
}

type PermissionUsecase struct {
	repo PermissionRepo
}

func NewPermissionUsecase(repo PermissionRepo) *PermissionUsecase {
	return &PermissionUsecase{repo: repo}
}

func (uc *PermissionUsecase) GetPermission(ctx context.Context, id int64) (*Permission, error) {
	if id <= 0 {
		return nil, ErrPermissionInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *PermissionUsecase) ListPermissions(ctx context.Context, opts ...PermissionListOption) ([]*Permission, error) {
	return uc.repo.ListPermissions(ctx, opts...)
}

func (uc *PermissionUsecase) CreatePermission(ctx context.Context, permission *Permission) (*Permission, error) {
	if err := validatePermission(permission); err != nil {
		return nil, err
	}
	return uc.repo.CreatePermission(ctx, permission)
}

func (uc *PermissionUsecase) UpdatePermission(ctx context.Context, permission *Permission) (*Permission, error) {
	if permission == nil || permission.ID <= 0 {
		return nil, ErrPermissionInvalidArgument
	}
	if err := validatePermission(permission); err != nil {
		return nil, err
	}
	return uc.repo.UpdatePermission(ctx, permission)
}

func (uc *PermissionUsecase) DeletePermission(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrPermissionInvalidArgument
	}
	return uc.repo.DeletePermission(ctx, id)
}

func validatePermission(permission *Permission) error {
	if permission == nil {
		return ErrPermissionInvalidArgument
	}
	permission.Title = strings.TrimSpace(permission.Title)
	permission.Handle = strings.TrimSpace(permission.Handle)
	if permission.ParentID < 0 || permission.Title == "" || permission.Handle == "" || permission.Weight < 0 {
		return ErrPermissionInvalidArgument
	}
	return nil
}
