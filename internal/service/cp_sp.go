package service

import (
	"context"

	pb "crow/api/cdn/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CpSpService struct {
	pb.UnimplementedCpSpServiceServer

	uc *biz.CpSpUsecase
}

func NewCpSpService(uc *biz.CpSpUsecase) *CpSpService {
	return &CpSpService{uc: uc}
}

func (s *CpSpService) CreateCpSp(ctx context.Context, req *pb.CreateCpSpRequest) (*pb.CpSp, error) {
	relation, err := s.uc.CreateCpSp(ctx, convertCpSp(req.GetCpSp()))
	if err != nil {
		return nil, err
	}
	return convertCpSpReply(relation), nil
}

func (s *CpSpService) GetCpSp(ctx context.Context, req *pb.GetCpSpRequest) (*pb.CpSp, error) {
	relation, err := s.uc.GetCpSp(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertCpSpReply(relation), nil
}

func (s *CpSpService) ListCpSps(ctx context.Context, req *pb.ListCpSpsRequest) (*pb.CpSpSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	relations, err := s.uc.ListCpSps(ctx,
		biz.CpSpListLimit(int(req.PageSize)),
		biz.CpSpListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.CpSpSet{
		CpSps: make([]*pb.CpSp, 0, len(relations)),
	}
	if len(relations) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, relation := range relations {
		set.CpSps = append(set.CpSps, convertCpSpReply(relation))
	}
	return set, nil
}

func (s *CpSpService) UpdateCpSp(ctx context.Context, req *pb.UpdateCpSpRequest) (*pb.CpSp, error) {
	if req.GetCpSp().GetId() <= 0 || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, biz.ErrCpSpInvalidArgument
	}
	if err := validateCpSpMaskPaths(req.GetUpdateMask().GetPaths()); err != nil {
		return nil, err
	}
	current, err := s.uc.GetCpSp(ctx, req.GetCpSp().GetId())
	if err != nil {
		return nil, err
	}
	updated, err := applyCpSpUpdateMask(current, req.GetCpSp(), req.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	relation, err := s.uc.UpdateCpSp(ctx, updated)
	if err != nil {
		return nil, err
	}
	return convertCpSpReply(relation), nil
}

func (s *CpSpService) DeleteCpSp(ctx context.Context, req *pb.DeleteCpSpRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteCpSp(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func applyCpSpUpdateMask(current *biz.CpSp, patch *pb.CpSp, paths []string) (*biz.CpSp, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrCpSpInvalidArgument
	}
	updated := *current
	for _, path := range expandCpSpMaskPaths(paths) {
		switch path {
		case "cp_id":
			updated.CpID = patch.GetCpId()
		case "sp_id":
			updated.SpID = patch.GetSpId()
		case "status":
			updated.Status = patch.GetStatus()
		default:
			return nil, biz.ErrCpSpInvalidArgument
		}
	}
	return &updated, nil
}

func validateCpSpMaskPaths(paths []string) error {
	for _, path := range expandCpSpMaskPaths(paths) {
		switch path {
		case "cp_id", "sp_id", "status":
		default:
			return biz.ErrCpSpInvalidArgument
		}
	}
	return nil
}

func expandCpSpMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "cp_id", "sp_id", "status")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertCpSp(in *pb.CpSp) *biz.CpSp {
	if in == nil {
		return nil
	}
	return &biz.CpSp{
		ID:     in.GetId(),
		CpID:   in.GetCpId(),
		SpID:   in.GetSpId(),
		Status: in.GetStatus(),
	}
}

func convertCpSpReply(in *biz.CpSp) *pb.CpSp {
	if in == nil {
		return nil
	}
	return &pb.CpSp{
		Id:         in.ID,
		CpId:       in.CpID,
		SpId:       in.SpID,
		Status:     in.Status,
		CreateTime: timestamppb.New(in.CreateTime),
		UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
