package schluesselbund_test

// **Ein Lauf fasst den Anmeldebund nicht an** — gemessen am 2026-09-14.
//
// # Wofür dieser Lauf da ist
//
// **Weil die Regel allein nicht getragen hat.** §3 verlangt seit jeher, dass
// ein Lauf weder Netz noch echtes Zuhause noch Schlüsselbund anfasst, und
// `MSE_SCHLUESSELBUND` steht dafür bereit. Eine Messung an diesem Tag zählte
// trotzdem **84** Zugriffe auf den Anmeldebund dieses Rechners aus **acht**
// Paketen — darunter zwei Schreibvorgänge von `peeridentitaet`. macOS fragte
// den Menschen dreimal um Erlaubnis, mitten in einem Lauf.
//
// # Was er hält und was nicht
//
// **Er hält, dass jeder der drei Wege abweist.** Was er **nicht** hält: ein
// Kindprozess, den ein Lauf startet — der heißt nicht `.test`. Dort setzt
// `internal/flaechen` die Umgebung, und die Begründung steht dort.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestEinLaufFasstDenAnmeldebundNichtAn(t *testing.T) {
	bund := schluesselbund.Anmeldebund()

	t.Run("lesen", func(t *testing.T) {
		_, fehler := bund.Lies(t.Context(), "probe", "konto")
		verlangeAbweisung(t, fehler)
	})

	t.Run("schreiben", func(t *testing.T) {
		verlangeAbweisung(t, bund.Schreibe(t.Context(), "probe", "konto", "x"))
	})

	t.Run("löschen", func(t *testing.T) {
		verlangeAbweisung(t, bund.Loesche(t.Context(), "probe", "konto"))
	})
}

// **Und ein benannter Bund kommt durch** — sonst wäre die Sperre eine
// Abschaltung und keine Umlenkung.
//
// **Die Gegenprobe zur Gegenprobe**: hier wird ein Ort gesetzt, den es nicht
// gibt, und die Meldung muss ihn nennen und nicht den Anmeldebund. Beide
// Wege scheitern; **woran**, ist der ganze Unterschied.
func TestEinBenannterBundKommtDurch(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "gibtesnicht.keychain")

	t.Setenv(schluesselbund.Bundort, ort)

	_, fehler := schluesselbund.Anmeldebund().Lies(t.Context(), "probe", "konto")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet war ErrNotFound, kam: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), ort) {
		t.Errorf("die Meldung nennt den Ort nicht: %v", fehler)
	}
}

func verlangeAbweisung(t *testing.T, fehler error) {
	t.Helper()

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet war eine Abweisung, kam: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), schluesselbund.Bundort) {
		t.Errorf("die Meldung sagt nicht, was zu tun ist: %v", fehler)
	}
}
