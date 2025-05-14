package databaseUtils

import (
	"time"
)

type Discipline struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Name string `gorm:"primaryKey" json:"name"`
}
