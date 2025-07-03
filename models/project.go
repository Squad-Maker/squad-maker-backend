package models

import (
	"errors"
	"squad-maker/database"
	pbSquad "squad-maker/generated/squad"
	otherUtils "squad-maker/utils/other"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Project struct {
	database.BaseModelWithSoftDelete
	SubjectId   int64 `gorm:"not null"`
	Subject     *Subject
	Name        string         `gorm:"not null"`
	Description string         `gorm:"not null"`
	Tools       pq.StringArray `gorm:"type:text[]"`

	Positions        []*ProjectPosition
	CompetenceLevels []*ProjectCompetenceLevel
	Students         []*ProjectStudent
}

func (p *Project) ConvertToProtobufMessage(tx *gorm.DB) (*pbSquad.Project, error) {
	message := &pbSquad.Project{}
	message.Id = p.Id
	message.Name = p.Name
	message.Description = p.Description
	message.Tools = p.Tools

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

	var competenceLevels []*ProjectCompetenceLevel
	r = tx.InnerJoins("CompetenceLevel").Where(ProjectCompetenceLevel{
		ProjectId: p.Id,
	}, "ProjectId").Find(&competenceLevels)
	if r.Error != nil {
		return nil, r.Error
	}

	for _, competenceLevel := range competenceLevels {
		message.CompetenceLevels = append(message.CompetenceLevels, &pbSquad.Project_CompetenceLevel{
			Id:    competenceLevel.CompetenceLevelId,
			Name:  competenceLevel.CompetenceLevel.Name,
			Count: competenceLevel.Count,
		})
	}

	return message, nil
}
