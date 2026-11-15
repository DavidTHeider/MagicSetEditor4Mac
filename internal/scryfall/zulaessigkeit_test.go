package scryfall_test

// **Die Zulässigkeiten einer Karte** — Werkstück 5 von Phase 4.6
// (#1293).
//
// Sie sind die **einzige** Quelle für „ist dieser Druck in Format X zulässig?"
// (Vermerk 0061). Was nicht von hier kommt, bleibt `ungeprüft`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Die vier Wörter kommen durch, wie sie dastehen.**
//
// **Nicht übersetzt.** Es sind Scryfalls eigene Wörter, und eine Übersetzung
// hier machte aus einer Auskunft eine Auslegung.
func TestDieZulaessigkeitenWerdenGelesen(t *testing.T) {
	t.Parallel()

	karte, befunde, fehler := scryfall.LiesKarte(beispiel(t, "zulaessigkeit.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %v", befunde)
	}

	for format, erwartet := range map[string]string{
		"standard": "not_legal",
		"modern":   "legal",
		"vintage":  "restricted",
		"legacy":   "banned",
	} {
		if karte.Legalitaeten[format] != erwartet {
			t.Errorf("%s: %q, erwartet %q", format, karte.Legalitaeten[format], erwartet)
		}
	}
}

// **Eine Karte ohne `legalities` hat keine — und das ist kein Fehler.**
//
// Scryfall liefert sie an jeder Karte; ein Beispiel ohne sie ist trotzdem
// lesbar, und ein `nil` dort ist die ehrliche Antwort: **nichts bekannt**.
func TestEineKarteOhneZulaessigkeiten(t *testing.T) {
	t.Parallel()

	karte, _, fehler := scryfall.LiesKarte(beispiel(t, "einfach.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(karte.Legalitaeten) != 0 {
		t.Errorf("es kamen Zulässigkeiten aus dem Nichts: %v", karte.Legalitaeten)
	}
}

// **Ein Wort, das dieses Programm nicht kennt, ist ein Befund** — und bricht
// nicht.
//
// Dieselbe Haltung wie bei `layout` und `rarity`: ein unbekannter
// Aufzählungswert ist eine Bedeutung, die hier fehlt. Wer ihn verschluckt,
// nennt ein Deck zulässig, über das er nichts weiß.
func TestEinUnbekanntesZulaessigkeitswort(t *testing.T) {
	t.Parallel()

	karte, befunde, fehler := scryfall.LiesKarte(
		beispiel(t, "zulaessigkeit-unbekannt.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	// **Der Befund nennt das Format und das Wort**, sonst müsste jemand die
	// ganze Antwort durchsehen, um zu wissen, worum es geht.
	if !strings.Contains(befunde[0].Meldung, "kuechentisch") ||
		!strings.Contains(befunde[0].Meldung, "vielleicht") {
		t.Errorf("die Meldung sagt zu wenig: %q", befunde[0].Meldung)
	}

	// **Und die Karte kommt trotzdem durch**, samt dem, was lesbar war.
	if karte.Legalitaeten["standard"] != "legal" {
		t.Errorf("das Lesbare ging verloren: %v", karte.Legalitaeten)
	}
}
