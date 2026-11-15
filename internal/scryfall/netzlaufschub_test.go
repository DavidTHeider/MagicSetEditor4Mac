//go:build netzlauf || dienstdrift

// **Die Messung zur Bündelung** — Abnahme aus
// #1373:
// „Ein Deck mit 100 verschiedenen Drucken wird bei **kaltem**
// Zwischenspeicher in **unter fünf Sekunden** geprüft — gemessen, nicht
// gerechnet."
//
// # Warum die Messung hier steht und nicht im Dienst
//
// Weil die Zahl, die zählt, an der **Drossel** hängt und nicht am Rechner.
// Ein Test gegen ein Doppel misst, wie schnell eine Map antwortet; ein Test
// gegen `httptest` misst dasselbe über localhost. Die Minute, gegen die
// Befund 6 stand, entstand aus 600 ms echter Wartezeit je Anfrage — und die
// gibt es nur gegen den echten Dienst.
//
// **Er läuft nicht in der Pipeline**, aus denselben drei Gründen wie
// `netzlauf_test.go`. Gefahren wird er so:
//
//	go test -tags netzlauf ./internal/scryfall/ -run TestHundertDruckeInWenigerAlsFuenfSekunden -v
package scryfall_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Hundert Drucke, kalter Speicher, unter fünf Sekunden.**
//
// # Woher die hundert Kennungen kommen
//
// Aus `/cards/search` über ein Set — echte Drucke, deren Kennungen dieser
// Test nicht mitbringen muss. Die Suche ist selbst eine gedrosselte Anfrage
// und geht **nicht** in die gemessene Zeit ein: gemessen wird das, was die
// Formatprüfung tut, wenn die Kennungen schon dastehen.
func TestHundertDruckeInWenigerAlsFuenfSekunden(t *testing.T) {
	klient := echterKlient(t)

	// **Eine Suche gibt eine Seite** — sie lädt nicht von selbst nach, und
	// das ist so gewollt (`Treffer.Weitere`). Eine Seite über ein ganzes Set
	// trägt mehr als hundert Drucke.
	treffer, _, _, fehler := klient.Druckesuche(t.Context(), "set:dmu")
	if fehler != nil {
		t.Fatalf("die Suche nach Kennungen: %v", fehler)
	}

	if len(treffer.Karten) < 100 {
		t.Fatalf("die Suche gab %d Drucke — der Test misst mit hundert",
			len(treffer.Karten))
	}

	kennungen := make([]scryfall.Kennung, 0, 100)

	for _, karte := range treffer.Karten[:100] {
		kennungen = append(kennungen, scryfall.Kennung{ID: karte.Kennung})
	}

	if len(kennungen) != 100 {
		t.Fatalf("%d Kennungen — der Test misst nur mit hundert", len(kennungen))
	}

	// **Ab hier wird gemessen** — zwei Schübe zu 75 und 25.
	begonnen := time.Now()

	var geholt int

	for anfang := 0; anfang < len(kennungen); anfang += scryfall.Sammelgrenze {
		ende := min(anfang+scryfall.Sammelgrenze, len(kennungen))

		antwort, _, _, fehler := klient.Sammlung(t.Context(), kennungen[anfang:ende])
		if fehler != nil {
			t.Fatalf("der Schub ab %d: %v", anfang, fehler)
		}

		geholt += len(antwort.Karten)
	}

	gedauert := time.Since(begonnen)

	t.Logf("%d Drucke in %d Anfragen: %s", geholt,
		(len(kennungen)+scryfall.Sammelgrenze-1)/scryfall.Sammelgrenze, gedauert)

	if geholt != 100 {
		t.Errorf("%d Karten kamen zurück, erwartet 100", geholt)
	}

	// **Fünf Sekunden**, wie das Issue sie nennt. Der Einzelweg brauchte
	// allein an Drosselwartezeit rund eine Minute.
	if gedauert > 5*time.Second {
		t.Errorf("die Prüfung dauerte %s — die Abnahme nennt fünf Sekunden", gedauert)
	}
}
