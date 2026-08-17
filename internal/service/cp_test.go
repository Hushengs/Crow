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

type stubCpRepo struct {
	nextID int64
	cps    map[int64]*biz.Cp
}

func newTestCpService() *CpService {
	return NewCpService(biz.NewCpUsecase(&stubCpRepo{
		nextID: 1,
		cps:    make(map[int64]*biz.Cp),
	}))
}

func (r *stubCpRepo) FindByID(_ context.Context, id int64) (*biz.Cp, error) {
	cp, ok := r.cps[id]
	if !ok {
		return nil, biz.ErrCpNotFound
	}
	return cloneCpForTest(cp), nil
}

func (r *stubCpRepo) ListCps(_ context.Context, opts ...biz.CpListOption) ([]*biz.Cp, error) {
	options := biz.CpListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrCpInvalidArgument
	}

	list := make([]*biz.Cp, 0, len(r.cps))
	for _, cp := range r.cps {
		list = append(list, cloneCpForTest(cp))
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID > list[j].ID
	})

	if options.Offset >= len(list) {
		return []*biz.Cp{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubCpRepo) CreateCp(_ context.Context, cp *biz.Cp) (*biz.Cp, error) {
	for _, existing := range r.cps {
		if existing.CpCode == cp.CpCode {
			return nil, biz.ErrCpCodeConflict
		}
	}
	now := time.Now()
	created := cloneCpForTest(cp)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.cps[created.ID] = cloneCpForTest(created)
	r.nextID++
	return cloneCpForTest(created), nil
}

func (r *stubCpRepo) UpdateCp(_ context.Context, cp *biz.Cp) (*biz.Cp, error) {
	current, ok := r.cps[cp.ID]
	if !ok {
		return nil, biz.ErrCpNotFound
	}
	for _, existing := range r.cps {
		if existing.ID != cp.ID && existing.CpCode == cp.CpCode {
			return nil, biz.ErrCpCodeConflict
		}
	}
	updated := cloneCpForTest(cp)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.cps[updated.ID] = cloneCpForTest(updated)
	return cloneCpForTest(updated), nil
}

func (r *stubCpRepo) DeleteCp(_ context.Context, id int64) error {
	if _, ok := r.cps[id]; !ok {
		return biz.ErrCpNotFound
	}
	delete(r.cps, id)
	return nil
}

func TestCpServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestCpService()

	created, err := svc.CreateCp(ctx, &v1.CreateCpRequest{
		Cp: &v1.Cp{
			CpCode: "CNTV",
			CpName: "中央电视台",
			Status: biz.CpStatusNormal,
		},
	})
	if err != nil {
		t.Fatalf("CreateCp() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreateCp() id = %d, want 1", created.GetId())
	}
	if created.GetCpCode() != "CNTV" || created.GetCpName() != "中央电视台" {
		t.Fatalf("CreateCp() = %+v, want created cp", created)
	}

	got, err := svc.GetCp(ctx, &v1.GetCpRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetCp() error = %v", err)
	}
	if got.GetCpCode() != "CNTV" || got.GetStatus() != biz.CpStatusNormal {
		t.Fatalf("GetCp() = %+v, want created cp", got)
	}

	updated, err := svc.UpdateCp(ctx, &v1.UpdateCpRequest{
		Cp: &v1.Cp{
			Id:     created.GetId(),
			CpName: "中国网络电视台",
			Status: biz.CpStatusFrozen,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"cp_name", "status"}},
	})
	if err != nil {
		t.Fatalf("UpdateCp() error = %v", err)
	}
	if updated.GetCpCode() != "CNTV" || updated.GetCpName() != "中国网络电视台" || updated.GetStatus() != biz.CpStatusFrozen {
		t.Fatalf("UpdateCp() = %+v, want updated fields", updated)
	}

	if _, err := svc.DeleteCp(ctx, &v1.DeleteCpRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteCp() error = %v", err)
	}
	if _, err := svc.GetCp(ctx, &v1.GetCpRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetCp() after delete error = %v, want not found", err)
	}
}

func TestCpServiceListCpsPagination(t *testing.T) {
	ctx := context.Background()
	svc := newTestCpService()

	for _, code := range []string{"alpha", "beta", "gamma"} {
		if _, err := svc.CreateCp(ctx, &v1.CreateCpRequest{
			Cp: &v1.Cp{
				CpCode: code,
				CpName: code + " 内容提供商",
				Status: biz.CpStatusNormal,
			},
		}); err != nil {
			t.Fatalf("CreateCp(%q) error = %v", code, err)
		}
	}

	firstPage, err := svc.ListCps(ctx, &v1.ListCpsRequest{PageSize: 2})
	if err != nil {
		t.Fatalf("ListCps(first page) error = %v", err)
	}
	if len(firstPage.GetCps()) != 2 {
		t.Fatalf("ListCps(first page) len = %d, want 2", len(firstPage.GetCps()))
	}
	if firstPage.GetCps()[0].GetCpCode() != "gamma" {
		t.Fatalf("ListCps(first page) first cp_code = %q, want gamma", firstPage.GetCps()[0].GetCpCode())
	}
	if firstPage.GetNextPageToken() == "" {
		t.Fatal("ListCps(first page) next_page_token is empty")
	}
}

func TestCpServiceValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestCpService()

	if _, err := svc.CreateCp(ctx, &v1.CreateCpRequest{
		Cp: &v1.Cp{CpName: "缺少编码", Status: biz.CpStatusNormal},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateCp(missing code) error = %v, want bad request", err)
	}
	if _, err := svc.CreateCp(ctx, &v1.CreateCpRequest{
		Cp: &v1.Cp{CpCode: "BAD", CpName: "非法状态", Status: 9},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateCp(invalid status) error = %v, want bad request", err)
	}
	if _, err := svc.UpdateCp(ctx, &v1.UpdateCpRequest{
		Cp:         &v1.Cp{Id: 1, CpCode: "CNTV"},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("UpdateCp(unknown path) error = %v, want bad request", err)
	}

	if _, err := svc.CreateCp(ctx, &v1.CreateCpRequest{
		Cp: &v1.Cp{CpCode: "CNTV", CpName: "中央电视台", Status: biz.CpStatusNormal},
	}); err != nil {
		t.Fatalf("CreateCp() error = %v", err)
	}
	if _, err := svc.CreateCp(ctx, &v1.CreateCpRequest{
		Cp: &v1.Cp{CpCode: "CNTV", CpName: "重复编码", Status: biz.CpStatusNormal},
	}); !kratoserrors.IsConflict(err) {
		t.Fatalf("CreateCp(duplicate code) error = %v, want conflict", err)
	}
}

func cloneCpForTest(in *biz.Cp) *biz.Cp {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}
