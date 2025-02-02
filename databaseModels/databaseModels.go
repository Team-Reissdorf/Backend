package databaseModels

import (
	"time"
)

type Athlete struct {
	AthleteId uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	FirstName string `json:"first_name" gorm:"uniqueIndex:unique_combination"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date" gorm:"type:date;uniqueIndex:unique_combination"`
	Sex       string `json:"sex"`
	Email     string `json:"email" gorm:"uniqueIndex:unique_combination"`

	TrainerEmail string `json:"trainer_email" gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type Trainer struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Email    string `gorm:"primaryKey" json:"email"`
	Password string `json:"password"`
}

type Discipline struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Name string `gorm:"primaryKey" json:"name"`
}

type Exercise struct {
	ExerciseId uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `gorm:"index"`

	Name string `json:"name"`
	Unit string `json:"unit"`

	DisciplineName string `json:"discipline_name" gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type Performance struct {
	PerformanceId uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `gorm:"index"`

	Points uint64 `json:"points"`
	Date   string `json:"date" gorm:"type:date"`

	ExerciseId uint `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	AthleteId  uint `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
