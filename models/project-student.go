package models

type ProjectStudent struct {
	ProjectId  int64 `gorm:"primaryKey;autoIncrement:false"`
	Project    *Project
	StudentId  int64 `gorm:"primaryKey;autoIncrement:false"`
	Student    *User
	PositionId int64 `gorm:"not null"`
	Position   *Position
}
