package service

import (
	"context"

	pb "crow/api/inject/v1"
	"crow/internal/biz"

	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InjectService struct {
	pb.UnimplementedInjectServiceServer
	uc *biz.InjectUsecase
}

func NewInjectService(uc *biz.InjectUsecase) *InjectService {
	return &InjectService{uc: uc}
}

func (s *InjectService) ListInjectContents(ctx context.Context, req *pb.ListInjectContentsRequest) (*pb.InjectContentSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	items, err := s.uc.ListContents(ctx, biz.InjectListLimit(int(req.PageSize)), biz.InjectListOffset(int(pageToken.Offset)))
	if err != nil {
		return nil, err
	}
	set := &pb.InjectContentSet{Contents: make([]*pb.InjectContent, 0, len(items))}
	if len(items) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, item := range items {
		set.Contents = append(set.Contents, injectContentReply(item))
	}
	return set, nil
}

func (s *InjectService) GetInjectContent(ctx context.Context, req *pb.GetInjectContentRequest) (*pb.InjectContent, error) {
	item, err := s.uc.GetContent(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return injectContentReply(item), nil
}

func (s *InjectService) RetryInjectContent(ctx context.Context, req *pb.RetryInjectContentRequest) (*pb.InjectContent, error) {
	item, err := s.uc.RetryContent(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return injectContentReply(item), nil
}

func (s *InjectService) ListInjectTasks(ctx context.Context, req *pb.ListInjectTasksRequest) (*pb.InjectTaskSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	items, err := s.uc.ListTasks(ctx, biz.InjectListLimit(int(req.PageSize)), biz.InjectListOffset(int(pageToken.Offset)))
	if err != nil {
		return nil, err
	}
	set := &pb.InjectTaskSet{Tasks: make([]*pb.InjectTask, 0, len(items))}
	if len(items) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, item := range items {
		set.Tasks = append(set.Tasks, injectTaskReply(item))
	}
	return set, nil
}

func (s *InjectService) GetInjectTask(ctx context.Context, req *pb.GetInjectTaskRequest) (*pb.InjectTask, error) {
	item, err := s.uc.GetTask(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return injectTaskReply(item), nil
}

func (s *InjectService) ListInjectLogs(ctx context.Context, req *pb.ListInjectLogsRequest) (*pb.InjectLogSet, error) {
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	items, err := s.uc.ListLogs(ctx, biz.InjectListLimit(int(req.PageSize)), biz.InjectListOffset(int(pageToken.Offset)))
	if err != nil {
		return nil, err
	}
	set := &pb.InjectLogSet{Logs: make([]*pb.InjectLog, 0, len(items))}
	if len(items) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, item := range items {
		set.Logs = append(set.Logs, injectLogReply(item))
	}
	return set, nil
}

func injectContentReply(in *biz.InjectContent) *pb.InjectContent {
	return &pb.InjectContent{
		Id: in.ID, ResourceType: in.ResourceType, ResourceId: in.ResourceID,
		VideoId: in.VideoID, EpisodeId: in.EpisodeID, MediaId: in.MediaID,
		Action: in.Action, LastTaskId: in.LastTaskID, Status: in.Status, FailReason: in.FailReason,
		CreateTime: timestamppb.New(in.CreateTime), UpdateTime: timestamppb.New(in.UpdateTime),
	}
}

func injectTaskReply(in *biz.InjectTask) *pb.InjectTask {
	return &pb.InjectTask{
		Id: in.ID, ContentId: in.ContentID, ResourceType: in.ResourceType, ResourceId: in.ResourceID,
		VideoId: in.VideoID, EpisodeId: in.EpisodeID, MediaId: in.MediaID, Action: in.Action, Status: in.Status,
		CreateTime: timestamppb.New(in.CreateTime), UpdateTime: timestamppb.New(in.UpdateTime),
	}
}

func injectLogReply(in *biz.InjectLog) *pb.InjectLog {
	return &pb.InjectLog{
		Id: in.ID, TaskId: in.TaskID, ContentId: in.ContentID, SpId: in.SpID,
		ResourceType: in.ResourceType, ResourceId: in.ResourceID, VideoId: in.VideoID, EpisodeId: in.EpisodeID, MediaId: in.MediaID,
		Action: in.Action, CorrelateId: in.CorrelateID, SyncStatus: in.SyncStatus, AsyncStatus: in.AsyncStatus,
		SyncMessage: in.SyncMessage, AsyncMessage: in.AsyncMessage,
		CreateTime: timestamppb.New(in.CreateTime), UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
