package formatHelper

import (
	"context"

	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
)

const (
	localEmailCheckRegexString = ".*\\.[a-zA-Z]{2,}(\\.)?$"
	dateFormatCheckRegexString = "^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$"
)

var (
	possibleSexValues    = []string{"m", "f", "d"}
	localEmailCheckRegex *regexp.Regexp
	dateFormatCheckRegex *regexp.Regexp

	InvalidSexLengthError          = errors.New("Sex should be one character only")
	InvalidSexValue                = errors.New("Sex can only be <m|f|d>")
	InvalidEmailAddressFormatError = errors.New("Email address format is invalid")
	EmailAddressContainsNameError  = errors.New("Email address should not contain the name")
	EmailAddressInvalidTldError    = errors.New("Email address TLD is invalid")
	DateFormatInvalidError         = errors.New("Date format is invalid")
	DateInFutureError              = errors.New("Date is in the future")
	EmptyStringError               = errors.New("Empty String")
)

func init() {
	ctx := context.Background()

	var err1 error
	localEmailCheckRegex, err1 = regexp.Compile(localEmailCheckRegexString)
	if err1 != nil {
		endpoints.Logger.Fatal(ctx, "Unable to compile local email address regex", err1)
	}

	var err2 error
	dateFormatCheckRegex, err2 = regexp.Compile(dateFormatCheckRegexString)
	if err2 != nil {
		endpoints.Logger.Fatal(ctx, "Unable to compile date format regex", err2)
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

// IsDate checks if the given date is in the required format (YYYY-MM-DD).
// Throws: DateFormatInvalidError
func IsDate(date string) error {
	// Check if the date format matches
	if !dateFormatCheckRegex.MatchString(date) {
		return DateFormatInvalidError
	}
	return nil
}

func IsBefore(date1 string, date2 string) error {
	parsedDate1, err := time.Parse("2006-01-02", date1)
	if err != nil {
		return errors.Wrap(err, "Fehler beim Parsen von date1")
	}

	parsedDate2, err := time.Parse("2006-01-02", date2)
	if err != nil {
		return errors.Wrap(err, "Fehler beim Parsen von date2")
	}
	if parsedDate1.Before(parsedDate2) {
		return errors.New("date1 is before date2")
	}

	return nil
}

// IsFuture checks if the given date is in the future and throws an error if it is.
func IsFuture(date string) error {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return errors.Wrap(err, "Failed to parse date")
	}
	if !parsedDate.Before(time.Now()) {
		return DateInFutureError
	}
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

func IsEmpty(bodyPart string) error {
	if len(bodyPart) == 0 {
		return EmptyStringError
	}
	return nil
}

// IsDuration prüft, ob s eine gültige Zeitdauer im Format MM:SS oder HH:MM:SS ist.
// Beispiele gültig: "03:25", "3:5", "01:03:25", "1:3:5"
// Minuten und Sekunden müssen dabei 0–59 sein.
func IsDuration(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("empty duration")
	}

	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return fmt.Errorf("duration must be MM:SS or HH:MM:SS")
	}

	// Überprüfe jedes Segment
	for i, seg := range parts {
		if seg == "" {
			return fmt.Errorf("empty segment in duration")
		}
		n, err := strconv.Atoi(seg)
		if err != nil {
			return fmt.Errorf("duration contains non-numeric segment: %q", seg)
		}
		// Nur die letzten beiden Segmente (Minuten, Sekunden) brauchen 0–59-Range
		if i >= len(parts)-2 {
			if n < 0 || n > 59 {
				return fmt.Errorf("minutes/seconds out of range: %d", n)
			}
		} else {
			// Stunden dürfen >= 0 sein (kein Upper-Bound)
			if n < 0 {
				return fmt.Errorf("hours cannot be negative: %d", n)
			}
		}
	}

	return nil
}

// FormatToMilliseconds converts a time string like "03:25" or "1:02:30" into milliseconds.
func FormatToMilliseconds(input string) (int, error) {
	parts := strings.Split(input, ":")
	var totalMs int

	switch len(parts) {
	case 2: // mm:ss
		min, err1 := strconv.Atoi(parts[0])
		sec, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return -1, errors.New("invalid time format (mm:ss)")
		}
		totalMs = (min*60 + sec) * 1000

	case 3: // hh:mm:ss
		hour, err1 := strconv.Atoi(parts[0])
		min, err2 := strconv.Atoi(parts[1])
		sec, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			return -1, errors.New("invalid time format (hh:mm:ss)")
		}
		totalMs = (hour*3600 + min*60 + sec) * 1000

	default:
		return -1, errors.New("unsupported time format")
	}

	return totalMs, nil
}

// FormatToCentimeters converts inputs like "2.40 m" or "800 m" into centimeters as string.
func FormatToCentimeters(input string) (int, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	// Match format: "<value> m" (z. B. "2.40 m" oder "800 m")
	re := regexp.MustCompile(`^([\d.,]+)\s*m$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) != 2 {
		return -1, errors.New("invalid format: expected number followed by 'm'")
	}

	// Replace , with .
	numericPart := strings.ReplaceAll(matches[1], ",", ".")

	// Parse float meters -> int centimeters
	meters, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		return -1, errors.New("invalid number format for meters")
	}
	cm := int(meters * 100)
	return cm, nil
}

// NormalizeResult standardizes the result into ms or cm depending on the unit.
func NormalizeResult(raw string, unit string) (int, error) {
	unit = strings.ToLower(strings.TrimSpace(unit))

	switch unit {
	case "sekunden", "s", "second", "seconds":
		ms, err := FormatToMilliseconds(raw)
		if err != nil {
			return 0, errors.New("failed to normalize time")
		}
		return ms, nil

	case "minute", "minuten", "min":
		ms, err := FormatToMilliseconds(raw)
		if err != nil {
			return 0, errors.New("failed to normalize time")
		}
		return ms, nil

	case "meter", "m":
		cm, err := FormatToCentimeters(raw)
		if err != nil {
			return 0, errors.New("failed to normalize distance")
		}
		return cm, nil

	case "zentimeter", "cm":
		val, err := strconv.Atoi(raw)
		if err != nil {
			return 0, errors.New("invalid centimeter value")
		}
		return val, nil

	case "bool", "boolean":
		normalized := strings.ToLower(strings.TrimSpace(raw))
		if normalized == "ja" || normalized == "true" || normalized == "yes" {
			return 1, nil
		} else if normalized == "nein" || normalized == "false" || normalized == "no" {
			return 0, nil
		}
		return 0, errors.New("invalid boolean value")

	case "punkte", "points", "int":
		val, err := strconv.Atoi(raw)
		if err != nil {
			return 0, errors.New("invalid points value")
		}
		return val, nil

	default:
		val, err := strconv.Atoi(raw)
		if err != nil {
			return 0, errors.New("unknown unit and failed to parse as int")
		}
		return val, nil
	}
}
