//go:build darwin

// **Keine Dienstmarke im Diagnosepaket** — SEC02
// (#2355),
// Abnahme „Supportpaket enthält null Marken".
//
// # Was dieser Lauf belegt
//
// **Die Absicherung ist wieder die Auswahl**: das Paket liest weder
// `dienst.txt` noch `dienst.log`. Dieser Lauf legt beide mit Marke an und
// schreibt die ganze Ansage zusätzlich ins Protokoll — so, als hätte jemand
// sie dort mitgeschrieben —, und prüft, dass keine Marke durchkommt.
//
// **Und eine Marke ist eine Hexfolge, keine Mischung**: sie kann aus Ziffern
// allein bestehen oder aus `a` bis `f` allein. Die Schwärzung verlangte
// Ziffer **und** Buchstabe; zwei der drei Fälle hier fielen durch.
package app_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestKeineDienstmarkeImDiagnosepaket(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	gewuerfelt := make([]byte, 32)
	_, _ = rand.Read(gewuerfelt) //nolint:errcheck // crypto/rand scheitert nicht

	marken := map[string]string{
		"gewürfelt":      hex.EncodeToString(gewuerfelt),
		"nur Ziffern":    hex.EncodeToString(bytes.Repeat([]byte{0x11}, 32)),
		"nur Buchstaben": hex.EncodeToString(bytes.Repeat([]byte{0xab}, 32)),
	}

	// **Wo der eingerichtete Dienst sie hinschreibt** — neben dem Bestand
	// dieses Zuhauses (`schnittstellenziele.go`).
	eigene := filepath.Join(filepath.Dir(bestand), "schnittstellen")

	for name, marke := range marken {
		ansage := app.Dienstansage{Adresse: "127.0.0.1:1", Marke: marke}

		if fehler := app.AnsageInDatei(filepath.Join(eigene, "dienst.txt"), ansage); fehler != nil {
			t.Fatalf("%s: die Ansage schreiben: %v", name, fehler)
		}

		// #nosec G703 -- der Ordner liegt unter dem Zuhause aus t.TempDir
		if fehler := os.WriteFile(filepath.Join(eigene, "dienst.log"),
			[]byte(ansage.Text()), 0o600); fehler != nil {
			t.Fatalf("%s: das Protokoll des Dienstes schreiben: %v", name, fehler)
		}

		// **Im `grund`**, wie ein Fehler aus dem Unterbau ihn trüge — ein
		// Merkmal ohne Erklärung fiele ganz weg und prüfte die Schwärzung nicht.
		slog.Default().Error("der Dienst sagte an", "grund", name+": "+ansage.Text())
	}

	vorschau := anwendung.Diagnosepaket(t.Context())
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(t.Context(), ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	inhalte := archivinhalte(t, ziel)

	// **Die Gegenprobe**: die Zeilen erreichen das Paket wirklich — sonst
	// wäre der Lauf auch ohne jede Schwärzung grün (§10).
	if strings.Count(inhalte["protokoll.txt"], "der Dienst sagte an") != len(marken) ||
		strings.Count(inhalte["protokoll.txt"], `Marke `) < len(marken) {
		t.Fatalf("die Ansagen stehen nicht im Protokoll des Pakets:\n%s", inhalte["protokoll.txt"])
	}

	for teil, inhalt := range inhalte {
		for name, marke := range marken {
			if strings.Contains(inhalt, marke) {
				t.Errorf("%s trägt die Marke %q:\n%s", teil, name, inhalt)
			}
		}
	}
}
