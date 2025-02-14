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
