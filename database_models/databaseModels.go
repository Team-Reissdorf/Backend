package database_models

type Athlete struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	BirthDate string `json:"birth_date"`
	Sex       string `json:"sex"`
}
