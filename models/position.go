package models

import (
	"squad-maker/database"
	pbSquad "squad-maker/generated/squad"

	"gorm.io/gorm"
)

type Position struct {
	database.BaseModelWithSoftDelete
	SubjectId int64 `gorm:"not null"`
	Subject   *Subject
	Name      string `gorm:"not null"`
}

func (p *Position) ConvertToProtobufMessage(tx *gorm.DB) (*pbSquad.Position, error) {
	message := &pbSquad.Position{}
	message.Id = p.Id
	message.SubjectId = p.SubjectId
	message.Name = p.Name

	return message, nil
}
