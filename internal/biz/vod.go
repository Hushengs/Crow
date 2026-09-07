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
	ID, CategoryID, CpID                                              int64
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
	UpdateVideo(context.Context, *Video) (*Video, error)
	FindVideo(context.Context, int64) (*Video, error)
	ListVideos(context.Context, int64, string) ([]*Video, error)
	DeleteVideo(context.Context, int64) error
	CreateEpisode(context.Context, *Episode) (*Episode, error)
	ListEpisodes(context.Context, int64) ([]*Episode, error)
	FindEpisode(context.Context, int64) (*Episode, error)
	DeleteEpisode(context.Context, int64) error
	CreateMedia(context.Context, *Media) (*Media, error)
	FindMedia(context.Context, int64) (*Media, error)
	ListMedia(context.Context, int64) ([]*Media, error)
	DeleteMedia(context.Context, int64) error
}

type VodUsecase struct {
	repo   VodRepo
	inject *InjectUsecase
}

func NewVodUsecase(repo VodRepo, inject *InjectUsecase) *VodUsecase {
	return &VodUsecase{repo: repo, inject: inject}
}

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
	if v == nil || v.CategoryID <= 0 || v.CpID <= 0 || strings.TrimSpace(v.VideoCode) == "" || strings.TrimSpace(v.Title) == "" || v.VideoType < 1 || v.VideoType > 4 || v.Status > 1 {
		return nil, ErrVodInvalid
	}
	normalizeVideo(v)
	item, err := uc.repo.CreateVideo(ctx, v)
	if err != nil {
		return nil, err
	}
	if err := uc.enqueueInject(ctx, InjectResourceVideo, item.ID, item.ID, 0, 0, InjectActionCreate); err != nil {
		return nil, err
	}
	return item, nil
}
func (uc *VodUsecase) UpdateVideo(ctx context.Context, v *Video) (*Video, error) {
	if v == nil || v.ID <= 0 || v.CategoryID <= 0 || v.CpID <= 0 || strings.TrimSpace(v.VideoCode) == "" || strings.TrimSpace(v.Title) == "" || v.VideoType < 1 || v.VideoType > 4 || v.Status > 1 {
		return nil, ErrVodInvalid
	}
	normalizeVideo(v)
	item, err := uc.repo.UpdateVideo(ctx, v)
	if err != nil {
		return nil, err
	}
	if err := uc.enqueueInject(ctx, InjectResourceVideo, item.ID, item.ID, 0, 0, InjectActionUpdate); err != nil {
		return nil, err
	}
	return item, nil
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
	video, err := uc.repo.FindVideo(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.enqueueInject(ctx, InjectResourceVideo, video.ID, video.ID, 0, 0, InjectActionDelete); err != nil {
		return err
	}
	return uc.repo.DeleteVideo(ctx, id)
}
func (uc *VodUsecase) CreateEpisode(ctx context.Context, v *Episode) (*Episode, error) {
	if v == nil || v.VideoID <= 0 || v.EpisodeNo == 0 || strings.TrimSpace(v.Title) == "" || v.Status > 1 {
		return nil, ErrVodInvalid
	}
	v.Title = strings.TrimSpace(v.Title)
	item, err := uc.repo.CreateEpisode(ctx, v)
	if err != nil {
		return nil, err
	}
	if err := uc.enqueueInject(ctx, InjectResourceEpisode, item.ID, item.VideoID, item.ID, 0, InjectActionCreate); err != nil {
		return nil, err
	}
	return item, nil
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
	episode, err := uc.repo.FindEpisode(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.enqueueInject(ctx, InjectResourceEpisode, episode.ID, episode.VideoID, episode.ID, 0, InjectActionDelete); err != nil {
		return err
	}
	return uc.repo.DeleteEpisode(ctx, id)
}
func (uc *VodUsecase) CreateMedia(ctx context.Context, v *Media) (*Media, error) {
	if v == nil || v.VideoID <= 0 || v.EpisodeID <= 0 || strings.TrimSpace(v.MediaID) == "" || strings.TrimSpace(v.MediaURL) == "" || v.Status > 1 {
		return nil, ErrVodInvalid
	}
	v.MediaID, v.MediaURL = strings.TrimSpace(v.MediaID), strings.TrimSpace(v.MediaURL)
	item, err := uc.repo.CreateMedia(ctx, v)
	if err != nil {
		return nil, err
	}
	if err := uc.enqueueInject(ctx, InjectResourceMedia, item.ID, item.VideoID, item.EpisodeID, item.ID, InjectActionCreate); err != nil {
		return nil, err
	}
	return item, nil
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
	media, err := uc.repo.FindMedia(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.enqueueInject(ctx, InjectResourceMedia, media.ID, media.VideoID, media.EpisodeID, media.ID, InjectActionDelete); err != nil {
		return err
	}
	return uc.repo.DeleteMedia(ctx, id)
}

func (uc *VodUsecase) enqueueInject(ctx context.Context, resourceType uint32, resourceID, videoID, episodeID, mediaID int64, action uint32) error {
	if uc.inject == nil {
		return nil
	}
	_, err := uc.inject.Enqueue(ctx, &InjectEnqueue{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		VideoID:      videoID,
		EpisodeID:    episodeID,
		MediaID:      mediaID,
		Action:       action,
	})
	return err
}

func normalizeVideo(v *Video) {
	v.VideoCode = strings.TrimSpace(v.VideoCode)
	v.Title = strings.TrimSpace(v.Title)
	v.Subtitle = strings.TrimSpace(v.Subtitle)
	v.PosterVerticalURL = strings.TrimSpace(v.PosterVerticalURL)
	v.PosterHorizontalURL = strings.TrimSpace(v.PosterHorizontalURL)
	v.ThumbnailURL = strings.TrimSpace(v.ThumbnailURL)
	v.Description = strings.TrimSpace(v.Description)
}
