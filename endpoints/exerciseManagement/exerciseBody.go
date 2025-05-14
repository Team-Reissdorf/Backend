package exerciseManagement

type ExerciseBodyWithId struct {
	ExerciseId     uint   `json:"exercise_id" example:"1"`
	Name           string `json:"name" example:"Exercise"`
	Unit           string `json:"unit" example:"minutes"`
	DisciplineName string `json:"discipline_name" example:"Discipline"`
	AgeSpecifics   string `json:"age_specifics" example:"Age specific description"`
}
