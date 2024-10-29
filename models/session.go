package models

import (
	"squad-maker/database"
	"time"
)

type Session struct {
	database.BaseModelWithSoftDelete
	UserId      int64 `gorm:"not null"`
	User        *User
	Token       string    `gorm:"index"`
	LastRefresh time.Time `gorm:"index"`
	DoNotExpire bool      `gorm:"not null;default:0"`
}
