package fenster

import (
	"math"
	"testing"
	"time"
)

// **Eine Verzögerung von außen läuft nicht um** (Befund 1 des Nachlaufs über
// J05, #2269): die Abnahme setzt sie über die Leitung, und ein Überlauf hätte
// die beabsichtigte Verzögerung aufgehoben.
func TestEineStoerverzoegerungLaeuftNichtUm(t *testing.T) {
	for ms, erwartet := range map[int]time.Duration{
		250:           250 * time.Millisecond,
		9223372036855: math.MaxInt64,
		0:             0,
	} {
		if gab := stoerdauer(ms); gab != erwartet {
			t.Errorf("stoerdauer(%d) = %v, erwartet %v", ms, gab, erwartet)
		}
	}
}
