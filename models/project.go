package models

import "squad-maker/database"

type Project struct {
	database.BaseModelWithSoftDelete
	SubjectId   int64 `gorm:"not null"`
	Subject     *Subject
	Name        string `gorm:"not null"`
	Description string `gorm:"not null"`
}
