// **Ein Preis ist eine Beobachtung** — Werkstück 3 von Phase 3.4
// (#1165).
package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der rohe Text kommt in den Lauf, nicht die geparste Zahl** (§3).
//
// Scryfall schickt `"4.17"` — eine **Zeichenkette**, nachgelesen am
// 2026-08-27: „…prices, as strings". Wer im Lauf schon `4.17` einsetzt, prüft
// die Umrechnung nicht, die im Betrieb stattfindet.
func TestEinPreisWirdAusSeinemTextGelesen(t *testing.T) {
	t.Parallel()

	for text, erwartet := range map[string]int64{
		"4.17":     417,
		"0.02":     2,
		"10":       1000,
		"1234.50":  123450,
		"0":        0,
		"0.00":     0,
		"  4.17  ": 417,
		"4.1":      410,
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			gefunden, err := domain.MarktpreisAus(text, "USD")
			if err != nil {
				t.Fatalf("MarktpreisAus(%q): %v", text, err)
			}

			if gefunden != erwartet {
				t.Errorf("%q ergab %d statt %d", text, gefunden, erwartet)
			}
		})
	}
}

// **Was kein Preis ist, wird abgewiesen — nicht gerundet.**
//
// **Die Fälle gelten weiter, und einer davon anders** (LOC-03.AK-02,
// #1689): »4.175« ist bei einer **zweistelligen** Währung immer noch ein
// Fehler — bei einer dreistelligen wäre es der letzte Fils. Welche gilt,
// sagt jetzt die Währung und nicht eine feste Zwei; der Lauf daneben zeigt
// beides.
//
// # Warum drei Nachkommastellen ein Fehler sind und keine Rundung
//
// Wer 4,175 auf 418 rundet, hat einen Betrag erfunden. Bei einer Sammlung mit
// 300 Zeilen summiert sich das, und niemand sieht, woher die Abweichung kommt.
// Ein Fehler sagt stattdessen, dass die Quelle etwas anderes schickt als
// erwartet — und genau das ist die Drifterkennung aus §16.
func TestWasKeinPreisIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, text := range map[string]string{
		"leer":                       "",
		"nur Leerraum":               "   ",
		"Buchstaben":                 "vier Euro",
		"negativ":                    "-1.00",
		"drei Nachkommastellen":      "4.175",
		"Komma statt Punkt":          "4,17",
		"zwei Punkte":                "4.1.7",
		"Währungszeichen dabei":      "€4.17",
		"Punkt ohne Nachkomma":       "4.",
		"Punkt ohne Vorkommateil":    ".17",
		"Buchstabe hinter dem Punkt": "4.x",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := domain.MarktpreisAus(text, "USD"); err == nil {
				t.Errorf("%q kam durch", text)
			}
		})
	}
}

// **Ein Preisstand nennt Quelle, Währung, Ausführung und Zeitpunkt.**
func TestEinGueltigerPreisstandKommtDurch(t *testing.T) {
	t.Parallel()

	stand := gueterPreisstand()

	if befunde := stand.Validate(); len(befunde) != 0 {
		t.Errorf("es gab Befunde: %+v", befunde)
	}
}

// gueterPreisstand gibt einen gültigen Preisstand.
func gueterPreisstand() domain.Preisstand {
	return domain.Preisstand{
		Printing: "8a1b0d5e-1e0f-4c0f-8b39-3d3a2f9a1c2d",
		Quelle:   domain.PreisquelleScryfall,

		// **Markt, Art und Lage gehören seit A3 dazu** (#1442): eine Reihe
		// ohne sie sagt nicht, **welche** Frage sie beantwortet.
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Lage:     domain.PreisGenannt,
		Waehrung: "EUR",
		Finish:   domain.FinishNonfoil,
		Wert:     417,
		Erhoben:  time.Date(2026, time.August, 27, 10, 0, 0, 0, time.UTC),
	}
}

