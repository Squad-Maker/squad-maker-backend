package models

import (
	"errors"
	"squad-maker/database"
	pbSquad "squad-maker/generated/squad"
	otherUtils "squad-maker/utils/other"

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

	var students []*ProjectStudent
	r := tx.InnerJoins("Student").InnerJoins("Position").Where(ProjectStudent{
		ProjectId: p.Id,
	}, "ProjectId").Find(&students)
	if r.Error != nil {
		return nil, r.Error
	}

	for _, student := range students {
		// se ficar lento, vamos ter que otimizar essa função...
		ssd := &StudentSubjectData{}
		r = tx.Joins("CompetenceLevel").Where(StudentSubjectData{
			StudentId: student.StudentId,
			SubjectId: p.SubjectId,
		}, "StudentId", "SubjectId").First(ssd)
		if r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				// por enquanto ignora se o student não tiver vínculo com o subject (o que provavelmente nunca vai acontecer)
				continue
			}
			return nil, r.Error
		}

		message.Students = append(message.Students, &pbSquad.Project_Student{
			Id:                  student.StudentId,
			Name:                student.Student.Name,
			Email:               student.Student.Email,
			CompetenceLevelId:   ssd.CompetenceLevelId,
			CompetenceLevelName: otherUtils.IIf(ssd.CompetenceLevel != nil, ssd.CompetenceLevel.Name, ""),
			Tools:               ssd.Tools,
			PositionId:          student.PositionId,
			PositionName:        student.Position.Name,
		})
	}

	var positions []*ProjectPosition
	r = tx.InnerJoins("Position").Where(ProjectPosition{
		ProjectId: p.Id,
	}, "ProjectId").Find(&positions)
	if r.Error != nil {
		return nil, r.Error
	}

	for _, position := range positions {
		message.Positions = append(message.Positions, &pbSquad.Project_Position{
			Id:    position.PositionId,
			Name:  position.Position.Name,
			Count: position.Count,
		})
	}

	return message, nil
}
