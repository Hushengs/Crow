package data

import (
	"context"
	"database/sql"
	"errors"

	"crow/internal/biz"
)

type cpSpRepo struct {
	data *Data
}

type cpSpPO struct {
	ID         int64
	CpID       int64
	SpID       int64
	Status     uint32
	CreateTime sql.NullTime
	UpdateTime sql.NullTime
}

func NewCpSpRepo(data *Data) biz.CpSpRepo {
	return &cpSpRepo{data: data}
}

func (r *cpSpRepo) FindByID(ctx context.Context, id int64) (*biz.CpSp, error) {
	row := &cpSpPO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, cp_id, sp_id, status, create_date, update_date
FROM cp_sp
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.CpID, &row.SpID, &row.Status, &row.CreateTime, &row.UpdateTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrCpSpNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *cpSpRepo) ListCpSps(ctx context.Context, opts ...biz.CpSpListOption) ([]*biz.CpSp, error) {
	options := biz.CpSpListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrCpSpInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, cp_id, sp_id, status, create_date, update_date
FROM cp_sp
ORDER BY id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	relations := make([]*biz.CpSp, 0, options.Limit)
	for rows.Next() {
		row := &cpSpPO{}
		if err := rows.Scan(&row.ID, &row.CpID, &row.SpID, &row.Status, &row.CreateTime, &row.UpdateTime); err != nil {
			return nil, err
		}
		relations = append(relations, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return relations, nil
}

func (r *cpSpRepo) CreateCpSp(ctx context.Context, relation *biz.CpSp) (*biz.CpSp, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO cp_sp (cp_id, sp_id, status)
VALUES (?, ?, ?)
`, relation.CpID, relation.SpID, relation.Status)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrCpSpConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *cpSpRepo) UpdateCpSp(ctx context.Context, relation *biz.CpSp) (*biz.CpSp, error) {
	result, err := r.data.db.ExecContext(ctx, `
UPDATE cp_sp
SET cp_id = ?, sp_id = ?, status = ?
WHERE id = ?
`, relation.CpID, relation.SpID, relation.Status, relation.ID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrCpSpConflict
		}
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, biz.ErrCpSpNotFound
	}
	return r.FindByID(ctx, relation.ID)
}

func (r *cpSpRepo) DeleteCpSp(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM cp_sp WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrCpSpNotFound
	}
	return nil
}

func (p *cpSpPO) toBiz() *biz.CpSp {
	if p == nil {
		return nil
	}
	relation := &biz.CpSp{
		ID:     p.ID,
		CpID:   p.CpID,
		SpID:   p.SpID,
		Status: p.Status,
	}
	if p.CreateTime.Valid {
		relation.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		relation.UpdateTime = p.UpdateTime.Time
	}
	return relation
}
