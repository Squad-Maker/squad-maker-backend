package models

type ProjectPosition struct {
	ProjectId  int64 `gorm:"primaryKey;autoIncrement:false"`
	Project    *Project
	PositionId int64 `gorm:"primaryKey;autoIncrement:false"`
	Position   *Position
	Count      int64 `gorm:"not null;default:0"`
}
