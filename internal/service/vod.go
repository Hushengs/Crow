package service

import (
	"context"

	pb "crow/api/vod/v1"
	"crow/internal/biz"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type VodService struct {
	pb.UnimplementedVodServiceServer
	uc *biz.VodUsecase
}

func NewVodService(uc *biz.VodUsecase) *VodService { return &VodService{uc: uc} }

func (s *VodService) ListCategories(ctx context.Context, _ *pb.ListCategoriesRequest) (*pb.CategorySet, error) {
	items, err := s.uc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := &pb.CategorySet{Categories: make([]*pb.Category, 0, len(items))}
	for _, item := range items {
		out.Categories = append(out.Categories, categoryReply(item))
	}
	return out, nil
}
func (s *VodService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	in := req.GetCategory()
	if in == nil {
		return nil, biz.ErrVodInvalid
	}
	item, err := s.uc.CreateCategory(ctx, &biz.VideoCategory{ID: in.Id, ParentID: in.ParentId, Name: in.Name, SortOrder: in.SortOrder, Status: in.Status})
	if err != nil {
		return nil, err
	}
	return categoryReply(item), nil
}
func (s *VodService) CreateVideo(ctx context.Context, req *pb.CreateVideoRequest) (*pb.Video, error) {
	item, err := s.uc.CreateVideo(ctx, videoBiz(req.GetVideo()))
	if err != nil {
		return nil, err
	}
	return videoReply(item), nil
}
func (s *VodService) GetVideo(ctx context.Context, req *pb.GetVideoRequest) (*pb.Video, error) {
	item, err := s.uc.GetVideo(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return videoReply(item), nil
}
func (s *VodService) ListVideos(ctx context.Context, req *pb.ListVideosRequest) (*pb.VideoSet, error) {
	items, err := s.uc.ListVideos(ctx, req.GetCategoryId(), req.GetKeyword())
	if err != nil {
		return nil, err
	}
	out := &pb.VideoSet{Videos: make([]*pb.Video, 0, len(items))}
	for _, item := range items {
		out.Videos = append(out.Videos, videoReply(item))
	}
	return out, nil
}
func (s *VodService) DeleteVideo(ctx context.Context, req *pb.DeleteVideoRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteVideo(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
func (s *VodService) CreateEpisode(ctx context.Context, req *pb.CreateEpisodeRequest) (*pb.Episode, error) {
	in := req.GetEpisode()
	if in == nil {
		return nil, biz.ErrVodInvalid
	}
	item, err := s.uc.CreateEpisode(ctx, &biz.Episode{ID: in.Id, VideoID: in.VideoId, EpisodeNo: in.EpisodeNo, Title: in.Title, Duration: in.Duration, Description: in.Description, Status: in.Status})
	if err != nil {
		return nil, err
	}
	return episodeReply(item), nil
}
func (s *VodService) ListEpisodes(ctx context.Context, req *pb.ListEpisodesRequest) (*pb.EpisodeSet, error) {
	items, err := s.uc.ListEpisodes(ctx, req.GetVideoId())
	if err != nil {
		return nil, err
	}
	out := &pb.EpisodeSet{Episodes: make([]*pb.Episode, 0, len(items))}
	for _, item := range items {
		out.Episodes = append(out.Episodes, episodeReply(item))
	}
	return out, nil
}
func (s *VodService) DeleteEpisode(ctx context.Context, req *pb.DeleteEpisodeRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteEpisode(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
func (s *VodService) CreateMedia(ctx context.Context, req *pb.CreateMediaRequest) (*pb.Media, error) {
	in := req.GetMedia()
	if in == nil {
		return nil, biz.ErrVodInvalid
	}
	item, err := s.uc.CreateMedia(ctx, &biz.Media{ID: in.Id, VideoID: in.VideoId, EpisodeID: in.EpisodeId, MediaID: in.MediaId, MediaURL: in.MediaUrl, FileFormat: in.FileFormat, Bitrate: in.Bitrate, Resolution: in.Resolution, FileSize: in.FileSize, Duration: in.Duration, Status: in.Status})
	if err != nil {
		return nil, err
	}
	return mediaReply(item), nil
}
func (s *VodService) ListMedia(ctx context.Context, req *pb.ListMediaRequest) (*pb.MediaSet, error) {
	items, err := s.uc.ListMedia(ctx, req.GetEpisodeId())
	if err != nil {
		return nil, err
	}
	out := &pb.MediaSet{Media: make([]*pb.Media, 0, len(items))}
	for _, item := range items {
		out.Media = append(out.Media, mediaReply(item))
	}
	return out, nil
}
func (s *VodService) DeleteMedia(ctx context.Context, req *pb.DeleteMediaRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteMedia(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func categoryReply(in *biz.VideoCategory) *pb.Category {
	return &pb.Category{Id: in.ID, ParentId: in.ParentID, Name: in.Name, SortOrder: in.SortOrder, Status: in.Status, CreateTime: timestamppb.New(in.CreateTime), UpdateTime: timestamppb.New(in.UpdateTime)}
}
func videoBiz(in *pb.Video) *biz.Video {
	if in == nil {
		return nil
	}
	return &biz.Video{ID: in.Id, CategoryID: in.CategoryId, VideoCode: in.VideoCode, Title: in.Title, Subtitle: in.Subtitle, VideoType: in.VideoType, PosterVerticalURL: in.PosterVerticalUrl, PosterHorizontalURL: in.PosterHorizontalUrl, ThumbnailURL: in.ThumbnailUrl, Description: in.Description, Year: in.Year, Duration: in.Duration, Status: in.Status}
}
func videoReply(in *biz.Video) *pb.Video {
	return &pb.Video{Id: in.ID, CategoryId: in.CategoryID, VideoCode: in.VideoCode, Title: in.Title, Subtitle: in.Subtitle, VideoType: in.VideoType, PosterVerticalUrl: in.PosterVerticalURL, PosterHorizontalUrl: in.PosterHorizontalURL, ThumbnailUrl: in.ThumbnailURL, Description: in.Description, Year: in.Year, Duration: in.Duration, Status: in.Status, CreateTime: timestamppb.New(in.CreateTime), UpdateTime: timestamppb.New(in.UpdateTime)}
}
func episodeReply(in *biz.Episode) *pb.Episode {
	return &pb.Episode{Id: in.ID, VideoId: in.VideoID, EpisodeNo: in.EpisodeNo, Title: in.Title, Duration: in.Duration, Description: in.Description, Status: in.Status, CreateTime: timestamppb.New(in.CreateTime), UpdateTime: timestamppb.New(in.UpdateTime)}
}
func mediaReply(in *biz.Media) *pb.Media {
	return &pb.Media{Id: in.ID, VideoId: in.VideoID, EpisodeId: in.EpisodeID, MediaId: in.MediaID, MediaUrl: in.MediaURL, FileFormat: in.FileFormat, Bitrate: in.Bitrate, Resolution: in.Resolution, FileSize: in.FileSize, Duration: in.Duration, Status: in.Status, CreateTime: timestamppb.New(in.CreateTime), UpdateTime: timestamppb.New(in.UpdateTime)}
}
