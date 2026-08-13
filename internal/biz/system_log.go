package biz

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	ErrSystemLogInvalidArgument = errors.BadRequest("SYSTEM_LOG_INVALID_ARGUMENT", "invalid system log argument")
)

// SystemLog is the domain model for backend system logs.
type SystemLog struct {
	ID         int64
	LogUID     string
	LogLevel   string
	Message    string
	FilePath   string
	LineNumber uint32
	CreateTime time.Time
}

// SystemLogRepo persists system logs.
type SystemLogRepo interface {
	Create(context.Context, *SystemLog) (*SystemLog, error)
	List(context.Context, ...SystemLogListOption) ([]*SystemLog, error)
}

// SystemLogListOption configures system log list queries.
type SystemLogListOption func(*SystemLogListOptions)

// SystemLogListOptions are system log list query options.
type SystemLogListOptions struct {
	Offset int
	Limit  int
}

// SystemLogListOffset sets an offset for system log queries.
func SystemLogListOffset(offset int) SystemLogListOption {
	return func(o *SystemLogListOptions) {
		o.Offset = offset
	}
}

// SystemLogListLimit sets a limit for system log queries.
func SystemLogListLimit(limit int) SystemLogListOption {
	return func(o *SystemLogListOptions) {
		o.Limit = limit
	}
}

// SystemLogUsecase writes and lists backend system logs.
type SystemLogUsecase struct {
	repo SystemLogRepo
}

// NewSystemLogUsecase creates a new SystemLogUsecase.
func NewSystemLogUsecase(repo SystemLogRepo) *SystemLogUsecase {
	return &SystemLogUsecase{repo: repo}
}

// Create stores a system log.
func (uc *SystemLogUsecase) Create(ctx context.Context, item *SystemLog) (*SystemLog, error) {
	if item == nil {
		return nil, ErrSystemLogInvalidArgument
	}
	item.LogUID = strings.TrimSpace(item.LogUID)
	item.LogLevel = strings.TrimSpace(strings.ToUpper(item.LogLevel))
	item.Message = strings.TrimSpace(item.Message)
	item.FilePath = strings.TrimSpace(item.FilePath)
	if item.LogUID == "" || item.LogLevel == "" || item.Message == "" {
		return nil, ErrSystemLogInvalidArgument
	}
	return uc.repo.Create(ctx, item)
}

// List returns a page of system logs.
func (uc *SystemLogUsecase) List(ctx context.Context, opts ...SystemLogListOption) ([]*SystemLog, error) {
	return uc.repo.List(ctx, opts...)
}
