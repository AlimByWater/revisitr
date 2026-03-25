package segments

import "errors"

var (
	ErrSegmentNotFound    = errors.New("segment not found")
	ErrNotSegmentOwner    = errors.New("not authorized")
	ErrNotCustomSegment   = errors.New("segment is not of type custom")
	ErrRuleNotFound       = errors.New("segment rule not found")
	ErrPredictionNotFound = errors.New("prediction not found")
)
