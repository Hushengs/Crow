package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"crow/internal/biz"

	"github.com/google/uuid"
)

type injectRepo struct {
	data *Data
}

func NewInjectRepo(data *Data) biz.InjectRepo {
	return &injectRepo{data: data}
}

const injectContentColumns = `id,resource_type,resource_id,video_id,episode_id,media_id,action,last_task_id,status,fail_reason,create_date,update_date`
const injectTaskColumns = `id,content_id,resource_type,resource_id,video_id,episode_id,media_id,action,status,create_date,update_date`
const injectLogColumns = `id,task_id,content_id,sp_id,resource_type,resource_id,video_id,episode_id,media_id,action,correlate_id,sync_status,async_status,sync_message,async_message,create_date,update_date`

func (r *injectRepo) Enqueue(ctx context.Context, job *biz.InjectEnqueue) (*biz.InjectContent, *biz.InjectTask, error) {
	tx, err := r.data.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO inject_content (resource_type, resource_id, video_id, episode_id, media_id, action, last_task_id, status, fail_reason)
VALUES (?, ?, ?, ?, ?, ?, 0, 0, '')
ON DUPLICATE KEY UPDATE
  video_id = VALUES(video_id),
  episode_id = VALUES(episode_id),
  media_id = VALUES(media_id),
  action = VALUES(action),
  status = 0,
  fail_reason = ''
`, job.ResourceType, job.ResourceID, job.VideoID, job.EpisodeID, job.MediaID, job.Action); err != nil {
		return nil, nil, err
	}

	var contentID int64
	if err := tx.QueryRowContext(ctx, `
SELECT id FROM inject_content WHERE resource_type = ? AND resource_id = ? LIMIT 1
`, job.ResourceType, job.ResourceID).Scan(&contentID); err != nil {
		return nil, nil, err
	}

	res, err := tx.ExecContext(ctx, `
INSERT INTO inject_task (content_id, resource_type, resource_id, video_id, episode_id, media_id, action, status)
VALUES (?, ?, ?, ?, ?, ?, ?, 0)
`, contentID, job.ResourceType, job.ResourceID, job.VideoID, job.EpisodeID, job.MediaID, job.Action)
	if err != nil {
		return nil, nil, err
	}
	taskID, err := res.LastInsertId()
	if err != nil {
		return nil, nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE inject_content SET last_task_id = ? WHERE id = ?`, taskID, contentID); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	content, err := r.FindContent(ctx, contentID)
	if err != nil {
		return nil, nil, err
	}
	task, err := r.FindTask(ctx, taskID)
	if err != nil {
		return content, nil, err
	}
	return content, task, nil
}

