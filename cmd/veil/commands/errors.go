package commands

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownCommand = errors.New("unknown command")
)

// UsageError indicates incorrect command usage.
type UsageError struct {
	Command string
	Usage   string
}

func (e *UsageError) Error() string {
	return fmt.Sprintf("Usage: %s", e.Usage)
}
