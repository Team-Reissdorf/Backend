package rulesetManagement

const CSVCOLUMNCOUNT = 11

var POSSIBLEUNITS = []string{"centimeter", "meter", "second", "minute", "bool", "point"}

type RulesetBody struct {
	RulesetYear    string `json:"ruleset_year"`
	DisciplineName string `json:"discipline_name"`
	ExerciseName   string `json:"exercise_name"`
	Unit           string `json:"unit"`
	Sex            string `json:"sex"`
	FromAge        uint   `json:"from_age"`
	ToAge          uint   `json:"to_age"`
	Bronze         uint64 `json:"bronze"`
	Silver         uint64 `json:"silver"`
	Gold           uint64 `json:"gold"`
	Description    string `json:"description"`
}
