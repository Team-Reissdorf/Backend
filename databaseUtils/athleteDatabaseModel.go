package databaseUtils

import (
	"time"
)

type Athlete struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	FirstName string `json:"first_name" gorm:"uniqueIndex:unique_combination_athletes"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date" gorm:"type:date;uniqueIndex:unique_combination_athletes"`
	Sex       string `json:"sex" gorm:"type:char(1);not null"`
	Email     string `json:"email" gorm:"uniqueIndex:unique_combination_athletes"`

	TrainerEmail string `json:"trainer_email" gorm:"index"`
	// BelongsTo Trainer (FK: TrainerEmail -> Trainer.Email)
	Trainer Trainer `json:"-" gorm:"foreignKey:TrainerEmail;references:Email;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
