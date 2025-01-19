package endpoints

type SuccessResponse struct {
	Message string `json:"message" example:"<task> successful"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"<task> failed"`
}
