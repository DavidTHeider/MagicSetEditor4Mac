package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Anzeigename eines Pakets** — der Kurzname, großgeschrieben; ohne ihn
// der Name der Datei ohne Endung (`Packaged::validate`,
// `src/util/io/package.cpp`, Zeilen 622–626, Fassung f9d9356 vom 2021-07-06).
func TestDerAnzeigenameEinesPakets(t *testing.T) {
	faelle := []struct {
		name   string
		paket  domain.Paket
		erwart string
	}{
		{"mit Kurznamen", domain.Paket{Name: "magic-m15", Kurzname: "m15 normal"}, "M15 Normal"},
		{"ohne Kurznamen", domain.Paket{Name: "magic-m15"}, "Magic-m15"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if got := service.Anzeigename(fall.paket); got != fall.erwart {
				t.Errorf("Anzeigename = %q, erwartet %q", got, fall.erwart)
			}
		})
	}
}
