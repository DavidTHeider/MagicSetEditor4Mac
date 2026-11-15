package app

import (
	"slices"
	"testing"
)

func TestOhnePortLaesstEinenNamenOhnePortStehen(t *testing.T) {
	t.Parallel()

	if ohne := ohnePort([]string{"tisch.local", "[fe80::1]:7778", "192.168.1.5:7778"}); !slices.Equal(ohne,
		[]string{"tisch.local", "fe80::1", "192.168.1.5"}) {
		t.Errorf("gekommen ist %v", ohne)
	}
}
