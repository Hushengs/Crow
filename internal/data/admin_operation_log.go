package data

import (
	"context"
	"database/sql"

	"crow/internal/biz"
)

type adminOperationLogRepo struct {
	data *Data
}

type adminOperationLogPO struct {
	ID            int64
	AdminID       int64
	AdminName     string
	Module        string
	Action        string
	Description   string
	RequestMethod string
	RequestURL    string
	RequestParams sql.NullString
	CreateTime    sql.NullTime
}

// NewAdminOperationLogRepo creates a new operation log repo.
func NewAdminOperationLogRepo(data *Data) biz.AdminOperationLogRepo {
	return &adminOperationLogRepo{data: data}
}

func (r *adminOperationLogRepo) Create(ctx context.Context, log *biz.AdminOperationLog) (*biz.AdminOperationLog, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO admin_operation_log (admin_id, admin_name, module, action, description, request_method, request_url, request_params)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, log.AdminID, log.AdminName, log.Module, log.Action, log.Description, log.RequestMethod, log.RequestURL, nullableString(log.RequestParams))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	row := &adminOperationLogPO{}
	err = r.data.db.QueryRowContext(ctx, `
SELECT id, admin_id, admin_name, module, action, description, request_method, request_url, request_params, create_date
FROM admin_operation_log
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.AdminID, &row.AdminName, &row.Module, &row.Action, &row.Description, &row.RequestMethod, &row.RequestURL, &row.RequestParams, &row.CreateTime)
	if err != nil {
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *adminOperationLogRepo) List(ctx context.Context, opts ...biz.AdminOperationLogListOption) ([]*biz.AdminOperationLog, error) {
	options := biz.AdminOperationLogListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrAdminOperationLogInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, admin_id, admin_name, module, action, description, request_method, request_url, request_params, create_date
FROM admin_operation_log
ORDER BY create_date DESC, id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*biz.AdminOperationLog, 0, options.Limit)
	for rows.Next() {
		row := &adminOperationLogPO{}
		if err := rows.Scan(&row.ID, &row.AdminID, &row.AdminName, &row.Module, &row.Action, &row.Description, &row.RequestMethod, &row.RequestURL, &row.RequestParams, &row.CreateTime); err != nil {
			return nil, err
		}
		logs = append(logs, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (p *adminOperationLogPO) toBiz() *biz.AdminOperationLog {
	if p == nil {
		return nil
	}
	log := &biz.AdminOperationLog{
		ID:            p.ID,
		AdminID:       p.AdminID,
		AdminName:     p.AdminName,
		Module:        p.Module,
		Action:        p.Action,
		Description:   p.Description,
		RequestMethod: p.RequestMethod,
		RequestURL:    p.RequestURL,
	}
	if p.RequestParams.Valid {
		log.RequestParams = p.RequestParams.String
	}
	if p.CreateTime.Valid {
		log.CreateTime = p.CreateTime.Time
	}
	return log
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