func (r *injectRepo) FindContent(ctx context.Context, id int64) (*biz.InjectContent, error) {
	row, err := scanInjectContent(r.data.db.QueryRowContext(ctx, `SELECT `+injectContentColumns+` FROM inject_content WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, biz.ErrInjectNotFound
	}
	return row, err
}

func (r *injectRepo) ListContents(ctx context.Context, opts ...biz.InjectListOption) ([]*biz.InjectContent, error) {
	limit, offset, err := injectListBounds(opts)
	if err != nil {
		return nil, err
	}
	rows, err := r.data.db.QueryContext(ctx, `SELECT `+injectContentColumns+` FROM inject_content ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*biz.InjectContent, 0, limit)
	for rows.Next() {
		item, err := scanInjectContent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *injectRepo) FindTask(ctx context.Context, id int64) (*biz.InjectTask, error) {
	row, err := scanInjectTask(r.data.db.QueryRowContext(ctx, `SELECT `+injectTaskColumns+` FROM inject_task WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, biz.ErrInjectNotFound
	}
	return row, err
}

func (r *injectRepo) ListTasks(ctx context.Context, opts ...biz.InjectListOption) ([]*biz.InjectTask, error) {
	limit, offset, err := injectListBounds(opts)
	if err != nil {
		return nil, err
	}
	rows, err := r.data.db.QueryContext(ctx, `SELECT `+injectTaskColumns+` FROM inject_task ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*biz.InjectTask, 0, limit)
	for rows.Next() {
		item, err := scanInjectTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *injectRepo) ListLogs(ctx context.Context, opts ...biz.InjectListOption) ([]*biz.InjectLog, error) {
	limit, offset, err := injectListBounds(opts)
	if err != nil {
		return nil, err
	}
	rows, err := r.data.db.QueryContext(ctx, `SELECT `+injectLogColumns+` FROM inject_log ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*biz.InjectLog, 0, limit)
	for rows.Next() {
		item, err := scanInjectLog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *injectRepo) ClaimNextTask(ctx context.Context) (*biz.InjectTask, error) {
	tx, err := r.data.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var id int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM inject_task WHERE status = 0 ORDER BY id ASC LIMIT 1 FOR UPDATE`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, biz.ErrInjectNoPending
	}
	if err != nil {
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE inject_task SET status = 1 WHERE id = ? AND status = 0`, id)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, biz.ErrInjectNoPending
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.FindTask(ctx, id)
}

func (r *injectRepo) DeleteTask(ctx context.Context, id int64) error {
	_, err := r.data.db.ExecContext(ctx, `DELETE FROM inject_task WHERE id = ?`, id)
	return err
}

func (r *injectRepo) UpdateContentStatus(ctx context.Context, id int64, status uint32, failReason string) error {
	res, err := r.data.db.ExecContext(ctx, `UPDATE inject_content SET status = ?, fail_reason = ? WHERE id = ?`, status, failReason, id)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func (r *injectRepo) ListActiveSpIDsForVideo(ctx context.Context, videoID int64) ([]int64, error) {
	if videoID <= 0 {
		return nil, biz.ErrInjectInvalidArgument
	}
	rows, err := r.data.db.QueryContext(ctx, `
SELECT DISTINCT s.id
FROM video v
INNER JOIN cp ON cp.id = v.cp_id AND cp.status = 1
INNER JOIN cp_sp cs ON cs.cp_id = v.cp_id AND cs.status = 1
INNER JOIN sp s ON s.id = cs.sp_id AND s.status = 1
WHERE v.id = ?
ORDER BY s.id
`, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *injectRepo) CreateLog(ctx context.Context, logRow *biz.InjectLog) error {
	if logRow.CorrelateID == "" {
		logRow.CorrelateID = uuid.NewString()
	}
	_, err := r.data.db.ExecContext(ctx, `
INSERT INTO inject_log (task_id, content_id, sp_id, resource_type, resource_id, video_id, episode_id, media_id, action, correlate_id, sync_status, async_status, sync_message, async_message)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, logRow.TaskID, logRow.ContentID, logRow.SpID, logRow.ResourceType, logRow.ResourceID, logRow.VideoID, logRow.EpisodeID, logRow.MediaID, logRow.Action, logRow.CorrelateID, logRow.SyncStatus, logRow.AsyncStatus, logRow.SyncMessage, logRow.AsyncMessage)
	return err
}

func injectListBounds(opts []biz.InjectListOption) (int, int, error) {
	options := biz.InjectListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return 0, 0, biz.ErrInjectInvalidArgument
	}
	return options.Limit, options.Offset, nil
}

func scanInjectContent(s rowScanner) (*biz.InjectContent, error) {
	v := &biz.InjectContent{}
	var created, updated sql.NullTime
	if err := s.Scan(&v.ID, &v.ResourceType, &v.ResourceID, &v.VideoID, &v.EpisodeID, &v.MediaID, &v.Action, &v.LastTaskID, &v.Status, &v.FailReason, &created, &updated); err != nil {
		return nil, err
	}
	v.CreateTime, v.UpdateTime = nullTimes(created, updated)
	return v, nil
}

func scanInjectTask(s rowScanner) (*biz.InjectTask, error) {
	v := &biz.InjectTask{}
	var created, updated sql.NullTime
	if err := s.Scan(&v.ID, &v.ContentID, &v.ResourceType, &v.ResourceID, &v.VideoID, &v.EpisodeID, &v.MediaID, &v.Action, &v.Status, &created, &updated); err != nil {
		return nil, err
	}
	v.CreateTime, v.UpdateTime = nullTimes(created, updated)
	return v, nil
}

func scanInjectLog(s rowScanner) (*biz.InjectLog, error) {
	v := &biz.InjectLog{}
	var created, updated sql.NullTime
	if err := s.Scan(&v.ID, &v.TaskID, &v.ContentID, &v.SpID, &v.ResourceType, &v.ResourceID, &v.VideoID, &v.EpisodeID, &v.MediaID, &v.Action, &v.CorrelateID, &v.SyncStatus, &v.AsyncStatus, &v.SyncMessage, &v.AsyncMessage, &created, &updated); err != nil {
		return nil, err
	}
	v.CreateTime, v.UpdateTime = nullTimes(created, updated)
	return v, nil
}

func nullTimes(created, updated sql.NullTime) (time.Time, time.Time) {
	var c, u time.Time
	if created.Valid {
		c = created.Time
	}
	if updated.Valid {
		u = updated.Time
	}
	return c, u
}
