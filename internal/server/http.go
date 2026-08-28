package server

import (
	adminv1 "crow/api/admin/v1"
	cdnv1 "crow/api/cdn/v1"
	loginv1 "crow/api/login/v1"
	todov1 "crow/api/todo/v1"
	vodv1 "crow/api/vod/v1"
	"crow/internal/biz"
	"crow/internal/conf"
	appmiddleware "crow/internal/middleware"
	"crow/internal/service"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"github.com/go-kratos/kratos/v3/transport/http"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, auth *conf.Auth, operationLogUC *biz.AdminOperationLogUsecase, todo *service.TodoService, login *service.LoginService, admin *service.AdminService, cp *service.CpService, sp *service.SpService, cpSp *service.CpSpService, vod *service.VodService) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			validate.Validator(func(req any) error {
				if msg, ok := req.(proto.Message); ok {
					if err := fieldbehavior.ValidateRequiredFields(msg); err != nil {
						return err
					}
				}
				return nil
			}),
			appmiddleware.NewAdminAuth(auth),
			appmiddleware.NewAdminOperationLog(operationLogUC),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	todov1.RegisterTodoServiceHTTPServer(srv, todo)
	loginv1.RegisterLoginServiceHTTPServer(srv, login)
	adminv1.RegisterAdminServiceHTTPServer(srv, admin)
	cdnv1.RegisterCpServiceHTTPServer(srv, cp)
	cdnv1.RegisterSpServiceHTTPServer(srv, sp)
	cdnv1.RegisterCpSpServiceHTTPServer(srv, cpSp)
	vodv1.RegisterVodServiceHTTPServer(srv, vod)
	registerPosterRoutes(srv)
	return srv
}
