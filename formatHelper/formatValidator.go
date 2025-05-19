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

	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
)

const (
	localEmailCheckRegexString = ".*\\.[a-zA-Z]{2,}(\\.)?$"
	dateFormatCheckRegexString = "^\\d{4}\\-[0-1][0-9]\\-[0-3][0-9]$"
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
	// Parse the date in local timezone
	parsedDate, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return errors.Wrap(err, "Failed to parse date")
	}

	// Get current time and normalize both dates to start of day
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	dateStart := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.Local)

	// Simple comparison - is the normalized date after today?
	if dateStart.After(todayStart) {
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

// IsDuration checks whether s is a valid duration in the format S, SS, M:SS or MM:SS.
// Seconds must be 0-59; any non-negative number can be used for minutes,
// but only 1 or 2 digits in the string.
func IsDuration(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("empty duration")
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		// "S" or "SS"
		seg := parts[0]
		if len(seg) < 1 || len(seg) > 2 {
			return fmt.Errorf("seconds must be 1 or 2 digits")
		}
		sec, err := strconv.Atoi(seg)
		if err != nil {
			return fmt.Errorf("seconds must be numeric: %q", seg)
		}
		if sec < 0 || sec > 59 {
			return fmt.Errorf("seconds must be between 0 and 59: %d", sec)
		}
		return nil

	case 2:
		// "M:SS" or "MM:SS"
		minSeg, secSeg := parts[0], parts[1]

		// Minutes: 1–2 digits, >=0
		if len(minSeg) < 1 || len(minSeg) > 2 {
			return fmt.Errorf("minutes must be 1 or 2 digits")
		}
		min, errM := strconv.Atoi(minSeg)
		if errM != nil {
			return fmt.Errorf("minutes must be numeric: %q", minSeg)
		}
		if min < 0 {
			return fmt.Errorf("minutes cannot be negative: %d", min)
		}

		// Seconds: **exact** 2 digits, between 0–59
		if len(secSeg) != 2 {
			return fmt.Errorf("seconds must be exactly 2 digits: %q", secSeg)
		}
		sec, errS := strconv.Atoi(secSeg)
		if errS != nil {
			return fmt.Errorf("seconds must be numeric: %q", secSeg)
		}
		if sec < 0 || sec > 59 {
			return fmt.Errorf("seconds must be between 0 and 59: %d", sec)
		}
		return nil

	default:
		return fmt.Errorf("duration must be S, SS, M:SS or MM:SS")
	}
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
	input = strings.TrimSpace(input)

	// Match format: "<value> m" (z. B. "2.40 m" oder "800 m")
	//re := regexp.MustCompile(`^([\d.,]+)\s*m$`)
	//matches := re.FindStringSubmatch(input)
	/*if len(matches) != 2 {
		return -1, errors.New("invalid format: expected number followed by 'm'")
	}*/

	// Replace , with .
	numericPart := strings.ReplaceAll(input, ",", ".")

	// Parse float meters -> int centimeters
	meters, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		return -1, errors.New("invalid number format for meters")
	}
	cm := int(meters * 100)
	return cm, nil
}

// NormalizeResult standardizes the result into ms or cm depending on the unit.
// For "second" and "minute", accepts S, SS, M:SS or MM:SS and converts to milliseconds.
func NormalizeResult(raw string, unit string) (int, error) {
	unit = strings.ToLower(strings.TrimSpace(unit))
	FlowWatch.GetLogHelper().Debug(context.Background(), "Normal input", raw, unit)

	switch unit {
	case "second", "minute":
		ms, err := parseToMilliseconds(raw)
		if err != nil {
			return 0, errors.New("failed to normalize time")
		}
		return ms, nil

	case "meter":
		cm, err := FormatToCentimeters(raw)
		if err != nil {
			return 0, errors.New("failed to normalize distance")
		}
		return cm, nil

	case "centimeter":
		val, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return 0, errors.New("invalid centimeter value")
		}
		return val, nil

	case "bool":
		normalized := strings.ToLower(strings.TrimSpace(raw))
		if normalized == "ja" || normalized == "true" || normalized == "yes" {
			return 1, nil
		} else if normalized == "nein" || normalized == "false" || normalized == "no" {
			return 0, nil
		}
		return 0, errors.New("invalid boolean value")

	case "points":
		p, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return 0, errors.New("invalid points value")
		}
		return p, nil

	default:
		// Fallback: try int parse
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return 0, errors.New("unknown unit and failed to parse as int")
		}
		return v, nil
	}
}

// parseToMilliseconds converts "S", "SS", "M:SS" or "MM:SS" into milliseconds.
func parseToMilliseconds(raw string) (int, error) {
	seg := strings.Split(strings.TrimSpace(raw), ":")
	switch len(seg) {
	case 1:
		// "S" or "SS"
		sec, err := strconv.Atoi(seg[0])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %q", seg[0])
		}
		if sec < 0 || sec > 59 {
			return 0, fmt.Errorf("seconds must be 0-59: %d", sec)
		}
		return sec * 1000, nil

	case 2:
		// "M:SS" or "MM:SS"
		min, errMin := strconv.Atoi(seg[0])
		sec, errSec := strconv.Atoi(seg[1])
		if errMin != nil {
			return 0, fmt.Errorf("invalid minutes: %q", seg[0])
		}
		if errSec != nil {
			return 0, fmt.Errorf("invalid seconds: %q", seg[1])
		}
		if min < 0 {
			return 0, fmt.Errorf("minutes cannot be negative: %d", min)
		}
		if sec < 0 || sec > 59 {
			return 0, fmt.Errorf("seconds must be 0-59: %d", sec)
		}
		return (min*60 + sec) * 1000, nil

	default:
		return 0, fmt.Errorf("invalid time format, use S, SS, M:SS or MM:SS: %q", raw)
	}
}
