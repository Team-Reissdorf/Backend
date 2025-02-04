package disciplineManagement

import (
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

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
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetAllDisciplines")
	defer span.End()

	// Get all disciplines from the database
	var disciplines []databaseUtils.Discipline
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(databaseUtils.Discipline{}).Find(&disciplines).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get all disciplines")
		endpoints.Logger.Error(ctx, err1)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the disciplines"})
		return
	}

	// Translate the database objects to a string array
	disciplineNames := make([]string, len(disciplines))
	for idx, discipline := range disciplines {
		disciplineNames[idx] = discipline.Name
	}

	c.JSON(
		http.StatusOK,
		DisciplinesResponse{
			Message:         "Request successful",
			DisciplineNames: disciplineNames,
		},
	)
}
