package data

import (
	"context"
	"database/sql"
	"errors"

	"crow/internal/biz"
)

type cpRepo struct {
	data *Data
}

type cpPO struct {
	ID         int64
	CpCode     string
	CpName     string
	Status     uint32
	CreateTime sql.NullTime
	UpdateTime sql.NullTime
}

// NewCpRepo creates a new CpRepo instance.
func NewCpRepo(data *Data) biz.CpRepo {
	return &cpRepo{data: data}
}

func (r *cpRepo) FindByID(ctx context.Context, id int64) (*biz.Cp, error) {
	row := &cpPO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, cp_code, cp_name, status, create_date, update_date
FROM cp
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.CpCode, &row.CpName, &row.Status, &row.CreateTime, &row.UpdateTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrCpNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *cpRepo) ListCps(ctx context.Context, opts ...biz.CpListOption) ([]*biz.Cp, error) {
	options := biz.CpListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrCpInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, cp_code, cp_name, status, create_date, update_date
FROM cp
ORDER BY id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cps := make([]*biz.Cp, 0, options.Limit)
	for rows.Next() {
		row := &cpPO{}
		if err := rows.Scan(&row.ID, &row.CpCode, &row.CpName, &row.Status, &row.CreateTime, &row.UpdateTime); err != nil {
			return nil, err
		}
		cps = append(cps, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cps, nil
}

func (r *cpRepo) CreateCp(ctx context.Context, cp *biz.Cp) (*biz.Cp, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO cp (cp_code, cp_name, status)
VALUES (?, ?, ?)
`, cp.CpCode, cp.CpName, cp.Status)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrCpCodeConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *cpRepo) UpdateCp(ctx context.Context, cp *biz.Cp) (*biz.Cp, error) {
	_, err := r.data.db.ExecContext(ctx, `
UPDATE cp
SET cp_code = ?, cp_name = ?, status = ?
WHERE id = ?
`, cp.CpCode, cp.CpName, cp.Status, cp.ID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrCpCodeConflict
		}
		return nil, err
	}
	// Prefer FindByID over RowsAffected: unchanged values yield 0 rows on MySQL.
	return r.FindByID(ctx, cp.ID)
}

func (r *cpRepo) DeleteCp(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM cp WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrCpNotFound
	}
	return nil
}

func (p *cpPO) toBiz() *biz.Cp {
	if p == nil {
		return nil
	}
	cp := &biz.Cp{
		ID:     p.ID,
		CpCode: p.CpCode,
		CpName: p.CpName,
		Status: p.Status,
	}
	if p.CreateTime.Valid {
		cp.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		cp.UpdateTime = p.UpdateTime.Time
	}
	return cp
}
