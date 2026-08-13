package data

import (
	"context"
	"database/sql"
	"errors"

	"crow/internal/biz"
)

type groupPermissionRepo struct {
	data *Data
}

type groupPermissionPO struct {
	ID           int64
	GroupID      int64
	PermissionID int64
	CreateTime   sql.NullTime
	UpdateTime   sql.NullTime
}

func NewGroupPermissionRepo(data *Data) biz.GroupPermissionRepo {
	return &groupPermissionRepo{data: data}
}

func (r *groupPermissionRepo) FindByID(ctx context.Context, id int64) (*biz.GroupPermission, error) {
	row := &groupPermissionPO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, role_id, permission_id, create_date, update_date
FROM role_permission
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.GroupID, &row.PermissionID, &row.CreateTime, &row.UpdateTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrGroupPermissionNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *groupPermissionRepo) ListGroupPermissions(ctx context.Context, opts ...biz.GroupPermissionListOption) ([]*biz.GroupPermission, error) {
	options := biz.GroupPermissionListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrGroupPermissionInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, role_id, permission_id, create_date, update_date
FROM role_permission
ORDER BY id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	relations := make([]*biz.GroupPermission, 0, options.Limit)
	for rows.Next() {
		row := &groupPermissionPO{}
		if err := rows.Scan(&row.ID, &row.GroupID, &row.PermissionID, &row.CreateTime, &row.UpdateTime); err != nil {
			return nil, err
		}
		relations = append(relations, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return relations, nil
}

func (r *groupPermissionRepo) CreateGroupPermission(ctx context.Context, relation *biz.GroupPermission) (*biz.GroupPermission, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO role_permission (role_id, permission_id)
VALUES (?, ?)
`, relation.GroupID, relation.PermissionID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrGroupPermissionConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *groupPermissionRepo) UpdateGroupPermission(ctx context.Context, relation *biz.GroupPermission) (*biz.GroupPermission, error) {
	result, err := r.data.db.ExecContext(ctx, `
UPDATE role_permission
SET role_id = ?, permission_id = ?
WHERE id = ?
`, relation.GroupID, relation.PermissionID, relation.ID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrGroupPermissionConflict
		}
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, biz.ErrGroupPermissionNotFound
	}
	return r.FindByID(ctx, relation.ID)
}

func (r *groupPermissionRepo) DeleteGroupPermission(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM role_permission WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrGroupPermissionNotFound
	}
	return nil
}

func (p *groupPermissionPO) toBiz() *biz.GroupPermission {
	if p == nil {
		return nil
	}
	relation := &biz.GroupPermission{
		ID:           p.ID,
		GroupID:      p.GroupID,
		PermissionID: p.PermissionID,
	}
	if p.CreateTime.Valid {
		relation.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		relation.UpdateTime = p.UpdateTime.Time
	}
	return relation
}
