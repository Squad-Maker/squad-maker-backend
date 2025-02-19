package models

import (
	"squad-maker/database"
	pbSquad "squad-maker/generated/squad"

	"gorm.io/gorm"
)

type CompetenceLevel struct {
	database.BaseModelWithSoftDelete
	SubjectId int64 `gorm:"not null"`
	Subject   *Subject
	Name      string `gorm:"not null"`
	// Weight    float64 `gorm:"not null;default:0"`
}

func (cl *CompetenceLevel) ConvertToProtobufMessage(tx *gorm.DB) (*pbSquad.CompetenceLevel, error) {
	message := &pbSquad.CompetenceLevel{}
	message.Id = cl.Id
	message.Name = cl.Name
	// message.Weight = cl.Weight

	return message, nil
}
