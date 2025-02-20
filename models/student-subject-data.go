package models

import (
	"errors"
	pbSquad "squad-maker/generated/squad"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type StudentSubjectData struct {
	StudentId          int64 `gorm:"primaryKey;autoIncrement:false"`
	Student            *User
	SubjectId          int64 `gorm:"primaryKey;autoIncrement:false"`
	Subject            *Subject
	Tools              pq.StringArray `gorm:"type:text[]"`
	PositionOption1Id  int64          `gorm:"not null"`
	PositionOption1    *Position
	PositionOption2Id  *int64
	PositionOption2    *Position
	PreferredProjectId *int64
	PreferredProject   *Project
	CompetenceLevelId  int64 `gorm:"not null"`
	CompetenceLevel    *CompetenceLevel
	HadFirstUpdate     bool `gorm:"not null;default:false"`
}

func (StudentSubjectData) TableName() string {
	return "student_subject_data"
}

func (ssd *StudentSubjectData) ConvertToProtobufMessage(tx *gorm.DB) (*pbSquad.StudentInSubject, error) {
	if ssd.Student == nil || ssd.Student.Id != ssd.StudentId {
		ssd.Student = &User{}
		r := tx.First(ssd.Student, ssd.StudentId)
		if r.Error != nil {
			return nil, r.Error
		}
	}

	if ssd.CompetenceLevel == nil || ssd.CompetenceLevel.Id != ssd.CompetenceLevelId {
		ssd.CompetenceLevel = &CompetenceLevel{}
		r := tx.First(ssd.CompetenceLevel, ssd.CompetenceLevelId)
		if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, r.Error
		}
	}

	if ssd.PositionOption1 == nil || ssd.PositionOption1.Id != ssd.PositionOption1Id {
		ssd.PositionOption1 = &Position{}
		r := tx.First(ssd.PositionOption1, ssd.PositionOption1Id)
		if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, r.Error
		}
	}

	if ssd.PositionOption2Id != nil && (ssd.PositionOption2 == nil || ssd.PositionOption2.Id != *ssd.PositionOption2Id) {
		ssd.PositionOption2 = &Position{}
		r := tx.First(ssd.PositionOption2, *ssd.PositionOption2Id)
		if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, r.Error
		}
	}

	if ssd.PreferredProjectId != nil && (ssd.PreferredProject == nil || ssd.PreferredProject.Id != *ssd.PreferredProjectId) {
		ssd.PreferredProject = &Project{}
		r := tx.First(ssd.PreferredProject, *ssd.PreferredProjectId)
		if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, r.Error
		}
	}

	var projects []*ProjectStudent
	r := tx.InnerJoins("Project").InnerJoins("Position").Where(ProjectStudent{
		StudentId: ssd.StudentId,
	}, "StudentId").Find(&projects)
	if r.Error != nil {
		return nil, r.Error
	}

	message := &pbSquad.StudentInSubject{}
	message.Id = ssd.StudentId
	message.Name = ssd.Student.Name
	message.Email = ssd.Student.Email
	message.CompetenceLevelId = ssd.CompetenceLevelId
	message.CompetenceLevelName = ssd.CompetenceLevel.Name
	message.Tools = ssd.Tools
	message.PositionOption_1Id = ssd.PositionOption1Id
	message.PositionOption_1Name = ssd.PositionOption1.Name
	if message.PositionOption_2Id != nil {
		message.PositionOption_2Id = ssd.PositionOption2Id
		message.PositionOption_2Name = &ssd.PositionOption2.Name
	}
	if message.PreferredProjectId != nil {
		message.PreferredProjectId = ssd.PreferredProjectId
		message.PreferredProjectName = &ssd.PreferredProject.Name
	}

	for _, project := range projects {
		message.InProjects = append(message.InProjects, &pbSquad.StudentInSubject_Project{
			Id:           project.ProjectId,
			Name:         project.Project.Name,
			PositionId:   project.PositionId,
			PositionName: project.Position.Name,
		})
	}

	return message, nil
}
