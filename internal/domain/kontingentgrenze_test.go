package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineKontingentgrenzeNenntIhreWartezeit(t *testing.T) {
	t.Parallel()

	for warten, erwartet := range map[time.Duration]string{
		4500 * time.Millisecond: "zu viele Rufe — in 5 s wieder",
		0:                       "zu viele Rufe — in 1 s wieder",
	} {
		fehler := domain.NeueKontingentgrenze(warten)

		var grenze *domain.Kontingentgrenze
		if !errors.Is(fehler, domain.ErrKontingent) || !errors.As(fehler, &grenze) || fehler.Error() != erwartet {
			t.Errorf("%v: %q, erwartet %q", warten, fehler, erwartet)
		}
	}
}
