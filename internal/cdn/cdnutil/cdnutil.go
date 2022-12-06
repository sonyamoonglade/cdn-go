package cdnutil

import (
	"fmt"

	"github.com/pkg/errors"
)

// WrapInternal returns new error based on err and context.
// Adds 'internal error {context}: ' prefix.
func WrapInternal(err error, context string) error {
	return fmt.Errorf("internal error at [%s]: %w", context, err)
}

// ChainInternal returns new error based on err and context.
// Adds '{context}: ' prefix.
func ChainInternal(err error, context string) error {
	return fmt.Errorf("[%s]: %w", context, err)
}

func IsErrorOf(targetError error) func(error) bool {
	return func(actual error) bool {
		return errors.Is(targetError, actual)
	}
}

func IsAvailable(hosts []string, self string) (string, bool) {
	flag := false
	available := ""
	for _, host := range hosts {
		if host == self {
			flag = true
		} else {
			available = host
		}
	}
	return available, flag
}
