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

type programPermissionSpec struct {
	Key       string
	ParentKey string
	Title     string
	Handle    string
	Weight    int32
}

var programPermissionSpecs = []programPermissionSpec{
	{Key: "system", Title: "系统管理", Handle: "module://system", Weight: 10},
	{Key: "admins", ParentKey: "system", Title: "管理员管理", Handle: "module://system/admins", Weight: 10},
	{Key: "admins:list", ParentKey: "admins", Title: "管理员列表", Handle: "/v1/admins", Weight: 10},
	{Key: "admins:create", ParentKey: "admins", Title: "管理员新增", Handle: "/v1/admins/create", Weight: 20},
	{Key: "admins:get", ParentKey: "admins", Title: "管理员详情", Handle: "/v1/admins/{id}", Weight: 30},
	{Key: "admins:update", ParentKey: "admins", Title: "管理员编辑", Handle: "/v1/admins/update", Weight: 40},
	{Key: "admins:delete", ParentKey: "admins", Title: "管理员删除", Handle: "DELETE:/v1/admins/{id}", Weight: 50},
	{Key: "roles", ParentKey: "system", Title: "角色管理", Handle: "module://system/roles", Weight: 20},
	{Key: "roles:list", ParentKey: "roles", Title: "角色列表", Handle: "/v1/roles", Weight: 10},
	{Key: "roles:create", ParentKey: "roles", Title: "角色新增", Handle: "/v1/roles/create", Weight: 20},
	{Key: "roles:get", ParentKey: "roles", Title: "角色详情", Handle: "/v1/roles/{id}", Weight: 30},
	{Key: "roles:update", ParentKey: "roles", Title: "角色编辑", Handle: "/v1/roles/update", Weight: 40},
	{Key: "roles:delete", ParentKey: "roles", Title: "角色删除", Handle: "DELETE:/v1/roles/{id}", Weight: 50},
	{Key: "permissions", ParentKey: "system", Title: "权限查看", Handle: "module://system/permissions", Weight: 30},
	{Key: "permissions:list", ParentKey: "permissions", Title: "权限列表", Handle: "/v1/permissions", Weight: 10},
	{Key: "permissions:get", ParentKey: "permissions", Title: "权限详情", Handle: "/v1/permissions/{id}", Weight: 20},
	{Key: "operation_logs", ParentKey: "system", Title: "操作日志", Handle: "module://system/admin-operation-logs", Weight: 40},
	{Key: "operation_logs:list", ParentKey: "operation_logs", Title: "操作日志列表", Handle: "/v1/admin-operation-logs", Weight: 10},
	{Key: "system_logs", ParentKey: "system", Title: "系统日志", Handle: "module://system/system-logs", Weight: 50},
	{Key: "system_logs:list", ParentKey: "system_logs", Title: "系统日志列表", Handle: "/v1/system-logs", Weight: 10},
	{Key: "cdn", Title: "内容管理", Handle: "module://cdn", Weight: 20},
	{Key: "cps", ParentKey: "cdn", Title: "内容提供商", Handle: "module://cdn/cps", Weight: 10},
	{Key: "cps:list", ParentKey: "cps", Title: "内容提供商列表", Handle: "/v1/cps", Weight: 10},
	{Key: "cps:create", ParentKey: "cps", Title: "内容提供商新增", Handle: "/v1/cps/create", Weight: 20},
	{Key: "cps:get", ParentKey: "cps", Title: "内容提供商详情", Handle: "/v1/cps/{id}", Weight: 30},
	{Key: "cps:update", ParentKey: "cps", Title: "内容提供商编辑", Handle: "/v1/cps/update", Weight: 40},
	{Key: "cps:delete", ParentKey: "cps", Title: "内容提供商删除", Handle: "DELETE:/v1/cps/{id}", Weight: 50},
	{Key: "sps", ParentKey: "cdn", Title: "内容服务商", Handle: "module://cdn/sps", Weight: 20},
	{Key: "sps:list", ParentKey: "sps", Title: "内容服务商列表", Handle: "/v1/sps", Weight: 10},
	{Key: "sps:create", ParentKey: "sps", Title: "内容服务商新增", Handle: "/v1/sps/create", Weight: 20},
	{Key: "sps:get", ParentKey: "sps", Title: "内容服务商详情", Handle: "/v1/sps/{id}", Weight: 30},
	{Key: "sps:update", ParentKey: "sps", Title: "内容服务商编辑", Handle: "/v1/sps/update", Weight: 40},
	{Key: "sps:delete", ParentKey: "sps", Title: "内容服务商删除", Handle: "DELETE:/v1/sps/{id}", Weight: 50},
	{Key: "cp_sps", ParentKey: "cps", Title: "CP-SP 绑定", Handle: "module://cdn/cp-sps", Weight: 60},
	{Key: "cp_sps:list", ParentKey: "cp_sps", Title: "CP-SP 绑定列表", Handle: "/v1/cp-sps", Weight: 10},
	{Key: "cp_sps:create", ParentKey: "cp_sps", Title: "CP-SP 绑定新增", Handle: "/v1/cp-sps/create", Weight: 20},
	{Key: "cp_sps:get", ParentKey: "cp_sps", Title: "CP-SP 绑定详情", Handle: "/v1/cp-sps/{id}", Weight: 30},
	{Key: "cp_sps:update", ParentKey: "cp_sps", Title: "CP-SP 绑定编辑", Handle: "/v1/cp-sps/update", Weight: 40},
	{Key: "cp_sps:delete", ParentKey: "cp_sps", Title: "CP-SP 绑定删除", Handle: "DELETE:/v1/cp-sps/{id}", Weight: 50},
	{Key: "videos", ParentKey: "cdn", Title: "影片库", Handle: "module://cdn/videos", Weight: 30},
	{Key: "videos:list", ParentKey: "videos", Title: "影片列表", Handle: "/v1/videos", Weight: 10},
	{Key: "videos:create", ParentKey: "videos", Title: "影片新增", Handle: "/v1/videos/create", Weight: 20},
	{Key: "videos:get", ParentKey: "videos", Title: "影片详情", Handle: "/v1/videos/{id}", Weight: 30},
	{Key: "videos:delete", ParentKey: "videos", Title: "影片删除", Handle: "DELETE:/v1/videos/{id}", Weight: 40},
	{Key: "video_categories:list", ParentKey: "videos", Title: "影片分类列表", Handle: "/v1/video-categories", Weight: 50},
	{Key: "video_categories:create", ParentKey: "videos", Title: "影片分类新增", Handle: "/v1/video-categories/create", Weight: 60},
	{Key: "episodes:list", ParentKey: "videos", Title: "节目列表", Handle: "/v1/episodes", Weight: 70},
	{Key: "episodes:create", ParentKey: "videos", Title: "节目新增", Handle: "/v1/episodes/create", Weight: 80},
	{Key: "episodes:delete", ParentKey: "videos", Title: "节目删除", Handle: "DELETE:/v1/episodes/{id}", Weight: 90},
	{Key: "media:list", ParentKey: "videos", Title: "媒体列表", Handle: "/v1/media", Weight: 100},
	{Key: "media:create", ParentKey: "videos", Title: "媒体新增", Handle: "/v1/media/create", Weight: 110},
	{Key: "media:delete", ParentKey: "videos", Title: "媒体删除", Handle: "DELETE:/v1/media/{id}", Weight: 120},
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

func (uc *PermissionUsecase) SyncProgramPermissions(ctx context.Context) error {
	existing, err := uc.repo.ListPermissions(ctx, PermissionListLimit(1000))
	if err != nil {
		return err
	}

	byHandle := make(map[string]*Permission, len(existing))
	byKey := make(map[string]*Permission, len(programPermissionSpecs))
	for _, permission := range existing {
		byHandle[permission.Handle] = permission
	}

	for _, spec := range programPermissionSpecs {
		parentID := int64(0)
		if spec.ParentKey != "" {
			parent := byKey[spec.ParentKey]
			if parent == nil {
				return ErrPermissionInvalidArgument
			}
			parentID = parent.ID
		}

		current := byHandle[spec.Handle]
		if current == nil {
			created, err := uc.repo.CreatePermission(ctx, &Permission{
				ParentID: parentID,
				Title:    spec.Title,
				Handle:   spec.Handle,
				Weight:   spec.Weight,
			})
			if err != nil {
				return err
			}
			byHandle[spec.Handle] = created
			byKey[spec.Key] = created
			continue
		}

		if current.ParentID != parentID || current.Title != spec.Title || current.Weight != spec.Weight {
			updated, err := uc.repo.UpdatePermission(ctx, &Permission{
				ID:       current.ID,
				ParentID: parentID,
				Title:    spec.Title,
				Handle:   spec.Handle,
				Weight:   spec.Weight,
			})
			if err != nil {
				return err
			}
			current = updated
			byHandle[spec.Handle] = updated
		}

		byKey[spec.Key] = current
	}

	return nil
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
