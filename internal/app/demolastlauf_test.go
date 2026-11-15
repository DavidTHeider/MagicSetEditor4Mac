//go:build fremdpakete

// Der **Lastlauf** über den Demokartensatz
// (#476).
//
// Er beantwortet zwei Fragen, die `TestDemokartenZeichnen` nicht stellt:
//
//   - **Wie effizient?** Wie lange braucht eine Karte — getrennt nach Planbau
//     und Rasterung.
//   - **Wie stabil?** Kommt zweimal dasselbe Bild heraus, und bleibt das über
//     viele Läufe so.
//
// # Was er misst und was nicht (§10)
//
// **Er misst diesen Rechner.** Eine Zahl von hier sagt nichts über einen
// anderen; sie ist als **Verhältnis** zwischen den Karten und als Verlaufsmaß
// über die Zeit brauchbar, nicht als Zusage.
//
// **Er misst bei Vergrößerung 1** — die Bildschirmauflösung, 375 × 523. Der
// Druckbogen rechnet mit dem Vierfachen, und das ist nicht dasselbe mal vier:
// die Glyphenrasterung hängt an der Größe.
//
// **Er hat keinen Deckel.** Eine Schranke auf einer Wanduhrzeit wäre auf einem
// belasteten Rechner rot, ohne dass sich etwas geändert hat. Was hier steht,
// ist ein Protokoll — wer eine Verschlechterung sucht, vergleicht zwei Läufe.
package app_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// LaeufeJeKarte ist, wie oft eine Karte im Lastlauf gezeichnet wird.
//
// **Fünf, und der Grund gehört dazu**: der erste Lauf zahlt das Öffnen der
// Schriften und das Laden der Rahmenbilder, die folgenden nicht. Mit fünf ist
// der Median vom ersten unabhängig, und einundzwanzig Karten bleiben unter
// einer Minute.
const LaeufeJeKarte = 5

// TestDerLastlaufUeberDenDemokartensatz misst Zeit und Stabilität.
func TestDerLastlaufUeberDenDemokartensatz(t *testing.T) {
	wurzel := fremdpaketwurzel(t)

	var (
		gesamtPlan, gesamtRaster time.Duration
		karten                   int
	)

	for _, karte := range demokarten() {
		t.Run(karte.name, func(t *testing.T) {
			plan, raster := lastJeKarte(t, wurzel, karte)

			gesamtPlan += plan
			gesamtRaster += raster
			karten++
		})
	}

	if karten == 0 {
		return
	}

	t.Logf("zusammen: %d Karten, je %d Läufe (der erste zählt nicht mit) — "+
		"Planbau %v im Mittel, Rasterung %v im Mittel, eine Karte %v",
		karten, LaeufeJeKarte,
		gerundet(gesamtPlan/time.Duration(karten)),
		gerundet(gesamtRaster/time.Duration(karten)),
		gerundet((gesamtPlan+gesamtRaster)/time.Duration(karten)))
}

// lastJeKarte zeichnet eine Karte mehrfach und gibt die Mediane.
func lastJeKarte(t *testing.T, wurzel string, karte demokarte) (plan, raster time.Duration) {
	t.Helper()

	zeichne := demozeichner(t, wurzel, karte)

	var (
		planzeiten, rasterzeiten []time.Duration
		erstes                   *image.NRGBA
	)

	for lauf := range LaeufeJeKarte {
		bild, planzeit, rasterzeit := zeichne(t)

		planzeiten = append(planzeiten, planzeit)
		rasterzeiten = append(rasterzeiten, rasterzeit)

		if lauf == 0 {
			erstes = bild

			continue
		}

		// **Zweimal dasselbe Bild.** Eine Abbildungsiteration, ein
		// Zwischenspeicher, der von der Reihenfolge abhängt, oder eine
		// nebenläufige Ablage fällt hier auf und sonst nirgends.
		if !bytes.Equal(erstes.Pix, bild.Pix) {
			t.Fatalf("Lauf %d gibt ein anderes Bild als Lauf 1 — die Zeichnung "+
				"ist nicht wiederholbar", lauf+1)
		}
	}

	// **Der erste Lauf steht für sich.** Er zahlt das Öffnen der Schriftdatei
	// und das Laden der Rahmenbilder; die folgenden finden beides vor. Ihn in
	// denselben Median zu werfen hieße, eine Zahl zu nennen, die keinen Fall
	// beschreibt — weder den kalten noch den warmen (§10).
	t.Logf("  kalt: Planbau %v, Rasterung %v — warm: Planbau %v (min %v, "+
		"max %v), Rasterung %v (min %v, max %v)",
		gerundet(planzeiten[0]), gerundet(rasterzeiten[0]),
		gerundet(mitte(planzeiten[1:])), gerundet(kleinste(planzeiten[1:])),
		gerundet(groesste(planzeiten[1:])),
		gerundet(mitte(rasterzeiten[1:])), gerundet(kleinste(rasterzeiten[1:])),
		gerundet(groesste(rasterzeiten[1:])))

	return mitte(planzeiten[1:]), mitte(rasterzeiten[1:])
}

