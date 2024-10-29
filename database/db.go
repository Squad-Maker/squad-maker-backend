package database

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"squad-maker/utils/env"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var (
	db                        *gorm.DB
	logLevel                  = logger.Warn
	ignoreRecordNotFoundError = true
)

func GetConnection() (*gorm.DB, error) {
	err := ensureDBInitialized()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetConnectionWithContext(ctx context.Context) (*gorm.DB, error) {
	err := ensureDBInitialized()
	if err != nil {
		return nil, err
	}

	return db.WithContext(ctx), nil
}

func ensureDBInitialized() error {
	if db != nil {
		return nil
	}

	con, err := gorm.Open(getDriverConnection(),
		&gorm.Config{
			NamingStrategy: NamingStrategy{},
			Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logLevel,
				IgnoreRecordNotFoundError: ignoreRecordNotFoundError,
				Colorful:                  true,
			}),
			QueryFields: true,
		})

	if err == nil {
		err = con.Use(pingPluginSingleton)
		if err == nil {
			internalDb, err := con.DB()
			if err == nil {
				internalDb.SetConnMaxIdleTime(230 * time.Second)
				// https://github.com/golang/go/issues/41114
				internalDb.SetConnMaxLifetime(30 * time.Minute)
			}
		}
	}

	db = con

	return err
}

func getDriverConnection() gorm.Dialector {
	return postgres.New(postgres.Config{
		DSN:                  GetPostgresDSN(),
		PreferSimpleProtocol: false,
	})
}

func GetLockingClause() clause.Locking {
	hint := clause.Locking{Strength: "NO KEY UPDATE"}
	hint.Table = clause.Table{Name: clause.CurrentTable}

	return hint
}

func GetPostgresDSN() string {
	host, _ := env.GetStr("DB_HOST")
	port, _ := env.GetInt32("DB_PORT")
	database, _ := env.GetStr("DB_NAME")
	username, _ := env.GetStr("DB_USER")
	password, _ := env.GetStr("DB_PASS")

	return "user=" + username + " password=" + password + " dbname=" + database + " port=" + strconv.FormatInt(int64(port), 10) + " host=" + host + " default_query_exec_mode=cache_describe Timezone=America/Sao_Paulo"
}
