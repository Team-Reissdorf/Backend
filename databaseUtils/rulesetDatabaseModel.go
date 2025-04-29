package databaseUtils

import (
	"time"
)

type Ruleset struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Year string `gorm:"primaryKey" json:"year"`
}
