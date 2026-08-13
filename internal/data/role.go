package data

import (
	"context"
	"database/sql"
	"errors"

	"crow/internal/biz"
)

type roleRepo struct {
	data *Data
}

type rolePO struct {
	ID         int64
	RoleName   string
	CreateTime sql.NullTime
	UpdateTime sql.NullTime
}

func NewRoleRepo(data *Data) biz.RoleRepo {
	return &roleRepo{data: data}
}

func (r *roleRepo) FindByID(ctx context.Context, id int64) (*biz.Role, error) {
	row := &rolePO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, role_name, create_date, update_date
FROM role
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.RoleName, &row.CreateTime, &row.UpdateTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrRoleNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *roleRepo) ListRoles(ctx context.Context, opts ...biz.RoleListOption) ([]*biz.Role, error) {
	options := biz.RoleListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrRoleInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, role_name, create_date, update_date
FROM role
ORDER BY id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]*biz.Role, 0, options.Limit)
	for rows.Next() {
		row := &rolePO{}
		if err := rows.Scan(&row.ID, &row.RoleName, &row.CreateTime, &row.UpdateTime); err != nil {
			return nil, err
		}
		roles = append(roles, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepo) CreateRole(ctx context.Context, role *biz.Role) (*biz.Role, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO role (role_name)
VALUES (?)
`, role.RoleName)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrRoleNameConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *roleRepo) UpdateRole(ctx context.Context, role *biz.Role) (*biz.Role, error) {
	result, err := r.data.db.ExecContext(ctx, `
UPDATE role
SET role_name = ?
WHERE id = ?
`, role.RoleName, role.ID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return nil, biz.ErrRoleNameConflict
		}
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, biz.ErrRoleNotFound
	}
	return r.FindByID(ctx, role.ID)
}

func (r *roleRepo) DeleteRole(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM role WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrRoleNotFound
	}
	return nil
}

func (p *rolePO) toBiz() *biz.Role {
	if p == nil {
		return nil
	}
	role := &biz.Role{
		ID:       p.ID,
		RoleName: p.RoleName,
	}
	if p.CreateTime.Valid {
		role.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		role.UpdateTime = p.UpdateTime.Time
	}
	return role
}
