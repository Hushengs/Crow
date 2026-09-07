package biz

import (
	"context"
	"time"

	v1 "crow/api/inject/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

const (
	InjectResourceVideo   uint32 = 1
	InjectResourceEpisode uint32 = 2
	InjectResourceMedia   uint32 = 3

	InjectActionCreate uint32 = 1
	InjectActionUpdate uint32 = 2
	InjectActionDelete uint32 = 3

	InjectTaskPending    uint32 = 0
	InjectTaskProcessing uint32 = 1

	InjectContentWaiting    uint32 = 0
	InjectContentProcessing uint32 = 1
	InjectContentSuccess    uint32 = 2
	InjectContentFailed     uint32 = 3

	InjectLogWaiting uint32 = 0
	InjectLogSuccess uint32 = 1
	InjectLogFailed  uint32 = 2
)

var (
	ErrInjectNotFound        = errors.NotFound(v1.ErrorReason_INJECT_NOT_FOUND.String(), "inject resource not found")
	ErrInjectInvalidArgument = errors.BadRequest(v1.ErrorReason_INJECT_INVALID_ARGUMENT.String(), "invalid inject argument")
	ErrInjectNoPending       = errors.NotFound(v1.ErrorReason_INJECT_NOT_FOUND.String(), "no pending inject task")
)

type InjectContent struct {
	ID, ResourceID, VideoID, EpisodeID, MediaID, LastTaskID int64
	ResourceType, Action, Status                            uint32
	FailReason                                              string
	CreateTime, UpdateTime                                  time.Time
}

type InjectTask struct {
	ID, ContentID, ResourceID, VideoID, EpisodeID, MediaID int64
	ResourceType, Action, Status                           uint32
	CreateTime, UpdateTime                                 time.Time
}

type InjectLog struct {
	ID, TaskID, ContentID, SpID, ResourceID, VideoID, EpisodeID, MediaID int64
	ResourceType, Action, SyncStatus, AsyncStatus                        uint32
	CorrelateID, SyncMessage, AsyncMessage                               string
	CreateTime, UpdateTime                                               time.Time
}

type InjectEnqueue struct {
	ResourceType, Action                    uint32
	ResourceID, VideoID, EpisodeID, MediaID int64
}

type InjectRepo interface {
	Enqueue(context.Context, *InjectEnqueue) (*InjectContent, *InjectTask, error)
	FindContent(context.Context, int64) (*InjectContent, error)
	ListContents(context.Context, ...InjectListOption) ([]*InjectContent, error)
	FindTask(context.Context, int64) (*InjectTask, error)
	ListTasks(context.Context, ...InjectListOption) ([]*InjectTask, error)
	ListLogs(context.Context, ...InjectListOption) ([]*InjectLog, error)
	ClaimNextTask(context.Context) (*InjectTask, error)
	DeleteTask(context.Context, int64) error
	UpdateContentStatus(context.Context, int64, uint32, string) error
	ListActiveSpIDsForVideo(context.Context, int64) ([]int64, error)
	CreateLog(context.Context, *InjectLog) error
}

type InjectListOption func(*InjectListOptions)

type InjectListOptions struct {
	Offset int
	Limit  int
}

func InjectListOffset(offset int) InjectListOption {
	return func(o *InjectListOptions) { o.Offset = offset }
}

func InjectListLimit(limit int) InjectListOption {
	return func(o *InjectListOptions) { o.Limit = limit }
}

type InjectUsecase struct {
	repo InjectRepo
}

func NewInjectUsecase(repo InjectRepo) *InjectUsecase {
	return &InjectUsecase{repo: repo}
}

func (uc *InjectUsecase) Enqueue(ctx context.Context, job *InjectEnqueue) (*InjectContent, error) {
	if err := validateInjectEnqueue(job); err != nil {
		return nil, err
	}
	content, _, err := uc.repo.Enqueue(ctx, job)
	return content, err
}

func (uc *InjectUsecase) GetContent(ctx context.Context, id int64) (*InjectContent, error) {
	if id <= 0 {
		return nil, ErrInjectInvalidArgument
	}
	return uc.repo.FindContent(ctx, id)
}

func (uc *InjectUsecase) ListContents(ctx context.Context, opts ...InjectListOption) ([]*InjectContent, error) {
	return uc.repo.ListContents(ctx, opts...)
}

func (uc *InjectUsecase) GetTask(ctx context.Context, id int64) (*InjectTask, error) {
	if id <= 0 {
		return nil, ErrInjectInvalidArgument
	}
	return uc.repo.FindTask(ctx, id)
}

func (uc *InjectUsecase) ListTasks(ctx context.Context, opts ...InjectListOption) ([]*InjectTask, error) {
	return uc.repo.ListTasks(ctx, opts...)
}

func (uc *InjectUsecase) ListLogs(ctx context.Context, opts ...InjectListOption) ([]*InjectLog, error) {
	return uc.repo.ListLogs(ctx, opts...)
}

func (uc *InjectUsecase) RetryContent(ctx context.Context, id int64) (*InjectContent, error) {
	content, err := uc.GetContent(ctx, id)
	if err != nil {
		return nil, err
	}
	return uc.Enqueue(ctx, &InjectEnqueue{
		ResourceType: content.ResourceType,
		ResourceID:   content.ResourceID,
		VideoID:      content.VideoID,
		EpisodeID:    content.EpisodeID,
		MediaID:      content.MediaID,
		Action:       content.Action,
	})
}

func (uc *InjectUsecase) DispatchOnce(ctx context.Context) error {
	task, err := uc.repo.ClaimNextTask(ctx)
	if err != nil {
		if errors.Is(err, ErrInjectNoPending) {
			return nil
		}
		return err
	}
	if err := uc.repo.UpdateContentStatus(ctx, task.ContentID, InjectContentProcessing, ""); err != nil {
		return err
	}
	spIDs, err := uc.repo.ListActiveSpIDsForVideo(ctx, task.VideoID)
	if err != nil {
		_ = uc.repo.UpdateContentStatus(ctx, task.ContentID, InjectContentFailed, err.Error())
		_ = uc.repo.DeleteTask(ctx, task.ID)
		return err
	}
	if len(spIDs) == 0 {
		reason := "no active SP bound to this video's content provider"
		_ = uc.repo.UpdateContentStatus(ctx, task.ContentID, InjectContentFailed, reason)
		return uc.repo.DeleteTask(ctx, task.ID)
	}
	failed := false
	var failReason string
	for _, spID := range spIDs {
		logRow := &InjectLog{
			TaskID:       task.ID,
			ContentID:    task.ContentID,
			SpID:         spID,
			ResourceType: task.ResourceType,
			ResourceID:   task.ResourceID,
			VideoID:      task.VideoID,
			EpisodeID:    task.EpisodeID,
			MediaID:      task.MediaID,
			Action:       task.Action,
			CorrelateID:  "",
			SyncStatus:   InjectLogSuccess,
			AsyncStatus:  InjectLogSuccess,
			SyncMessage:  "accepted",
			AsyncMessage: "injected",
		}
		if err := uc.repo.CreateLog(ctx, logRow); err != nil {
			failed = true
			failReason = err.Error()
		}
	}
	status := InjectContentSuccess
	reason := ""
	if failed {
		status = InjectContentFailed
		reason = failReason
	}
	if err := uc.repo.UpdateContentStatus(ctx, task.ContentID, status, reason); err != nil {
		return err
	}
	return uc.repo.DeleteTask(ctx, task.ID)
}

func validateInjectEnqueue(job *InjectEnqueue) error {
	if job == nil || job.ResourceID <= 0 {
		return ErrInjectInvalidArgument
	}
	switch job.Action {
	case InjectActionCreate, InjectActionUpdate, InjectActionDelete:
	default:
		return ErrInjectInvalidArgument
	}
	switch job.ResourceType {
	case InjectResourceVideo:
		if job.VideoID != job.ResourceID || job.EpisodeID != 0 || job.MediaID != 0 {
			return ErrInjectInvalidArgument
		}
	case InjectResourceEpisode:
		if job.EpisodeID != job.ResourceID || job.VideoID <= 0 || job.MediaID != 0 {
			return ErrInjectInvalidArgument
		}
	case InjectResourceMedia:
		if job.MediaID != job.ResourceID || job.VideoID <= 0 || job.EpisodeID <= 0 {
			return ErrInjectInvalidArgument
		}
	default:
		return ErrInjectInvalidArgument
	}
	return nil
}
