package data

import (
	"context"
	"database/sql"
	"errors"

	"crow/internal/biz"
)

type permissionRepo struct {
	data *Data
}

type permissionPO struct {
	ID         int64
	ParentID   int64
	Title      string
	Handle     string
	Weight     int32
	CreateTime sql.NullTime
	UpdateTime sql.NullTime
}

func NewPermissionRepo(data *Data) biz.PermissionRepo {
	return &permissionRepo{data: data}
}

func (r *permissionRepo) FindByID(ctx context.Context, id int64) (*biz.Permission, error) {
	row := &permissionPO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, parent_id, title, handle, weight, create_date, update_date
FROM permission
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.ParentID, &row.Title, &row.Handle, &row.Weight, &row.CreateTime, &row.UpdateTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrPermissionNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *permissionRepo) ListPermissions(ctx context.Context, opts ...biz.PermissionListOption) ([]*biz.Permission, error) {
	options := biz.PermissionListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrPermissionInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, parent_id, title, handle, weight, create_date, update_date
FROM permission
ORDER BY parent_id ASC, weight ASC, id ASC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]*biz.Permission, 0, options.Limit)
	for rows.Next() {
		row := &permissionPO{}
		if err := rows.Scan(&row.ID, &row.ParentID, &row.Title, &row.Handle, &row.Weight, &row.CreateTime, &row.UpdateTime); err != nil {
			return nil, err
		}
		permissions = append(permissions, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepo) CreatePermission(ctx context.Context, permission *biz.Permission) (*biz.Permission, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO permission (parent_id, title, handle, weight)
VALUES (?, ?, ?, ?)
`, permission.ParentID, permission.Title, permission.Handle, permission.Weight)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrPermissionHandleConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *permissionRepo) UpdatePermission(ctx context.Context, permission *biz.Permission) (*biz.Permission, error) {
	result, err := r.data.db.ExecContext(ctx, `
UPDATE permission
SET parent_id = ?, title = ?, handle = ?, weight = ?
WHERE id = ?
`, permission.ParentID, permission.Title, permission.Handle, permission.Weight, permission.ID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrPermissionHandleConflict
		}
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, biz.ErrPermissionNotFound
	}
	return r.FindByID(ctx, permission.ID)
}

func (r *permissionRepo) DeletePermission(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM permission WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrPermissionNotFound
	}
	return nil
}

func (p *permissionPO) toBiz() *biz.Permission {
	if p == nil {
		return nil
	}
	permission := &biz.Permission{
		ID:       p.ID,
		ParentID: p.ParentID,
		Title:    p.Title,
		Handle:   p.Handle,
		Weight:   p.Weight,
	}
	if p.CreateTime.Valid {
		permission.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		permission.UpdateTime = p.UpdateTime.Time
	}
	return permission
}
