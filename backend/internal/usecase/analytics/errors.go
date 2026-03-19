package analytics

import "errors"

var ErrInvalidDateRange = errors.New("invalid date range: 'from' must be before 'to'")
