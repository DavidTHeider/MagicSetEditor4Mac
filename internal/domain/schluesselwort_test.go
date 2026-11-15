package domain_test

// **Was eine Karte an Schlüsselwörtern trägt** — Werkstück M14 von Phase 5.5
// (#1616).
//
// # Warum die Zeile zählt und nicht das Wort
//
// **Weil „kommt vor" nicht „hat" heißt.** „Kann nicht von Kreaturen ohne
// Flugfähigkeit geblockt werden" nennt Flugfähigkeit und gibt sie niemandem.
// `M6` las das Wort als Teilzeichenkette und führte den Preis als Grenze mit;
// **gemessen am 2026-09-18** über die zehn kanonischen Sets kostet das bei
// `flying` 3.349 gegen 2.315 Karten — 45 Prozent zu viel
// (`docs/SCHLUESSELWOERTER.md`).
//
// **CR 702.1** sagt, wie es dasteht: ein Objekt nennt nur den **Namen** der
// Fähigkeit. Sie steht deshalb auf
// einer eigenen Zeile, mehrere durch Kommas getrennt. **CR 207.2a** erlaubt
// dahinter Erinnerungstext in Klammern.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Zeile aus Schlüsselwörtern gibt sie alle** — CR 702.1.
func TestEineZeileAusSchluesselwoerternGibtSieAlle(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		text     string
		hat      []string
		hatNicht []string
	}{
		{
			name: "eines allein",
			text: "Flying",
			hat:  []string{domain.SchluesselFlug},
		},
		{
			name: "mehrere durch Kommas",
			text: "Flying, first strike, lifelink",
			hat: []string{
				domain.SchluesselFlug, domain.SchluesselErstschlag,
				domain.SchluesselLebensverknuepfung,
			},
		},
		{
			name: "mit Erinnerungstext",
			text: "Deathtouch (Any amount of damage this deals to a " +
				"creature is enough to destroy it.)",
			hat: []string{domain.SchluesselTodesberuehrung},
		},
		{
			// **Der Fall, um dessentwillen es diese Lesart gibt.**
			name:     "ein Satz nennt es und gibt es nicht",
			text:     "This creature can't be blocked except by creatures with flying.",
			hatNicht: []string{domain.SchluesselFlug},
		},
		{
			name: "eine zweite Zeile zählt mit",
			text: "Vigilance\nWhenever this creature attacks, draw a card.",
			hat:  []string{domain.SchluesselWachsamkeit},
		},
		{
			// **Auszeichnungen des Pakets fallen weg** (§8): MSE schreibt
			// `<i-flavor>` und `<soft-line>` in denselben Text.
			name: "mit Auszeichnung",
			text: "<b>Trample</b>",
			hat:  []string{domain.SchluesselTrampelschaden},
		},
		{
			name:     "eine Aufzählung mit einem Fremdwort gibt nichts",
			text:     "Flying, buzzes a lot",
			hatNicht: []string{domain.SchluesselFlug},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			woerter := domain.SchluesselwoerterAus(fall.text)

			for _, wort := range fall.hat {
				if !woerter[wort] {
					t.Errorf("%q fehlt in %v", wort, woerter)
				}
			}

			for _, wort := range fall.hatNicht {
				if woerter[wort] {
					t.Errorf("%q steht da und sollte nicht", wort)
				}
			}
		})
	}
}

// **Die Liste ist die des Schnitts** — und sie ist gezählt, nicht gefühlt.
func TestDerSchnittFuehrtZwoelfSchluesselwoerter(t *testing.T) {
	t.Parallel()

	if anzahl := len(domain.Schluesselwoerter()); anzahl != 12 {
		t.Errorf("%d Schlüsselwörter: %v",
			anzahl, domain.Schluesselwoerter())
	}
}
