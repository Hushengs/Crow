package service

import (
	"context"

	pb "crow/api/admin/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminService) CreateGroupPermission(ctx context.Context, req *pb.CreateGroupPermissionRequest) (*pb.GroupPermission, error) {
	relation, err := s.groupPermissionUC.CreateGroupPermission(ctx, convertGroupPermission(req.GetGroupPermission()))
	if err != nil {
		return nil, err
	}
	return convertGroupPermissionReply(relation), nil
}

func (s *AdminService) GetGroupPermission(ctx context.Context, req *pb.GetGroupPermissionRequest) (*pb.GroupPermission, error) {
	relation, err := s.groupPermissionUC.GetGroupPermission(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertGroupPermissionReply(relation), nil
}

func (s *AdminService) ListGroupPermissions(ctx context.Context, req *pb.ListGroupPermissionsRequest) (*pb.GroupPermissionSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	relations, err := s.groupPermissionUC.ListGroupPermissions(ctx,
		biz.GroupPermissionListLimit(int(req.PageSize)),
		biz.GroupPermissionListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.GroupPermissionSet{
		GroupPermissions: make([]*pb.GroupPermission, 0, len(relations)),
	}
	if len(relations) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, relation := range relations {
		set.GroupPermissions = append(set.GroupPermissions, convertGroupPermissionReply(relation))
	}
	return set, nil
}

func (s *AdminService) UpdateGroupPermission(ctx context.Context, req *pb.UpdateGroupPermissionRequest) (*pb.GroupPermission, error) {
	if req.GetGroupPermission().GetId() <= 0 || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, biz.ErrGroupPermissionInvalidArgument
	}
	if err := validateGroupPermissionMaskPaths(req.GetUpdateMask().GetPaths()); err != nil {
		return nil, err
	}
	current, err := s.groupPermissionUC.GetGroupPermission(ctx, req.GetGroupPermission().GetId())
	if err != nil {
		return nil, err
	}
	updated, err := applyGroupPermissionUpdateMask(current, req.GetGroupPermission(), req.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	relation, err := s.groupPermissionUC.UpdateGroupPermission(ctx, updated)
	if err != nil {
		return nil, err
	}
	return convertGroupPermissionReply(relation), nil
}

func (s *AdminService) DeleteGroupPermission(ctx context.Context, req *pb.DeleteGroupPermissionRequest) (*emptypb.Empty, error) {
	if err := s.groupPermissionUC.DeleteGroupPermission(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func applyGroupPermissionUpdateMask(current *biz.GroupPermission, patch *pb.GroupPermission, paths []string) (*biz.GroupPermission, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrGroupPermissionInvalidArgument
	}
	updated := *current
	for _, path := range expandGroupPermissionMaskPaths(paths) {
		switch path {
		case "group_id":
			updated.GroupID = patch.GetGroupId()
		case "permission_id":
			updated.PermissionID = patch.GetPermissionId()
		default:
			return nil, biz.ErrGroupPermissionInvalidArgument
		}
	}
	return &updated, nil
}

func validateGroupPermissionMaskPaths(paths []string) error {
	for _, path := range expandGroupPermissionMaskPaths(paths) {
		switch path {
		case "group_id", "permission_id":
		default:
			return biz.ErrGroupPermissionInvalidArgument
		}
	}
	return nil
}

func expandGroupPermissionMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "group_id", "permission_id")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertGroupPermission(in *pb.GroupPermission) *biz.GroupPermission {
	if in == nil {
		return nil
	}
	return &biz.GroupPermission{
		ID:           in.GetId(),
		GroupID:      in.GetGroupId(),
		PermissionID: in.GetPermissionId(),
	}
}

func convertGroupPermissionReply(in *biz.GroupPermission) *pb.GroupPermission {
	if in == nil {
		return nil
	}
	return &pb.GroupPermission{
		Id:           in.ID,
		GroupId:      in.GroupID,
		PermissionId: in.PermissionID,
		CreateTime:   timestamppb.New(in.CreateTime),
		UpdateTime:   timestamppb.New(in.UpdateTime),
	}
}
