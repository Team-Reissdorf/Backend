package databaseModels

import "gorm.io/gorm"

type Person struct {
	gorm.Model
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type Athlete struct {
	PersonID  uint   `json:"person_id" gorm:"primaryKey"` // FK to person, also PK because athlete is a specialization of person
	Person    Person `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BirthDate string `json:"birth_date"`
	Sex       string `json:"sex"`
	// ToDo: Add FK to trainer
}

type Trainer struct {
	PersonID uint   `json:"person_id" gorm:"primaryKey"` // FK to person, also PK because trainer is a specialization of person
	Person   Person `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Password string `json:"password"`
}
