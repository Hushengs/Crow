package service

import (
	"context"

	pb "crow/api/admin/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminService) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.Role, error) {
	role, err := s.roleUC.CreateRole(ctx, convertRole(req.GetRole()))
	if err != nil {
		return nil, err
	}
	return convertRoleReply(role), nil
}

func (s *AdminService) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.Role, error) {
	role, err := s.roleUC.GetRole(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertRoleReply(role), nil
}

func (s *AdminService) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.RoleSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	roles, err := s.roleUC.ListRoles(ctx,
		biz.RoleListLimit(int(req.PageSize)),
		biz.RoleListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.RoleSet{
		Roles: make([]*pb.Role, 0, len(roles)),
	}
	if len(roles) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, role := range roles {
		set.Roles = append(set.Roles, convertRoleReply(role))
	}
	return set, nil
}

func (s *AdminService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.Role, error) {
	if req.GetRole().GetId() <= 0 || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, biz.ErrRoleInvalidArgument
	}
	if err := validateRoleMaskPaths(req.GetUpdateMask().GetPaths()); err != nil {
		return nil, err
	}
	current, err := s.roleUC.GetRole(ctx, req.GetRole().GetId())
	if err != nil {
		return nil, err
	}
	updated, err := applyRoleUpdateMask(current, req.GetRole(), req.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	role, err := s.roleUC.UpdateRole(ctx, updated)
	if err != nil {
		return nil, err
	}
	return convertRoleReply(role), nil
}

func (s *AdminService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*emptypb.Empty, error) {
	if err := s.roleUC.DeleteRole(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func applyRoleUpdateMask(current *biz.Role, patch *pb.Role, paths []string) (*biz.Role, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrRoleInvalidArgument
	}
	updated := *current
	for _, path := range expandRoleMaskPaths(paths) {
		switch path {
		case "role_name":
			updated.RoleName = patch.GetRoleName()
		default:
			return nil, biz.ErrRoleInvalidArgument
		}
	}
	return &updated, nil
}

func validateRoleMaskPaths(paths []string) error {
	for _, path := range expandRoleMaskPaths(paths) {
		switch path {
		case "role_name":
		default:
			return biz.ErrRoleInvalidArgument
		}
	}
	return nil
}

func expandRoleMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "role_name")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertRole(in *pb.Role) *biz.Role {
	if in == nil {
		return nil
	}
	return &biz.Role{
		ID:       in.GetId(),
		RoleName: in.GetRoleName(),
	}
}

func convertRoleReply(in *biz.Role) *pb.Role {
	if in == nil {
		return nil
	}
	return &pb.Role{
		Id:         in.ID,
		RoleName:   in.RoleName,
		CreateTime: timestamppb.New(in.CreateTime),
		UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
