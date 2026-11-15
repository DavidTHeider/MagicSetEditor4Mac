package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Beispielset ist einen Klick entfernt** —
// #702.
//
// # Wofür
//
// Der erste Weg einer frisch geladenen Anwendung führte durch einen
// Dateidialog — auf einen Rechner, auf dem noch kein `.mse-set` liegt. Seit die
// Vorlagen mitgeliefert werden dürfen, liegt eines im Bündel.
func TestBeispielsetsNennenWasImBuendelLiegt(t *testing.T) {
	ort := t.TempDir()

	for _, name := range []string{"Vanguard.mse-set", "Probe.mse-set", "beiwerk.txt"} {
		if err := os.WriteFile(filepath.Join(ort, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	anwendung := app.NeueApp("")
	anwendung.MitBeispielsetort(func() string { return ort })

	antwort := anwendung.Beispielsets(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Nur `.mse-set`, und in Namensordnung.** Was daneben liegt, ist Beiwerk
	// — ein `.DS_Store` als Beispielset anzubieten wäre eine Zusage, die beim
	// Klick bricht.
	namen := make([]string, 0, len(antwort.Sets))
	for _, satz := range antwort.Sets {
		namen = append(namen, satz.Name)
	}

	if strings.Join(namen, "|") != "Probe|Vanguard" {
		t.Errorf("Sets = %q", namen)
	}

	if antwort.Sets[1].Pfad != filepath.Join(ort, "Vanguard.mse-set") {
		t.Errorf("Pfad %q", antwort.Sets[1].Pfad)
	}
}

// **Ohne Bündel gibt es keine, und das ist kein Fehler.**
//
// Wer die Anwendung entwickelt, baut ohne den Freigabeweg; eine Meldung
// „keine Beispielsets" bei jedem Start wäre eine, die niemand mehr liest.
func TestBeispielsetsOhneBuendelSindLeer(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitBeispielsetort(func() string { return "" })

	antwort := anwendung.Beispielsets(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Eine leere Liste und kein `nil`** — dieselbe Begründung wie überall:
	// das Frontend prüft auf eines statt auf zwei.
	if antwort.Sets == nil || len(antwort.Sets) != 0 {
		t.Errorf("Sets = %v", antwort.Sets)
	}
}

// **Ein unlesbares Verzeichnis wird gemeldet, nicht verschwiegen.**
func TestBeispielsetsMeldenEinenLesefehler(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitBeispielsetort(func() string {
		return filepath.Join(t.TempDir(), "gibtsnicht")
	})

	antwort := anwendung.Beispielsets(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ein Pfad, der nicht in der Liste steht, wird abgewiesen** —
// die Liste ist der Filter, nicht der Aufrufer.
//
// Ohne diese Prüfung gäbe es einen Weg, über den ein Skript jede Datei dieses
// Rechners in den Bestand zöge: `BeispielsetOeffnen("/etc/passwd")` ginge
// unmittelbar an den Setleser.
func TestBeispielsetOeffnenWeistFremdePfadeAb(t *testing.T) {
	ort := t.TempDir()
	if err := os.WriteFile(filepath.Join(ort, "Probe.mse-set"), []byte("x"), 0o600); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	anwendung := app.NeueApp("")
	anwendung.MitBeispielsetort(func() string { return ort })

	bindungen := app.NeueBindungen(anwendung)

	antwort := bindungen.BeispielsetOeffnen(t.Context(), filepath.Join(ort, "..", "woanders.mse-set"))
	if antwort.Fehler == nil {
		t.Fatal("ein fremder Pfad wurde angenommen")
	}

	// **Und die Liste geht durch dieselbe Bindung** — sie ist der Filter, und
	// ein Test, der nur die Fassade prüfte, ließe die Bindung ungeprüft.
	liste := bindungen.Beispielsets(t.Context())
	if liste.Fehler != nil || len(liste.Sets) != 1 {
		t.Errorf("Liste = %+v", liste)
	}
}

// **Die Naht zu den Vorlagen wird vor dem Start gesetzt** — danach bliebe sie
// ungelesen, und das sähe aus wie ein leeres Bündel.
func TestMitVorlagenortWirktVorDemStart(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	ort := t.TempDir()
	anwendung.MitVorlagenort(func() string { return ort })

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})
}

// **Ein Set in der Saat ist ein Beispielset** — Werkstück U7 von Phase 5.1
// (#1418):
// die Testkarten der Werkstatt liegen als `probe-werkstatt.mse-set` neben
// den Saatpaketen, und wer sie öffnen will, findet sie unter „Beispielsets"
// — auch im Entwicklungslauf, der kein `beispielsets/` neben dem Programm hat.
func TestBeispielsetsNennenAuchDieSetsDerSaat(t *testing.T) {
	saat := t.TempDir()

	for _, name := range []string{"probe-werkstatt.mse-set", "probe.mse-game"} {
		if err := os.MkdirAll(filepath.Join(saat, name), 0o750); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	anwendung := app.NeueApp("")
	anwendung.MitBeispielsetort(func() string { return "" })
	anwendung.MitSaatort(func() string { return saat })

	antwort := anwendung.Beispielsets(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Sets) != 1 || antwort.Sets[0].Name != "probe-werkstatt" ||
		antwort.Sets[0].Pfad != filepath.Join(saat, "probe-werkstatt.mse-set") {
		t.Errorf("gefunden %+v", antwort.Sets)
	}
}
