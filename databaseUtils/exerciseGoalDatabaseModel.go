package databaseUtils

import (
	"time"
)

type ExerciseGoal struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	RulesetId uint `gorm:"index;uniqueIndex:unique_combination_exercise_goals"`
	// BelongsTo ExerciseRuleset (FK: RulesetId -> ExerciseRuleset.Id)
	ExerciseRuleset ExerciseRuleset `json:"-" gorm:"foreignKey:RulesetId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	FromAge     uint   `json:"from_age" gorm:"uniqueIndex:unique_combination_exercise_goals"`
	ToAge       uint   `json:"to_age" gorm:"uniqueIndex:unique_combination_exercise_goals"`
	Sex         string `json:"sex" gorm:"uniqueIndex:unique_combination_exercise_goals;type:char(1);not null"`
	Bronze      uint64 `json:"bronze"` // Time: ms; Distance: cm; Points; Bool: <0|1>
	Silver      uint64 `json:"silver"` // Time: ms; Distance: cm; Points; Bool: <0|1>
	Gold        uint64 `json:"gold"`   // Time: ms; Distance: cm; Points; Bool: <0|1>
	Description string `json:"description"`
}
