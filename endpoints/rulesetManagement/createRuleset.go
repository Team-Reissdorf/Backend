package rulesetManagement

import (
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	possibleUnits = []string{"centimeter", "meter", "second", "minute", "bool", "point"}
)

// CreateRuleset creates new ruleset entries in the db from a csv file
// @Summary Creates new ruleset entries from csv file
// @Description Upload a CSV file to create multiple ruleset entries. Needs to contain 11 columns.
// @Tags Ruleset Management
// @Accept multipart/form-data
// @Produce json
// @Param RulesetEntries formData file true "CSV file containing details of the ruleset"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 409 {object} endpoints.ErrorResponse "All ruleset entries already exist; none have been created"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/ruleset/create [post]
func CreateRuleset(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreateRulesetEntries")
	defer span.End()

	// Bind body to csv file
	file, err1 := c.FormFile("RulesetEntries")
	if err1 != nil || file == nil {
		err1 = errors.Wrap(err1, "Failed to get the file")
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File is missing or invalid"})
		return
	}

	// Check MIME type
	fileHeader := file.Header.Get("Content-Type")
	if !strings.HasPrefix(fileHeader, "text/csv") && !strings.HasPrefix(fileHeader, "application/vnd.ms-excel") {
		err := errors.New("Invalid file type, only CSV files are allowed")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
		return
	}

	// Get the user id from the context
	// trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Open the CSV file
	fileContent, err2 := file.Open()
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to open file")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Could not open file"})
		return
	}
	defer func(fileContent multipart.File) {
		err := fileContent.Close()
		if err != nil {
			err = errors.Wrap(err, "Failed to close file")
			endpoints.Logger.Error(ctx, err)
		}
	}(fileContent)

	// Read the CSV file
	reader := csv.NewReader(fileContent)
	reader.Comma = ';'
	records, err3 := reader.ReadAll()
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to read csv. Invalid CSV format?")
		endpoints.Logger.Warn(ctx, err3)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File could not be read. Invalid CSV format?"})
		return
	}

	// Parse data
	var rulesets []RulesetBody
	for _, record := range records {
		// Ensure the column count is correct
		if len(record) != CSVCOLUMNCOUNT {
			err := errors.New("Inconsistent number of columns in the CSV file")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
			return
		}

		// Parse age values
		FromAge, errA := strconv.Atoi(record[5])
		if errA != nil {
			errA = errors.Wrap(errA, "Failed to parse from age")
			FlowWatch.GetLogHelper().Debug(ctx, errA)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: fmt.Sprintf("Invalid from age value: %s", record[5])})
			return
		}

		ToAge, errB := strconv.Atoi(record[6])
		if errB != nil {
			errB = errors.Wrap(errB, "Failed to parse to age")
			FlowWatch.GetLogHelper().Debug(ctx, errB)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: fmt.Sprintf("Invalid to age value: %s", record[6])})
			return
		}

		// Parse goal values
		Bronze, errC := strconv.Atoi(record[7])
		if errC != nil {
			errC = errors.Wrap(errC, "Failed to parse Bronze")
			FlowWatch.GetLogHelper().Debug(ctx, errC)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: fmt.Sprintf("Invalid Bronze value: %s", record[7])})
			return
		}

		Silver, errD := strconv.Atoi(record[8])
		if errD != nil {
			errD = errors.Wrap(errD, "Failed to parse Silver")
			FlowWatch.GetLogHelper().Debug(ctx, errD)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: fmt.Sprintf("Invalid Silver value: %s", record[8])})
			return
		}

		Gold, errE := strconv.Atoi(record[9])
		if errE != nil {
			errE = errors.Wrap(errE, "Failed to parse Gold")
			FlowWatch.GetLogHelper().Debug(ctx, errE)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: fmt.Sprintf("Invalid Gold value: %s", record[9])})
			return
		}

		// Parse the ruleset record
		rulesetBody := RulesetBody{
			RulesetYear:    record[0],
			DisciplineName: record[1],
			ExerciseName:   record[2],
			Unit:           record[3],
			Sex:            record[4],
			FromAge:        uint(FromAge),
			ToAge:          uint(ToAge),
			Bronze:         uint64(Bronze),
			Silver:         uint64(Silver),
			Gold:           uint64(Gold),
			Description:    record[10],
		}

		rulesets = append(rulesets, rulesetBody)
	}

	// Write ruleset data to the database
	db := DatabaseFlow.GetDB(ctx)
	for idx, ruleset := range rulesets {
		// Ensure the ruleset year exists
		var rulesetYearCount int64
		errA := db.Model(&databaseUtils.Ruleset{}).
			Where("year = ?", ruleset.RulesetYear).
			Count(&rulesetYearCount).
			Error
		if errA != nil {
			errA := errors.Wrap(errA, "Failed to check the ruleset year")
			FlowWatch.GetLogHelper().Debug(ctx, errA)
			c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Ruleset year is invalid: %s", ruleset))
			return
		}

		// Create new ruleset year if needed
		if rulesetYearCount == 0 {
			errB := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
				err := tx.Model(&databaseUtils.Ruleset{}).
					Create(&databaseUtils.Ruleset{Year: ruleset.RulesetYear}).
					Error
				return err
			})
			if errB != nil {
				errB = errors.Wrap(errB, "Failed to create the ruleset year")
				FlowWatch.GetLogHelper().Error(ctx, errB)
				c.AbortWithStatusJSON(http.StatusInternalServerError, fmt.Sprintf("Failed to create the ruleset year: %s", ruleset.RulesetYear))
				return
			}
		}

		// Ensure the discipline exists
		disciplineName := CapitalizeFirst(ruleset.DisciplineName)
		var disciplineCount int64
		errC := db.Model(&databaseUtils.Discipline{}).
			Where("name = ?", disciplineName).
			Count(&disciplineCount).
			Error
		if errC != nil {
			errC = errors.Wrap(errC, "Failed to check the discipline")
			FlowWatch.GetLogHelper().Debug(ctx, errC)
			c.AbortWithStatusJSON(http.StatusNotFound, fmt.Sprintf("Discipline does not exist: %s", disciplineName))
			return
		}

		// Ensure the exercise exists
		var exerciseCount int64
		errD := db.Model(&databaseUtils.Exercise{}).
			Where("name = ? AND discipline_name = ?", ruleset.ExerciseName, disciplineName).
			Count(&exerciseCount).
			Error
		if errD != nil {
			errD = errors.Wrap(errD, "Failed to check the exercise")
			FlowWatch.GetLogHelper().Debug(ctx, errD)
			c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Exercise is invalid: %s", ruleset.ExerciseName))
			return
		}

		// Create the exercise if needed
		if exerciseCount == 0 {
			// Validate the unit field
			unit := strings.ToLower(ruleset.Unit)
			if !contains(possibleUnits, unit) {
				err := errors.New(fmt.Sprintf("Invalid unit in dataset %d", idx))
				FlowWatch.GetLogHelper().Debug(ctx, err)
				c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
				return
			}

			// Write exerciseCount to the database
			errE := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
				err := tx.Model(&databaseUtils.Exercise{}).
					Create(&databaseUtils.Exercise{
						Name:           ruleset.ExerciseName,
						Unit:           unit,
						DisciplineName: disciplineName,
					}).
					Error
				return err
			})
			if errE != nil {
				errE = errors.Wrap(errE, "Failed to create the exercise")
				FlowWatch.GetLogHelper().Error(ctx, errE)
				c.AbortWithStatusJSON(http.StatusInternalServerError, fmt.Sprintf("Failed to create exercise: %s", ruleset.ExerciseName))
			}
		}

		// Get the exercise id
		var exercise databaseUtils.Exercise
		errF := db.Model(&databaseUtils.Exercise{}).
			Where("name = ? AND discipline_name = ?", ruleset.ExerciseName, disciplineName).
			First(&exercise).
			Error
		if errF != nil {
			errF = errors.Wrap(errF, "Failed to check the exercise")
			FlowWatch.GetLogHelper().Error(ctx, errF)
			c.AbortWithStatusJSON(http.StatusInternalServerError, fmt.Sprintf("Failed to get exercise: %s", ruleset.ExerciseName))
			return
		}

		// Ensure the exercise ruleset exists
		var exerciseRulesetCount int64
		errG := db.Model(&databaseUtils.ExerciseRuleset{}).
			Where("ruleset_year = ? AND exercise_id = ?", ruleset.RulesetYear, exercise.ID).
			Count(&exerciseRulesetCount).
			Error
		if errG != nil {
			errG = errors.Wrap(errG, "Failed to check the ruleset")
			FlowWatch.GetLogHelper().Debug(ctx, errG)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Failed to check the exercise ruleset"})
			return
		}

		// Create the exercise ruleset if needed
		if exerciseRulesetCount == 0 {
			errH := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
				err := tx.Model(&databaseUtils.ExerciseRuleset{}).
					Create(&databaseUtils.ExerciseRuleset{
						RulesetYear: ruleset.RulesetYear,
						ExerciseId:  exercise.ID,
					}).
					Error
				return err
			})
			if errH != nil {
				msg := fmt.Sprintf("Failed to create the exercise ruleset: %s - %s", ruleset.ExerciseName, ruleset.RulesetYear)
				errH = errors.Wrap(errH, msg)
				FlowWatch.GetLogHelper().Debug(ctx, errH)
				c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: msg})
				return
			}
		}

		// Get the ruleset id
		var exerciseRuleset databaseUtils.ExerciseRuleset
		errI := db.Model(&databaseUtils.ExerciseRuleset{}).
			Where("ruleset_year = ? AND exercise_id = ?", ruleset.RulesetYear, exercise.ID).
			First(&exerciseRuleset).
			Error
		if errI != nil {
			errI = errors.Wrap(errI, "Failed to get the ruleset")
			FlowWatch.GetLogHelper().Error(ctx, errI)
			c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Failed to get the exercise ruleset: %s - %s", ruleset.ExerciseName, ruleset.RulesetYear))
			return
		}

		// Check if the exercise goal already exists
		var exerciseGoalCount int64
		errJ := db.Model(&databaseUtils.ExerciseGoal{}).
			Where("ruleset_id = ? AND from_age = ? AND to_age = ? AND sex = ?",
				exerciseRuleset.ID, ruleset.FromAge, ruleset.ToAge, ruleset.Sex).
			Count(&exerciseGoalCount).
			Error
		if errJ != nil {
			errJ = errors.Wrap(errJ, "Failed to check the exercise goal")
			FlowWatch.GetLogHelper().Debug(ctx, errJ)
			c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Exercise goal is invalid: %s", ruleset.ExerciseName))
			return
		}

		// Create exercise goal if needed
		if exerciseGoalCount == 0 {
			errK := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
				err := tx.Model(&databaseUtils.ExerciseGoal{}).
					Create(&databaseUtils.ExerciseGoal{
						RulesetId:   exerciseRuleset.ID,
						FromAge:     ruleset.FromAge,
						ToAge:       ruleset.ToAge,
						Sex:         ruleset.Sex,
						Bronze:      ruleset.Bronze,
						Silver:      ruleset.Silver,
						Gold:        ruleset.Gold,
						Description: ruleset.Description,
					}).
					Error
				return err
			})
			if errK != nil {
				msg := fmt.Sprintf("Failed to create the exercise goal: %s - %s - %d - %d - %s",
					ruleset.ExerciseName, ruleset.RulesetYear, ruleset.FromAge, ruleset.ToAge, ruleset.Sex)
				errK = errors.Wrap(errK, msg)
				FlowWatch.GetLogHelper().Error(ctx, errK)
				c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: msg})
				return
			}
		} else if exerciseGoalCount > 0 { // Update existing exercise goal
			errL := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
				err := tx.Model(&databaseUtils.ExerciseGoal{}).
					Where("ruleset_id = ? AND from_age = ? AND to_age = ? AND sex = ?",
						exerciseRuleset.ID, ruleset.FromAge, ruleset.ToAge, ruleset.Sex).
					Updates(map[string]interface{}{
						"bronze":      ruleset.Bronze,
						"silver":      ruleset.Silver,
						"gold":        ruleset.Gold,
						"description": ruleset.Description,
					}).Error
				return err
			})
			if errL != nil {
				msg := fmt.Sprintf("Failed to update the exercise goal: %s - %s - %d - %d - %s",
					ruleset.ExerciseName, ruleset.RulesetYear, ruleset.FromAge, ruleset.ToAge, ruleset.Sex)
				errL = errors.Wrap(errL, msg)
				FlowWatch.GetLogHelper().Error(ctx, errL)
				c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: msg})
				return
			}
		}
	}

	// Return success message
	c.JSON(
		http.StatusOK,
		endpoints.SuccessResponse{
			Message: "Creations successful",
		},
	)
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(strings.ToLower(s))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
