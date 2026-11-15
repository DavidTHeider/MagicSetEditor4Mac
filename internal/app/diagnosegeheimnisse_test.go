//go:build darwin

// **Kein Geheimnis im Paket** — Werkstück O10 von Phase 5.2
// (#1434).
//
// Die Abnahme aus dem Issue: „Mit absichtlich in allen Datenwegen gesetzten
// Testgeheimnissen enthält das Paket keinen Klartextmarker."
//
// # Was dieser Lauf belegt — und was die eigentliche Absicherung ist
//
// **Die Absicherung ist die Auswahl**, nicht die Schwärzung: das Paket liest
// die Tokentafel, den Schlüsselbund, die Kennwörter, die TOTP-Geheimnisse und
// die Antworten fremder Anbieter **gar nicht**. Dieser Lauf prüft, dass das
// auch dann gilt, wenn überall ein Marker steht — und zusätzlich, dass der
// eine Teil mit freiem Text, das Protokoll, ihn nicht durchreicht.
package app_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// marker ist lang genug, dass die Schwärzung ihn als Zufallsfolge erkennt —
// und trägt Buchstaben **und** Ziffern, wie ein echtes Geheimnis.
const marker = "GEHEIMMARKER4711abcdefghijklmnop"

func TestKeinGeheimnisImDiagnosepaket(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	// **In jeden Datenweg, den es gibt**: eine Umgebungsvariable, wie sie der
	// Dienst für seine Marke liest; ein Protokolleintrag mit dem Marker im
	// `grund`, so wie ihn ein Fehler aus dem Unterbau trüge; und derselbe
	// Marker als Nachweis in einer Kopfzeile.
	t.Setenv("MSE_MARKE", marker)
	t.Setenv("MSE_TOKEN", marker)

	slog.Default().Error("ein Ruf wurde abgewiesen",
		"grund", "das Token "+marker+" ist unbekannt")
	slog.Default().Error("ein zweiter Ruf wurde abgewiesen",
		"grund", "Authorization: Bearer "+marker)

	vorschau := anwendung.Diagnosepaket(t.Context())
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	for name, inhalt := range archivinhalte(t, ziel) {
		if strings.Contains(inhalt, marker) {
			t.Errorf("%s trägt den Marker:\n%s", name, inhalt)
		}
	}
}

// **Die Gegenprobe zur Gegenprobe**: ohne Schwärzung stünde der Marker da.
// Ohne diesen Lauf wäre der obige auch dann grün, wenn das Protokoll gar
// nicht ins Paket käme — und dann bewiese er nichts (§10).
func TestDasProtokollKaemeOhneSchwaerzungMitDemMarker(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	slog.Default().Error("ein Ruf wurde abgewiesen",
		"grund", "das Token "+marker+" ist unbekannt")

	vorschau := anwendung.Diagnosepaket(t.Context())
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	protokoll := archivinhalte(t, ziel)["protokoll.txt"]

	switch {
	case !strings.Contains(protokoll, "ein Ruf wurde abgewiesen"):
		t.Fatalf("die Zeile steht gar nicht im Paket — der Marker hätte "+
			"keine Gelegenheit gehabt:\n%s", protokoll)
	case !strings.Contains(protokoll, app.Geschwaerzt):
		t.Errorf("die Zeile ist ungeschwärzt durchgekommen:\n%s", protokoll)
	}
}

// schreibeInsProtokoll richtet den Prozessschreiber auf die Protokolldatei
// dieses Zuhauses.
//
// **Ohne diese Zeilen schreibt `slog.Default()` nach stderr** — und ein Lauf,
// der „kein Marker im Paket" behauptet, während die Zeile das Paket nie
// erreicht, ist grün und beweist nichts. Gefunden hat es die Gegenprobe
// darunter, nicht das Lesen (§10).
func schreibeInsProtokoll(t *testing.T) {
	t.Helper()

	pfad, fehler := app.Protokollpfad()
	if fehler != nil {
		t.Fatalf("Protokollpfad: %v", fehler)
	}

	schreiber, schliessen, anlegefehler := app.Protokoll(pfad)
	if anlegefehler != nil {
		t.Fatalf("Protokoll anlegen: %v", anlegefehler)
	}

	vorher := slog.Default()

	slog.SetDefault(schreiber)

	t.Cleanup(func() {
		slog.SetDefault(vorher)

		if schliessfehler := schliessen(); schliessfehler != nil {
			t.Errorf("schließen: %v", schliessfehler)
		}
	})
}

// **Der freie Wert einer Panik geht nicht ins Paket** (#2399, Vermerk 0193)
// — ein kurzes Kennwort aus Buchstaben fiel durch die Schwärzung nach
// Gestalt. Ein Wert des Laufzeitsystems bleibt, denn er sagt, was geschah.
func TestKeinPaniktextImDiagnosepaket(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	slog.Default().Error("ein Schritt brach mit einer Panik ab", "panik", "Kennwort hunter")
	slog.Default().Error("ein Rückruf brach mit einer Panik ab", "panik", "runtime error: integer divide by zero")

	vorschau := anwendung.Diagnosepaket(t.Context())
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	protokoll := archivinhalte(t, ziel)["protokoll.txt"]
	if strings.Contains(protokoll, "hunter") || !strings.Contains(protokoll, "integer divide by zero") ||
		!strings.Contains(protokoll, "ein Schritt brach mit einer Panik ab") {
		t.Errorf("im Paket:\n%s", protokoll)
	}
}

// **Auch im Format JSON** — Zweitblick über Phase 5, Befund 1: die Weiche
// schreibt JSON, wenn `protokoll.format` es sagt, und das Merkmal heißt dann
// `"panik":"…"`. Geschrieben wird hier mit demselben Handler, den die Weiche
// nimmt (`slog.NewJSONHandler`), in dieselbe Datei.
func TestKeinPaniktextImDiagnosepaketAuchAlsJSON(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	pfad, fehler := app.Protokollpfad()
	if fehler != nil {
		t.Fatalf("Protokollpfad: %v", fehler)
	}

	datei, fehler := os.OpenFile(pfad, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) //nolint:gosec // der Protokollpfad des Laufs
	if fehler != nil {
		t.Fatal(fehler)
	}

	schreiber := slog.New(slog.NewJSONHandler(datei, nil))
	schreiber.Error("ein Schritt brach mit einer Panik ab", "panik", "Kennwort hunter")
	schreiber.Error("ein Rückruf brach mit einer Panik ab", "panik", "runtime error: integer divide by zero")

	if fehler := datei.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	protokoll := archivinhalte(t, ziel)["protokoll.txt"]
	if strings.Contains(protokoll, "hunter") || !strings.Contains(protokoll, "integer divide by zero") {
		t.Errorf("im Paket:\n%s", protokoll)
	}
}
