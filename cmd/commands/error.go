package commands

import (
	"errors"
	"fmt"
)

var errHelp = errors.New("help requested")

type CommandError struct {
	Err error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command error: %v", e.Err)
}

func (*CommandError) Is(e error) bool {
	_, ok := e.(*CommandError)
	return ok
}

func IsCommandError(err error) bool {
	return errors.Is(err, &CommandError{})
}
