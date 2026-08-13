package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"crow/internal/biz"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type adminRepo struct {
	data *Data
}

type adminPO struct {
	ID                int64
	Username          string
	Password          string
	RealName          string
	RoleID            int64
	Status            uint32
	LastLoginTime     sql.NullTime
	PasswordUpdatedAt sql.NullTime
	Remark            string
	CreateTime        sql.NullTime
	UpdateTime        sql.NullTime
}

// NewAdminRepo creates a new AdminRepo instance.
func NewAdminRepo(data *Data) biz.AdminRepo {
	return &adminRepo{data: data}
}

func (r *adminRepo) FindByID(ctx context.Context, id int64) (*biz.Admin, error) {
	row := &adminPO{}
	err := r.data.db.QueryRowContext(ctx, `
SELECT id, username, password, real_name, role_id, status, last_login_time, password_updated_at, remark, create_date, update_date
FROM admin
WHERE id = ?
LIMIT 1
`, id).Scan(
		&row.ID,
		&row.Username,
		&row.Password,
		&row.RealName,
		&row.RoleID,
		&row.Status,
		&row.LastLoginTime,
		&row.PasswordUpdatedAt,
		&row.Remark,
		&row.CreateTime,
		&row.UpdateTime,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrAdminNotFound
		}
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *adminRepo) ListAdmins(ctx context.Context, opts ...biz.AdminListOption) ([]*biz.Admin, error) {
	options := biz.AdminListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrAdminInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, username, password, real_name, role_id, status, last_login_time, password_updated_at, remark, create_date, update_date
FROM admin
ORDER BY id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	admins := make([]*biz.Admin, 0, options.Limit)
	for rows.Next() {
		row := &adminPO{}
		if err := rows.Scan(
			&row.ID,
			&row.Username,
			&row.Password,
			&row.RealName,
			&row.RoleID,
			&row.Status,
			&row.LastLoginTime,
			&row.PasswordUpdatedAt,
			&row.Remark,
			&row.CreateTime,
			&row.UpdateTime,
		); err != nil {
			return nil, err
		}
		admins = append(admins, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *adminRepo) CreateAdmin(ctx context.Context, admin *biz.Admin) (*biz.Admin, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO admin (username, password, real_name, role_id, status, password_updated_at, remark)
VALUES (?, ?, ?, ?, ?, ?, ?)
`, admin.Username, admin.PasswordHash, admin.RealName, admin.RoleID, admin.Status, admin.PasswordUpdatedAt, admin.Remark)
	if err != nil {
		if isDuplicateUsernameError(err) {
			return nil, biz.ErrAdminUsernameConflict
		}
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *adminRepo) UpdateAdmin(ctx context.Context, admin *biz.Admin) (*biz.Admin, error) {
	result, err := r.data.db.ExecContext(ctx, `
UPDATE admin
SET username = ?, password = ?, real_name = ?, role_id = ?, status = ?, last_login_time = ?, password_updated_at = ?, remark = ?
WHERE id = ?
`, admin.Username, admin.PasswordHash, admin.RealName, admin.RoleID, admin.Status, nullableTime(admin.LastLoginTime), nullableTime(admin.PasswordUpdatedAt), admin.Remark, admin.ID)
	if err != nil {
		if isDuplicateUsernameError(err) {
			return nil, biz.ErrAdminUsernameConflict
		}
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, biz.ErrAdminNotFound
	}
	return r.FindByID(ctx, admin.ID)
}

func (r *adminRepo) DeleteAdmin(ctx context.Context, id int64) error {
	result, err := r.data.db.ExecContext(ctx, "DELETE FROM admin WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return biz.ErrAdminNotFound
	}
	return nil
}

func (p *adminPO) toBiz() *biz.Admin {
	if p == nil {
		return nil
	}
	admin := &biz.Admin{
		ID:           p.ID,
		Username:     p.Username,
		PasswordHash: p.Password,
		RealName:     p.RealName,
		RoleID:       p.RoleID,
		Status:       p.Status,
		Remark:       p.Remark,
	}
	if p.LastLoginTime.Valid {
		admin.LastLoginTime = &p.LastLoginTime.Time
	}
	if p.PasswordUpdatedAt.Valid {
		admin.PasswordUpdatedAt = &p.PasswordUpdatedAt.Time
	}
	if p.CreateTime.Valid {
		admin.CreateTime = p.CreateTime.Time
	}
	if p.UpdateTime.Valid {
		admin.UpdateTime = p.UpdateTime.Time
	}
	return admin
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func isDuplicateUsernameError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
