package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewTodoUsecase,
	NewLoginTokenGenerator,
	NewLoginUsecase,
	NewAdminUsecase,
	NewRoleUsecase,
	NewAdminRoleUsecase,
	NewPermissionUsecase,
	NewGroupPermissionUsecase,
	NewAdminOperationLogUsecase,
	NewSystemLogUsecase,
)
