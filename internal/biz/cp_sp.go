package biz

import (
	"context"
	"time"

	v1 "crow/api/cdn/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

const (
	CpSpStatusDisabled uint32 = 0
	CpSpStatusNormal   uint32 = 1
)

var (
	ErrCpSpNotFound        = errors.NotFound(v1.ErrorReason_CP_SP_NOT_FOUND.String(), "cp sp relation not found")
	ErrCpSpInvalidArgument = errors.BadRequest(v1.ErrorReason_CP_SP_INVALID_ARGUMENT.String(), "invalid cp sp argument")
	ErrCpSpConflict        = errors.Conflict(v1.ErrorReason_CP_SP_CONFLICT.String(), "cp sp relation already exists")
)

// CpSp is the domain model for a content provider to content service provider binding.
type CpSp struct {
	ID         int64
	CpID       int64
	SpID       int64
	Status     uint32
	CreateTime time.Time
	UpdateTime time.Time
}

// CpSpRepo is the persistence boundary for CP-SP bindings.
type CpSpRepo interface {
	FindByID(context.Context, int64) (*CpSp, error)
	ListCpSps(context.Context, ...CpSpListOption) ([]*CpSp, error)
	CreateCpSp(context.Context, *CpSp) (*CpSp, error)
	UpdateCpSp(context.Context, *CpSp) (*CpSp, error)
	DeleteCpSp(context.Context, int64) error
}

// CpSpListOption configures CP-SP list queries.
type CpSpListOption func(*CpSpListOptions)

// CpSpListOptions are CP-SP list query options.
type CpSpListOptions struct {
	Offset int
	Limit  int
}

// CpSpListOffset sets an offset for CP-SP list queries.
func CpSpListOffset(offset int) CpSpListOption {
	return func(o *CpSpListOptions) {
		o.Offset = offset
	}
}

// CpSpListLimit sets a limit for CP-SP list queries.
func CpSpListLimit(limit int) CpSpListOption {
	return func(o *CpSpListOptions) {
		o.Limit = limit
	}
}

// CpSpUsecase is the CP-SP binding usecase.
type CpSpUsecase struct {
	repo CpSpRepo
}

// NewCpSpUsecase creates a new CpSpUsecase.
func NewCpSpUsecase(repo CpSpRepo) *CpSpUsecase {
	return &CpSpUsecase{repo: repo}
}

// GetCpSp returns a CP-SP binding by ID.
func (uc *CpSpUsecase) GetCpSp(ctx context.Context, id int64) (*CpSp, error) {
	if id <= 0 {
		return nil, ErrCpSpInvalidArgument
	}
	return uc.repo.FindByID(ctx, id)
}

// ListCpSps returns a page of CP-SP bindings.
func (uc *CpSpUsecase) ListCpSps(ctx context.Context, opts ...CpSpListOption) ([]*CpSp, error) {
	return uc.repo.ListCpSps(ctx, opts...)
}

// CreateCpSp creates a new CP-SP binding.
func (uc *CpSpUsecase) CreateCpSp(ctx context.Context, relation *CpSp) (*CpSp, error) {
	if err := validateCpSp(relation); err != nil {
		return nil, err
	}
	return uc.repo.CreateCpSp(ctx, relation)
}

// UpdateCpSp updates an existing CP-SP binding.
func (uc *CpSpUsecase) UpdateCpSp(ctx context.Context, relation *CpSp) (*CpSp, error) {
	if relation == nil || relation.ID <= 0 {
		return nil, ErrCpSpInvalidArgument
	}
	if err := validateCpSp(relation); err != nil {
		return nil, err
	}
	return uc.repo.UpdateCpSp(ctx, relation)
}

// DeleteCpSp deletes a CP-SP binding by ID.
func (uc *CpSpUsecase) DeleteCpSp(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrCpSpInvalidArgument
	}
	return uc.repo.DeleteCpSp(ctx, id)
}

func validateCpSp(relation *CpSp) error {
	if relation == nil || relation.CpID <= 0 || relation.SpID <= 0 || !isValidCpSpStatus(relation.Status) {
		return ErrCpSpInvalidArgument
	}
	return nil
}

func isValidCpSpStatus(status uint32) bool {
	switch status {
	case CpSpStatusDisabled, CpSpStatusNormal:
		return true
	default:
		return false
	}
}
