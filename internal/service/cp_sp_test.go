package service

import (
	"context"
	"sort"
	"testing"
	"time"

	v1 "crow/api/cdn/v1"
	"crow/internal/biz"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type stubCpSpRepo struct {
	nextID int64
	cpSps  map[int64]*biz.CpSp
}

func newTestCpSpService() *CpSpService {
	return NewCpSpService(biz.NewCpSpUsecase(&stubCpSpRepo{
		nextID: 1,
		cpSps:  make(map[int64]*biz.CpSp),
	}))
}

func (r *stubCpSpRepo) FindByID(_ context.Context, id int64) (*biz.CpSp, error) {
	relation, ok := r.cpSps[id]
	if !ok {
		return nil, biz.ErrCpSpNotFound
	}
	return cloneCpSpForTest(relation), nil
}

func (r *stubCpSpRepo) ListCpSps(_ context.Context, opts ...biz.CpSpListOption) ([]*biz.CpSp, error) {
	options := biz.CpSpListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrCpSpInvalidArgument
	}

	list := make([]*biz.CpSp, 0, len(r.cpSps))
	for _, relation := range r.cpSps {
		list = append(list, cloneCpSpForTest(relation))
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID > list[j].ID
	})

	if options.Offset >= len(list) {
		return []*biz.CpSp{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubCpSpRepo) CreateCpSp(_ context.Context, relation *biz.CpSp) (*biz.CpSp, error) {
	for _, existing := range r.cpSps {
		if existing.CpID == relation.CpID && existing.SpID == relation.SpID {
			return nil, biz.ErrCpSpConflict
		}
	}
	now := time.Now()
	created := cloneCpSpForTest(relation)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.cpSps[created.ID] = cloneCpSpForTest(created)
	r.nextID++
	return cloneCpSpForTest(created), nil
}

func (r *stubCpSpRepo) UpdateCpSp(_ context.Context, relation *biz.CpSp) (*biz.CpSp, error) {
	current, ok := r.cpSps[relation.ID]
	if !ok {
		return nil, biz.ErrCpSpNotFound
	}
	for _, existing := range r.cpSps {
		if existing.ID != relation.ID && existing.CpID == relation.CpID && existing.SpID == relation.SpID {
			return nil, biz.ErrCpSpConflict
		}
	}
	updated := cloneCpSpForTest(relation)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.cpSps[updated.ID] = cloneCpSpForTest(updated)
	return cloneCpSpForTest(updated), nil
}

func (r *stubCpSpRepo) DeleteCpSp(_ context.Context, id int64) error {
	if _, ok := r.cpSps[id]; !ok {
		return biz.ErrCpSpNotFound
	}
	delete(r.cpSps, id)
	return nil
}

func TestCpSpServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestCpSpService()

	created, err := svc.CreateCpSp(ctx, &v1.CreateCpSpRequest{
		CpSp: &v1.CpSp{CpId: 1, SpId: 2, Status: biz.CpSpStatusNormal},
	})
	if err != nil {
		t.Fatalf("CreateCpSp() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreateCpSp() id = %d, want 1", created.GetId())
	}

	got, err := svc.GetCpSp(ctx, &v1.GetCpSpRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetCpSp() error = %v", err)
	}
	if got.GetCpId() != 1 || got.GetSpId() != 2 || got.GetStatus() != biz.CpSpStatusNormal {
		t.Fatalf("GetCpSp() = %+v, want cp_id=1 sp_id=2 status=1", got)
	}

	list, err := svc.ListCpSps(ctx, &v1.ListCpSpsRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("ListCpSps() error = %v", err)
	}
	if len(list.GetCpSps()) != 1 {
		t.Fatalf("ListCpSps() len = %d, want 1", len(list.GetCpSps()))
	}

	updated, err := svc.UpdateCpSp(ctx, &v1.UpdateCpSpRequest{
		CpSp:       &v1.CpSp{Id: created.GetId(), SpId: 3, Status: biz.CpSpStatusDisabled},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sp_id", "status"}},
	})
	if err != nil {
		t.Fatalf("UpdateCpSp() error = %v", err)
	}
	if updated.GetCpId() != 1 || updated.GetSpId() != 3 || updated.GetStatus() != biz.CpSpStatusDisabled {
		t.Fatalf("UpdateCpSp() = %+v, want cp_id=1 sp_id=3 status=0", updated)
	}

	if _, err := svc.DeleteCpSp(ctx, &v1.DeleteCpSpRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteCpSp() error = %v", err)
	}
	if _, err := svc.GetCpSp(ctx, &v1.GetCpSpRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetCpSp() after delete error = %v, want not found", err)
	}
}

func TestCpSpServiceValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestCpSpService()

	if _, err := svc.CreateCpSp(ctx, &v1.CreateCpSpRequest{
		CpSp: &v1.CpSp{CpId: 1, Status: biz.CpSpStatusNormal},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateCpSp(missing sp_id) error = %v, want bad request", err)
	}
	if _, err := svc.CreateCpSp(ctx, &v1.CreateCpSpRequest{
		CpSp: &v1.CpSp{CpId: 1, SpId: 2, Status: 9},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateCpSp(invalid status) error = %v, want bad request", err)
	}

	if _, err := svc.CreateCpSp(ctx, &v1.CreateCpSpRequest{
		CpSp: &v1.CpSp{CpId: 1, SpId: 2, Status: biz.CpSpStatusNormal},
	}); err != nil {
		t.Fatalf("CreateCpSp() error = %v", err)
	}
	if _, err := svc.CreateCpSp(ctx, &v1.CreateCpSpRequest{
		CpSp: &v1.CpSp{CpId: 1, SpId: 2, Status: biz.CpSpStatusNormal},
	}); !kratoserrors.IsConflict(err) {
		t.Fatalf("CreateCpSp(duplicate) error = %v, want conflict", err)
	}
}

func cloneCpSpForTest(in *biz.CpSp) *biz.CpSp {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}
