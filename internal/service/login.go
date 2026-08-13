package service

import (
	"context"
	"encoding/json"
	"strings"

	pb "crow/api/login/v1"
	"crow/internal/biz"

	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
)

type LoginService struct {
	pb.UnimplementedLoginServiceServer

	uc             *biz.LoginUsecase
	operationLogUC *biz.AdminOperationLogUsecase
}

func NewLoginService(uc *biz.LoginUsecase, operationLogUC *biz.AdminOperationLogUsecase) *LoginService {
	return &LoginService{uc: uc, operationLogUC: operationLogUC}
}

func (s *LoginService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	result, err := s.uc.Login(ctx, &biz.LoginInput{
		Account:  req.GetAccount(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	s.writeLoginOperationLog(ctx, req, result)
	return &pb.LoginReply{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		UserId:       result.UserID,
	}, nil
}

func (s *LoginService) writeLoginOperationLog(ctx context.Context, req *pb.LoginRequest, result *biz.LoginResult) {
	if s.operationLogUC == nil || req == nil || result == nil || result.UserID <= 0 {
		return
	}

	requestURL := "/v1/login"
	requestMethod := "POST"
	if httpReq, ok := kratoshttp.RequestFromServerContext(ctx); ok {
		requestURL = httpReq.URL.Path
		requestMethod = httpReq.Method
	}

	params, _ := json.Marshal(map[string]any{
		"account": strings.TrimSpace(req.GetAccount()),
	})

	_, _ = s.operationLogUC.Create(ctx, &biz.AdminOperationLog{
		AdminID:       result.UserID,
		AdminName:     strings.TrimSpace(req.GetAccount()),
		Module:        "auth",
		Action:        "login",
		Description:   "管理员登录",
		RequestMethod: requestMethod,
		RequestURL:    requestURL,
		RequestParams: string(params),
	})
}
