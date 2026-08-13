package service

import (
	"context"
	"time"

	pb "crow/api/admin/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AdminService struct {
	pb.UnimplementedAdminServiceServer

	adminUC           *biz.AdminUsecase
	roleUC            *biz.RoleUsecase
	adminRoleUC       *biz.AdminRoleUsecase
	permissionUC      *biz.PermissionUsecase
	groupPermissionUC *biz.GroupPermissionUsecase
	operationLogUC    *biz.AdminOperationLogUsecase
	systemLogUC       *biz.SystemLogUsecase
}

func NewAdminService(
	adminUC *biz.AdminUsecase,
	roleUC *biz.RoleUsecase,
	adminRoleUC *biz.AdminRoleUsecase,
	permissionUC *biz.PermissionUsecase,
	groupPermissionUC *biz.GroupPermissionUsecase,
	operationLogUC *biz.AdminOperationLogUsecase,
	systemLogUC *biz.SystemLogUsecase,
) *AdminService {
	return &AdminService{
		adminUC:           adminUC,
		roleUC:            roleUC,
		adminRoleUC:       adminRoleUC,
		permissionUC:      permissionUC,
		groupPermissionUC: groupPermissionUC,
		operationLogUC:    operationLogUC,
		systemLogUC:       systemLogUC,
	}
}

func (s *AdminService) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest) (*pb.Admin, error) {
	admin, err := s.adminUC.CreateAdmin(ctx, convertAdmin(req.GetAdmin()))
	if err != nil {
		return nil, err
	}
	return convertAdminReply(admin), nil
}

func (s *AdminService) GetAdmin(ctx context.Context, req *pb.GetAdminRequest) (*pb.Admin, error) {
	admin, err := s.adminUC.GetAdmin(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertAdminReply(admin), nil
}

func (s *AdminService) ListAdmins(ctx context.Context, req *pb.ListAdminsRequest) (*pb.AdminSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	admins, err := s.adminUC.ListAdmins(ctx,
		biz.AdminListLimit(int(req.PageSize)),
		biz.AdminListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.AdminSet{
		Admins: make([]*pb.Admin, 0, len(admins)),
	}
	if len(admins) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, admin := range admins {
		set.Admins = append(set.Admins, convertAdminReply(admin))
	}
	return set, nil
}

func (s *AdminService) UpdateAdmin(ctx context.Context, req *pb.UpdateAdminRequest) (*pb.Admin, error) {
	if req.GetAdmin().GetId() <= 0 || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, biz.ErrAdminInvalidArgument
	}
	if err := validateAdminMaskPaths(req.GetUpdateMask().GetPaths()); err != nil {
		return nil, err
	}
	current, err := s.adminUC.GetAdmin(ctx, req.GetAdmin().GetId())
	if err != nil {
		return nil, err
	}
	updated, err := applyAdminUpdateMask(current, req.GetAdmin(), req.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	admin, err := s.adminUC.UpdateAdmin(ctx, updated)
	if err != nil {
		return nil, err
	}
	return convertAdminReply(admin), nil
}

func (s *AdminService) DeleteAdmin(ctx context.Context, req *pb.DeleteAdminRequest) (*emptypb.Empty, error) {
	if err := s.adminUC.DeleteAdmin(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func applyAdminUpdateMask(current *biz.Admin, patch *pb.Admin, paths []string) (*biz.Admin, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrAdminInvalidArgument
	}
	updated := cloneAdminDomain(current)
	for _, path := range expandAdminMaskPaths(paths) {
		switch path {
		case "username":
			updated.Username = patch.GetUsername()
		case "password":
			updated.Password = patch.GetPassword()
		case "real_name":
			updated.RealName = patch.GetRealName()
		case "role_id":
			updated.RoleID = patch.GetRoleId()
		case "status":
			updated.Status = patch.GetStatus()
		case "remark":
			updated.Remark = patch.GetRemark()
		default:
			return nil, biz.ErrAdminInvalidArgument
		}
	}
	return updated, nil
}

func validateAdminMaskPaths(paths []string) error {
	for _, path := range expandAdminMaskPaths(paths) {
		switch path {
		case "username", "password", "real_name", "role_id", "status", "remark":
		default:
			return biz.ErrAdminInvalidArgument
		}
	}
	return nil
}

func expandAdminMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "username", "password", "real_name", "role_id", "status", "remark")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertAdmin(in *pb.Admin) *biz.Admin {
	if in == nil {
		return nil
	}
	return &biz.Admin{
		ID:       in.GetId(),
		Username: in.GetUsername(),
		Password: in.GetPassword(),
		RealName: in.GetRealName(),
		RoleID:   in.GetRoleId(),
		Status:   in.GetStatus(),
		Remark:   in.GetRemark(),
	}
}

func convertAdminReply(in *biz.Admin) *pb.Admin {
	if in == nil {
		return nil
	}
	return &pb.Admin{
		Id:                in.ID,
		Username:          in.Username,
		RealName:          in.RealName,
		RoleId:            in.RoleID,
		Status:            in.Status,
		LastLoginTime:     toTimestamp(in.LastLoginTime),
		PasswordUpdatedAt: toTimestamp(in.PasswordUpdatedAt),
		Remark:            in.Remark,
		CreateTime:        timestamppb.New(in.CreateTime),
		UpdateTime:        timestamppb.New(in.UpdateTime),
	}
}

func toTimestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}

func cloneAdminDomain(in *biz.Admin) *biz.Admin {
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
