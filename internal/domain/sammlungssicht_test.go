package domain_test

// **Die Sicht auf eine Sammlung** — Befund aus U8, Werkstück
// #1500.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// exemplar baut eines mit lesbarem Bezug.
func exemplar(name, setcode string, menge int, foil bool) *domain.Exemplar {
	return &domain.Exemplar{
		Printing: "scryfall:" + name, Name: name, Setcode: setcode,
		Menge: menge, Finish: domain.FinishAus(foil),
	}
}

// **Die Ordnung folgt der Sprache und nicht den Bytes.**
//
// **Das ist der Grund, dass sie in Go bleibt**: SQLites `COLLATE NOCASE` ist
// ASCII, und „Ätherfluss" stünde damit hinter „Zauber".
func TestDieOrdnungFolgtDerSprache(t *testing.T) {
	t.Parallel()

	sicht := domain.Sammlungssicht{Sortierung: domain.SortierungBezeichnung}

	gesehen := sicht.Gesehen([]*domain.Exemplar{
		exemplar("Zauber", "m11", 1, false),
		exemplar("Ätherfluss", "m11", 1, false),
		exemplar("Bolt", "m11", 1, false),
	})

	namen := make([]string, 0, len(gesehen))
	for _, eines := range gesehen {
		namen = append(namen, eines.Name)
	}

	if namen[0] != "Ätherfluss" || namen[1] != "Bolt" || namen[2] != "Zauber" {
		t.Errorf("die Ordnung ist %v", namen)
	}
}

// **Die Suche geht über vier Felder.**
func TestDieSucheGehtUeberVierFelder(t *testing.T) {
	t.Parallel()

	mitNotiz := exemplar("Bolt", "m11", 1, false)
	mitNotiz.Notiz = "getauscht mit Anna"

	alle := []*domain.Exemplar{
		mitNotiz,
		exemplar("Dragon", "war", 1, false),
	}

	for gesucht, erwartet := range map[string]int{
		"bolt":     1,
		"BOLT":     1,
		"war":      1,
		"getausch": 1,
		"scryfall": 2,
		"nichts":   0,
		"":         2,
	} {
		t.Run(gesucht, func(t *testing.T) {
			t.Parallel()

			sicht := domain.Sammlungssicht{Suche: gesucht}

			if gefunden := len(sicht.Gesehen(alle)); gefunden != erwartet {
				t.Errorf("%q findet %d, erwartet waren %d", gesucht, gefunden, erwartet)
			}
		})
	}
}

// **Die Ausführung filtert Foil gegen Nonfoil.**
func TestDieAusfuehrungFiltert(t *testing.T) {
	t.Parallel()

	alle := []*domain.Exemplar{
		exemplar("Bolt", "m11", 1, true),
		exemplar("Dragon", "war", 1, false),
	}

	for ausfuehrung, erwartet := range map[domain.Ausfuehrung]int{
		"":                        2,
		domain.AusfuehrungAlle:    2,
		domain.AusfuehrungFoil:    1,
		domain.AusfuehrungNonfoil: 1,
	} {
		sicht := domain.Sammlungssicht{Ausfuehrung: ausfuehrung}

		if gefunden := len(sicht.Gesehen(alle)); gefunden != erwartet {
			t.Errorf("%q findet %d, erwartet waren %d", ausfuehrung, gefunden, erwartet)
		}
	}
}

