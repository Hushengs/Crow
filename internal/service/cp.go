package service

import (
	"context"

	pb "crow/api/cdn/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CpService struct {
	pb.UnimplementedCpServiceServer

	uc *biz.CpUsecase
}

func NewCpService(uc *biz.CpUsecase) *CpService {
	return &CpService{uc: uc}
}

func (s *CpService) CreateCp(ctx context.Context, req *pb.CreateCpRequest) (*pb.Cp, error) {
	cp, err := s.uc.CreateCp(ctx, convertCp(req.GetCp()))
	if err != nil {
		return nil, err
	}
	return convertCpReply(cp), nil
}

func (s *CpService) GetCp(ctx context.Context, req *pb.GetCpRequest) (*pb.Cp, error) {
	cp, err := s.uc.GetCp(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertCpReply(cp), nil
}

func (s *CpService) ListCps(ctx context.Context, req *pb.ListCpsRequest) (*pb.CpSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	cps, err := s.uc.ListCps(ctx,
		biz.CpListLimit(int(req.PageSize)),
		biz.CpListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.CpSet{
		Cps: make([]*pb.Cp, 0, len(cps)),
	}
	if len(cps) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, cp := range cps {
		set.Cps = append(set.Cps, convertCpReply(cp))
	}
	return set, nil
}

func (s *CpService) UpdateCp(ctx context.Context, req *pb.UpdateCpRequest) (*pb.Cp, error) {
	if req.GetCp().GetId() <= 0 || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, biz.ErrCpInvalidArgument
	}
	if err := validateCpMaskPaths(req.GetUpdateMask().GetPaths()); err != nil {
		return nil, err
	}
	current, err := s.uc.GetCp(ctx, req.GetCp().GetId())
	if err != nil {
		return nil, err
	}
	updated, err := applyCpUpdateMask(current, req.GetCp(), req.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	cp, err := s.uc.UpdateCp(ctx, updated)
	if err != nil {
		return nil, err
	}
	return convertCpReply(cp), nil
}

func (s *CpService) DeleteCp(ctx context.Context, req *pb.DeleteCpRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteCp(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func applyCpUpdateMask(current *biz.Cp, patch *pb.Cp, paths []string) (*biz.Cp, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrCpInvalidArgument
	}
	updated := *current
	for _, path := range expandCpMaskPaths(paths) {
		switch path {
		case "cp_code":
			updated.CpCode = patch.GetCpCode()
		case "cp_name":
			updated.CpName = patch.GetCpName()
		case "status":
			updated.Status = patch.GetStatus()
		default:
			return nil, biz.ErrCpInvalidArgument
		}
	}
	return &updated, nil
}

func validateCpMaskPaths(paths []string) error {
	for _, path := range expandCpMaskPaths(paths) {
		switch path {
		case "cp_code", "cp_name", "status":
		default:
			return biz.ErrCpInvalidArgument
		}
	}
	return nil
}

func expandCpMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "cp_code", "cp_name", "status")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertCp(in *pb.Cp) *biz.Cp {
	if in == nil {
		return nil
	}
	return &biz.Cp{
		ID:     in.GetId(),
		CpCode: in.GetCpCode(),
		CpName: in.GetCpName(),
		Status: in.GetStatus(),
	}
}

func convertCpReply(in *biz.Cp) *pb.Cp {
	if in == nil {
		return nil
	}
	return &pb.Cp{
		Id:         in.ID,
		CpCode:     in.CpCode,
		CpName:     in.CpName,
		Status:     in.Status,
		CreateTime: timestamppb.New(in.CreateTime),
		UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
