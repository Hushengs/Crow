package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"crow/internal/biz"
)

type vodRepo struct{ data *Data }

func NewVodRepo(data *Data) biz.VodRepo { return &vodRepo{data: data} }

type rowScanner interface{ Scan(...any) error }

func scanCategory(s rowScanner) (*biz.VideoCategory, error) {
	v := &biz.VideoCategory{}
	var created, updated sql.NullTime
	if err := s.Scan(&v.ID, &v.ParentID, &v.Name, &v.SortOrder, &v.Status, &created, &updated); err != nil {
		return nil, err
	}
	if created.Valid {
		v.CreateTime = created.Time
	}
	if updated.Valid {
		v.UpdateTime = updated.Time
	}
	return v, nil
}

func scanVideo(s rowScanner) (*biz.Video, error) {
	v := &biz.Video{}
	var year sql.NullInt64
	var description sql.NullString
	var created, updated sql.NullTime
	if err := s.Scan(&v.ID, &v.CategoryID, &v.VideoCode, &v.Title, &v.Subtitle, &v.VideoType,
		&v.PosterVerticalURL, &v.PosterHorizontalURL, &v.ThumbnailURL, &description, &year,
		&v.Duration, &v.Status, &created, &updated); err != nil {
		return nil, err
	}
	if description.Valid {
		v.Description = description.String
	}
	if year.Valid {
		v.Year = uint32(year.Int64)
	}
	if created.Valid {
		v.CreateTime = created.Time
	}
	if updated.Valid {
		v.UpdateTime = updated.Time
	}
	return v, nil
}

func scanEpisode(s rowScanner) (*biz.Episode, error) {
	v := &biz.Episode{}
	var created, updated sql.NullTime
	if err := s.Scan(&v.ID, &v.VideoID, &v.EpisodeNo, &v.Title, &v.Duration, &v.Description, &v.Status, &created, &updated); err != nil {
		return nil, err
	}
	if created.Valid {
		v.CreateTime = created.Time
	}
	if updated.Valid {
		v.UpdateTime = updated.Time
	}
	return v, nil
}

func scanMedia(s rowScanner) (*biz.Media, error) {
	v := &biz.Media{}
	var created, updated sql.NullTime
	if err := s.Scan(&v.ID, &v.VideoID, &v.EpisodeID, &v.MediaID, &v.MediaURL, &v.FileFormat,
		&v.Bitrate, &v.Resolution, &v.FileSize, &v.Duration, &v.Status, &created, &updated); err != nil {
		return nil, err
	}
	if created.Valid {
		v.CreateTime = created.Time
	}
	if updated.Valid {
		v.UpdateTime = updated.Time
	}
	return v, nil
}

func mapVodWriteError(err error) error {
	if isMySQLDuplicateEntryError(err) {
		return biz.ErrVodConflict
	}
	return err
}

func (r *vodRepo) ListCategories(ctx context.Context) ([]*biz.VideoCategory, error) {
	rows, err := r.data.db.QueryContext(ctx, `SELECT id,parent_id,name,sort_order,status,create_date,update_date FROM video_category ORDER BY parent_id,sort_order,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*biz.VideoCategory, 0)
	for rows.Next() {
		v, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *vodRepo) CreateCategory(ctx context.Context, v *biz.VideoCategory) (*biz.VideoCategory, error) {
	if v.ParentID > 0 {
		var exists int
		if err := r.data.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM video_category WHERE id=?`, v.ParentID).Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, biz.ErrVodInvalid
		}
	}
	res, err := r.data.db.ExecContext(ctx, `INSERT INTO video_category(parent_id,name,sort_order,status) VALUES(?,?,?,?)`, v.ParentID, v.Name, v.SortOrder, v.Status)
	if err != nil {
		return nil, mapVodWriteError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return scanCategory(r.data.db.QueryRowContext(ctx, `SELECT id,parent_id,name,sort_order,status,create_date,update_date FROM video_category WHERE id=?`, id))
}

const videoColumns = `id,category_id,video_code,title,subtitle,video_type,poster_vertical_url,poster_horizontal_url,thumbnail_url,description,year,duration,status,create_date,update_date`

func (r *vodRepo) CreateVideo(ctx context.Context, v *biz.Video) (*biz.Video, error) {
	res, err := r.data.db.ExecContext(ctx, `INSERT INTO video(category_id,video_code,title,subtitle,video_type,poster_vertical_url,poster_horizontal_url,thumbnail_url,description,year,duration,status) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		v.CategoryID, v.VideoCode, v.Title, v.Subtitle, v.VideoType, v.PosterVerticalURL, v.PosterHorizontalURL, v.ThumbnailURL, v.Description, nullableYear(v.Year), v.Duration, v.Status)
	if err != nil {
		return nil, mapVodWriteError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindVideo(ctx, id)
}

func (r *vodRepo) UpdateVideo(ctx context.Context, v *biz.Video) (*biz.Video, error) {
	res, err := r.data.db.ExecContext(ctx, `UPDATE video SET category_id=?,video_code=?,title=?,subtitle=?,video_type=?,poster_vertical_url=?,poster_horizontal_url=?,thumbnail_url=?,description=?,year=?,duration=?,status=? WHERE id=?`,
		v.CategoryID, v.VideoCode, v.Title, v.Subtitle, v.VideoType, v.PosterVerticalURL, v.PosterHorizontalURL, v.ThumbnailURL, v.Description, nullableYear(v.Year), v.Duration, v.Status, v.ID)
	if err != nil {
		return nil, mapVodWriteError(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, biz.ErrVodNotFound
	}
	return r.FindVideo(ctx, v.ID)
}

func nullableYear(year uint32) any {
	if year == 0 {
		return nil
	}
	return year
}

func (r *vodRepo) FindVideo(ctx context.Context, id int64) (*biz.Video, error) {
	v, err := scanVideo(r.data.db.QueryRowContext(ctx, `SELECT `+videoColumns+` FROM video WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, biz.ErrVodNotFound
	}
	return v, err
}