// **Nach Menge absteigend, nach Datum das Jüngste zuerst** — und die Karte ist
// immer der zweite Schlüssel.
func TestDieBeidenAnderenOrdnungen(t *testing.T) {
	t.Parallel()

	wenig := exemplar("Alpha", "m11", 1, false)
	viel := exemplar("Zeta", "m11", 9, false)
	gleichviel := exemplar("Beta", "m11", 9, false)

	alt := exemplar("Alt", "m11", 1, false)
	alt.Erwerbsdatum = "2020-01-01"
	neu := exemplar("Neu", "m11", 1, false)
	neu.Erwerbsdatum = "2026-01-01"
	ohne := exemplar("Ohne", "m11", 1, false)

	nachMenge := domain.Sammlungssicht{Sortierung: domain.SortierungMenge}.
		Gesehen([]*domain.Exemplar{wenig, viel, gleichviel})

	if nachMenge[0].Name != "Beta" || nachMenge[1].Name != "Zeta" ||
		nachMenge[2].Name != "Alpha" {
		t.Errorf("nach Menge: %v", namenVon(nachMenge))
	}

	nachDatum := domain.Sammlungssicht{Sortierung: domain.SortierungErwerbsdatum}.
		Gesehen([]*domain.Exemplar{alt, ohne, neu})

	// **Ohne Datum ans Ende** — nicht nach vorn, weil die leere Zeichenkette
	// klein ist.
	if nachDatum[0].Name != "Neu" || nachDatum[1].Name != "Alt" ||
		nachDatum[2].Name != "Ohne" {
		t.Errorf("nach Datum: %v", namenVon(nachDatum))
	}
}

func namenVon(exemplare []*domain.Exemplar) []string {
	namen := make([]string, 0, len(exemplare))
	for _, eines := range exemplare {
		namen = append(namen, eines.Name)
	}

	return namen
}

// **Eine erfundene Sicht wird abgewiesen** — an beiden Feldern.
func TestEineErfundeneSichtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Sammlungssicht{}).Pruefe(); fehler != nil {
		t.Errorf("die leere Sicht: %v", fehler)
	}

	if fehler := (domain.Sammlungssicht{
		Ausfuehrung: domain.AusfuehrungFoil, Sortierung: domain.SortierungMenge,
	}).Pruefe(); fehler != nil {
		t.Errorf("eine gültige Sicht: %v", fehler)
	}

	if fehler := (domain.Sammlungssicht{Ausfuehrung: "glitzernd"}).Pruefe(); fehler == nil {
		t.Error("eine erfundene Ausführung kam durch")
	}

	if fehler := (domain.Sammlungssicht{Sortierung: "nach Farbe"}).Pruefe(); fehler == nil {
		t.Error("eine erfundene Ordnung kam durch")
	}
}

// **Der Ausschnitt gibt das Fenster und die Gesamtzahl.**
func TestDerAusschnittGibtFensterUndGesamtzahl(t *testing.T) {
	t.Parallel()

	alle := make([]*domain.Exemplar, 0, 10)
	for nummer := range 10 {
		alle = append(alle, exemplar(string(rune('A'+nummer)), "m11", 1, false))
	}

	for name, fall := range map[string]struct {
		ab, anzahl int
		erwartet   int
		ersterName string
	}{
		"die ersten drei": {ab: 0, anzahl: 3, erwartet: 3, ersterName: "A"},
		"aus der Mitte":   {ab: 4, anzahl: 3, erwartet: 3, ersterName: "E"},
		"über das Ende":   {ab: 8, anzahl: 5, erwartet: 2, ersterName: "I"},
		"hinter dem Ende": {ab: 20, anzahl: 5, erwartet: 0},
		"alles":           {ab: 0, anzahl: 0, erwartet: 10, ersterName: "A"},
		"ab null":         {ab: -3, anzahl: 2, erwartet: 2, ersterName: "A"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fenster, gesamt := domain.Ausschnitt(alle, fall.ab, fall.anzahl)

			if gesamt != 10 {
				t.Errorf("die Gesamtzahl ist %d", gesamt)
			}

			if len(fenster) != fall.erwartet {
				t.Fatalf("das Fenster hat %d Zeilen, erwartet waren %d",
					len(fenster), fall.erwartet)
			}

			if fall.ersterName != "" && fenster[0].Name != fall.ersterName {
				t.Errorf("die erste Zeile ist %q", fenster[0].Name)
			}
		})
	}
}