// **Was fehlen kann, wird benannt.**
func TestPreisstandValidateBefunde(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		aendern func(*domain.Preisstand)
		feld    string
	}{
		"ohne Druck":        {func(p *domain.Preisstand) { p.Printing = "" }, "printing"},
		"ohne Quelle":       {func(p *domain.Preisstand) { p.Quelle = "" }, "quelle"},
		"ohne Währung":      {func(p *domain.Preisstand) { p.Waehrung = "" }, "waehrung"},
		"Ausführung falsch": {func(p *domain.Preisstand) { p.Finish = "glänzend" }, "finish"},
		"negativ":           {func(p *domain.Preisstand) { p.Wert = -1 }, "wert"},
		"ohne Zeitpunkt":    {func(p *domain.Preisstand) { p.Erhoben = time.Time{} }, "erhoben"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stand := gueterPreisstand()
			fall.aendern(&stand)

			befunde := stand.Validate()
			if len(befunde) != 1 {
				t.Fatalf("es gab %d Befunde: %+v", len(befunde), befunde)
			}

			if befunde[0].Field != fall.feld {
				t.Errorf("der Befund hängt an %q statt an %q",
					befunde[0].Field, fall.feld)
			}

			if !errors.Is(befunde[0], domain.ErrValidation) {
				t.Errorf("der Befund ist kein ErrValidation: %v", befunde[0])
			}
		})
	}
}

// **Ein Preis sagt, wie alt er ist** (§16, Satz 5).
//
// # Warum das eine Rechnung der Domäne ist
//
// „Wie alt" ist die Frage, an der eine Oberfläche entscheidet, ob sie warnt.
// Sie zweimal zu rechnen — im Fenster und in der Kommandozeile — hieße, zwei
// Antworten auf dieselbe Frage zu haben.
func TestEinPreisstandSagtWieAltErIst(t *testing.T) {
	t.Parallel()

	stand := gueterPreisstand()

	jetzt := stand.Erhoben.Add(50 * time.Hour)

	if alter := stand.Alter(jetzt); alter != 50*time.Hour {
		t.Errorf("das Alter ist %v", alter)
	}
}

// **Ein Zeitpunkt in der Zukunft ergibt kein negatives Alter.**
//
// Die Uhr eines Rechners kann falsch gehen, und eine Angabe „vor −3 Stunden"
// hilft niemandem. Sie ist dann eben frisch.
func TestEinPreisstandAusDerZukunftIstFrisch(t *testing.T) {
	t.Parallel()

	stand := gueterPreisstand()

	if alter := stand.Alter(stand.Erhoben.Add(-time.Hour)); alter != 0 {
		t.Errorf("das Alter ist %v", alter)
	}
}

// **Und die Stellen kommen aus der Währung** — LOC-03.AK-02
// (#1689).
//
// **Der Unterschied zu vorher**: `PreisAusText` las **immer** zwei
// Nachkommastellen. Für den Yen erfand er damit eine Genauigkeit, die es
// nicht gibt, und für den Dinar wies er die dritte Stelle als Drift ab,
// obwohl sie dort die letzte ist.
func TestDerMarktpreisFolgtDenStellenSeinerWaehrung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		text      string
		waehrung  string
		betrag    int64
		gehtDurch bool
	}{
		"Dollar mit zwei Stellen": {"4.17", "USD", 417, true},
		"Dollar mit drei Stellen": {"4.175", "USD", 0, false},
		"Yen ohne Stellen":        {"150", "JPY", 150, true},
		"Yen mit einer Stelle":    {"150.5", "JPY", 0, false},
		"Dinar mit drei Stellen":  {"1.234", "BHD", 1234, true},
		"Dinar mit vier Stellen":  {"1.2345", "BHD", 0, false},
		"negativer Marktpreis":    {"-1.00", "USD", 0, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			betrag, fehler := domain.MarktpreisAus(fall.text, fall.waehrung)

			if fall.gehtDurch {
				if fehler != nil {
					t.Fatalf("%q in %s: %v", fall.text, fall.waehrung, fehler)
				}

				if betrag != fall.betrag {
					t.Errorf("gelesen wurde %d statt %d", betrag, fall.betrag)
				}

				return
			}

			if fehler == nil {
				t.Errorf("%q in %s kam durch", fall.text, fall.waehrung)
			}
		})
	}
}
