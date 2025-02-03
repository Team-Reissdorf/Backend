package disciplineManagement

import "github.com/gin-gonic/gin"

type DisciplinesResponse struct {
	Message         string   `json:"message" example:"Request successful"`
	DisciplineNames []string `json:"discipline_names" example:"Discipline name"`
}

// GetAllDisciplines returns all discipline names
// @Summary Returns all discipline names
// @Description All discipline names are returned from the database
// @Tags Discipline Management
// @Produce json
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} DisciplinesResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/discipline/get-all [get]
func GetAllDisciplines(c *gin.Context) {

}
