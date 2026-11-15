package service_test

// **Die Typzeile und der Manawert je Platz** — Werkstück U10 von Phase 5.1
// (#1421).
//
// Ohne sie kann die Auswertung keine Typen zählen, keine Länder erkennen und
// keinen Durchschnitt rechnen. Beides kommt auf **demselben Weg** wie der
// Name — aus derselben Antwort, ohne zweiten Ruf.
//
// **Der Manawert ist nachgelesen, nicht gerechnet** (§15): Scryfall führt ihn
// als `cmc`, und bei einer zweiseitigen Karte sind die Kosten leer, während
// der Wert dasteht. Wer ihn aus den Kosten rechnete, bekäme dort eine Null.

import (
	"testing"
)

func TestEinPlatzTraegtTypzeileUndManawert(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.typzeilen[druckA] = "Creature — Elf Druid"
	werk.drucke.manawerte[druckA] = 2

	werk.legeDruck(t, druckA, 4)

	deck, err := werk.bau.Deck(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(deck.Plaetze) != 1 {
		t.Fatalf("%d Plätze", len(deck.Plaetze))
	}

	platz := deck.Plaetze[0]
	if platz.Typzeile != "Creature — Elf Druid" || platz.Manawert != 2 {
		t.Errorf("%+v", platz)
	}
}

// **Eine eigene Karte holt ihre Typzeile aus dem Bestand.**
//
// **Und das Feld heißt `type`** (`docs/FORMATE.md`): MSE rechnet es aus
// `super type` und `sub type` und schreibt es mit `save value: false` — steht
// es nicht im Bestand, bleibt die Zeile leer, und das ist die Antwort.
func TestEineEigeneKarteTraegtIhreTypzeile(t *testing.T) {
	werk := neuesPruefwerk(t)
	werk.bau.MitKartennamen(&kartennamendoppel{
		namen:  map[int64]string{eigeneKarte: "Elfe"},
		felder: map[string]string{"casting cost": "{1}{G}", "type": "Creature — Elf"},
	})

	werk.legeKarte(t, 1)

	deck, err := werk.bau.Deck(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	platz := deck.Plaetze[0]
	if platz.Typzeile != "Creature — Elf" || platz.Manakosten != "{1}{G}" {
		t.Errorf("%+v", platz)
	}
}

// **Ohne Antwort bleibt beides leer** — und das ist kein Fehler (§16, Regel 5).
func TestOhneAntwortBleibenTypzeileUndManawertLeer(t *testing.T) {
	werk := neuesPruefwerk(t)
	werk.drucke.unbekannt[druckA] = true

	werk.legeDruck(t, druckA, 4)

	deck, err := werk.bau.Deck(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if platz := deck.Plaetze[0]; platz.Typzeile != "" || platz.Manawert != 0 {
		t.Errorf("%+v", platz)
	}
}
