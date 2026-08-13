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

type stubAdminRepo struct {
	nextID int64
	admins map[int64]*biz.Admin
}

func newTestAdminService() *AdminService {
	adminRepo := &stubAdminRepo{
		nextID: 1,
		admins: make(map[int64]*biz.Admin),
	}
	return NewAdminService(
		biz.NewAdminUsecase(adminRepo),
		biz.NewRoleUsecase(&stubRoleRepo{nextID: 1, roles: make(map[int64]*biz.Role)}),
		biz.NewAdminRoleUsecase(&stubAdminRoleRepo{nextID: 1, adminRoles: make(map[int64]*biz.AdminRole)}),
		biz.NewPermissionUsecase(&stubPermissionRepo{nextID: 1, permissions: make(map[int64]*biz.Permission)}),
		biz.NewGroupPermissionUsecase(&stubGroupPermissionRepo{nextID: 1, groupPermissions: make(map[int64]*biz.GroupPermission)}),
		biz.NewAdminOperationLogUsecase(&stubAdminOperationLogRepo{nextID: 1, logs: make(map[int64]*biz.AdminOperationLog)}),
	)
}

func (r *stubAdminRepo) FindByID(_ context.Context, id int64) (*biz.Admin, error) {
	admin, ok := r.admins[id]
	if !ok {
		return nil, biz.ErrAdminNotFound
	}
	return cloneAdminForTest(admin), nil
}

func (r *stubAdminRepo) ListAdmins(_ context.Context, opts ...biz.AdminListOption) ([]*biz.Admin, error) {
	options := biz.AdminListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrAdminInvalidArgument
	}

	list := make([]*biz.Admin, 0, len(r.admins))
	for _, admin := range r.admins {
		list = append(list, cloneAdminForTest(admin))
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID > list[j].ID
	})

	if options.Offset >= len(list) {
		return []*biz.Admin{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubAdminRepo) CreateAdmin(_ context.Context, admin *biz.Admin) (*biz.Admin, error) {
	now := time.Now()
	created := cloneAdminForTest(admin)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.admins[created.ID] = cloneAdminForTest(created)
	r.nextID++
	return cloneAdminForTest(created), nil
}

func (r *stubAdminRepo) UpdateAdmin(_ context.Context, admin *biz.Admin) (*biz.Admin, error) {
	current, ok := r.admins[admin.ID]
	if !ok {
		return nil, biz.ErrAdminNotFound
	}
	updated := cloneAdminForTest(admin)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.admins[updated.ID] = cloneAdminForTest(updated)
	return cloneAdminForTest(updated), nil
}

func (r *stubAdminRepo) DeleteAdmin(_ context.Context, id int64) error {
	if _, ok := r.admins[id]; !ok {
		return biz.ErrAdminNotFound
	}
	delete(r.admins, id)
	return nil
}

func TestAdminServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestAdminService()

	created, err := svc.CreateAdmin(ctx, &v1.CreateAdminRequest{
		Admin: &v1.Admin{
			Username: "admin",
			Password: "123456",
			RealName: "系统管理员",
			RoleId:   1,
			Status:   biz.AdminStatusNormal,
			Remark:   "seed",
		},
	})
	if err != nil {
		t.Fatalf("CreateAdmin() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreateAdmin() id = %d, want 1", created.GetId())
	}
	if created.GetPassword() != "" {
		t.Fatalf("CreateAdmin() password = %q, want empty", created.GetPassword())
	}
	if created.GetPasswordUpdatedAt() == nil {
		t.Fatal("CreateAdmin() password_updated_at is nil")
	}

	got, err := svc.GetAdmin(ctx, &v1.GetAdminRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetAdmin() error = %v", err)
	}
	if got.GetUsername() != "admin" || got.GetRealName() != "系统管理员" {
		t.Fatalf("GetAdmin() = %+v, want created admin", got)
	}

	updated, err := svc.UpdateAdmin(ctx, &v1.UpdateAdminRequest{
		Admin: &v1.Admin{
			Id:       created.GetId(),
			Password: "654321",
			RealName: "超级管理员",
			Status:   biz.AdminStatusDisabled,
			Remark:   "updated",
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"password", "real_name", "status", "remark"}},
	})
	if err != nil {
		t.Fatalf("UpdateAdmin() error = %v", err)
	}
	if updated.GetPassword() != "" {
		t.Fatalf("UpdateAdmin() password = %q, want empty", updated.GetPassword())
	}
	if updated.GetRealName() != "超级管理员" || updated.GetStatus() != biz.AdminStatusDisabled {
		t.Fatalf("UpdateAdmin() = %+v, want updated fields", updated)
	}

	if _, err := svc.DeleteAdmin(ctx, &v1.DeleteAdminRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteAdmin() error = %v", err)
	}
	if _, err := svc.GetAdmin(ctx, &v1.GetAdminRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetAdmin() after delete error = %v, want not found", err)
	}
}

func TestAdminServiceListAdminsPagination(t *testing.T) {
	ctx := context.Background()
	svc := newTestAdminService()

	for _, username := range []string{"alpha", "beta", "gamma"} {
		if _, err := svc.CreateAdmin(ctx, &v1.CreateAdminRequest{
			Admin: &v1.Admin{
				Username: username,
				Password: "123456",
				Status:   biz.AdminStatusNormal,
			},
		}); err != nil {
			t.Fatalf("CreateAdmin(%q) error = %v", username, err)
		}
	}

	firstPage, err := svc.ListAdmins(ctx, &v1.ListAdminsRequest{PageSize: 2})
	if err != nil {
		t.Fatalf("ListAdmins(first page) error = %v", err)
	}
	if len(firstPage.GetAdmins()) != 2 {
		t.Fatalf("ListAdmins(first page) len = %d, want 2", len(firstPage.GetAdmins()))
	}
	if firstPage.GetAdmins()[0].GetUsername() != "gamma" {
		t.Fatalf("ListAdmins(first page) first username = %q, want gamma", firstPage.GetAdmins()[0].GetUsername())
	}
	if firstPage.GetNextPageToken() == "" {
		t.Fatal("ListAdmins(first page) next_page_token is empty")
	}
}

func TestAdminServiceValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestAdminService()

	if _, err := svc.CreateAdmin(ctx, &v1.CreateAdminRequest{
		Admin: &v1.Admin{Username: "admin", Status: biz.AdminStatusNormal},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateAdmin(missing password) error = %v, want bad request", err)
	}
	if _, err := svc.UpdateAdmin(ctx, &v1.UpdateAdminRequest{
		Admin:      &v1.Admin{Id: 1, Username: "admin"},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("UpdateAdmin(unknown path) error = %v, want bad request", err)
	}
}

func cloneAdminForTest(in *biz.Admin) *biz.Admin {
	if in == nil {
		return nil
	}
	cloned := *in
	if in.LastLoginTime != nil {
		t := *in.LastLoginTime
		cloned.LastLoginTime = &t
	}
	if in.PasswordUpdatedAt != nil {
		t := *in.PasswordUpdatedAt
		cloned.PasswordUpdatedAt = &t
	}
	return &cloned
}
