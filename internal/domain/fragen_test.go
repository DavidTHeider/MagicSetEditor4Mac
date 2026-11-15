package domain_test

// **Die Wahl je Frage** — Werkstück 8 von Phase 4.8
// (#1321),
// nach [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md),
// Entscheidung 1.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Suche hat eine Wahl, und sie steht in der Einstellungsreihe.**
//
// Sie wird erst jetzt gebaut: vorher gab es **einen** Dienst, der sie
// beantworten konnte, und eine Wahl mit einer Möglichkeit ist keine.
func TestDieSucheHatEineWahl(t *testing.T) {
	t.Parallel()

	schluessel := domain.FrageEinstellung(domain.FrageSuche)

	if !slices.Contains(domain.Einstellungsreihe(), schluessel) {
		t.Fatalf("%q steht nicht in der Einstellungsreihe", schluessel)
	}

	// **Scryfall ist die Vorgabe** — es ist der Dienst, der heute sucht.
	if vorgabe := domain.EinstellungsVorgabe(schluessel); vorgabe != string(domain.DienstScryfall) {
		t.Errorf("die Vorgabe ist %q, erwartet scryfall", vorgabe)
	}
}

// **Zur Wahl stehen nur die Dienste, die suchen können.**
//
// **Alle vier können es** (0063, die Tabelle): Scryfall, magicthegathering.io,
// MTGJSON und Scrydex. Wer einen fünften aufnimmt, der es nicht kann, trägt
// ihn nicht ein — und die Liste hier ist dann kürzer als [domain.Dienste].
func TestZurWahlStehenNurDieDieSuchenKoennen(t *testing.T) {
	t.Parallel()

	werte := domain.ErlaubteEinstellungswerte(
		domain.FrageEinstellung(domain.FrageSuche))

	erwartet := []string{"scryfall", "mtgapi", "mtgjson", "scrydex"}

	if !slices.Equal(werte, erwartet) {
		t.Errorf("zur Wahl stehen %v, erwartet %v", werte, erwartet)
	}
}

// **Jede Frage, die eine Wahl hat, hat auch eine Zeile.**
//
// Wer eine zweite Frage aufnimmt — Bilder, Preise —, bekommt Schalter,
// Vorgabe und Prüfung daraus. Eine zweite Liste daneben wäre die Falle, die
// bei den Einstellungen schon zweimal zugeschnappt ist.
func TestJedeFrageHatIhreZeile(t *testing.T) {
	t.Parallel()

	if len(domain.Fragen()) == 0 {
		t.Fatal("es gibt keine Frage")
	}

	for _, frage := range domain.Fragen() {
		schluessel := domain.FrageEinstellung(frage)

		if !slices.Contains(domain.Einstellungsreihe(), schluessel) {
			t.Errorf("%q hat keine Zeile in der Einstellungsreihe", schluessel)
		}

		if len(domain.ErlaubteEinstellungswerte(schluessel)) == 0 {
			t.Errorf("%q hat keine Werte zur Wahl", schluessel)
		}
	}
}

// **Und wer eine Frage stellt, erfährt, welche Dienste sie beantworten.**
func TestEineFrageNenntIhreDienste(t *testing.T) {
	t.Parallel()

	if dienste := domain.DiensteFuer(domain.FrageSuche); len(dienste) != 4 {
		t.Errorf("die Suche kennt %d Dienste, erwartet 4", len(dienste))
	}

	// **Eine unbekannte Frage kennt keine** — und stürzt nicht ab.
	if dienste := domain.DiensteFuer(domain.Frage("gibtsnicht")); len(dienste) != 0 {
		t.Errorf("eine unbekannte Frage nennt %v", dienste)
	}
}
