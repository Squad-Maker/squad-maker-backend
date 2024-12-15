package models

import "github.com/lib/pq"

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
}

func (StudentSubjectData) TableName() string {
	return "student_subject_data"
}
