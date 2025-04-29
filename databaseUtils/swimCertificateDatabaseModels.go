package databaseUtils

import (
	"time"
)

type SwimCertificate struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Date             time.Time
	DocumentPath     string
	OriginalFileName string

	AthleteId uint `gorm:"index"`
	// BelongsTo Athlete (FK: AthleteId -> Athlete.Id)
	Athlete Athlete `json:"-" gorm:"foreignKey:AthleteId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
