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

type stubSpRepo struct {
	nextID int64
	sps    map[int64]*biz.Sp
}

func newTestSpService() *SpService {
	return NewSpService(biz.NewSpUsecase(&stubSpRepo{
		nextID: 1,
		sps:    make(map[int64]*biz.Sp),
	}))
}

func (r *stubSpRepo) FindByID(_ context.Context, id int64) (*biz.Sp, error) {
	sp, ok := r.sps[id]
	if !ok {
		return nil, biz.ErrSpNotFound
	}
	return cloneSpForTest(sp), nil
}

func (r *stubSpRepo) ListSps(_ context.Context, opts ...biz.SpListOption) ([]*biz.Sp, error) {
	options := biz.SpListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrSpInvalidArgument
	}

	list := make([]*biz.Sp, 0, len(r.sps))
	for _, sp := range r.sps {
		list = append(list, cloneSpForTest(sp))
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID > list[j].ID
	})

	if options.Offset >= len(list) {
		return []*biz.Sp{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(list) {
		end = len(list)
	}
	return list[options.Offset:end], nil
}

func (r *stubSpRepo) CreateSp(_ context.Context, sp *biz.Sp) (*biz.Sp, error) {
	for _, existing := range r.sps {
		if existing.SpCode == sp.SpCode {
			return nil, biz.ErrSpCodeConflict
		}
	}
	now := time.Now()
	created := cloneSpForTest(sp)
	created.ID = r.nextID
	created.CreateTime = now
	created.UpdateTime = now
	r.sps[created.ID] = cloneSpForTest(created)
	r.nextID++
	return cloneSpForTest(created), nil
}

func (r *stubSpRepo) UpdateSp(_ context.Context, sp *biz.Sp) (*biz.Sp, error) {
	current, ok := r.sps[sp.ID]
	if !ok {
		return nil, biz.ErrSpNotFound
	}
	for _, existing := range r.sps {
		if existing.ID != sp.ID && existing.SpCode == sp.SpCode {
			return nil, biz.ErrSpCodeConflict
		}
	}
	updated := cloneSpForTest(sp)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.sps[updated.ID] = cloneSpForTest(updated)
	return cloneSpForTest(updated), nil
}

func (r *stubSpRepo) DeleteSp(_ context.Context, id int64) error {
	if _, ok := r.sps[id]; !ok {
		return biz.ErrSpNotFound
	}
	delete(r.sps, id)
	return nil
}

func TestSpServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestSpService()

	created, err := svc.CreateSp(ctx, &v1.CreateSpRequest{
		Sp: &v1.Sp{
			SpCode:   "HUNAN_IPTV",
			SpName:   "湖南IPTV",
			SpConfig: `{"play_domain":"play.example.com"}`,
			Status:   biz.SpStatusNormal,
		},
	})
	if err != nil {
		t.Fatalf("CreateSp() error = %v", err)
	}
	if created.GetId() != 1 {
		t.Fatalf("CreateSp() id = %d, want 1", created.GetId())
	}
	if created.GetSpCode() != "HUNAN_IPTV" || created.GetSpName() != "湖南IPTV" {
		t.Fatalf("CreateSp() = %+v, want created sp", created)
	}
	if created.GetSpConfig() != `{"play_domain":"play.example.com"}` {
		t.Fatalf("CreateSp() sp_config = %q", created.GetSpConfig())
	}

	got, err := svc.GetSp(ctx, &v1.GetSpRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetSp() error = %v", err)
	}
	if got.GetSpCode() != "HUNAN_IPTV" || got.GetStatus() != biz.SpStatusNormal {
		t.Fatalf("GetSp() = %+v, want created sp", got)
	}

	updated, err := svc.UpdateSp(ctx, &v1.UpdateSpRequest{
		Sp: &v1.Sp{
			Id:       created.GetId(),
			SpName:   "湖南广电 IPTV",
			SpConfig: `{"play_domain":"iptv.example.com"}`,
			Status:   biz.SpStatusFrozen,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sp_name", "sp_config", "status"}},
	})
	if err != nil {
		t.Fatalf("UpdateSp() error = %v", err)
	}
	if updated.GetSpCode() != "HUNAN_IPTV" || updated.GetSpName() != "湖南广电 IPTV" || updated.GetStatus() != biz.SpStatusFrozen {
		t.Fatalf("UpdateSp() = %+v, want updated fields", updated)
	}

	if _, err := svc.DeleteSp(ctx, &v1.DeleteSpRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteSp() error = %v", err)
	}
	if _, err := svc.GetSp(ctx, &v1.GetSpRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetSp() after delete error = %v, want not found", err)
	}
}

func TestSpServiceListSpsPagination(t *testing.T) {
	ctx := context.Background()
	svc := newTestSpService()

	for _, code := range []string{"alpha", "beta", "gamma"} {
		if _, err := svc.CreateSp(ctx, &v1.CreateSpRequest{
			Sp: &v1.Sp{
				SpCode: code,
				SpName: code + " 内容服务商",
				Status: biz.SpStatusNormal,
			},
		}); err != nil {
			t.Fatalf("CreateSp(%q) error = %v", code, err)
		}
	}

	firstPage, err := svc.ListSps(ctx, &v1.ListSpsRequest{PageSize: 2})
	if err != nil {
		t.Fatalf("ListSps(first page) error = %v", err)
	}
	if len(firstPage.GetSps()) != 2 {
		t.Fatalf("ListSps(first page) len = %d, want 2", len(firstPage.GetSps()))
	}
	if firstPage.GetSps()[0].GetSpCode() != "gamma" {
		t.Fatalf("ListSps(first page) first sp_code = %q, want gamma", firstPage.GetSps()[0].GetSpCode())
	}
	if firstPage.GetNextPageToken() == "" {
		t.Fatal("ListSps(first page) next_page_token is empty")
	}
}

func TestSpServiceValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestSpService()

	if _, err := svc.CreateSp(ctx, &v1.CreateSpRequest{
		Sp: &v1.Sp{SpName: "缺少编码", Status: biz.SpStatusNormal},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateSp(missing code) error = %v, want bad request", err)
	}
	if _, err := svc.CreateSp(ctx, &v1.CreateSpRequest{
		Sp: &v1.Sp{SpCode: "BAD", SpName: "非法状态", Status: 9},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateSp(invalid status) error = %v, want bad request", err)
	}
	if _, err := svc.CreateSp(ctx, &v1.CreateSpRequest{
		Sp: &v1.Sp{SpCode: "BADJSON", SpName: "非法配置", SpConfig: "{", Status: biz.SpStatusNormal},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateSp(invalid json) error = %v, want bad request", err)
	}
	if _, err := svc.UpdateSp(ctx, &v1.UpdateSpRequest{
		Sp:         &v1.Sp{Id: 1, SpCode: "HUNAN_IPTV"},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("UpdateSp(unknown path) error = %v, want bad request", err)
	}

	if _, err := svc.CreateSp(ctx, &v1.CreateSpRequest{
		Sp: &v1.Sp{SpCode: "HUNAN_IPTV", SpName: "湖南IPTV", Status: biz.SpStatusNormal},
	}); err != nil {
		t.Fatalf("CreateSp() error = %v", err)
	}
	if _, err := svc.CreateSp(ctx, &v1.CreateSpRequest{
		Sp: &v1.Sp{SpCode: "HUNAN_IPTV", SpName: "重复编码", Status: biz.SpStatusNormal},
	}); !kratoserrors.IsConflict(err) {
		t.Fatalf("CreateSp(duplicate code) error = %v, want conflict", err)
	}
}

func cloneSpForTest(in *biz.Sp) *biz.Sp {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}
