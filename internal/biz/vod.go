package biz

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	ErrVodNotFound    = errors.NotFound("VOD_NOT_FOUND", "vod resource not found")
	ErrVodInvalid     = errors.BadRequest("VOD_INVALID_ARGUMENT", "invalid vod argument")
	ErrVodConflict    = errors.Conflict("VOD_CONFLICT", "vod resource already exists")
	ErrVodHasChildren = errors.Conflict("VOD_HAS_CHILDREN", "resource still has children")
)

type VideoCategory struct {
	ID, ParentID           int64
	Name                   string
	SortOrder              int32
	Status                 uint32
	CreateTime, UpdateTime time.Time
}

type Video struct {
	ID, CategoryID                                                    int64
	VideoCode, Title, Subtitle                                        string
	VideoType                                                         uint32
	PosterVerticalURL, PosterHorizontalURL, ThumbnailURL, Description string
	Year, Duration, Status                                            uint32
	CreateTime, UpdateTime                                            time.Time
}

type Episode struct {
	ID, VideoID            int64
	EpisodeNo              uint32
	Title                  string
	Duration               uint32
	Description            string
	Status                 uint32
	CreateTime, UpdateTime time.Time
}

type Media struct {
	ID, VideoID, EpisodeID        int64
	MediaID, MediaURL, FileFormat string
	Bitrate                       uint32
	Resolution                    string
	FileSize                      uint64
	Duration, Status              uint32
	CreateTime, UpdateTime        time.Time
}

type VodRepo interface {
	ListCategories(context.Context) ([]*VideoCategory, error)
	CreateCategory(context.Context, *VideoCategory) (*VideoCategory, error)
	CreateVideo(context.Context, *Video) (*Video, error)
	FindVideo(context.Context, int64) (*Video, error)
	ListVideos(context.Context, int64, string) ([]*Video, error)
	DeleteVideo(context.Context, int64) error
	CreateEpisode(context.Context, *Episode) (*Episode, error)
	ListEpisodes(context.Context, int64) ([]*Episode, error)
	DeleteEpisode(context.Context, int64) error
	CreateMedia(context.Context, *Media) (*Media, error)
	ListMedia(context.Context, int64) ([]*Media, error)
	DeleteMedia(context.Context, int64) error
}

type VodUsecase struct{ repo VodRepo }

func NewVodUsecase(repo VodRepo) *VodUsecase { return &VodUsecase{repo: repo} }

func (uc *VodUsecase) ListCategories(ctx context.Context) ([]*VideoCategory, error) {
	return uc.repo.ListCategories(ctx)
}
func (uc *VodUsecase) CreateCategory(ctx context.Context, v *VideoCategory) (*VideoCategory, error) {
	if v == nil || strings.TrimSpace(v.Name) == "" || v.ParentID < 0 {
		return nil, ErrVodInvalid
	}
	v.Name = strings.TrimSpace(v.Name)
	if v.Status > 1 {
		return nil, ErrVodInvalid
	}
	return uc.repo.CreateCategory(ctx, v)
}
func (uc *VodUsecase) CreateVideo(ctx context.Context, v *Video) (*Video, error) {
	if v == nil || v.CategoryID <= 0 || strings.TrimSpace(v.VideoCode) == "" || strings.TrimSpace(v.Title) == "" || v.VideoType < 1 || v.VideoType > 4 || v.Status > 1 {
		return nil, ErrVodInvalid
	}
	v.VideoCode, v.Title = strings.TrimSpace(v.VideoCode), strings.TrimSpace(v.Title)
	return uc.repo.CreateVideo(ctx, v)
}
func (uc *VodUsecase) GetVideo(ctx context.Context, id int64) (*Video, error) {
	if id <= 0 {
		return nil, ErrVodInvalid
	}
	return uc.repo.FindVideo(ctx, id)
}
func (uc *VodUsecase) ListVideos(ctx context.Context, categoryID int64, keyword string) ([]*Video, error) {
	if categoryID < 0 {
		return nil, ErrVodInvalid
	}
	return uc.repo.ListVideos(ctx, categoryID, strings.TrimSpace(keyword))
}
func (uc *VodUsecase) DeleteVideo(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrVodInvalid
	}
	return uc.repo.DeleteVideo(ctx, id)
}
func (uc *VodUsecase) CreateEpisode(ctx context.Context, v *Episode) (*Episode, error) {
	if v == nil || v.VideoID <= 0 || v.EpisodeNo == 0 || strings.TrimSpace(v.Title) == "" || v.Status > 1 {
		return nil, ErrVodInvalid
	}
	v.Title = strings.TrimSpace(v.Title)
	return uc.repo.CreateEpisode(ctx, v)
}
func (uc *VodUsecase) ListEpisodes(ctx context.Context, videoID int64) ([]*Episode, error) {
	if videoID <= 0 {
		return nil, ErrVodInvalid
	}
	return uc.repo.ListEpisodes(ctx, videoID)
}
func (uc *VodUsecase) DeleteEpisode(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrVodInvalid
	}
	return uc.repo.DeleteEpisode(ctx, id)
}
func (uc *VodUsecase) CreateMedia(ctx context.Context, v *Media) (*Media, error) {
	if v == nil || v.VideoID <= 0 || v.EpisodeID <= 0 || strings.TrimSpace(v.MediaID) == "" || strings.TrimSpace(v.MediaURL) == "" || v.Status > 1 {
		return nil, ErrVodInvalid
	}
	v.MediaID, v.MediaURL = strings.TrimSpace(v.MediaID), strings.TrimSpace(v.MediaURL)
	return uc.repo.CreateMedia(ctx, v)
}
func (uc *VodUsecase) ListMedia(ctx context.Context, episodeID int64) ([]*Media, error) {
	if episodeID <= 0 {
		return nil, ErrVodInvalid
	}
	return uc.repo.ListMedia(ctx, episodeID)
}
func (uc *VodUsecase) DeleteMedia(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrVodInvalid
	}
	return uc.repo.DeleteMedia(ctx, id)
}
