package cdnutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapInternal(t *testing.T) {

	err := errors.New("im error")

	err = WrapInternal(err, "context.important")

	expected := "internal error at [context.important]: im error"
	require.Equal(t, expected, err.Error())
}

func TestChainInternal(t *testing.T) {

	errfunc := func() error {
		return errors.New("database connection is lost")
	}

	f1 := func() error {
		err := errfunc()
		return WrapInternal(err, "f1.errfunc")
	}

	f2 := func() error {
		// f2 is service calling another function that returns already wrapped error.
		// So chain it. Not WrapInternal again.
		err := f1()
		return ChainInternal(err, "f2->f1")
	}

	// Test related
	err := f2()

	expected := "[f2->f1]: internal error at [f1.errfunc]: database connection is lost"
	require.Equal(t, expected, err.Error())

}
