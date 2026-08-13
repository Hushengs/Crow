package data

import (
	"context"
	"database/sql"
	"errors"

	"crow/internal/biz"
)

type adminRoleRepo struct {
	data *Data
}

type adminRolePO struct {
	ID         int64
	AdminID    int64
	RoleID     int64
	CreateTime sql.NullTime
	UpdateTime sql.NullTime
}

func NewAdminRoleRepo(data *Data) biz.AdminRoleRepo {
	return &adminRoleRepo{data: data}
}

func (r *adminRoleRepo) FindByID(ctx context.Context, id int64) (*biz.AdminRole, error) {
	row := &adminRolePO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, admin_id, role_id, create_date, update_date
FROM admin_role
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.AdminID, &row.RoleID, &row.CreateTime, &row.UpdateTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrAdminRoleNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *adminRoleRepo) ListAdminRoles(ctx context.Context, opts ...biz.AdminRoleListOption) ([]*biz.AdminRole, error) {
	options := biz.AdminRoleListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrAdminRoleInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, admin_id, role_id, create_date, update_date
FROM admin_role
ORDER BY id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	relations := make([]*biz.AdminRole, 0, options.Limit)
	for rows.Next() {
		row := &adminRolePO{}
		if err := rows.Scan(&row.ID, &row.AdminID, &row.RoleID, &row.CreateTime, &row.UpdateTime); err != nil {
			return nil, err
		}
		relations = append(relations, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return relations, nil
}

func (r *adminRoleRepo) CreateAdminRole(ctx context.Context, relation *biz.AdminRole) (*biz.AdminRole, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO admin_role (admin_id, role_id)
VALUES (?, ?)
`, relation.AdminID, relation.RoleID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrAdminRoleConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *adminRoleRepo) UpdateAdminRole(ctx context.Context, relation *biz.AdminRole) (*biz.AdminRole, error) {
	result, err := r.data.db.ExecContext(ctx, `
UPDATE admin_role
SET admin_id = ?, role_id = ?
WHERE id = ?
`, relation.AdminID, relation.RoleID, relation.ID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrAdminRoleConflict
		}
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, biz.ErrAdminRoleNotFound
	}
	return r.FindByID(ctx, relation.ID)
}

func (r *adminRoleRepo) DeleteAdminRole(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM admin_role WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrAdminRoleNotFound
	}
	return nil
}

func (p *adminRolePO) toBiz() *biz.AdminRole {
	if p == nil {
		return nil
	}
	relation := &biz.AdminRole{
		ID:      p.ID,
		AdminID: p.AdminID,
		RoleID:  p.RoleID,
	}
	if p.CreateTime.Valid {
		relation.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		relation.UpdateTime = p.UpdateTime.Time
	}
	return relation
}
