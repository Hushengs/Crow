package biz

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	ErrAdminOperationLogInvalidArgument = errors.BadRequest("ADMIN_OPERATION_LOG_INVALID_ARGUMENT", "invalid admin operation log argument")
)

// AdminOperationLog is the domain model for administrator operation logs.
type AdminOperationLog struct {
	ID            int64
	AdminID       int64
	AdminName     string
	Module        string
	Action        string
	Description   string
	RequestMethod string
	RequestURL    string
	RequestParams string
	CreateTime    time.Time
}

// AdminOperationLogListOption configures operation log list queries.
type AdminOperationLogListOption func(*AdminOperationLogListOptions)

// AdminOperationLogListOptions are log list query options.
type AdminOperationLogListOptions struct {
	Offset int
	Limit  int
}

// AdminOperationLogListOffset sets an offset for log list queries.
func AdminOperationLogListOffset(offset int) AdminOperationLogListOption {
	return func(o *AdminOperationLogListOptions) {
		o.Offset = offset
	}
}

// AdminOperationLogListLimit sets a limit for log list queries.
func AdminOperationLogListLimit(limit int) AdminOperationLogListOption {
	return func(o *AdminOperationLogListOptions) {
		o.Limit = limit
	}
}

// AdminOperationLogRepo persists operation logs.
type AdminOperationLogRepo interface {
	Create(context.Context, *AdminOperationLog) (*AdminOperationLog, error)
	List(context.Context, ...AdminOperationLogListOption) ([]*AdminOperationLog, error)
}

// AdminOperationLogUsecase writes administrator operation logs.
type AdminOperationLogUsecase struct {
	repo AdminOperationLogRepo
}

// NewAdminOperationLogUsecase creates a new AdminOperationLogUsecase.
func NewAdminOperationLogUsecase(repo AdminOperationLogRepo) *AdminOperationLogUsecase {
	return &AdminOperationLogUsecase{repo: repo}
}

// Create writes an operation log.
func (uc *AdminOperationLogUsecase) Create(ctx context.Context, log *AdminOperationLog) (*AdminOperationLog, error) {
	if log == nil {
		return nil, ErrAdminOperationLogInvalidArgument
	}
	log.AdminName = strings.TrimSpace(log.AdminName)
	log.Module = strings.TrimSpace(log.Module)
	log.Action = strings.TrimSpace(log.Action)
	log.Description = strings.TrimSpace(log.Description)
	log.RequestMethod = strings.TrimSpace(strings.ToUpper(log.RequestMethod))
	log.RequestURL = strings.TrimSpace(log.RequestURL)
	log.RequestParams = strings.TrimSpace(log.RequestParams)

	if log.AdminID <= 0 || log.AdminName == "" || log.Module == "" || log.Action == "" || log.RequestMethod == "" || log.RequestURL == "" {
		return nil, ErrAdminOperationLogInvalidArgument
	}
	return uc.repo.Create(ctx, log)
}

// List returns a page of operation logs.
func (uc *AdminOperationLogUsecase) List(ctx context.Context, opts ...AdminOperationLogListOption) ([]*AdminOperationLog, error) {
	return uc.repo.List(ctx, opts...)
}
