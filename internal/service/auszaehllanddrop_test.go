//go:build fremdpakete

package service_test

// **Die beiden `land drop hit %` gegen das Orakel** —
// #1356.
//
// # Warum eigens
//
// Sie sind die zwei Dimensionen, für die es dieses Issue gibt: sie rechnen
// über **das ganze Set** und geben ihr Ergebnis nur an der ersten Karte aus.
// Dass sie **rechnen**, zählt `auszaehlfremd_test.go`; dass sie **dasselbe**
// rechnen wie MSE2, steht hier.
//
//	go test -tags fremdpakete ./internal/service/ -run LandDrop -v
//
// # Warum `Colorless` und kein anderes Set
//
// **Weil die Dimension sonst gar nicht rechnet.** Sie hat zwei Wächter, und
// beide geben eine Meldung statt einer Zahl:
//
//	Archenemy   !! Not enough non-commander non-promo cards … !!   (70 Karten)
//	White       !! No lands found， 0% chance to hit land drops !!
//	Multicolor  dieselbe Meldung
//
// Gemessen am 2026-08-31. Zwei gleiche Warnmeldungen zu vergleichen sähe aus
// wie eine Übereinstimmung und wäre keine — es ist genau die Falle aus §10:
// **die Reichweite einer Messung ist die Reichweite ihrer Aussage.**
//
// `Colorless.mse-set` führt Länder und rechnet: `Turn 1 — 59.4%`,
// `Turn 2 — 24.3%`, … über 3.216 Karten in 21 s. Der Fall unten hält
// ausdrücklich fest, dass keine Warnmeldung verglichen wird.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// landdroporakel fragt MSE2 nach der Dimension für jede Karte.
func landdroporakel(t *testing.T, setpfad string, beginn int) []string {
	t.Helper()

	if _, fehler := os.Stat(orakel); fehler != nil {
		t.Skipf("das Orakel ist nicht gebaut: %v", fehler)
	}

	frist, abbrechen := context.WithTimeout(t.Context(), 10*time.Minute)
	defer abbrechen()

	// #nosec G204 -- der Pfad ist eine Konstante dieser Datei, das Set kommt
	// aus dem Zuhause dieses Rechners, und der Lauf steht hinter `fremdpakete`.
	befehl := exec.CommandContext(frist, orakel, "--cli", setpfad, "-q", "--raw")
	befehl.Stdin = strings.NewReader(
		"n := length(set.cards)-1\n" +
			"for x from 0 to n do [ land_drop_hit_percentage_statistic(" +
			"start: " + string(rune('0'+beginn)) + ", card: set.cards[x]) ]\n")

	roh, fehler := befehl.Output()
	if fehler != nil {
		t.Fatalf("das Orakel: %v", fehler)
	}

	zeilen := strings.Split(strings.TrimSpace(string(roh)), "\n")

	var antwort []string
	if fehler := json.Unmarshal(
		[]byte(zeilen[len(zeilen)-1]), &antwort); fehler != nil {
		t.Fatalf("die Antwort des Orakels ist keine Liste: %v (%s)",
			fehler, zeilen[len(zeilen)-1])
	}

	return antwort
}

// TestDieLandDropDimensionenStimmenMitDemOrakelUeberein ist die Abnahme.
func TestDieLandDropDimensionenStimmenMitDemOrakelUeberein(t *testing.T) {
	setpfad := kanonischesSet(t, "Colorless.mse-set")

	paket, fehler := msepaket.Oeffne(filepath.Join(packort, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(
		msepaket.NeuerPaketleser(packort)); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	spiel, _ := paket.Spielmodell()

	karten, setangaben := kartenUndAngabenAusSet(t, setpfad)

	speicher := service.NeuerStilspeicher(&setkopfprobe{
		kopf:  domain.SetHead{ID: 1, Game: "magic", Style: "m15-altered"},
		werte: setangaben,
	}, packort)

	vorlage := magicvorlage(spiel)
	vorlage.Setangaben = setangaben

	for _, fall := range []struct {
		name   string
		beginn int
	}{
		{"land drop hit % (On The Play)", 6},
		{"land drop hit % (On The Draw)", 7},
	} {
		t.Run(fall.name, func(t *testing.T) {
			var dimension msepaket.Auszaehldimension

			for _, eine := range paket.Auszaehldimensionen() {
				if eine.Name == fall.name {
					dimension = eine
				}
			}

			if dimension.Skript == "" {
				t.Fatalf("die Dimension %q steht nicht im Paket", fall.name)
			}

			vonMSE := landdroporakel(t, setpfad, fall.beginn)

			if len(karten) != len(vonMSE) {
				t.Fatalf("%d Karten hier, %d beim Orakel", len(karten), len(vonMSE))
			}

			begonnen := time.Now()

			werte, befunde := service.Auszaehlwerte(
				t.Context(), karten, vorlage,
				func(karte *domain.CardRow) service.Vorlage {
					eigene, _, da := speicher.StilFuerKarte(t.Context(), karte)
					if !da {
						return vorlage
					}

					eigene.Spielskript = vorlage.Spielskript
					eigene.Setfelder = vorlage.Setfelder
					eigene.Kartenfelder = vorlage.Kartenfelder
					eigene.Setangaben = setangaben

					return eigene
				},
				service.Auszaehlangabe{
					Name: dimension.Name, Skript: dimension.Skript, LeereZeigen: true,
				})

			t.Logf("%d Karten in %s gerechnet", len(karten),
				time.Since(begonnen).Round(time.Millisecond))

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %+v", befunde)
			}

			if len(werte) != len(vonMSE) {
				t.Fatalf("%d Werte hier, %d beim Orakel", len(werte), len(vonMSE))
			}

			// **Die erste Karte trägt die Antwort, alle anderen sind leer.**
			// Genau das ist die Gestalt, um die es geht.
			abweichungen := 0

			for i := range werte {
				if werte[i] != vonMSE[i] {
					abweichungen++

					if abweichungen <= 3 {
						t.Errorf("Karte %d: hier %q, MSE %q", i, werte[i], vonMSE[i])
					}
				}
			}

			if abweichungen > 0 {
				t.Errorf("%d von %d Karten weichen ab", abweichungen, len(werte))
			}

			if werte[0] == "" {
				t.Error("die erste Karte gibt nichts — die Dimension hat nicht gerechnet")
			}

			// **Und sie hat wirklich gerechnet.**
			//
			// Die Dimension gibt bei zu wenigen Karten oder ohne Länder eine
			// Meldung zwischen `!!`. Zwei gleiche Meldungen zu vergleichen
			// sähe aus wie eine Übereinstimmung und wäre keine.
			if strings.HasPrefix(werte[0], "!!") {
				t.Errorf("die Dimension gibt eine Warnung statt einer Rechnung: %q", werte[0])
			}

			if !strings.Contains(werte[0], "%") {
				t.Errorf("die Antwort enthält keinen Prozentsatz: %q", werte[0])
			}
		})
	}
}
