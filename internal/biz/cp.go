package biz

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	v1 "crow/api/cdn/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

const (
	CpStatusDisabled uint32 = 0
	CpStatusNormal   uint32 = 1
	CpStatusFrozen   uint32 = 2

	maxCpCodeLength = 32
	maxCpNameLength = 64
)

var (
	ErrCpNotFound        = errors.NotFound(v1.ErrorReason_CP_NOT_FOUND.String(), "cp not found")
	ErrCpInvalidArgument = errors.BadRequest(v1.ErrorReason_CP_INVALID_ARGUMENT.String(), "invalid cp argument")
	ErrCpCodeConflict    = errors.Conflict(v1.ErrorReason_CP_CODE_CONFLICT.String(), "cp code already exists")
)

// Cp is the domain model for a content provider.
type Cp struct {
	ID         int64
	CpCode     string
	CpName     string
	Status     uint32
	CreateTime time.Time
	UpdateTime time.Time
}

// CpRepo is the persistence boundary for content providers.
type CpRepo interface {
	FindByID(context.Context, int64) (*Cp, error)
	ListCps(context.Context, ...CpListOption) ([]*Cp, error)
	CreateCp(context.Context, *Cp) (*Cp, error)
	UpdateCp(context.Context, *Cp) (*Cp, error)
	DeleteCp(context.Context, int64) error
}

// CpListOption configures content provider list queries.
type CpListOption func(*CpListOptions)

// CpListOptions are content provider list query options.
type CpListOptions struct {
	Offset int
	Limit  int
}

// CpListOffset sets an offset for content provider list queries.
func CpListOffset(offset int) CpListOption {
	return func(o *CpListOptions) {
		o.Offset = offset
	}
}

// CpListLimit sets a limit for content provider list queries.
func CpListLimit(limit int) CpListOption {
	return func(o *CpListOptions) {
		o.Limit = limit
	}
}

// CpUsecase is the content provider usecase.
type CpUsecase struct {
	repo CpRepo
}

// NewCpUsecase creates a new CpUsecase.
func NewCpUsecase(repo CpRepo) *CpUsecase {
	return &CpUsecase{repo: repo}
}

// GetCp returns a content provider by ID.
func (uc *CpUsecase) GetCp(ctx context.Context, id int64) (*Cp, error) {
	if id <= 0 {
		return nil, ErrCpInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

// ListCps returns a page of content providers.
func (uc *CpUsecase) ListCps(ctx context.Context, opts ...CpListOption) ([]*Cp, error) {
	return uc.repo.ListCps(ctx, opts...)
}

// CreateCp creates a new content provider.
func (uc *CpUsecase) CreateCp(ctx context.Context, cp *Cp) (*Cp, error) {
	if err := normalizeAndValidateCp(cp); err != nil {
		return nil, err
	}
	return uc.repo.CreateCp(ctx, cp)
}

// UpdateCp updates an existing content provider.
func (uc *CpUsecase) UpdateCp(ctx context.Context, cp *Cp) (*Cp, error) {
	if cp == nil || cp.ID <= 0 {
		return nil, ErrCpInvalidArgument
	}
	if err := normalizeAndValidateCp(cp); err != nil {
		return nil, err
	}
	return uc.repo.UpdateCp(ctx, cp)
}

// DeleteCp deletes a content provider by ID.
func (uc *CpUsecase) DeleteCp(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrCpInvalidArgument
	}
	return uc.repo.DeleteCp(ctx, id)
}

func normalizeAndValidateCp(cp *Cp) error {
	if cp == nil {
		return ErrCpInvalidArgument
	}
	cp.CpCode = strings.TrimSpace(cp.CpCode)
	cp.CpName = strings.TrimSpace(cp.CpName)
	if cp.CpCode == "" || cp.CpName == "" || !isValidCpStatus(cp.Status) {
		return ErrCpInvalidArgument
	}
	if utf8.RuneCountInString(cp.CpCode) > maxCpCodeLength || utf8.RuneCountInString(cp.CpName) > maxCpNameLength {
		return ErrCpInvalidArgument
	}
	return nil
}

func isValidCpStatus(status uint32) bool {
	switch status {
	case CpStatusDisabled, CpStatusNormal, CpStatusFrozen:
		return true
	default:
		return false
	}
}
