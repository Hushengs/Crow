package data

import (
	"errors"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func isMySQLDuplicateEntryError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
