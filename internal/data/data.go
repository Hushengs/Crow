package data

import (
	"context"
	"crow/internal/conf"
	"database/sql"

	"github.com/go-kratos/kratos/v3/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewTodoRepo,
	NewLoginRepo,
	NewAdminRepo,
	NewRoleRepo,
	NewAdminRoleRepo,
	NewPermissionRepo,
	NewGroupPermissionRepo,
	NewAdminOperationLogRepo,
	NewSystemLogRepo,
)

// Data .
type Data struct {
	db *sql.DB
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	db, err := sql.Open(c.GetDatabase().GetDriver(), c.GetDatabase().GetSource())
	if err != nil {
		return nil, nil, err
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	cleanup := func() {
		log.Info("closing the data resources")
		_ = db.Close()
	}
	return &Data{db: db}, cleanup, nil
}
