package data

import (
	"context"
	"database/sql"

	"crow/internal/biz"
)

type systemLogRepo struct {
	data *Data
}

type systemLogPO struct {
	ID         int64
	LogUID     string
	LogLevel   string
	Message    string
	FilePath   sql.NullString
	LineNumber sql.NullInt64
	CreateTime sql.NullTime
}

// NewSystemLogRepo creates a new system log repo.
func NewSystemLogRepo(data *Data) biz.SystemLogRepo {
	return &systemLogRepo{data: data}
}

func (r *systemLogRepo) Create(ctx context.Context, item *biz.SystemLog) (*biz.SystemLog, error) {
	result, err := r.data.db.ExecContext(ctx, `
INSERT INTO system_log (log_uid, log_level, message, file_path, line_number)
VALUES (?, ?, ?, ?, ?)
`, item.LogUID, item.LogLevel, item.Message, nullableString(item.FilePath), nullableInt(item.LineNumber))
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	row := &systemLogPO{}
	err = r.data.db.QueryRowContext(ctx, `
SELECT id, log_uid, log_level, message, file_path, line_number, create_date
FROM system_log
WHERE id = ?
LIMIT 1
`, id).Scan(&row.ID, &row.LogUID, &row.LogLevel, &row.Message, &row.FilePath, &row.LineNumber, &row.CreateTime)
	if err != nil {
		return nil, err
	}
	return row.toBiz(), nil
}

func (r *systemLogRepo) List(ctx context.Context, opts ...biz.SystemLogListOption) ([]*biz.SystemLog, error) {
	options := biz.SystemLogListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, biz.ErrSystemLogInvalidArgument
	}

	rows, err := r.data.db.QueryContext(ctx, `
SELECT id, log_uid, log_level, message, file_path, line_number, create_date
FROM system_log
ORDER BY create_date DESC, id DESC
LIMIT ? OFFSET ?
`, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*biz.SystemLog, 0, options.Limit)
	for rows.Next() {
		row := &systemLogPO{}
		if err := rows.Scan(&row.ID, &row.LogUID, &row.LogLevel, &row.Message, &row.FilePath, &row.LineNumber, &row.CreateTime); err != nil {
			return nil, err
		}
		logs = append(logs, row.toBiz())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (p *systemLogPO) toBiz() *biz.SystemLog {
	if p == nil {
		return nil
	}
	item := &biz.SystemLog{
		ID:       p.ID,
		LogUID:   p.LogUID,
		LogLevel: p.LogLevel,
		Message:  p.Message,
	}
	if p.FilePath.Valid {
		item.FilePath = p.FilePath.String
	}
	if p.LineNumber.Valid {
		item.LineNumber = uint32(p.LineNumber.Int64)
	}
	if p.CreateTime.Valid {
		item.CreateTime = p.CreateTime.Time
	}
	return item
}

func nullableInt(value uint32) any {
	if value == 0 {
		return nil
	}
	return value
}
