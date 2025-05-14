package databaseUtils

import (
	"time"
)

type Performance struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Points uint64 `json:"points"`
	Medal  string `json:"medal"`
	Date   string `json:"date" gorm:"type:date"`

	ExerciseId uint `gorm:"index"`
	// BelongsTo Exercise (FK: ExerciseId -> Exercise.Id)
	Exercise Exercise `json:"-" gorm:"foreignKey:ExerciseId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	AthleteId uint `gorm:"index"`
	// BelongsTo Athlete (FK: AthleteId -> Athlete.Id)
	Athlete Athlete `json:"-" gorm:"foreignKey:AthleteId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
