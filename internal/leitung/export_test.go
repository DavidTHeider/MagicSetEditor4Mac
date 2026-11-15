package leitung

import "testing"

// LesenTauschen legt einen Griff um das Lesen der PKI-Dateien — **nur für
// Tests**, und nur für Läufe ohne `t.Parallel`: die Variable gilt für das
// ganze Paket.
func LesenTauschen(t testing.TB, neu func(pfad string, lies func(string) ([]byte, error)) ([]byte, error)) {
	t.Helper()

	alt := dateiLesen
	dateiLesen = func(pfad string) ([]byte, error) { return neu(pfad, alt) }

	t.Cleanup(func() { dateiLesen = alt })
}
