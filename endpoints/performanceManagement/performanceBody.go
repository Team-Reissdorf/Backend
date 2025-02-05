package performanceManagement

type PerformanceBody struct {
	Points     uint64 `json:"points" example:"1"`
	Date       string `json:"date" example:"YYYY-MM-DD"`
	ExerciseId uint   `json:"exercise_id" example:"1"`
	AthleteId  uint   `json:"athlete_id" example:"1"`
}

type PerformanceBodyWithId struct {
	PerformanceId uint   `json:"performance_id" example:"1"`
	Points        uint64 `json:"points" example:"1"`
	Date          string `json:"date" example:"YYYY-MM-DD"`
	ExerciseId    uint   `json:"exercise_id" example:"1"`
	AthleteId     uint   `json:"athlete_id" example:"1"`
	Medal         string `json:"medal" example:"gold"`
}
