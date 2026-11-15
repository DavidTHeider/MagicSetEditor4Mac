package domain

// **Die vier Pfade eines eigenen Dienstzeugnisses** — NET-03.AK-01
// (#1707),
// entschieden in
// [0154](../../docs/entscheidungen/0154-das-dienstzeugnis-ist-nicht-die-identitaet.md).
//
// **Alle leer heißt: das Zeugnis auf der Peeridentität, wie bisher.** Die
// Einstellungen sind die Konfigurationsdatei der Dienste — sie liegen im
// Bestand dieses Rechners, das Fenster und `mse einstellung` schreiben sie,
// `mse-rest` und `mse-mcp` lesen sie beim Start.

import (
	"fmt"
	"strings"
)

const (
	// DienstTLSZertifikat ist das Blatt des Dienstes, als PEM.
	DienstTLSZertifikat = "dienst.tls.zertifikat"

	// DienstTLSSchluessel ist sein privater Schlüssel — nur für den
	// Eigentümer lesbar.
	DienstTLSSchluessel = "dienst.tls.schluessel"

	// DienstTLSKette sind Zwischenzertifikate — leer heißt: keine, oder sie
	// stehen im Zertifikat.
	DienstTLSKette = "dienst.tls.kette"

	// DienstTLSAnker sind die Vertrauensanker der Kette — leer heißt: die des
	// Systems.
	DienstTLSAnker = "dienst.tls.anker"
)

// Zeugnispfade sind die vier Schlüssel, in der Reihenfolge der Einstellungen.
func Zeugnispfade() []string {
	return []string{DienstTLSZertifikat, DienstTLSSchluessel, DienstTLSKette, DienstTLSAnker}
}

// PruefeZeugnispfad weist ab, was kein absoluter Pfad ist.
//
// **Eine Tilde wird nicht aufgelöst.** Der Dienst liest den Pfad, und sein
// Zuhause muss nicht das sein, das der Mensch beim Tippen meinte.
func PruefeZeugnispfad(wert string) error {
	if wert == "" {
		return nil
	}

	if !strings.HasPrefix(wert, "/") {
		return fmt.Errorf("%w: %q ist kein absoluter Pfad — gib ihn ab / an",
			ErrValidation, wert)
	}

	if strings.ContainsRune(wert, 0) {
		return fmt.Errorf("%w: der Pfad enthält ein Nullzeichen", ErrValidation)
	}

	return nil
}
