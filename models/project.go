package models

import (
	"squad-maker/database"
	pbSquad "squad-maker/generated/squad"

	"gorm.io/gorm"
)

type Project struct {
	database.BaseModelWithSoftDelete
	SubjectId   int64 `gorm:"not null"`
	Subject     *Subject
	Name        string `gorm:"not null"`
	Description string `gorm:"not null"`

	Positions []*ProjectPosition
	Students  []*ProjectStudent
}

func (p *Project) ConvertToProtobufMessage(tx *gorm.DB) (*pbSquad.Project, error) {
	message := &pbSquad.Project{}
	message.Id = p.Id
	message.Name = p.Name
	message.Description = p.Description

	return message, nil
}
