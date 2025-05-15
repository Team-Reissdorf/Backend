package setup

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/rulesetManagement"
	"gorm.io/gorm"
)

func CreateStandardRulesets(ctx context.Context) {
	// read files in
	path := os.Getenv("RULESET_DIR")
	rulesCSVpath, err := filepath.Glob(path + "*.csv")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range rulesCSVpath {
		file, err := os.Open(path + f)
		if err != nil {
			log.Fatal(err)
		}

		rulesets, errRS := read_csv_to_struct(file)
		if errRS != nil {
			log.Fatal(err)
		}

		for _, set := range rulesets {
			err := write_db(set)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

}

func read_csv_to_struct(file *os.File) ([]rulesetManagement.RulesetBody, error) {
	reader := csv.NewReader(file)
	reader.Comma = ';'

	entries, err := reader.ReadAll()
	if err != nil {
		// holy cancer
		return []rulesetManagement.RulesetBody{}, err
	}

	var rulesets []rulesetManagement.RulesetBody

	for _, entry := range entries {
		if len(entry) != rulesetManagement.CSVCOLUMNCOUNT {
			return []rulesetManagement.RulesetBody{}, errors.New("inconsistent number of columns in the CSV file (setup)")
		}

		// Parse age values
		FromAge, err := strconv.Atoi(entry[5])

		ToAge, err := strconv.Atoi(entry[6])

		// Parse goal values
		Bronze, err := strconv.Atoi(entry[7])

		Silver, err := strconv.Atoi(entry[8])

		Gold, err := strconv.Atoi(entry[9])

		// just check here, idc where it failed
		if err != nil {
			log.Fatal(err)
		}

		rulesets = append(rulesets, rulesetManagement.RulesetBody{
			RulesetYear:    entry[0],
			DisciplineName: entry[1],
			ExerciseName:   entry[2],
			Unit:           entry[3],
			Sex:            entry[4],
			FromAge:        uint(FromAge),
			ToAge:          uint(ToAge),
			Bronze:         uint64(Bronze),
			Silver:         uint64(Silver),
			Gold:           uint64(Gold),
			Description:    entry[10],
		})

	}

	return rulesets, nil
}

func write_db(ruleset rulesetManagement.RulesetBody, ctx context.Context) error {
	db := DatabaseFlow.GetDB(ctx)

	// Ensure the ruleset year exists
	var rulesetYearCount int64
	errA := db.Model(&databaseUtils.Ruleset{}).
		Where("year = ?", ruleset.RulesetYear).
		Count(&rulesetYearCount).
		Error
	if errA != nil {
		errA := errors.Wrap(errA, "Failed to check the ruleset year")
		return errA
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
	disciplineName := rulesetManagement.CapitalizeFirst(ruleset.DisciplineName)
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
