package service_test

// **Die Typzeile einer eigenen Karte kommt aus dem Bestand** —
// #1622.
//
// # Der Befund, für den es diese Datei gibt
//
// **Gefunden am 2026-09-17 bei der ersten Sichtprüfung des Spielkerns mit
// einem echten Deck**: zwanzig `Mountain` aus dem kanonischen Set wurden
// gegen die Vierergrenze gezählt. CR 100.2a nimmt Standardländer aus, und
// dieses Programm weiß das — es fragte die Typzeile nur bei **Scryfall**,
// und eine eigene Karte ohne Druckverweis hat dort keine.
//
// **Es ist der Normalfall dieses Programms.** Ein Karteneditor, in dem
// jemand eigene Karten baut, prüft Decks aus eigenen Karten — und §16,
// Regel 5 sagt: ohne Netz arbeitet die Anwendung weiter.

import (
	"strings"
	"testing"
)

// **Zwanzig eigene Gebirge sind kein Verstoß** — CR 100.2a mit CR 205.4a.
func TestZwanzigEigeneGebirgeSindKeinVerstoss(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.karten.felder["super type"] = "Basic Land"
	werk.karten.felder["sub type"] = "Mountain"

	werk.legeKarte(t, 20)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" {
			t.Errorf("die Gebirge wurden beanstandet: %q", befund.Meldung)
		}
	}
}

// **Und zwanzig Dryad Arbor sind einer** — CR 205.4c: der Untertyp `Forest`
// macht aus einer Kreatur kein Standardland, weil der **Supertyp** fehlt.
//
// **Die Gegenprobe zur Zeile darüber.** Ohne sie sagte der Lauf nur, dass
// irgendetwas durchgeht.
func TestZwanzigEigeneDryadArborSindEinVerstoss(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.karten.felder["super type"] = "Legendary Creature Land"
	werk.karten.felder["sub type"] = "Dryad Forest"

	werk.legeKarte(t, 20)

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
		t.Error("zwanzig gleiche Nichtländer gingen durch")
	}
}

// **Auch ohne Scryfallnamen wird die Höchstzahl geprüft.**
//
// **Die Vierergrenze ist eine Regel dieses Programms** und keine Auskunft
// von Scryfall — ein Format ohne Brücke dorthin hat sie trotzdem.
func TestOhneScryfallnamenGiltDieHoechstzahlWeiter(t *testing.T) {
	werk := neuesPruefwerk(t)
	werk.formatOhneScryfallnamen(t)

	werk.karten.felder["super type"] = "Basic Land"
	werk.karten.felder["sub type"] = "Mountain"

	werk.legeKarte(t, 20)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" {
			t.Errorf("die Gebirge wurden beanstandet: %q", befund.Meldung)
		}
	}
}

// **Eine leere Typzeile bleibt ungeprüft** — §10: eine Ausnahme auf Verdacht
// ist keine.
//
// **Nicht jedes Spiel führt einen Supertyp**, und ein leeres Feld ist keine
// Auskunft. Wer daraus „Standardland" machte, ließe jede Karte ohne Typzeile
// in beliebiger Zahl durch.
func TestEineEigeneKarteOhneTypzeileWirdGezaehlt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.legeKarte(t, 20)

	urteil, _, err := werk.pruefung.Pruefe(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("prüfen: %v", err)
	}

	gefunden := false

	for _, befund := range urteil.Groessen {
		if befund.Feld == "gleiche" &&
			strings.Contains(befund.Meldung, "20") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("zwanzig namenlose Karten gingen durch: %+v", urteil.Groessen)
	}
}

// **Und was Scryfall sagt, gilt weiter** — der Bestand tritt nur ein, wo
// dort nichts steht.
//
// **Das ist die Reihenfolge und nicht die Vorliebe**: ein echter Druck hat
// eine gedruckte Typzeile, und die ist die Auskunft. Der Bestand kennt für
// ihn ohnehin keine.
func TestDerBestandTrittNurEinWoScryfallSchweigt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.drucke.legalitaeten[druckB] = map[string]string{"standard": "legal"}
	werk.drucke.typzeilen[druckB] = "Instant"

	// **Der Bestand sagt „Standardland", und er wird nicht gefragt** — der
	// Platz nennt einen Druck, und für den antwortet Scryfall.
	werk.karten.felder["super type"] = "Basic Land"
	werk.karten.felder["sub type"] = "Mountain"

	werk.legeDruck(t, druckB, 20)

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
		t.Error("zwanzig gleiche Spontanzauber gingen durch")
	}
}
