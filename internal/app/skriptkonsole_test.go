package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// konsolenprobe ist ein Double der Auswertung (§3).
type konsolenprobe struct {
	ergebnis app.KonsolenDTO
	fehler   error

	// namen ist, was die Umgebung kennt — für `Skriptnamen` (#1164).
	namen []string

	karteID  int64
	ausdruck string
}

// Namen gibt, was ein Ausdruck an dieser Karte nennen darf.
func (k *konsolenprobe) Namen(_ context.Context, karteID int64) ([]string, error) {
	k.karteID = karteID

	return k.namen, k.fehler
}

func (k *konsolenprobe) Auswerten(
	_ context.Context, karteID int64, ausdruck string,
) (app.KonsolenDTO, error) {
	k.karteID, k.ausdruck = karteID, ausdruck

	return k.ergebnis, k.fehler
}

// **Der Ausdruck kommt an, das Ergebnis zurück** —
// #1070.
//
// **Der Quelltext, nicht der Text** — so zeigt MSE es (`src/cli/cli_main.cpp`,
// Zeile 234, Fassung f9d9356 vom 2021-07-06): wer eine Konsole benutzt, will
// eine Zeichenkette von einem Namen unterscheiden.
func TestDieKonsolenbindungWertetAus(t *testing.T) {
	probe := &konsolenprobe{ergebnis: app.KonsolenDTO{
		Quelltext: `"ab"`, Text: "ab", Art: "text",
		Befunde: []string{"kein Stilpaket"},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(probe)

	antwort := anwendung.Skript(t.Context(), "7", `"a" + "b"`)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.karteID != 7 || probe.ausdruck != `"a" + "b"` {
		t.Errorf("angekommen ist %d / %q", probe.karteID, probe.ausdruck)
	}

	if antwort.Quelltext != `"ab"` || antwort.Text != "ab" || antwort.Art != "text" {
		t.Errorf("die Antwort ist %+v", antwort)
	}

	if len(antwort.Befunde) != 1 {
		t.Errorf("die Befunde sind %+v", antwort.Befunde)
	}
}

// **Leere Befunde kommen als leere Liste** — über die Wails-Bindung stünde
// dort sonst `null`.
func TestDieKonsolenbindungGibtEineLeereBefundliste(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(&konsolenprobe{})

	if antwort := anwendung.Skript(t.Context(), "7", "1"); antwort.Befunde == nil {
		t.Error("nil statt einer leeren Liste")
	}
}

// **Jeder Fehlerweg meldet, und keiner stürzt ab.**
func TestDieKonsolenbindungWeistAb(t *testing.T) {
	faelle := map[string]struct {
		kennung string
		probe   app.Skriptkonsolenquelle
	}{
		"keine Zahl":               {"abc", &konsolenprobe{}},
		"die Auswertung scheitert": {"7", &konsolenprobe{fehler: errors.New("1 +")}},
		"die Konsole fehlt ganz":   {"7", nil},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			anwendung := app.NeueApp("")
			if fall.probe != nil {
				anwendung.MitSkriptkonsole(fall.probe)
			}

			antwort := anwendung.Skript(t.Context(), fall.kennung, "1")
			if antwort.Fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}

			if antwort.Befunde == nil {
				t.Error("nil statt einer leeren Liste")
			}
		})
	}
}
