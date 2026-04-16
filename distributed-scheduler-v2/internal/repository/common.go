package repository

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func IsDuplicateErr(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
