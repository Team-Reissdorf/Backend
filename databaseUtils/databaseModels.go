package databaseUtils

import (
	"time"
)

type Athlete struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	FirstName string `json:"first_name" gorm:"uniqueIndex:unique_combination"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date" gorm:"type:date;uniqueIndex:unique_combination"`
	Sex       string `json:"sex"`
	Email     string `json:"email" gorm:"uniqueIndex:unique_combination"`

	TrainerEmail string `json:"trainer_email" gorm:"index"`
	// BelongsTo Trainer (FK: TrainerEmail -> Trainer.Email)
	Trainer Trainer `json:"-" gorm:"foreignKey:TrainerEmail;references:Email;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
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

type ExerciseSpecific struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	ExerciseId uint `gorm:"index"`
	// BelongsTo Exercise (FK: ExerciseId -> Exercise.ExerciseId)
	Exercise Exercise `json:"-" gorm:"foreignKey:ExerciseId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	FromAge     uint   `json:"from_age"`
	ToAge       uint   `json:"to_age"`
	Description string `json:"description"`
}

type Performance struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Points uint64 `json:"points"`
	Date   string `json:"date" gorm:"type:date"`

	ExerciseId uint `gorm:"index"`
	// BelongsTo Exercise (FK: ExerciseId -> Exercise.ExerciseId)
	Exercise Exercise `json:"-" gorm:"foreignKey:ExerciseId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	AthleteId uint `gorm:"index"`
	// BelongsTo Athlete (FK: AthleteId -> Athlete.AthleteId)
	Athlete Athlete `json:"-" gorm:"foreignKey:AthleteId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
