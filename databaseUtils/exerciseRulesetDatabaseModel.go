package databaseUtils

import (
	"time"
)

type ExerciseRuleset struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	RulesetYear string `gorm:"uniqueIndex:unique_combination_exercise_ruleset"`
	// BelongsTo Ruleset (FK: RulesetYear -> Ruleset.Year)
	Ruleset Ruleset `json:"-" gorm:"foreignKey:RulesetYear;references:Year;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ExerciseId uint `gorm:"uniqueIndex:unique_combination_exercise_ruleset"`
	// BelongsTo Exercise (FK: ExerciseId -> Exercise.Id)
	Exercise Exercise `json:"-" gorm:"foreignKey:ExerciseId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
