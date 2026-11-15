//go:build darwin

package app_test

// **Die Absturzspuren gehen ins Diagnosepaket — geschwärzt** — F04
// (#2343).
//
// Künstliche Spuren mit je einem Geheimnis in einer anderen Gestalt: ein
// Token in der Panikmeldung, ein Nachweis in einer Kopfzeile, ein
// Base32-Geheimnis neben einem Pfad — **und zwei, die der Schwärzer nicht
// erkennt**: ein kurzes Kennwort und eine Passphrase nur aus Buchstaben
// (Zweitblick über Phase 1 von 0.9.0, Befund 7). Im Paket steht jede Spur
// mit Stapel und Laufzeitfehler, keins der Geheimnisse und kein freier
// Paniktext. **Keine Telemetrie**: das Paket entsteht nur auf Wunsch.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieAbsturzspurenStehenGeschwaerztImPaket(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	geheimnisse := []string{
		"TOKENGEHEIM4711abcdefghijklmnopqrstu",
		"BEARERGEHEIM0815abcdefghijklmnopqrs",
		"JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP7Q",
		"hunter2",
		"korrektpferdbatterieheftklammer",
	}
	spuren := []string{
		"panic: das Token " + geheimnisse[0] + " ist unbekannt\n\ngoroutine 1 [running]:\n" +
			"main.lauf(0x14000123456, 0x7)\n\t/Users/probe/main.go:10 +0x1c\n",
		"fatal error: Authorization: Bearer " + geheimnisse[1] + "\n",
		"panic: /Users/probe/Library/geheim " + geheimnisse[2] + "\n",
		"panic: das Kennwort " + geheimnisse[3] + " passt nicht [recovered]\n" +
			"\tpanic: " + geheimnisse[4] + "\n",
		"panic: runtime error: index out of range [5] with length 3\n\ngoroutine 7 [running]:\n",
	}

	ordner := t.TempDir()
	for i, spur := range spuren {
		name := filepath.Join(ordner, "absturz-"+string(rune('1'+i))+"-1.txt")
		if err := os.WriteFile(name, []byte(spur), 0o600); err != nil {
			t.Fatal(err)
		}

		geschrieben := time.Date(2026, 10, 13, 12, i, 0, 0, time.UTC)
		if err := os.Chtimes(name, geschrieben, geschrieben); err != nil {
			t.Fatal(err)
		}
	}

	if err := anwendung.StartmarkeSetzen(ordner); err != nil {
		t.Fatal(err)
	}

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	alles := ""
	for _, inhalt := range archivinhalte(t, ziel) {
		alles += inhalt
	}

	if strings.Count(alles, "Absturzspur") < len(spuren) {
		t.Errorf("die Spuren stehen nicht im Paket:\n%s", alles)
	}

	// **Was das Laufzeitsystem schreibt, bleibt** — Stapel, Funktionen und
	// ein Laufzeitfehler; **was der Code als Wert warf, nicht**.
	for _, bleibt := range []string{"main.lauf(", "goroutine 7 [running]:", "runtime error: index out of range [5] with length 3"} {
		if !strings.Contains(alles, bleibt) {
			t.Errorf("im Paket fehlt %q:\n%s", bleibt, alles)
		}
	}

	if strings.Contains(alles, "ist unbekannt") || strings.Contains(alles, "passt nicht") {
		t.Errorf("ein freier Paniktext steht im Paket:\n%s", alles)
	}

	// **Der Kopf nennt, wann die Spur geschrieben wurde** — nicht ihren
	// Dateinamen: den nahm der Schwärzer am nativen Fenster als Kennung, und
	// übrig blieb „Absturzspur «geschwärzt».txt" (wiederanlauf09, 2026-10-14).
	if !strings.Contains(alles, "--- Absturzspur, geschrieben 2026-10-13T12:04:00Z ---") {
		t.Errorf("der Kopf der jüngsten Spur nennt ihre Zeit nicht:\n%s", alles)
	}

	for _, geheim := range geheimnisse {
		if strings.Contains(alles, geheim) {
			t.Errorf("das Paket trägt ein Geheimnis: %s", geheim)
		}
	}

	_ = anwendung.Beenden() //nolint:errcheck // nimmt die Absturzausgabe wieder weg
}
