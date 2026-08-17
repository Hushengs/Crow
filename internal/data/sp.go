package data

import (
	"context"
	"database/sql"
	"errors"

	"crow/internal/biz"
)

type spRepo struct {
	data *Data
}

type spPO struct {
	ID         int64
	SpCode     string
	SpName     string
	SpConfig   sql.NullString
	Status     uint32
	CreateTime sql.NullTime
	UpdateTime sql.NullTime
}

// NewSpRepo creates a new SpRepo instance.
func NewSpRepo(data *Data) biz.SpRepo {
	return &spRepo{data: data}
}

func (r *spRepo) FindByID(ctx context.Context, id int64) (*biz.Sp, error) {
	row := &spPO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, sp_code, sp_name, sp_config, status, create_date, update_date
FROM sp
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.SpCode, &row.SpName, &row.SpConfig, &row.Status, &row.CreateTime, &row.UpdateTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrSpNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *spRepo) ListSps(ctx context.Context, opts ...biz.SpListOption) ([]*biz.Sp, error) {
	options := biz.SpListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrSpInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, sp_code, sp_name, sp_config, status, create_date, update_date
FROM sp
ORDER BY id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sps := make([]*biz.Sp, 0, options.Limit)
	for rows.Next() {
		row := &spPO{}
		if err := rows.Scan(&row.ID, &row.SpCode, &row.SpName, &row.SpConfig, &row.Status, &row.CreateTime, &row.UpdateTime); err != nil {
			return nil, err
		}
		sps = append(sps, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sps, nil
}

func (r *spRepo) CreateSp(ctx context.Context, sp *biz.Sp) (*biz.Sp, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO sp (sp_code, sp_name, sp_config, status)
VALUES (?, ?, ?, ?)
`, sp.SpCode, sp.SpName, nullableJSON(sp.SpConfig), sp.Status)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrSpCodeConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *spRepo) UpdateSp(ctx context.Context, sp *biz.Sp) (*biz.Sp, error) {
	result, err := r.data.db.ExecContext(ctx, `
UPDATE sp
SET sp_code = ?, sp_name = ?, sp_config = ?, status = ?
WHERE id = ?
`, sp.SpCode, sp.SpName, nullableJSON(sp.SpConfig), sp.Status, sp.ID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrSpCodeConflict
		}
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, biz.ErrSpNotFound
	}
	return r.FindByID(ctx, sp.ID)
}

func (r *spRepo) DeleteSp(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM sp WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrSpNotFound
	}
	return nil
}

func (p *spPO) toBiz() *biz.Sp {
	if p == nil {
		return nil
	}
	sp := &biz.Sp{
		ID:       p.ID,
		SpCode:   p.SpCode,
		SpName:   p.SpName,
		SpConfig: p.SpConfig.String,
		Status:   p.Status,
	}
	if p.CreateTime.Valid {
		sp.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		sp.UpdateTime = p.UpdateTime.Time
	}
	return sp
}

func nullableJSON(value string) any {
	if value == "" {
		return nil
	}
	return value
}
