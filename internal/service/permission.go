package service

import (
	"context"

	pb "crow/api/admin/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminService) CreatePermission(ctx context.Context, req *pb.CreatePermissionRequest) (*pb.Permission, error) {
	return nil, biz.ErrPermissionInvalidArgument
}

func (s *AdminService) GetPermission(ctx context.Context, req *pb.GetPermissionRequest) (*pb.Permission, error) {
	permission, err := s.permissionUC.GetPermission(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertPermissionReply(permission), nil
}

func (s *AdminService) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.PermissionSet, error) {
	if err := s.permissionUC.SyncProgramPermissions(ctx); err != nil {
		return nil, err
	}
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	permissions, err := s.permissionUC.ListPermissions(ctx,
		biz.PermissionListLimit(int(req.PageSize)),
		biz.PermissionListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.PermissionSet{
		Permissions: make([]*pb.Permission, 0, len(permissions)),
	}
	if len(permissions) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, permission := range permissions {
		set.Permissions = append(set.Permissions, convertPermissionReply(permission))
	}
	return set, nil
}

func (s *AdminService) UpdatePermission(ctx context.Context, req *pb.UpdatePermissionRequest) (*pb.Permission, error) {
	return nil, biz.ErrPermissionInvalidArgument
}

func (s *AdminService) DeletePermission(ctx context.Context, req *pb.DeletePermissionRequest) (*emptypb.Empty, error) {
	return nil, biz.ErrPermissionInvalidArgument
}

func applyPermissionUpdateMask(current *biz.Permission, patch *pb.Permission, paths []string) (*biz.Permission, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrPermissionInvalidArgument
	}
	updated := *current
	for _, path := range expandPermissionMaskPaths(paths) {
		switch path {
		case "parent_id":
			updated.ParentID = patch.GetParentId()
		case "title":
			updated.Title = patch.GetTitle()
		case "handle":
			updated.Handle = patch.GetHandle()
		case "weight":
			updated.Weight = patch.GetWeight()
		default:
			return nil, biz.ErrPermissionInvalidArgument
		}
	}
	return &updated, nil
}

func validatePermissionMaskPaths(paths []string) error {
	for _, path := range expandPermissionMaskPaths(paths) {
		switch path {
		case "parent_id", "title", "handle", "weight":
		default:
			return biz.ErrPermissionInvalidArgument
		}
	}
	return nil
}

func expandPermissionMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "parent_id", "title", "handle", "weight")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertPermission(in *pb.Permission) *biz.Permission {
	if in == nil {
		return nil
	}
	return &biz.Permission{
		ID:       in.GetId(),
		ParentID: in.GetParentId(),
		Title:    in.GetTitle(),
		Handle:   in.GetHandle(),
		Weight:   in.GetWeight(),
	}
}

func convertPermissionReply(in *biz.Permission) *pb.Permission {
	if in == nil {
		return nil
	}
	return &pb.Permission{
		Id:         in.ID,
		ParentId:   in.ParentID,
		Title:      in.Title,
		Handle:     in.Handle,
		Weight:     in.Weight,
		CreateTime: timestamppb.New(in.CreateTime),
		UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