func (r *vodRepo) ListVideos(ctx context.Context, categoryID int64, keyword string) ([]*biz.Video, error) {
	query := `SELECT ` + videoColumns + ` FROM video WHERE 1=1`
	args := make([]any, 0, 3)
	if categoryID > 0 {
		query += ` AND category_id=?`
		args = append(args, categoryID)
	}
	if keyword != "" {
		query += ` AND (title LIKE ? OR subtitle LIKE ? OR video_code LIKE ?)`
		like := "%" + keyword + "%"
		args = append(args, like, like, like)
	}
	query += ` ORDER BY id DESC`
	rows, err := r.data.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*biz.Video, 0)
	for rows.Next() {
		v, err := scanVideo(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *vodRepo) DeleteVideo(ctx context.Context, id int64) error {
	var children int
	if err := r.data.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM episode WHERE video_id=?`, id).Scan(&children); err != nil {
		return err
	}
	if children > 0 {
		return biz.ErrVodHasChildren
	}
	return deleteVodRow(ctx, r.data.db, "video", id)
}

func (r *vodRepo) CreateEpisode(ctx context.Context, v *biz.Episode) (*biz.Episode, error) {
	res, err := r.data.db.ExecContext(ctx, `INSERT INTO episode(video_id,episode_no,title,duration,description,status) VALUES(?,?,?,?,?,?)`,
		v.VideoID, v.EpisodeNo, v.Title, v.Duration, v.Description, v.Status)
	if err != nil {
		return nil, mapVodWriteError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return scanEpisode(r.data.db.QueryRowContext(ctx, `SELECT id,video_id,episode_no,title,duration,description,status,create_date,update_date FROM episode WHERE id=?`, id))
}

func (r *vodRepo) ListEpisodes(ctx context.Context, videoID int64) ([]*biz.Episode, error) {
	rows, err := r.data.db.QueryContext(ctx, `SELECT id,video_id,episode_no,title,duration,description,status,create_date,update_date FROM episode WHERE video_id=? ORDER BY episode_no,id`, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*biz.Episode, 0)
	for rows.Next() {
		v, err := scanEpisode(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *vodRepo) DeleteEpisode(ctx context.Context, id int64) error {
	var children int
	if err := r.data.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM media WHERE episode_id=?`, id).Scan(&children); err != nil {
		return err
	}
	if children > 0 {
		return biz.ErrVodHasChildren
	}
	return deleteVodRow(ctx, r.data.db, "episode", id)
}

func (r *vodRepo) CreateMedia(ctx context.Context, v *biz.Media) (*biz.Media, error) {
	var episodeVideoID int64
	if err := r.data.db.QueryRowContext(ctx, `SELECT video_id FROM episode WHERE id=?`, v.EpisodeID).Scan(&episodeVideoID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrVodInvalid
		}
		return nil, err
	}
	if episodeVideoID != v.VideoID {
		return nil, biz.ErrVodInvalid
	}
	res, err := r.data.db.ExecContext(ctx, `INSERT INTO media(video_id,episode_id,media_id,media_url,file_format,bitrate,resolution,file_size,duration,status) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		v.VideoID, v.EpisodeID, v.MediaID, v.MediaURL, v.FileFormat, v.Bitrate, v.Resolution, v.FileSize, v.Duration, v.Status)
	if err != nil {
		return nil, mapVodWriteError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return scanMedia(r.data.db.QueryRowContext(ctx, `SELECT id,video_id,episode_id,media_id,media_url,file_format,bitrate,resolution,file_size,duration,status,create_date,update_date FROM media WHERE id=?`, id))
}

func (r *vodRepo) ListMedia(ctx context.Context, episodeID int64) ([]*biz.Media, error) {
	rows, err := r.data.db.QueryContext(ctx, `SELECT id,video_id,episode_id,media_id,media_url,file_format,bitrate,resolution,file_size,duration,status,create_date,update_date FROM media WHERE episode_id=? ORDER BY bitrate DESC,id`, episodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*biz.Media, 0)
	for rows.Next() {
		v, err := scanMedia(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *vodRepo) DeleteMedia(ctx context.Context, id int64) error {
	return deleteVodRow(ctx, r.data.db, "media", id)
}

func deleteVodRow(ctx context.Context, db *sql.DB, table string, id int64) error {
	switch table {
	case "video", "episode", "media":
	default:
		return fmt.Errorf("unsupported vod table %q", table)
	}
	res, err := db.ExecContext(ctx, `DELETE FROM `+table+` WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return biz.ErrVodNotFound
	}
	return nil
}
