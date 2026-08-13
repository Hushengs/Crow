package service

import (
	"context"

	pb "crow/api/admin/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AdminService) ListAdminOperationLogs(ctx context.Context, req *pb.ListAdminOperationLogsRequest) (*pb.AdminOperationLogSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	logs, err := s.operationLogUC.List(ctx,
		biz.AdminOperationLogListLimit(int(req.PageSize)),
		biz.AdminOperationLogListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}

	set := &pb.AdminOperationLogSet{
		AdminOperationLogs: make([]*pb.AdminOperationLog, 0, len(logs)),
	}
	if len(logs) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, item := range logs {
		set.AdminOperationLogs = append(set.AdminOperationLogs, convertAdminOperationLogReply(item))
	}
	return set, nil
}

func convertAdminOperationLogReply(in *biz.AdminOperationLog) *pb.AdminOperationLog {
	if in == nil {
		return nil
	}
	return &pb.AdminOperationLog{
		Id:            in.ID,
		AdminId:       in.AdminID,
		AdminName:     in.AdminName,
		Module:        in.Module,
		Action:        in.Action,
		Description:   in.Description,
		RequestMethod: in.RequestMethod,
		RequestUrl:    in.RequestURL,
		RequestParams: in.RequestParams,
		CreateTime:    timestamppb.New(in.CreateTime),
	}
}
