package biz

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	v1 "crow/api/cdn/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

const (
	SpStatusDisabled uint32 = 0
	SpStatusNormal   uint32 = 1
	SpStatusFrozen   uint32 = 2

	maxSpCodeLength = 32
	maxSpNameLength = 64
)

var (
	ErrSpNotFound        = errors.NotFound(v1.ErrorReason_SP_NOT_FOUND.String(), "sp not found")
	ErrSpInvalidArgument = errors.BadRequest(v1.ErrorReason_SP_INVALID_ARGUMENT.String(), "invalid sp argument")
	ErrSpCodeConflict    = errors.Conflict(v1.ErrorReason_SP_CODE_CONFLICT.String(), "sp code already exists")
)

// Sp is the domain model for a content service provider.
type Sp struct {
	ID         int64
	SpCode     string
	SpName     string
	SpConfig   string
	Status     uint32
	CreateTime time.Time
	UpdateTime time.Time
}

// SpRepo is the persistence boundary for content service providers.
type SpRepo interface {
	FindByID(context.Context, int64) (*Sp, error)
	ListSps(context.Context, ...SpListOption) ([]*Sp, error)
	CreateSp(context.Context, *Sp) (*Sp, error)
	UpdateSp(context.Context, *Sp) (*Sp, error)
	DeleteSp(context.Context, int64) error
}

// SpListOption configures content service provider list queries.
type SpListOption func(*SpListOptions)

// SpListOptions are content service provider list query options.
type SpListOptions struct {
	Offset int
	Limit  int
}

// SpListOffset sets an offset for content service provider list queries.
func SpListOffset(offset int) SpListOption {
	return func(o *SpListOptions) {
		o.Offset = offset
	}
}

// SpListLimit sets a limit for content service provider list queries.
func SpListLimit(limit int) SpListOption {
	return func(o *SpListOptions) {
		o.Limit = limit
	}
}

// SpUsecase is the content service provider usecase.
type SpUsecase struct {
	repo SpRepo
}

// NewSpUsecase creates a new SpUsecase.
func NewSpUsecase(repo SpRepo) *SpUsecase {
	return &SpUsecase{repo: repo}
}

// GetSp returns a content service provider by ID.
func (uc *SpUsecase) GetSp(ctx context.Context, id int64) (*Sp, error) {
	if id <= 0 {
		return nil, ErrSpInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

// ListSps returns a page of content service providers.
func (uc *SpUsecase) ListSps(ctx context.Context, opts ...SpListOption) ([]*Sp, error) {
	return uc.repo.ListSps(ctx, opts...)
}

// CreateSp creates a new content service provider.
func (uc *SpUsecase) CreateSp(ctx context.Context, sp *Sp) (*Sp, error) {
	if err := normalizeAndValidateSp(sp); err != nil {
		return nil, err
	}
	return uc.repo.CreateSp(ctx, sp)
}

// UpdateSp updates an existing content service provider.
func (uc *SpUsecase) UpdateSp(ctx context.Context, sp *Sp) (*Sp, error) {
	if sp == nil || sp.ID <= 0 {
		return nil, ErrSpInvalidArgument
	}
	if err := normalizeAndValidateSp(sp); err != nil {
		return nil, err
	}
	return uc.repo.UpdateSp(ctx, sp)
}

// DeleteSp deletes a content service provider by ID.
func (uc *SpUsecase) DeleteSp(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrSpInvalidArgument
	}
	return uc.repo.DeleteSp(ctx, id)
}

func normalizeAndValidateSp(sp *Sp) error {
	if sp == nil {
		return ErrSpInvalidArgument
	}
	sp.SpCode = strings.TrimSpace(sp.SpCode)
	sp.SpName = strings.TrimSpace(sp.SpName)
	sp.SpConfig = strings.TrimSpace(sp.SpConfig)
	if sp.SpCode == "" || sp.SpName == "" || !isValidSpStatus(sp.Status) {
		return ErrSpInvalidArgument
	}
	if utf8.RuneCountInString(sp.SpCode) > maxSpCodeLength || utf8.RuneCountInString(sp.SpName) > maxSpNameLength {
		return ErrSpInvalidArgument
	}
	if sp.SpConfig != "" {
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(sp.SpConfig), &raw); err != nil {
			return ErrSpInvalidArgument
		}
		sp.SpConfig = string(raw)
	}
	return nil
}

func isValidSpStatus(status uint32) bool {
	switch status {
	case SpStatusDisabled, SpStatusNormal, SpStatusFrozen:
		return true
	default:
		return false
	}
}
