package app

import (
	"math"
	"testing"
	"time"
)

// **Eine Stundenzahl von außen läuft nicht um** — Befund 1 des Nachlaufs
// über J05 (#2269): `MassePreise` reichte `hoechstensStunden` ungeprüft in
// die Multiplikation, und 5.124.096 Stunden wurden zu rund 25 Minuten — die
// Frischeprüfung holte danach Preise neu, die nach der verlangten Grenze
// frisch waren.
func TestDasHoechstalterEinesPreisesLaeuftNichtUm(t *testing.T) {
	for stunden, erwartet := range map[int]time.Duration{
		0:       0,
		24:      24 * time.Hour,
		5124096: math.MaxInt64,
		-3:      0,
	} {
		if gab := preisalter(stunden); gab != erwartet {
			t.Errorf("preisalter(%d) = %v, erwartet %v", stunden, gab, erwartet)
		}
	}
}
