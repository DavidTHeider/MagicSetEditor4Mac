//go:build fremdpakete

package app

// **Wie oft ändert sich die Feldordnung beim Blättern?** — #2309, Folge aus
// dem Nachlauf über J05 (#2269), Befund 7.
//
// Das Formular übernimmt beim Blättern die Flächen der vorigen Karte, wenn
// beide denselben Stil tragen, bis die eigenen kommen
// (`frontend/src/seiten/formularruhe.ts`). Ein Stil darf seine Felder je
// Karte per Skript legen — 361 der 376 Stile des Full-Magic-Pack tun es
// mindestens einmal. **Ob das die Reihenfolge im Formular ändert**, zählt
// dieser Lauf: er blättert in jedem Kanonset durch die ersten Karten, holt
// die Flächen wie das Fenster (`Vorschau`) und vergleicht die Leseordnung
// zweier aufeinanderfolgender Karten desselben Stils.
//
// **Die Leseordnung ist die des Fensters** (`leseordnung` in
// `frontend/src/seiten/angebotenefelder.ts`): Zeilen nach der Mitte des
// ersten Feldes, in der Zeile von links nach rechts, Felder über 90 % der
// Hülle als Grund am Ende.
//
//	MSE_FREMDPAKETE=… MSE_KANON=… go test -tags fremdpakete -run TestWieOftSichDieFeldordnungBeimBlaetternAendert -v ./internal/app/

import (
	"os"
	"slices"
	"sort"
	"testing"
)

// blaetterKarten ist, wie viele Karten je Set geblättert werden.
const blaetterKarten = 120

func TestWieOftSichDieFeldordnungBeimBlaetternAendert(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	var paare, gleicheLage, andereOrdnung, umgestellt int

	for _, satz := range kanonsaetze {
		geoeffnet := anwendung.SetOeffnen(t.Context(), kanonsatzpfad(t, satz))
		if geoeffnet.Fehler != nil {
			t.Errorf("%s öffnen: %+v", satz, geoeffnet.Fehler)

			continue
		}

		karten := anwendung.Karten(t.Context(), geoeffnet.SetID).Karten
		if len(karten) > blaetterKarten {
			karten = karten[:blaetterKarten]
		}

		var (
			vorige                []FeldflaecheDTO
			vorigerStil           string
			satzPaare, satzAnders int
		)

		for _, karte := range karten {
			antwort := anwendung.Vorschau(t.Context(), karte.ID, 1)
			if antwort.Fehler != nil || antwort.Vorschau == nil {
				vorige = nil

				continue
			}

			flaechen := antwort.Vorschau.Flaechen
			stil := eigenerStil(karte)

			if vorige != nil && stil == vorigerStil {
				paare++
				satzPaare++

				if slices.Equal(flaechen, vorige) {
					gleicheLage++
				}

				if !slices.Equal(leseordnungVon(flaechen), leseordnungVon(vorige)) {
					andereOrdnung++
					satzAnders++

					// **Umgestellt oder nur dazu und weg?** Gezählt wird, ob
					// die Felder, die beide Karten zeichnen, anders stehen.
					if !slices.Equal(gemeinsam(leseordnungVon(flaechen), leseordnungVon(vorige)),
						gemeinsam(leseordnungVon(vorige), leseordnungVon(flaechen))) {
						umgestellt++
					}
				}
			}

			vorige, vorigerStil = flaechen, stil
		}

		t.Logf("  %-12s %3d Paare desselben Stils, %3d mit anderer Ordnung", satz, satzPaare, satzAnders)
	}

	t.Logf("zusammen: %d Paare, %d mit gleicher Lage, %d mit anderer Ordnung, "+
		"davon %d mit umgestellten gemeinsamen Feldern", paare, gleicheLage, andereOrdnung, umgestellt)
}

// gemeinsam ist die Ordnung von `a`, beschränkt auf die Felder, die auch `b`
// nennt.
func gemeinsam(a, b []string) []string {
	var aus []string

	for _, feld := range a {
		if slices.Contains(b, feld) {
			aus = append(aus, feld)
		}
	}

	return aus
}

// eigenerStil ist der Wert `stylesheet` der Karte — leer heißt „der des
// Sets", wie in `Karteneditor.tsx`.
func eigenerStil(karte KarteDTO) string {
	for _, wert := range karte.Werte {
		if wert.Feld == "stylesheet" {
			return wert.Wert
		}
	}

	return ""
}

// leseordnungVon ist die Leseordnung des Fensters, in Go nachgebaut.
func leseordnungVon(flaechen []FeldflaecheDTO) []string {
	var breite, hoehe float64
	for _, flaeche := range flaechen {
		breite = max(breite, flaeche.Links+flaeche.Breite)
		hoehe = max(hoehe, flaeche.Oben+flaeche.Hoehe)
	}

	gesehen := map[string]bool{}

	var gesetzt, grund []FeldflaecheDTO

	for _, flaeche := range flaechen {
		if gesehen[flaeche.Feld] {
			continue
		}

		gesehen[flaeche.Feld] = true

		if flaeche.Breite*flaeche.Hoehe >= 0.9*breite*hoehe {
			grund = append(grund, flaeche)

			continue
		}

		gesetzt = append(gesetzt, flaeche)
	}

	sort.SliceStable(gesetzt, func(i, j int) bool { return gesetzt[i].Oben < gesetzt[j].Oben })

	type platz struct {
		feld string
		wert float64
	}

	ordnung := make([]platz, 0, len(gesetzt))
	zeile, mitte := 0.0, -1e18

	for _, flaeche := range gesetzt {
		if flaeche.Oben >= mitte {
			zeile++
			mitte = flaeche.Oben + flaeche.Hoehe/2
		}

		ordnung = append(ordnung, platz{flaeche.Feld, zeile*1e6 + flaeche.Links})
	}

	sort.SliceStable(ordnung, func(i, j int) bool { return ordnung[i].wert < ordnung[j].wert })

	namen := make([]string, 0, len(flaechen))
	for _, eintrag := range ordnung {
		namen = append(namen, eintrag.feld)
	}

	for _, flaeche := range grund {
		namen = append(namen, flaeche.Feld)
	}

	return namen
}
