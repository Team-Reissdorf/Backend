package setup

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints/rulesetManagement"
	"gorm.io/gorm"
)

func CreateStandardRulesets(ctx context.Context) {

	// read files in
	path := os.Getenv("RULESET_DIR")
	if path == "" {
		FlowWatch.GetLogHelper().Fatal(ctx, "ruleset dir unset")
	}
	FlowWatch.GetLogHelper().Info(ctx, "got ruleset dir "+path)

	rulesCSVpath, err := filepath.Glob(path + "*.csv")
	if err != nil {
		FlowWatch.GetLogHelper().Fatal(ctx, err)
	}

	for _, f := range rulesCSVpath {
		file, err := os.Open(f)
		if err != nil {
			FlowWatch.GetLogHelper().Fatal(ctx, err)
		}

		FlowWatch.GetLogHelper().Info(ctx, "writing ruleset "+file.Name()+" to db")

		rulesets, errRS := read_csv_to_struct(ctx, file)
		if errRS != nil {
			FlowWatch.GetLogHelper().Fatal(ctx, err)
		}

		for _, set := range rulesets {
			err := write_db(set, ctx)
			if err != nil {
				FlowWatch.GetLogHelper().Fatal(ctx, err)
			}
		}

	}

	FlowWatch.GetLogHelper().Info(ctx, "done creating default rulesets.")

}

func read_csv_to_struct(ctx context.Context, file *os.File) ([]rulesetManagement.RulesetBody, error) {
	reader := csv.NewReader(file)
	reader.Comma = ';'

	entries, errread := reader.ReadAll()
	if errread != nil {
		FlowWatch.GetLogHelper().Fatal(ctx, errread)
		// holy cancer
		return []rulesetManagement.RulesetBody{}, errread
	}

	var rulesets []rulesetManagement.RulesetBody

	for _, entry := range entries {
		if len(entry) != rulesetManagement.CSVCOLUMNCOUNT {

			return []rulesetManagement.RulesetBody{}, errors.New("inconsistent number of columns in the CSV file (setup)")
		}

		// Parse age values
		FromAge, err := strconv.Atoi(entry[5])
		if err != nil {
			FlowWatch.GetLogHelper().Fatal(ctx, err)
		}

		ToAge, err := strconv.Atoi(entry[6])

		if err != nil {
			FlowWatch.GetLogHelper().Fatal(ctx, err)
		}

		// Parse goal values
		Bronze, err := strconv.Atoi(entry[7])
		if err != nil {
			FlowWatch.GetLogHelper().Fatal(ctx, err)
		}

		Silver, err := strconv.Atoi(entry[8])
		if err != nil {
			FlowWatch.GetLogHelper().Fatal(ctx, err)
		}

		Gold, err := strconv.Atoi(entry[9])
		if err != nil {
			FlowWatch.GetLogHelper().Fatal(ctx, err)
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
			return errB
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
		return errC
	}

	// Ensure the exercise exists
	var exerciseCount int64
	errD := db.Model(&databaseUtils.Exercise{}).
		Where("name = ? AND discipline_name = ?", ruleset.ExerciseName, disciplineName).
		Count(&exerciseCount).
		Error
	if errD != nil {
		return errD
	}

	// Create the exercise if needed
	if exerciseCount == 0 {
		// Validate the unit field
		unit := strings.ToLower(ruleset.Unit)
		if !rulesetManagement.Contains(rulesetManagement.POSSIBLEUNITS, unit) {
			err := fmt.Errorf("invalid unit in dataset")
			return err
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
			return errE
		}
	}

	// Get the exercise id
	var exercise databaseUtils.Exercise
	errF := db.Model(&databaseUtils.Exercise{}).
		Where("name = ? AND discipline_name = ?", ruleset.ExerciseName, disciplineName).
		First(&exercise).
		Error
	if errF != nil {
		return errF
	}

	// Ensure the exercise ruleset exists
	var exerciseRulesetCount int64
	errG := db.Model(&databaseUtils.ExerciseRuleset{}).
		Where("ruleset_year = ? AND exercise_id = ?", ruleset.RulesetYear, exercise.ID).
		Count(&exerciseRulesetCount).
		Error
	if errG != nil {
		return errG
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
			FlowWatch.GetLogHelper().Info(ctx, msg)
			return errH
		}
	}

	// Get the ruleset id
	var exerciseRuleset databaseUtils.ExerciseRuleset
	errI := db.Model(&databaseUtils.ExerciseRuleset{}).
		Where("ruleset_year = ? AND exercise_id = ?", ruleset.RulesetYear, exercise.ID).
		First(&exerciseRuleset).
		Error
	if errI != nil {
		return errI
	}

	// Check if the exercise goal already exists
	var exerciseGoalCount int64
	errJ := db.Model(&databaseUtils.ExerciseGoal{}).
		Where("ruleset_id = ? AND from_age = ? AND to_age = ? AND sex = ?",
			exerciseRuleset.ID, ruleset.FromAge, ruleset.ToAge, ruleset.Sex).
		Count(&exerciseGoalCount).
		Error
	if errJ != nil {
		return errJ
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

			FlowWatch.GetLogHelper().Info(ctx, msg)

			return errK
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

			FlowWatch.GetLogHelper().Info(ctx, msg)
			return errL
		}
	}

	// yup it all worked :)
	return nil
}
