package service

import (
	"context"

	pb "crow/api/login/v1"
	"crow/internal/biz"
)

type LoginService struct {
	pb.UnimplementedLoginServiceServer

	uc *biz.LoginUsecase
}

func NewLoginService(uc *biz.LoginUsecase) *LoginService {
	return &LoginService{uc: uc}
}

func (s *LoginService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	result, err := s.uc.Login(ctx, &biz.LoginInput{
		Account:  req.GetAccount(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.LoginReply{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		UserId:       result.UserID,
	}, nil
}
