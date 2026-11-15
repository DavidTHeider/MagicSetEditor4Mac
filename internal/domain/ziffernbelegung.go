// **Welche Seite ⌘0 bis ⌘9 öffnen** — UX-02.AK-03
// (#1674).
//
// # Warum es sie gibt
//
// **Weil die Ziffern gewandert sind.** Mit den Bereichen der Seitenleiste
// (#1753) gehören ⌘1 bis ⌘6 dem Bereich und nicht mehr der Seite; ⌘7 bis ⌘0
// sind frei geworden. Das Kriterium verlangt dafür zweierlei: eine
// dokumentierte Migration — sie steht in `CHANGELOG.md` — **und eine
// konfigurierbare Belegung**. Wer mit ⌘3 zu den Karten will, wie vor 0.6.0,
// stellt es hier ein.
//
// # Was die Domäne prüft und was nicht
//
// **Die Gestalt**: `ziffer=seite`, durch Kommas getrennt, jede Ziffer einmal,
// eine Seite in Kleinbuchstaben — oder leer, dann gilt die Vorgabe. **Und dass
// es die Seite gibt.**
//
// Hier stand bis zum 2026-09-26 das Gegenteil: „Die Domäne kennt keine
// Seiten", das Fenster lasse eine unbekannte wegfallen. Der Zweitblick über
// Phase 6.3 fand, was das kostet (G13): MCP, Kommandozeile und REST haben kein
// Fenster, `3=kartne` kam durch, und das Fenster löschte den Eintrag bei der
// nächsten Änderung still. **Die Namen sind ein Wertevorrat wie jeder andere**
// — was eine Seite zeigt, weiß die Domäne weiterhin nicht. Dass die Liste
// dieselbe ist wie in `navigation.ts`, hält
// `internal/werkzeuge/fensterseiten_test.go`.

package domain

import (
	"regexp"
	"slices"
	"strings"
)

// EinstellungZiffernbelegung ist der Schlüssel der Belegung.
const EinstellungZiffernbelegung = "kuerzel.ziffern"

// ziffernpaar ist ein Eintrag der Belegung: eine Ziffer, ein Gleichheitszeichen
// und eine Seite — **oder keine**: „2=" räumt die Ziffer.
var ziffernpaar = regexp.MustCompile(`^(\d)=([a-z]*)$`)

// fensterseiten sind die Seiten, die eine Ziffer öffnen kann — sortiert.
var fensterseiten = []string{
	"ausgeben",
	"bilder",
	"decks",
	"druckbogen",
	"einstellungen",
	"ersetzen",
	"handeln",
	"hilfe",
	"karten",
	"konsole",
	"nutzer",
	"pack",
	"pakete",
	"sammlung",
	"schluesselwoerter",
	"setangaben",
	"sets",
	"spielen",
	"start",
	"statistik",
	"stil",
	"suche",
	"symbol",
	"turniere",
	"vorlagen",
}

// Fensterseiten gibt die Seiten des Fensters, sortiert — eine Kopie.
func Fensterseiten() []string { return slices.Clone(fensterseiten) }

// PruefeZiffernbelegung urteilt über den Wert der Belegung.
func PruefeZiffernbelegung(wert string) error {
	if wert == "" {
		return nil
	}

	gesehen := map[string]bool{}

	for _, eintrag := range strings.Split(wert, ",") {
		treffer := ziffernpaar.FindStringSubmatch(eintrag)
		if treffer == nil {
			return NewFieldError("wert", SchluesselValidierung,
				"»%s« ist kein Eintrag der Belegung — erwartet wird eine Ziffer, ein Gleichheitszeichen und eine Seite, etwa 3=karten.",
				eintrag)
		}

		if treffer[2] != "" && !slices.Contains(fensterseiten, treffer[2]) {
			return NewFieldError("wert", SchluesselValidierung,
				"»%s« ist keine Seite dieser Anwendung.", treffer[2])
		}

		if gesehen[treffer[1]] {
			return NewFieldError("wert", SchluesselValidierung,
				"Die Ziffer %s ist zweimal belegt.", treffer[1])
		}

		gesehen[treffer[1]] = true
	}

	return nil
}
