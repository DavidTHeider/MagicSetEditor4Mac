package storage_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Kennzahlen kommen aus Abfragen, nicht aus einer Schleife.**
//
// Der Beleg dafür steht nicht in einer Zusicherung, sondern in der Form: der
// Leser gibt sie fertig zurück und lädt keine Karte dafür. Was hier geprüft
// wird, sind die Zahlen.
func TestKennzahlenEinesSets(t *testing.T) {
	leser, _ := suchbestand(t)

	zahlen, err := leser.Kennzahlen(probekontext(t), 1)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if zahlen.Cards != 2 {
		t.Errorf("%d Karten", zahlen.Cards)
	}

	// Drei verschiedene Felder: name, typ, text.
	if zahlen.Fields != 3 {
		t.Errorf("%d Felder", zahlen.Fields)
	}

	if zahlen.Empty != 0 {
		t.Errorf("%d leere Karten", zahlen.Empty)
	}
}

// **Die Belegung sagt, wie vollständig ein Set ist.** „text" hat nur eine der
// beiden Karten — genau das ist die Auskunft, für die jemand herkommt.
func TestDieBelegungJeFeld(t *testing.T) {
	leser, _ := suchbestand(t)

	belegung, err := leser.Belegung(probekontext(t), 1)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	erwartet := []domain.FieldCount{
		{Field: "name", Cards: 2},
		{Field: "typ", Cards: 2},
		{Field: "text", Cards: 1},
	}

	if len(belegung) != len(erwartet) {
		t.Fatalf("%+v", belegung)
	}

	for i, feld := range erwartet {
		if belegung[i] != feld {
			t.Errorf("%d: %+v, erwartet war %+v", i, belegung[i], feld)
		}
	}
}

// Ein leerer Wert zählt nicht als belegt — sonst sähe ein Set vollständig aus,
// dessen Felder alle leer sind.
func TestEinLeererWertZaehltNichtAlsBelegt(t *testing.T) {
	db := frischeDatenbank(t)

	set := &domain.Set{Name: "s", Game: "g", Cards: []*domain.Card{
		{Values: []domain.Value{
			{Field: "name", Text: "Da"},
			{Field: "typ", Text: "   "},
		}},
		{Values: []domain.Value{{Field: "name", Text: ""}}},
	}}

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), set, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	belegung, err := leser.Belegung(probekontext(t), 1)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(belegung) != 1 || belegung[0] != (domain.FieldCount{Field: "name", Cards: 1}) {
		t.Errorf("%+v", belegung)
	}

	zahlen, err := leser.Kennzahlen(probekontext(t), 1)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// Die zweite Karte trägt keinen einzigen nichtleeren Wert.
	if zahlen.Empty != 1 {
		t.Errorf("%d leere Karten", zahlen.Empty)
	}
}

// **Die Verteilung eines Felds** — wie viele Karten welchen Wert tragen.
func TestVerteilungEinesFelds(t *testing.T) {
	leser, _ := suchbestand(t)

	verteilung, err := leser.Verteilung(probekontext(t), 1, "typ")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(verteilung) != 1 || verteilung[0].Value != "Zauber" ||
		verteilung[0].Cards != 2 {
		t.Errorf("%+v", verteilung)
	}

	// **Und jeder Wert nennt seine Karten** (UXR-F6, #2013): die zwei Karten
	// des Sets, aufsteigend.
	if !slices.Equal(verteilung[0].CardIDs, []int64{1, 2}) {
		t.Errorf("die Karten lauten %v", verteilung[0].CardIDs)
	}

	// Ein Feld, das es nicht gibt, ist eine leere Verteilung — kein Fehler.
	leer, err := leser.Verteilung(probekontext(t), 1, "gibtsnicht")
	if err != nil || len(leer) != 0 {
		t.Errorf("%+v / %v", leer, err)
	}
}

// **Die Karten eines Werts sind genau die, die ihn tragen** — UXR-F6
// (#2013).
//
// Mehrere Werte in einem Feld, ein leerer dazwischen, und die Kennungen nicht
// in der Reihenfolge, in der die Karten angelegt wurden: die Liste unter dem
// Balken zeigt genau diese Karten, und eine, die fehlt oder zu viel ist, wäre
// ein Balken, der etwas anderes sagt als seine Zahl.
func TestDieVerteilungNenntDieKartenJeWert(t *testing.T) {
	db := frischeDatenbank(t)

	set := &domain.Set{Name: "s", Game: "g", Cards: []*domain.Card{
		{Values: []domain.Value{{Field: "typ", Text: "Kreatur"}}},
		{Values: []domain.Value{{Field: "typ", Text: "Zauber"}}},
		{Values: []domain.Value{{Field: "typ", Text: "  "}}},
		{Values: []domain.Value{{Field: "typ", Text: "Kreatur"}}},
		{Values: []domain.Value{{Field: "name", Text: "ohne Typ"}}},
		{Values: []domain.Value{{Field: "typ", Text: "Kreatur"}}},
	}}

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), set, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	verteilung, err := storage.NeuerSetLeser(db).Verteilung(probekontext(t), 1, "typ")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// Kennungen vergibt der Bestand fortlaufend: 1, 2, 3, … in Kartenfolge.
	if len(verteilung) != 2 {
		t.Fatalf("%+v", verteilung)
	}

	if verteilung[0].Value != "Kreatur" || !slices.Equal(verteilung[0].CardIDs, []int64{1, 4, 6}) {
		t.Errorf("Kreatur: %+v", verteilung[0])
	}

	if verteilung[1].Value != "Zauber" || !slices.Equal(verteilung[1].CardIDs, []int64{2}) {
		t.Errorf("Zauber: %+v", verteilung[1])
	}
}

// Ein Set ohne Karten ist kein Fehler, sondern ein Set ohne Karten.
func TestKennzahlenEinesLeerenSets(t *testing.T) {
	db := frischeDatenbank(t)

	if _, err := storage.NeuerSetSpeicher(db).
		SetAnlegen(probekontext(t), &domain.Set{Name: "s", Game: "g"}, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	zahlen, err := leser.Kennzahlen(probekontext(t), 1)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if zahlen.Cards != 0 || zahlen.Fields != 0 {
		t.Errorf("%+v", zahlen)
	}

	belegung, err := leser.Belegung(probekontext(t), 1)
	if err != nil || len(belegung) != 0 {
		t.Errorf("%+v / %v", belegung, err)
	}
}

func TestDieStatistikMeldetEineGeschlosseneDatenbank(t *testing.T) {
	_, db := suchbestand(t)

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	if _, err := leser.Kennzahlen(probekontext(t), 1); err == nil {
		t.Error("Kennzahlen kam ohne Fehler zurück")
	}

	if _, err := leser.Belegung(probekontext(t), 1); err == nil {
		t.Error("Belegung kam ohne Fehler zurück")
	}

	if _, err := leser.Verteilung(probekontext(t), 1, "typ"); err == nil {
		t.Error("Verteilung kam ohne Fehler zurück")
	}
}
