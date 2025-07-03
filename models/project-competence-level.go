package models

type ProjectCompetenceLevel struct {
	ProjectId         int64 `gorm:"primaryKey;autoIncrement:false"`
	Project           *Project
	CompetenceLevelId int64 `gorm:"primaryKey;autoIncrement:false"`
	CompetenceLevel   *CompetenceLevel
	Count             int64 `gorm:"not null;default:0"`
}
