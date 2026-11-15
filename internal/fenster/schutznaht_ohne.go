//go:build !server && (!abnahme || !darwin)

package fenster

// cspZusatz gibt außerhalb des Abnahmebaus keine weitere Quelle — der
// Freigabebau trägt die Richtlinie, wie sie in [schutzrichtlinie] steht.
func cspZusatz() string { return "" }

// cspBeobachten setzt außerhalb des Abnahmebaus immer durch.
func cspBeobachten() bool { return false }
