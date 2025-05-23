package athleteManagement

type AthleteBody struct {
	FirstName string `json:"first_name" example:"Bob"`
	LastName  string `json:"last_name" example:"Alice"`
	Email     string `json:"email" example:"bob.alice@example.com"`
	BirthDate string `json:"birth_date" example:"YYYY-MM-DD"`
	Sex       string `json:"sex" example:"<m|f|d>"`
}

type AthleteBodyWithId struct {
	AthleteId uint   `json:"athlete_id" example:"1"`
	FirstName string `json:"first_name" example:"Bob"`
	LastName  string `json:"last_name" example:"Alice"`
	Email     string `json:"email" example:"bob.alice@example.com"`
	BirthDate string `json:"birth_date" example:"YYYY-MM-DD"`
	Sex       string `json:"sex" example:"<m|f|d>"`
	SwimCert  bool   `json:"swim_cert"`
}

type SwimCertificateWithID struct {
	ID        uint `json:"cert_id" example:"1"`
	AthleteId uint `json:"athlete_id" example:"1"`
}
