package biz

import (
	"context"
	"time"

	v1 "crow/api/admin/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	ErrGroupPermissionNotFound        = errors.NotFound(v1.ErrorReason_GROUP_PERMISSION_NOT_FOUND.String(), "group permission not found")
	ErrGroupPermissionInvalidArgument = errors.BadRequest(v1.ErrorReason_GROUP_PERMISSION_INVALID_ARGUMENT.String(), "invalid group permission argument")
	ErrGroupPermissionConflict        = errors.Conflict(v1.ErrorReason_GROUP_PERMISSION_CONFLICT.String(), "group permission relation already exists")
)

type GroupPermission struct {
	ID           int64
	GroupID      int64
	PermissionID int64
	CreateTime   time.Time
	UpdateTime   time.Time
}

type GroupPermissionRepo interface {
	FindByID(context.Context, int64) (*GroupPermission, error)
	ListGroupPermissions(context.Context, ...GroupPermissionListOption) ([]*GroupPermission, error)
	CreateGroupPermission(context.Context, *GroupPermission) (*GroupPermission, error)
	UpdateGroupPermission(context.Context, *GroupPermission) (*GroupPermission, error)
	DeleteGroupPermission(context.Context, int64) error
}

type GroupPermissionListOption func(*GroupPermissionListOptions)

type GroupPermissionListOptions struct {
	Offset int
	Limit  int
}

func GroupPermissionListOffset(offset int) GroupPermissionListOption {
	return func(o *GroupPermissionListOptions) {
		o.Offset = offset
	}
}

func GroupPermissionListLimit(limit int) GroupPermissionListOption {
	return func(o *GroupPermissionListOptions) {
		o.Limit = limit
	}
}

type GroupPermissionUsecase struct {
	repo GroupPermissionRepo
}

func NewGroupPermissionUsecase(repo GroupPermissionRepo) *GroupPermissionUsecase {
	return &GroupPermissionUsecase{repo: repo}
}

func (uc *GroupPermissionUsecase) GetGroupPermission(ctx context.Context, id int64) (*GroupPermission, error) {
	if id <= 0 {
		return nil, ErrGroupPermissionInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *GroupPermissionUsecase) ListGroupPermissions(ctx context.Context, opts ...GroupPermissionListOption) ([]*GroupPermission, error) {
	return uc.repo.ListGroupPermissions(ctx, opts...)
}

func (uc *GroupPermissionUsecase) CreateGroupPermission(ctx context.Context, relation *GroupPermission) (*GroupPermission, error) {
	if err := validateGroupPermission(relation); err != nil {
		return nil, err
	}
	return uc.repo.CreateGroupPermission(ctx, relation)
}

func (uc *GroupPermissionUsecase) UpdateGroupPermission(ctx context.Context, relation *GroupPermission) (*GroupPermission, error) {
	if relation == nil || relation.ID <= 0 {
		return nil, ErrGroupPermissionInvalidArgument
	}
	if err := validateGroupPermission(relation); err != nil {
		return nil, err
	}
	return uc.repo.UpdateGroupPermission(ctx, relation)
}

func (uc *GroupPermissionUsecase) DeleteGroupPermission(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrGroupPermissionInvalidArgument
	}
	return uc.repo.DeleteGroupPermission(ctx, id)
}

func validateGroupPermission(relation *GroupPermission) error {
	if relation == nil || relation.GroupID <= 0 || relation.PermissionID <= 0 {
		return ErrGroupPermissionInvalidArgument
	}
	return nil
}
