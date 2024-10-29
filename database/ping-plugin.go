package database

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	pingPluginSingleton *pingPlugin
)

type pingPlugin struct{}

func init() {
	pingPluginSingleton = &pingPlugin{}
}

func (p *pingPlugin) Name() string {
	return "zion:ping"
}

func (p *pingPlugin) Initialize(db *gorm.DB) error {
	transactionEnabled := func(db *gorm.DB) bool {
		return !db.SkipDefaultTransaction
	}
	transactionDisabled := func(db *gorm.DB) bool {
		return db.SkipDefaultTransaction
	}

	temp := db.Callback().Create()
	temp.Match(transactionEnabled).Before("gorm:begin_transaction").Register("zion:ping", pingCallback)
	temp.Match(transactionDisabled).Before("gorm:before_create").Register("zion:ping", pingCallback)
	temp = nil

	db.Callback().Query().Before("gorm:query").Register("zion:ping", pingCallback)

	temp = db.Callback().Delete()
	temp.Match(transactionEnabled).Before("gorm:begin_transaction").Register("zion:ping", pingCallback)
	temp.Match(transactionDisabled).Before("gorm:before_delete").Register("zion:ping", pingCallback)
	temp = nil

	temp = db.Callback().Update()
	temp.Match(transactionEnabled).Before("gorm:begin_transaction").Register("zion:ping", pingCallback)
	temp.Match(transactionDisabled).Before("gorm:setup_reflect_value").Register("zion:ping", pingCallback)
	temp = nil

	db.Callback().Row().Before("gorm:row").Register("zion:ping", pingCallback)
	db.Callback().Raw().Before("gorm:raw").Register("zion:ping", pingCallback)

	return nil
}

func pingCallback(db *gorm.DB) {
	// try until DeadlineExceeded or Timeout
	err := internalPingCallback(db)
	for err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {

			if db.Error == nil {
				db.Error = err
			}
			return
		}

		time.Sleep(20 * time.Millisecond)
		err = internalPingCallback(db)
	}
}

func internalPingCallback(db *gorm.DB) error {
	internalDb, err := db.DB()

	if err != nil {
		return err
	}

	if db.Statement != nil && db.Statement.Context != nil {
		err = internalDb.PingContext(db.Statement.Context)
	} else {
		err = internalDb.Ping()
	}

	return err
}
