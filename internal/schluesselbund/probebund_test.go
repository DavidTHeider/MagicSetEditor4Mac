//go:build darwin

package schluesselbund_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// probebund legt einen eigenen Schlüsselbund an — **nie den echten**.
//
// **Er liegt in `t.TempDir()` und wird danach gelöscht.** Ein Test, der in den
// Anmeldebund schriebe, hinterließe dort einen Eintrag und stolperte über den
// Zugriffsdialog von macOS (§3).
func probebund(t *testing.T) *schluesselbund.Bund {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.keychain")

	//nolint:gosec // das Programm steht als Literal, der Ort liegt in t.TempDir()
	if aus, err := exec.CommandContext(t.Context(), "/usr/bin/security",
		"create-keychain", "-p", "leer", ort).CombinedOutput(); err != nil {
		t.Fatalf("den Probebund anlegen: %v — %s", err, aus)
	}

	t.Cleanup(func() {
		// **Ein eigener Kontext**, weil der des Laufs beim Aufräumen schon
		// abgebrochen ist — dann räumte nichts ab.
		ctx, fertig := context.WithTimeout(context.Background(), 10*time.Second)
		defer fertig()

		//nolint:gosec,errcheck // dito; und ein Probebund, der stehen bleibt,
		// liegt in einem Verzeichnis, das der Lauf ohnehin abräumt
		_ = exec.CommandContext(ctx, "/usr/bin/security",
			"delete-keychain", ort).Run()
	})

	// **Aufgeschlossen, damit `security` nicht nach einem Passwort fragt.**
	// Angelegt wird der Bund mit leerem Passwort; ein Bund, der zufällt,
	// hielte den nächsten Aufruf an, bis die Frist ihn erschlägt.
	//
	//nolint:gosec,errcheck // dito; und ein Bund, der schon offen ist,
	// beantwortet das mit Erfolg
	_ = exec.CommandContext(t.Context(), "/usr/bin/security",
		"unlock-keychain", "-p", "leer", ort).Run()

	// **Acht Sekunden und nicht zwei.**
	//
	// Im Betrieb sind es zehn. Hier stand bis zum 2026-09-02 eine Frist von
	// **zwei** Sekunden, damit ein Lauf nicht dreizehn Sekunden auf ein
	// `security` wartet, das nach einem Passwort fragt.
	//
	// **Das war zu knapp.** Der gesunde Aufruf dauert auf dieser Maschine
	// 0,1 bis 0,35 Sekunden — unter Last aber länger als zwei, und dann
	// erschlägt die Frist ihn: „das Geheimnis lesen: signal: killed".
	// Gemessen: **einer von sechs** Läufen rot ohne Last, **zwei von sechs**
	// unter Last.
	//
	// Eine Vorrichtung, die **strenger** ist als der Betrieb, ist derselbe
	// Fehler wie eine, die großzügiger ist (§3) — sie misst die Maschine
	// statt die Sache.
	// **Eine lange Frist für die Probe.** Die übliche (zehn Sekunden) gilt dem
	// Fenster, das nicht hängen soll; unter `-race` neben allen anderen
	// Paketen brauchte `security` in der Pipeline länger, und der Lauf endete
	// mit »signal: killed« — ein Befund über die Maschine, nicht über den Bund.
	return schluesselbund.NeuerBund(ort).MitFrist(8 * time.Second).MitFrist(2 * time.Minute)
}
