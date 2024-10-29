package models

import (
	"squad-maker/database"
	pbAuth "squad-maker/generated/auth"
)

type User struct {
	database.BaseModelWithSoftDelete
	UtfprUsername  string          `gorm:"not null;uniqueIndex"` // se deletar o usuário, não vai dar mais de cadastrar por conta deste unique; ter isso em mente
	Type           pbAuth.UserType `gorm:"not null;type:integer"`
	Name           string          `gorm:"not null"`
	Email          string          `gorm:"not null;index"`
	HadFirstUpdate bool            `gorm:"not null;default:false"`
}
