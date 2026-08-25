package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewTodoService, NewLoginService, NewAdminService, NewCpService, NewSpService, NewCpSpService, NewVodService)
