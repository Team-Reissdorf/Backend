package databaseUtils

import (
	"time"
)

type Trainer struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Email    string `gorm:"primaryKey" json:"email"`
	Password string `json:"password"`
}
