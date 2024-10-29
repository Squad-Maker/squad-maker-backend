package database

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	Id        int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

type BaseModelWithSoftDelete struct {
	BaseModel
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func CompanyIdScope(companyId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Statement == nil {
			panic("statement nil") // TODO
		} else if db.Statement.Schema == nil {
			if db.Statement.Model != nil {
				db.Statement.Parse(db.Statement.Model)
				if db.Statement.Schema == nil {
					// panic("failed to parse schema 1") // TODO
					return db
				}
			} else if db.Statement.Dest != nil {
				db.Statement.Parse(db.Statement.Dest)
				if db.Statement.Schema == nil {
					// panic("failed to parse schema 2") // TODO
					return db
				}
			} else {
				// panic("struct for schema not found") // TODO
				return db
			}
		}

		companyField := db.Statement.Schema.FieldsByName["CompanyId"]
		if companyField != nil {
			return db.Where(companyField.DBName+" = ?", companyId)
		}

		return db
	}
}