// demozeichner richtet eine Karte ein und gibt einen Zeichner darauf.
//
// **Die Einrichtung steht außerhalb der Messung.** Ein Stilpaket zu öffnen
// heißt, eine Textdatei zu zerteilen und ihre Module nachzuladen; das in die
// Zeit einer Karte zu rechnen, machte den ersten Lauf zur Aussage über den
// Paketleser (§10).
func demozeichner(
	t *testing.T, wurzel string, karte demokarte,
) func(*testing.T) (*image.NRGBA, time.Duration, time.Duration) {
	t.Helper()

	stil := fremdstil(t, wurzel, karte.stil)
	suche := schriften.Neue(schriften.Orte()).MitPaketen(strings.TrimSuffix(wurzel, "/data"))
	bilder := bildquelleFuer(wurzel)

	vorlage := fremdvorlage(t, stil, wurzel)
	vorlage.Setangaben = demoset()

	if karte.kunst != "" {
		vorlage.Setpaket = kunstsetFuer(t, karte.kunst)
	}

	zeile := demokartenzeile(karte)

	return func(t *testing.T) (*image.NRGBA, time.Duration, time.Duration) {
		t.Helper()

		vorPlan := time.Now()

		plan, _, err := service.NeuerPlanbauer().
			MitSchriften(suche).MitBildern(bilder).
			Stilplan(probekontext(t), zeile, vorlage)

		planzeit := time.Since(vorPlan)

		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		vorRaster := time.Now()

		bild, _, err := render.NeuerRasterer(bilder).MitSchriften(suche).
			Zeichne(probekontext(t), plan, 1, render.GueteMittel)

		rasterzeit := time.Since(vorRaster)

		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		return bild, planzeit, rasterzeit
	}
}

// mitte gibt den Median.
//
// **Nicht das Mittel**: ein einzelner Ausreißer — der Rechner tut kurz etwas
// anderes — zöge es mit, und der Median nicht.
func mitte(dauern []time.Duration) time.Duration {
	if len(dauern) == 0 {
		return 0
	}

	sortiert := append([]time.Duration(nil), dauern...)
	sort.Slice(sortiert, func(a, b int) bool { return sortiert[a] < sortiert[b] })

	return sortiert[len(sortiert)/2]
}

func kleinste(dauern []time.Duration) time.Duration {
	sortiert := append([]time.Duration(nil), dauern...)
	sort.Slice(sortiert, func(a, b int) bool { return sortiert[a] < sortiert[b] })

	return sortiert[0]
}

func groesste(dauern []time.Duration) time.Duration {
	sortiert := append([]time.Duration(nil), dauern...)
	sort.Slice(sortiert, func(a, b int) bool { return sortiert[a] < sortiert[b] })

	return sortiert[len(sortiert)-1]
}

// gerundet macht aus einer Dauer eine, die man lesen kann.
func gerundet(dauer time.Duration) time.Duration {
	if dauer >= time.Millisecond {
		return dauer.Round(100 * time.Microsecond)
	}

	return dauer.Round(time.Microsecond)
}

// **Ein abgebrochener Lauf hört auf** (§7).
//
// Eine Kartenvorschau wird abgebrochen, nicht abgewartet: wer während des
// Tippens vier Läufe anstößt, zeigt am Ende den drittletzten Stand, wenn er
// den Lauf nicht an einen `context` bindet. Der Lastlauf sagt, was das wert
// ist — **77 Millisekunden je Karte** auf diesem Rechner, und das ist eine
// Zeit, in der jemand weitertippt.
func TestEinAbgebrochenerLaufHoertAuf(t *testing.T) {
	wurzel := fremdpaketwurzel(t)

	karte := demokarten()[0]
	stil := fremdstil(t, wurzel, karte.stil)
	suche := schriften.Neue(schriften.Orte()).MitPaketen(strings.TrimSuffix(wurzel, "/data"))
	bilder := bildquelleFuer(wurzel)

	vorlage := fremdvorlage(t, stil, wurzel)
	vorlage.Setangaben = demoset()

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(suche).MitBildern(bilder).
		Stilplan(probekontext(t), demokartenzeile(karte), vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	abgebrochen, brich := context.WithCancel(probekontext(t))
	brich()

	if _, _, err := render.NeuerRasterer(bilder).MitSchriften(suche).
		Zeichne(abgebrochen, plan, 1, render.GueteMittel); !errors.Is(err, context.Canceled) {
		t.Errorf("zeichnen gibt %v, erwartet context.Canceled", err)
	}
}
