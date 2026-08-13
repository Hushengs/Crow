package service

import (
	"context"
	"sort"
	"testing"
	"time"

	v1 "crow/api/admin/v1"
	"crow/internal/biz"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type stubRoleRepo struct {
	nextID int64
	roles  map[int64]*biz.Role
}

func (r *stubRoleRepo) FindByID(_ context.Context, id int64) (*biz.Role, error) {
	role, ok := r.roles[id]
	if !ok {
		return nil, biz.ErrRoleNotFound
	}
	return cloneRoleForTest(role), nil
}

func (r *stubRoleRepo) ListRoles(_ context.Context, opts ...biz.RoleListOption) ([]*biz.Role, error) {
	options := biz.RoleListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrRoleInvalidArgument
	}
	list := make([]*biz.Role, 0, len(r.roles))
	for _, role := range r.roles {
		list = append(list, cloneRoleForTest(role))
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID > list[j].ID })
	if options.Offset >= len(list) {
		return []*biz.Role{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubRoleRepo) CreateRole(_ context.Context, role *biz.Role) (*biz.Role, error) {
	now := time.Now()
	created := cloneRoleForTest(role)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.roles[created.ID] = cloneRoleForTest(created)
	r.nextID++
	return cloneRoleForTest(created), nil
}

func (r *stubRoleRepo) UpdateRole(_ context.Context, role *biz.Role) (*biz.Role, error) {
	current, ok := r.roles[role.ID]
	if !ok {
		return nil, biz.ErrRoleNotFound
	}
	updated := cloneRoleForTest(role)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.roles[updated.ID] = cloneRoleForTest(updated)
	return cloneRoleForTest(updated), nil
}

func (r *stubRoleRepo) DeleteRole(_ context.Context, id int64) error {
	if _, ok := r.roles[id]; !ok {
		return biz.ErrRoleNotFound
	}
	delete(r.roles, id)
	return nil
}

type stubAdminRoleRepo struct {
	nextID     int64
	adminRoles map[int64]*biz.AdminRole
}

func (r *stubAdminRoleRepo) FindByID(_ context.Context, id int64) (*biz.AdminRole, error) {
	relation, ok := r.adminRoles[id]
	if !ok {
		return nil, biz.ErrAdminRoleNotFound
	}
	return cloneAdminRoleForTest(relation), nil
}

func (r *stubAdminRoleRepo) ListAdminRoles(_ context.Context, opts ...biz.AdminRoleListOption) ([]*biz.AdminRole, error) {
	options := biz.AdminRoleListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrAdminRoleInvalidArgument
	}
	list := make([]*biz.AdminRole, 0, len(r.adminRoles))
	for _, relation := range r.adminRoles {
		list = append(list, cloneAdminRoleForTest(relation))
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID > list[j].ID })
	if options.Offset >= len(list) {
		return []*biz.AdminRole{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubAdminRoleRepo) CreateAdminRole(_ context.Context, relation *biz.AdminRole) (*biz.AdminRole, error) {
	now := time.Now()
	created := cloneAdminRoleForTest(relation)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.adminRoles[created.ID] = cloneAdminRoleForTest(created)
	r.nextID++
	return cloneAdminRoleForTest(created), nil
}

func (r *stubAdminRoleRepo) UpdateAdminRole(_ context.Context, relation *biz.AdminRole) (*biz.AdminRole, error) {
	current, ok := r.adminRoles[relation.ID]
	if !ok {
		return nil, biz.ErrAdminRoleNotFound
	}
	updated := cloneAdminRoleForTest(relation)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.adminRoles[updated.ID] = cloneAdminRoleForTest(updated)
	return cloneAdminRoleForTest(updated), nil
}

func (r *stubAdminRoleRepo) DeleteAdminRole(_ context.Context, id int64) error {
	if _, ok := r.adminRoles[id]; !ok {
		return biz.ErrAdminRoleNotFound
	}
	delete(r.adminRoles, id)
	return nil
}

type stubPermissionRepo struct {
	nextID      int64
	permissions map[int64]*biz.Permission
}

func (r *stubPermissionRepo) FindByID(_ context.Context, id int64) (*biz.Permission, error) {
	permission, ok := r.permissions[id]
	if !ok {
		return nil, biz.ErrPermissionNotFound
	}
	return clonePermissionForTest(permission), nil
}

func (r *stubPermissionRepo) ListPermissions(_ context.Context, opts ...biz.PermissionListOption) ([]*biz.Permission, error) {
	options := biz.PermissionListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrPermissionInvalidArgument
	}
	list := make([]*biz.Permission, 0, len(r.permissions))
	for _, permission := range r.permissions {
		list = append(list, clonePermissionForTest(permission))
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].ParentID != list[j].ParentID {
			return list[i].ParentID < list[j].ParentID
		}
		if list[i].Weight != list[j].Weight {
			return list[i].Weight < list[j].Weight
		}
		return list[i].ID < list[j].ID
	})
	if options.Offset >= len(list) {
		return []*biz.Permission{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubPermissionRepo) CreatePermission(_ context.Context, permission *biz.Permission) (*biz.Permission, error) {
	now := time.Now()
	created := clonePermissionForTest(permission)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.permissions[created.ID] = clonePermissionForTest(created)
	r.nextID++
	return clonePermissionForTest(created), nil
}

func (r *stubPermissionRepo) UpdatePermission(_ context.Context, permission *biz.Permission) (*biz.Permission, error) {
	current, ok := r.permissions[permission.ID]
	if !ok {
		return nil, biz.ErrPermissionNotFound
	}
	updated := clonePermissionForTest(permission)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.permissions[updated.ID] = clonePermissionForTest(updated)
	return clonePermissionForTest(updated), nil
}

func (r *stubPermissionRepo) DeletePermission(_ context.Context, id int64) error {
	if _, ok := r.permissions[id]; !ok {
		return biz.ErrPermissionNotFound
	}
	delete(r.permissions, id)
	return nil
}

type stubGroupPermissionRepo struct {
	nextID           int64
	groupPermissions map[int64]*biz.GroupPermission
}

func (r *stubGroupPermissionRepo) FindByID(_ context.Context, id int64) (*biz.GroupPermission, error) {
	relation, ok := r.groupPermissions[id]
	if !ok {
		return nil, biz.ErrGroupPermissionNotFound
	}
	return cloneGroupPermissionForTest(relation), nil
}

func (r *stubGroupPermissionRepo) ListGroupPermissions(_ context.Context, opts ...biz.GroupPermissionListOption) ([]*biz.GroupPermission, error) {
	options := biz.GroupPermissionListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrGroupPermissionInvalidArgument
	}
	list := make([]*biz.GroupPermission, 0, len(r.groupPermissions))
	for _, relation := range r.groupPermissions {
		list = append(list, cloneGroupPermissionForTest(relation))
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID > list[j].ID })
	if options.Offset >= len(list) {
		return []*biz.GroupPermission{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubGroupPermissionRepo) CreateGroupPermission(_ context.Context, relation *biz.GroupPermission) (*biz.GroupPermission, error) {
	now := time.Now()
	created := cloneGroupPermissionForTest(relation)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.groupPermissions[created.ID] = cloneGroupPermissionForTest(created)
	r.nextID++
	return cloneGroupPermissionForTest(created), nil
}

func (r *stubGroupPermissionRepo) UpdateGroupPermission(_ context.Context, relation *biz.GroupPermission) (*biz.GroupPermission, error) {
	current, ok := r.groupPermissions[relation.ID]
	if !ok {
		return nil, biz.ErrGroupPermissionNotFound
	}
	updated := cloneGroupPermissionForTest(relation)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.groupPermissions[updated.ID] = cloneGroupPermissionForTest(updated)
	return cloneGroupPermissionForTest(updated), nil
}

func (r *stubGroupPermissionRepo) DeleteGroupPermission(_ context.Context, id int64) error {
	if _, ok := r.groupPermissions[id]; !ok {
		return biz.ErrGroupPermissionNotFound
	}
	delete(r.groupPermissions, id)
	return nil
}

func newTestRBACService() *AdminService {
	return NewAdminService(
		biz.NewAdminUsecase(&stubAdminRepo{nextID: 1, admins: make(map[int64]*biz.Admin)}),
		biz.NewRoleUsecase(&stubRoleRepo{nextID: 1, roles: make(map[int64]*biz.Role)}),
		biz.NewAdminRoleUsecase(&stubAdminRoleRepo{nextID: 1, adminRoles: make(map[int64]*biz.AdminRole)}),
		biz.NewPermissionUsecase(&stubPermissionRepo{nextID: 1, permissions: make(map[int64]*biz.Permission)}),
		biz.NewGroupPermissionUsecase(&stubGroupPermissionRepo{nextID: 1, groupPermissions: make(map[int64]*biz.GroupPermission)}),
	)
}

func TestRoleServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestRBACService()

	created, err := svc.CreateRole(ctx, &v1.CreateRoleRequest{Role: &v1.Role{RoleName: "super-admin"}})
	if err != nil {
		t.Fatalf("CreateRole() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreateRole() id = %d, want 1", created.GetId())
	}

	got, err := svc.GetRole(ctx, &v1.GetRoleRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetRole() error = %v", err)
	}
	if got.GetRoleName() != "super-admin" {
		t.Fatalf("GetRole() role_name = %q, want super-admin", got.GetRoleName())
	}

	list, err := svc.ListRoles(ctx, &v1.ListRolesRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("ListRoles() error = %v", err)
	}
	if len(list.GetRoles()) != 1 {
		t.Fatalf("ListRoles() len = %d, want 1", len(list.GetRoles()))
	}

	updated, err := svc.UpdateRole(ctx, &v1.UpdateRoleRequest{
		Role:       &v1.Role{Id: created.GetId(), RoleName: "auditor"},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"role_name"}},
	})
	if err != nil {
		t.Fatalf("UpdateRole() error = %v", err)
	}
	if updated.GetRoleName() != "auditor" {
		t.Fatalf("UpdateRole() role_name = %q, want auditor", updated.GetRoleName())
	}

	if _, err := svc.DeleteRole(ctx, &v1.DeleteRoleRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteRole() error = %v", err)
	}
	if _, err := svc.GetRole(ctx, &v1.GetRoleRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetRole() after delete error = %v, want not found", err)
	}
}

func TestAdminRoleServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestRBACService()

	created, err := svc.CreateAdminRole(ctx, &v1.CreateAdminRoleRequest{
		AdminRole: &v1.AdminRole{AdminId: 1, RoleId: 2},
	})
	if err != nil {
		t.Fatalf("CreateAdminRole() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreateAdminRole() id = %d, want 1", created.GetId())
	}

	got, err := svc.GetAdminRole(ctx, &v1.GetAdminRoleRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetAdminRole() error = %v", err)
	}
	if got.GetAdminId() != 1 || got.GetRoleId() != 2 {
		t.Fatalf("GetAdminRole() = %+v, want admin_id=1 role_id=2", got)
	}

	list, err := svc.ListAdminRoles(ctx, &v1.ListAdminRolesRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("ListAdminRoles() error = %v", err)
	}
	if len(list.GetAdminRoles()) != 1 {
		t.Fatalf("ListAdminRoles() len = %d, want 1", len(list.GetAdminRoles()))
	}

	updated, err := svc.UpdateAdminRole(ctx, &v1.UpdateAdminRoleRequest{
		AdminRole:  &v1.AdminRole{Id: created.GetId(), RoleId: 3},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"role_id"}},
	})
	if err != nil {
		t.Fatalf("UpdateAdminRole() error = %v", err)
	}
	if updated.GetAdminId() != 1 || updated.GetRoleId() != 3 {
		t.Fatalf("UpdateAdminRole() = %+v, want admin_id=1 role_id=3", updated)
	}

	if _, err := svc.DeleteAdminRole(ctx, &v1.DeleteAdminRoleRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteAdminRole() error = %v", err)
	}
	if _, err := svc.GetAdminRole(ctx, &v1.GetAdminRoleRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetAdminRole() after delete error = %v, want not found", err)
	}
}

func TestPermissionServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestRBACService()

	created, err := svc.CreatePermission(ctx, &v1.CreatePermissionRequest{
		Permission: &v1.Permission{
			ParentId: 0,
			Title:    "用户列表",
			Handle:   "/users",
			Weight:   10,
		},
	})
	if err != nil {
		t.Fatalf("CreatePermission() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreatePermission() id = %d, want 1", created.GetId())
	}

	got, err := svc.GetPermission(ctx, &v1.GetPermissionRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetPermission() error = %v", err)
	}
	if got.GetHandle() != "/users" || got.GetTitle() != "用户列表" {
		t.Fatalf("GetPermission() = %+v, want created permission", got)
	}

	list, err := svc.ListPermissions(ctx, &v1.ListPermissionsRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("ListPermissions() error = %v", err)
	}
	if len(list.GetPermissions()) != 1 {
		t.Fatalf("ListPermissions() len = %d, want 1", len(list.GetPermissions()))
	}

	updated, err := svc.UpdatePermission(ctx, &v1.UpdatePermissionRequest{
		Permission: &v1.Permission{
			Id:     created.GetId(),
			Title:  "用户管理",
			Handle: "/admins/users",
			Weight: 20,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title", "handle", "weight"}},
	})
	if err != nil {
		t.Fatalf("UpdatePermission() error = %v", err)
	}
	if updated.GetTitle() != "用户管理" || updated.GetHandle() != "/admins/users" || updated.GetWeight() != 20 {
		t.Fatalf("UpdatePermission() = %+v, want updated fields", updated)
	}

	if _, err := svc.DeletePermission(ctx, &v1.DeletePermissionRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeletePermission() error = %v", err)
	}
	if _, err := svc.GetPermission(ctx, &v1.GetPermissionRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetPermission() after delete error = %v, want not found", err)
	}
}

func TestGroupPermissionServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestRBACService()

	created, err := svc.CreateGroupPermission(ctx, &v1.CreateGroupPermissionRequest{
		GroupPermission: &v1.GroupPermission{GroupId: 1, PermissionId: 2},
	})
	if err != nil {
		t.Fatalf("CreateGroupPermission() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreateGroupPermission() id = %d, want 1", created.GetId())
	}

	got, err := svc.GetGroupPermission(ctx, &v1.GetGroupPermissionRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetGroupPermission() error = %v", err)
	}
	if got.GetGroupId() != 1 || got.GetPermissionId() != 2 {
		t.Fatalf("GetGroupPermission() = %+v, want group_id=1 permission_id=2", got)
	}

	list, err := svc.ListGroupPermissions(ctx, &v1.ListGroupPermissionsRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("ListGroupPermissions() error = %v", err)
	}
	if len(list.GetGroupPermissions()) != 1 {
		t.Fatalf("ListGroupPermissions() len = %d, want 1", len(list.GetGroupPermissions()))
	}

	updated, err := svc.UpdateGroupPermission(ctx, &v1.UpdateGroupPermissionRequest{
		GroupPermission: &v1.GroupPermission{Id: created.GetId(), PermissionId: 5},
		UpdateMask:      &fieldmaskpb.FieldMask{Paths: []string{"permission_id"}},
	})
	if err != nil {
		t.Fatalf("UpdateGroupPermission() error = %v", err)
	}
	if updated.GetGroupId() != 1 || updated.GetPermissionId() != 5 {
		t.Fatalf("UpdateGroupPermission() = %+v, want group_id=1 permission_id=5", updated)
	}

	if _, err := svc.DeleteGroupPermission(ctx, &v1.DeleteGroupPermissionRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteGroupPermission() error = %v", err)
	}
	if _, err := svc.GetGroupPermission(ctx, &v1.GetGroupPermissionRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetGroupPermission() after delete error = %v, want not found", err)
	}
}

func TestRBACServiceValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestRBACService()

	if _, err := svc.CreateRole(ctx, &v1.CreateRoleRequest{Role: &v1.Role{}}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateRole() error = %v, want bad request", err)
	}
	if _, err := svc.UpdatePermission(ctx, &v1.UpdatePermissionRequest{
		Permission: &v1.Permission{Id: 1, Title: "x"},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("UpdatePermission() error = %v, want bad request", err)
	}
}

func cloneRoleForTest(in *biz.Role) *biz.Role {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}

func cloneAdminRoleForTest(in *biz.AdminRole) *biz.AdminRole {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}

func clonePermissionForTest(in *biz.Permission) *biz.Permission {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}

func cloneGroupPermissionForTest(in *biz.GroupPermission) *biz.GroupPermission {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}
