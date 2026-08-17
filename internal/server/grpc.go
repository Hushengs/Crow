package server

import (
	adminv1 "crow/api/admin/v1"
	cdnv1 "crow/api/cdn/v1"
	loginv1 "crow/api/login/v1"
	todov1 "crow/api/todo/v1"
	"crow/internal/conf"
	"crow/internal/service"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, todo *service.TodoService, login *service.LoginService, admin *service.AdminService, cp *service.CpService, sp *service.SpService, cpSp *service.CpSpService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	todov1.RegisterTodoServiceServer(srv, todo)
	loginv1.RegisterLoginServiceServer(srv, login)
	adminv1.RegisterAdminServiceServer(srv, admin)
	cdnv1.RegisterCpServiceServer(srv, cp)
	cdnv1.RegisterSpServiceServer(srv, sp)
	cdnv1.RegisterCpSpServiceServer(srv, cpSp)
	return srv
}
