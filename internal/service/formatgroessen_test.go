// **Was die Formatprüfung an Größen rechnet** — Werkstück 5 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, weil `formatpruefung_test.go` über die Zeilengrenze
// aus §2 ging. Der Schnitt folgt der Frage: dort steht, **was eine Karte
// bekommt**, hier, **was über das ganze Deck gerechnet wird** — Standardländer
// und gleiche Namen.
package service_test

import (
	"strings"
	"testing"
)

// **Vierundzwanzig Wald sind kein Verstoß** —
// #1370.
//
// Die Prüfung fragt Scryfall ohnehin nach jedem Druck; die Typzeile kommt
// dabei mit. Ohne sie meldete jedes gewöhnliche Deck einen Verstoß über
// gleiche Karten.
func TestVierundzwanzigWaldSindKeinVerstoss(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckA] = map[string]string{"standard": "legal"}
	werk.drucke.typzeilen[druckA] = "Basic Land — Forest"

	werk.drucke.legalitaeten[druckB] = map[string]string{"standard": "legal"}
	werk.drucke.typzeilen[druckB] = "Instant"

	werk.legeDruck(t, druckA, 24)
	werk.legeDruck(t, druckB, 4)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" {
			t.Errorf("die Wälder wurden beanstandet: %q", befund.Meldung)
		}
	}
}

// **Und fünf gleiche Nicht-Länder bleiben einer.**
func TestFuenfGleicheNichtlaenderBleibenEinVerstoss(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckB] = map[string]string{"standard": "legal"}
	werk.drucke.typzeilen[druckB] = "Instant"

	werk.legeDruck(t, druckB, 5)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	gefunden := false

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" {
			gefunden = true
		}
	}

	if !gefunden {
		t.Error("fünf gleiche Instants kamen durch")
	}
}

// **Eine Karte, deren Typ niemand kennt, wird gezählt.**
//
// Eine eigene Karte ohne Verweis hat keine Typzeile bei Scryfall — und eine
// Ausnahme auf Verdacht ließe ein Deck mit fünf gleichen durchgehen.
func TestEineKarteOhneTypWirdGezaehlt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.legeKarte(t, 5)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	gefunden := false

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" {
			gefunden = true
		}
	}

	if !gefunden {
		t.Error("eine eigene Karte wurde auf Verdacht ausgenommen")
	}
}

// **Vier verschiedene Drucke derselben Karte sind vier Karten** — Befund 1 des
// Zweitblicks über Phase 4.6.
//
// **Der Weg, den der Befund meinte**: der Name steht im Urteil, kam aber nicht
// dorthin, wo gezählt wird.
func TestVierDruckeDerselbenKarteWerdenZusammengezaehlt(t *testing.T) {
	werk := neuesPruefwerk(t)

	// **Fünf Drucke, ein Name.**
	for stelle, druck := range []string{
		"a1b2c3d4-0000-4000-8000-000000000001",
		"a1b2c3d4-0000-4000-8000-000000000002",
		"a1b2c3d4-0000-4000-8000-000000000003",
		"a1b2c3d4-0000-4000-8000-000000000004",
		"a1b2c3d4-0000-4000-8000-000000000005",
	} {
		werk.drucke.legalitaeten[druck] = map[string]string{"standard": "legal"}
		werk.drucke.typzeilen[druck] = "Instant"
		werk.drucke.namen[druck] = "Lightning Bolt"

		werk.legeDruck(t, druck, 1)

		_ = stelle
	}

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	gefunden := ""

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" {
			gefunden = befund.Meldung
		}
	}

	if gefunden == "" {
		t.Fatalf("fünf Drucke derselben Karte kamen durch: %+v", urteil.Groessen)
	}

	if !strings.Contains(gefunden, "Lightning Bolt") {
		t.Errorf("die Meldung nennt die Karte nicht: %q", gefunden)
	}
}
