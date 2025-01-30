package formatHelper

import (
	"context"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
	"net/mail"
	"regexp"
	"slices"
)

const (
	localEmailCheckRegexString = ".*\\.[a-zA-Z]{2,}(\\.)?$"
	dateFormatCheckRegexString = "^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$"
)

var (
	possibleSexValues    = []string{"m", "f", "d"}
	localEmailCheckRegex *regexp.Regexp

	InvalidSexLengthError          = errors.New("Sex should be one character only")
	InvalidSexValue                = errors.New("Sex can only be <m|f|d>")
	InvalidEmailAddressFormatError = errors.New("Email address format is invalid")
	EmailAddressContainsNameError  = errors.New("Email address should not contain the name")
	EmailAddressInvalidTldError    = errors.New("Email address TLD is invalid")
	DateFormatInvalidError         = errors.New("Date format is invalid")
)

func init() {
	ctx := context.Background()

	var err1 error
	localEmailCheckRegex, err1 = regexp.Compile(localEmailCheckRegexString)
	if err1 != nil {
		endpoints.Logger.Fatal(ctx, "Unable to compile local email address regex", err1)
	}
}

// IsEmail validates the email address using the mail.ParseAddress function
// and additionally checks if it has a valid TLD to block local addresses
func IsEmail(email string) error {
	// Validate the email address
	address, err := mail.ParseAddress(email)
	if err != nil {
		err = errors.Wrap(InvalidEmailAddressFormatError, err.Error())
		return err
	}

	// Check if the email address contains the name
	if address.Address != email {
		return EmailAddressContainsNameError
	}

	// Check if the email address is local
	if !localEmailCheckRegex.MatchString(email) {
		return EmailAddressInvalidTldError
	}

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

	if !slices.Contains(possibleSexValues, sex) {
		return InvalidSexValue
	}

	return nil
}
