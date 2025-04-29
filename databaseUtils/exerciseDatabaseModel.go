package databaseUtils

import (
	"time"
)

type Exercise struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Name string `json:"name"`
	Unit string `json:"unit"`

	DisciplineName string `json:"discipline_name" gorm:"index"`
	// BelongsTo Discipline (FK: DisciplineName -> Discipline.Name)
	Discipline Discipline `json:"-" gorm:"foreignKey:DisciplineName;references:Name;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
