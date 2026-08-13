package service

import (
	"context"

	pb "crow/api/admin/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminService) CreateAdminRole(ctx context.Context, req *pb.CreateAdminRoleRequest) (*pb.AdminRole, error) {
	relation, err := s.adminRoleUC.CreateAdminRole(ctx, convertAdminRole(req.GetAdminRole()))
	if err != nil {
		return nil, err
	}
	return convertAdminRoleReply(relation), nil
}

func (s *AdminService) GetAdminRole(ctx context.Context, req *pb.GetAdminRoleRequest) (*pb.AdminRole, error) {
	relation, err := s.adminRoleUC.GetAdminRole(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertAdminRoleReply(relation), nil
}

func (s *AdminService) ListAdminRoles(ctx context.Context, req *pb.ListAdminRolesRequest) (*pb.AdminRoleSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	relations, err := s.adminRoleUC.ListAdminRoles(ctx,
		biz.AdminRoleListLimit(int(req.PageSize)),
		biz.AdminRoleListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.AdminRoleSet{
		AdminRoles: make([]*pb.AdminRole, 0, len(relations)),
	}
	if len(relations) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, relation := range relations {
		set.AdminRoles = append(set.AdminRoles, convertAdminRoleReply(relation))
	}
	return set, nil
}

func (s *AdminService) UpdateAdminRole(ctx context.Context, req *pb.UpdateAdminRoleRequest) (*pb.AdminRole, error) {
	if req.GetAdminRole().GetId() <= 0 || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, biz.ErrAdminRoleInvalidArgument
	}
	if err := validateAdminRoleMaskPaths(req.GetUpdateMask().GetPaths()); err != nil {
		return nil, err
	}
	current, err := s.adminRoleUC.GetAdminRole(ctx, req.GetAdminRole().GetId())
	if err != nil {
		return nil, err
	}
	updated, err := applyAdminRoleUpdateMask(current, req.GetAdminRole(), req.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	relation, err := s.adminRoleUC.UpdateAdminRole(ctx, updated)
	if err != nil {
		return nil, err
	}
	return convertAdminRoleReply(relation), nil
}

func (s *AdminService) DeleteAdminRole(ctx context.Context, req *pb.DeleteAdminRoleRequest) (*emptypb.Empty, error) {
	if err := s.adminRoleUC.DeleteAdminRole(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func applyAdminRoleUpdateMask(current *biz.AdminRole, patch *pb.AdminRole, paths []string) (*biz.AdminRole, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrAdminRoleInvalidArgument
	}
	updated := *current
	for _, path := range expandAdminRoleMaskPaths(paths) {
		switch path {
		case "admin_id":
			updated.AdminID = patch.GetAdminId()
		case "role_id":
			updated.RoleID = patch.GetRoleId()
		default:
			return nil, biz.ErrAdminRoleInvalidArgument
		}
	}
	return &updated, nil
}

func validateAdminRoleMaskPaths(paths []string) error {
	for _, path := range expandAdminRoleMaskPaths(paths) {
		switch path {
		case "admin_id", "role_id":
		default:
			return biz.ErrAdminRoleInvalidArgument
		}
	}
	return nil
}

func expandAdminRoleMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "admin_id", "role_id")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertAdminRole(in *pb.AdminRole) *biz.AdminRole {
	if in == nil {
		return nil
	}
	return &biz.AdminRole{
		ID:      in.GetId(),
		AdminID: in.GetAdminId(),
		RoleID:  in.GetRoleId(),
	}
}

func convertAdminRoleReply(in *biz.AdminRole) *pb.AdminRole {
	if in == nil {
		return nil
	}
	return &pb.AdminRole{
		Id:         in.ID,
		AdminId:    in.AdminID,
		RoleId:     in.RoleID,
		CreateTime: timestamppb.New(in.CreateTime),
		UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
