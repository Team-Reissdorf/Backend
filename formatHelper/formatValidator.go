package formatHelper

import (
	"github.com/pkg/errors"
	"gorm.io/gorm/utils"
)

var (
	possibleSexValues = []string{"m", "f", "d"}

	InvalidSexLengthError = errors.New("Sex should be one character only")
	InvalidSexValue       = errors.New("Sex can only be <m|f|d>")
)

func IsEmail(email string) error {
	// ToDo: Implement
	return nil
}

func IsDate(date string) error {
	// ToDo: Implement
	return nil
}

// IsSex checks if the length is one character and one of the allowed values
func IsSex(sex string) error {
	if len(sex) != 1 {
		return InvalidSexLengthError
	}

	if !utils.Contains(possibleSexValues, sex) {
		return InvalidSexValue
	}

	return nil
}
