//go:build darwin

package app_test

// **Was die Anwendung über sich selbst sagt** — OPS-07
// (#1724):
// Fassungen, Signatur ohne Beschönigung, Aktualisierung, Orte und ein
// Prüfstand, der Geprüftes von bloß Vorhandenem trennt.

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieBetriebsauskunftSagtWasIst(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	auskunft := anwendung.Betriebsauskunft(t.Context())
	if auskunft.Fehler != nil {
		t.Fatalf("%+v", auskunft.Fehler)
	}

	// **Ad-hoc und nicht notarisiert** (AK-03, Entscheidung 0003) — die
	// Auskunft behauptet nichts, was nicht so ist.
	if auskunft.Fassung != app.Fassung || auskunft.Signatur != app.SignaturAdhoc || auskunft.Notarisiert {
		t.Errorf("Fassung %q, Signatur %q, notarisiert %v", auskunft.Fassung, auskunft.Signatur, auskunft.Notarisiert)
	}

	if auskunft.Migrationen < 1 || auskunft.LetzteMigration == "" || auskunft.Aktualisierung != "aus" {
		t.Errorf("Migrationen %d, letzte %q, Aktualisierung %q",
			auskunft.Migrationen, auskunft.LetzteMigration, auskunft.Aktualisierung)
	}

	namen := make([]string, 0, len(auskunft.Orte))
	for _, ort := range auskunft.Orte {
		namen = append(namen, ort.Name)

		// **Das Zuhause als Tilde** — dieselbe Regel wie in der Diagnose.
		if !strings.HasPrefix(ort.Pfad, "~/") {
			t.Errorf("%s: %q", ort.Name, ort.Pfad)
		}
	}

	for _, name := range []string{
		app.OrtBestand, app.OrtPakete, app.OrtBilder, app.OrtZwischenspeicher, app.OrtProtokoll, app.OrtProfile,
	} {
		if !slices.Contains(namen, name) {
			t.Errorf("%s fehlt in %v", name, namen)
		}
	}

	if len(auskunft.Pruefstand) == 0 {
		t.Error("kein Prüfstand")
	}
}

// **Ohne Bestand keine Migrationen** — und kein Fehler: die Auskunft steht
// auch dann, wenn nichts aufging.
func TestDieBetriebsauskunftOhneBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	auskunft := app.NeueApp("").Betriebsauskunft(t.Context())
	if auskunft.Fehler != nil || auskunft.Migrationen != 0 || auskunft.Aktualisierung != "aus" {
		t.Errorf("%+v", auskunft)
	}
}
