package service

import (
	"context"

	pb "crow/api/cdn/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SpService struct {
	pb.UnimplementedSpServiceServer

	uc *biz.SpUsecase
}

func NewSpService(uc *biz.SpUsecase) *SpService {
	return &SpService{uc: uc}
}

func (s *SpService) CreateSp(ctx context.Context, req *pb.CreateSpRequest) (*pb.Sp, error) {
	sp, err := s.uc.CreateSp(ctx, convertSp(req.GetSp()))
	if err != nil {
		return nil, err
	}
	return convertSpReply(sp), nil
}

func (s *SpService) GetSp(ctx context.Context, req *pb.GetSpRequest) (*pb.Sp, error) {
	sp, err := s.uc.GetSp(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertSpReply(sp), nil
}

func (s *SpService) ListSps(ctx context.Context, req *pb.ListSpsRequest) (*pb.SpSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	sps, err := s.uc.ListSps(ctx,
		biz.SpListLimit(int(req.PageSize)),
		biz.SpListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &pb.SpSet{
		Sps: make([]*pb.Sp, 0, len(sps)),
	}
	if len(sps) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, sp := range sps {
		set.Sps = append(set.Sps, convertSpReply(sp))
	}
	return set, nil
}

func (s *SpService) UpdateSp(ctx context.Context, req *pb.UpdateSpRequest) (*pb.Sp, error) {
	if req.GetSp().GetId() <= 0 || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, biz.ErrSpInvalidArgument
	}
	if err := validateSpMaskPaths(req.GetUpdateMask().GetPaths()); err != nil {
		return nil, err
	}
	current, err := s.uc.GetSp(ctx, req.GetSp().GetId())
	if err != nil {
		return nil, err
	}
	updated, err := applySpUpdateMask(current, req.GetSp(), req.GetUpdateMask().GetPaths())
	if err != nil {
		return nil, err
	}
	sp, err := s.uc.UpdateSp(ctx, updated)
	if err != nil {
		return nil, err
	}
	return convertSpReply(sp), nil
}

func (s *SpService) DeleteSp(ctx context.Context, req *pb.DeleteSpRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteSp(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func applySpUpdateMask(current *biz.Sp, patch *pb.Sp, paths []string) (*biz.Sp, error) {
	if current == nil || patch == nil {
		return nil, biz.ErrSpInvalidArgument
	}
	updated := *current
	for _, path := range expandSpMaskPaths(paths) {
		switch path {
		case "sp_code":
			updated.SpCode = patch.GetSpCode()
		case "sp_name":
			updated.SpName = patch.GetSpName()
		case "sp_config":
			updated.SpConfig = patch.GetSpConfig()
		case "status":
			updated.Status = patch.GetStatus()
		default:
			return nil, biz.ErrSpInvalidArgument
		}
	}
	return &updated, nil
}

func validateSpMaskPaths(paths []string) error {
	for _, path := range expandSpMaskPaths(paths) {
		switch path {
		case "sp_code", "sp_name", "sp_config", "status":
		default:
			return biz.ErrSpInvalidArgument
		}
	}
	return nil
}

func expandSpMaskPaths(paths []string) []string {
	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "*" {
			expanded = append(expanded, "sp_code", "sp_name", "sp_config", "status")
			continue
		}
		expanded = append(expanded, path)
	}
	return expanded
}

func convertSp(in *pb.Sp) *biz.Sp {
	if in == nil {
		return nil
	}
	return &biz.Sp{
		ID:       in.GetId(),
		SpCode:   in.GetSpCode(),
		SpName:   in.GetSpName(),
		SpConfig: in.GetSpConfig(),
		Status:   in.GetStatus(),
	}
}

func convertSpReply(in *biz.Sp) *pb.Sp {
	if in == nil {
		return nil
	}
	return &pb.Sp{
		Id:         in.ID,
		SpCode:     in.SpCode,
		SpName:     in.SpName,
		SpConfig:   in.SpConfig,
		Status:     in.Status,
		CreateTime: timestamppb.New(in.CreateTime),
		UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
