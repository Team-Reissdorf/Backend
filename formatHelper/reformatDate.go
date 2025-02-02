package formatHelper

import (
	"github.com/pkg/errors"
	"time"
)

// FormatDate converts a date string from RFC3339 (e.g., "2003-01-17T00:00:00Z") to the format "YYYY-MM-DD".
func FormatDate(date string) (string, error) {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse date: "+date)
		return "", err
	}
	return t.Format("2006-01-02"), nil
}
