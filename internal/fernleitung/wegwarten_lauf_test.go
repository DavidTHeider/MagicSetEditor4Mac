//go:build netzlauf || vmlauf

package fernleitung

import (
	"testing"
	"time"
)

// wartetAufDenWeg fragt nach, bis ein Weg dasteht, wie er erwartet ist — oder
// bis zur Frist. **Der direkte Weg entsteht nicht sofort**: erst Verkehr durch
// den Tunnel stößt die Suche an, und bis dahin geht es über das Relais.
func wartetAufDenWeg(t *testing.T, horcher *Horcher, nurRelais bool) []Fernweg {
	t.Helper()

	var wege []Fernweg

	for frist := time.Now().Add(15 * time.Second); time.Now().Before(frist); time.Sleep(250 * time.Millisecond) {
		wege = horcher.Wege()
		if len(wege) == 1 && (nurRelais && wege[0].Relais != "" || !nurRelais && wege[0].Direkt != "") {
			return wege
		}
	}

	return wege
}
