//go:build debug
// +build debug

package database

import (
	"gorm.io/gorm/logger"
)

func init() {
	logLevel = logger.Info
	ignoreRecordNotFoundError = false
}
