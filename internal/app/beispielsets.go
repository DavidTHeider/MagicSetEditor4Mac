package app

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// **Die Sets, die das Bündel mitbringt** —
// #702.
//
// # Wofür
//
// Der erste Weg einer frisch geladenen Anwendung führte durch einen
// Dateidialog — auf einen Rechner, auf dem noch kein `.mse-set` liegt. Wer
// nachsehen wollte, was dieses Programm kann, musste erst eines besorgen.
//
// # Warum eine eigene Bindung und nicht der Dateidialog mit Vorgabeort
//
// **Weil ein Dialog eine Frage stellt, und hier keine offen ist.** Ein Klick
// auf „Beispielset öffnen" hat genau eine Antwort; ein Dialog davor wäre ein
// Schritt, der nichts entscheidet.

// BeispielsetDTO ist ein mitgeliefertes Set.
type BeispielsetDTO struct {
	// Name ist der Dateiname ohne Endung — **unübersetzt** (§9): er steht so
	// in der Datei und ist der Name, unter dem es MSE kennt.
	Name string `json:"name"`

	Pfad string `json:"pfad"`
}

// BeispielsetsAntwort trägt die Liste oder einen Fehler.
type BeispielsetsAntwort struct {
	Sets   []BeispielsetDTO `json:"sets"`
	Fehler *Fehler          `json:"fehler"`
}

// Beispielsets nennt, was im Bündel liegt.
//
// **Ohne Bündel eine leere Liste und kein Fehler.** Wer die Anwendung
// entwickelt, baut ohne den Freigabeweg; eine Meldung „keine Beispielsets" bei
// jedem Start wäre eine, die niemand mehr liest.
//
// **Ein unlesbares Verzeichnis dagegen wird gemeldet.** Das ist der Fall
// „mitgeliefert, aber kaputt", und der gehört gesagt (§8).
func (a *App) Beispielsets(ctx context.Context) BeispielsetsAntwort {
	ctx = a.imProfil(ctx)

	sets := []BeispielsetDTO{}

	// **Zwei Orte, eine Liste** — die mitgelieferten Sets neben dem Programm
	// und die Sets in der Saat (U7, #1418): die Testkarten der Werkstatt
	// liegen als `probe-werkstatt.mse-set` bei den Saatpaketen, damit sie
	// auch der Entwicklungslauf hat, der kein `beispielsets/` kennt.
	for _, ort := range []string{a.beispielsetort(), a.saatort()} {
		if ort == "" {
			continue
		}

		eintraege, err := os.ReadDir(ort)
		if err != nil {
			return BeispielsetsAntwort{Sets: []BeispielsetDTO{}, Fehler: Uebersetze(ctx, err)}
		}

		for _, eintrag := range eintraege {
			name := eintrag.Name()
			if !strings.HasSuffix(name, ".mse-set") {
				continue
			}

			sets = append(sets, BeispielsetDTO{
				Name: strings.TrimSuffix(name, ".mse-set"),
				Pfad: filepath.Join(ort, name),
			})
		}
	}

	sort.Slice(sets, func(i, j int) bool { return sets[i].Name < sets[j].Name })

	return BeispielsetsAntwort{Sets: sets}
}

func (a *App) MitBeispielsetort(ort func() string) { a.beispielsetort = ort }

// MitSaatort setzt die Naht zur Saat — für Läufe mit eigener Saat.
func (a *App) MitSaatort(ort func() string) { a.saatort = ort }

// MitVorlagenort setzt die Naht zu den mitgelieferten Paketen.
//
// **Vor `Starten`**, denn die Paketwurzeln entstehen dort — danach gesetzt
// bliebe der Ort ungelesen, und das sähe aus wie ein leeres Bündel.
func (a *App) MitVorlagenort(ort func() string) { a.vorlagenort = ort }
