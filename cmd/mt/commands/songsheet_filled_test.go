package commands

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLocos(t *testing.T) {

	in := "    Foo Bar  Baz   "
	outExpected := []stringInt{
		stringInt{"Foo", 4},
		stringInt{"Bar", 8},
		stringInt{"Baz", 13},
	}
	outTest := getLocos(in)

	require.False(t, reflect.DeepEqual(outExpected, outTest),
		"expected: %+v\n got:  %+v\n", outExpected, outTest)
}
