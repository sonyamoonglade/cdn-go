package validate

import (
	"fmt"
	"testing"

	"animakuro/cdn/internal/entities"

	"github.com/stretchr/testify/require"
)

func TestBucketOperation(t *testing.T) {

	type OpTest struct {
		ops []*entities.Operation
		err error
	}

	tt := []OpTest{
		{ops: []*entities.Operation{
			{Type: "public", Name: "get"},
			{Type: "random", Name: "post"},
		}, err: fmt.Errorf("validation error: invalid type random")},

		{ops: []*entities.Operation{
			{Type: "public", Name: "mock"},
			{Type: "private", Name: "get"},
		}, err: fmt.Errorf("validation error: invalid operation mock")},

		{ops: []*entities.Operation{
			{Type: "public", Name: "get"},
			{Type: "public", Name: "post"},
		}, err: nil},
		{ops: []*entities.Operation{
			{Type: "public", Name: "delete"},
			{Type: "public", Name: "post"},
		}, err: nil},
	}

	for _, test := range tt {
		err := BucketOperation(test.ops)
		require.Equal(t, test.err, err)
	}

}
