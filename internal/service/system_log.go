package service

import (
	"context"

	pb "crow/api/admin/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminService) ListSystemLogs(ctx context.Context, req *pb.ListSystemLogsRequest) (*pb.SystemLogSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	items, err := s.systemLogUC.List(ctx,
		biz.SystemLogListLimit(int(req.PageSize)),
		biz.SystemLogListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}

	set := &pb.SystemLogSet{
		SystemLogs: make([]*pb.SystemLog, 0, len(items)),
	}
	if len(items) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, item := range items {
		set.SystemLogs = append(set.SystemLogs, convertSystemLogReply(item))
	}
	return set, nil
}

func convertSystemLogReply(in *biz.SystemLog) *pb.SystemLog {
	if in == nil {
		return nil
	}
	return &pb.SystemLog{
		Id:         in.ID,
		LogUid:     in.LogUID,
		LogLevel:   in.LogLevel,
		Message:    in.Message,
		FilePath:   in.FilePath,
		LineNumber: in.LineNumber,
		CreateTime: timestamppb.New(in.CreateTime),
	}
}
