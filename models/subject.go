package models

import "squad-maker/database"

type Subject struct {
	database.BaseModelWithSoftDelete
	Name string `gorm:"not null"`

	Students []*StudentSubjectData
}
