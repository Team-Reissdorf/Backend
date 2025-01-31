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
	BirthDate string `json:"birth_date" gorm:"type:date" gorm:"uniqueIndex:unique_combination"`
	Sex       string `json:"sex"`
	Email     string `json:"email" gorm:"uniqueIndex:unique_combination"`

	TrainerEmail string `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type Trainer struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Email    string `gorm:"primaryKey" json:"email"`
	Password string `json:"password"`
}
