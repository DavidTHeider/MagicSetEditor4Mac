//go:build !abnahme || !darwin

package fenster

// probepanik tut außerhalb des Abnahmebaus nichts — die Naht gibt es dort
// nicht (`probepanik_abnahme_darwin.go`).
func probepanik(string) {}

// probeabsturz tut außerhalb des Abnahmebaus nichts.
func probeabsturz() {}

func probeSyncpanik() {}
