package domain_test

// **Die Fehlerwege von Mana und Grundland** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Warum sie einzeln stehen
//
// **Weil jeder von ihnen eine eigene Auskunft gibt** (§8). „Es ging nicht"
// ist keine; wer ein Land tappen will, das nicht ihm gehört, soll das lesen
// und nicht raten. Und weil sie sonst unbetreten blieben — der Riegel zählt
// sie nach, und ein Zweig ohne Lauf ist ein Versprechen, das niemand hält
// (§10).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg, auf dem ein Land nicht tappt, sagt warum.**
func TestJederFehlerwegDesGrundlandsSagtWarum(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		partie domain.Partie
		wer    string
		objekt string
		wort   string
	}{
		// **Ein anderes Objekt davor** — die Suche geht weiter.
		"das Land steht nicht an erster Stelle": {
			partie: mitFremdemObjektDavor(t),
			wer:    "sitz-1", objekt: "sitz-1/land/0",
			wort: "",
		},

		"ein Objekt, das es nicht gibt": {
			partie: einLandtisch(t, "Forest"),
			wer:    "sitz-1", objekt: "gibtsnicht",
			wort: "gibt es in dieser Partie nicht",
		},

		"ein Land auf der Hand": {
			partie: landAufDieHand(einLandtisch(t, "Forest")),
			wer:    "sitz-1", objekt: "sitz-1/land/0",
			wort: "Schlachtfeld",
		},

		"eine Karte, die im Abzug fehlt": {
			partie: ohneAbzug(einLandtisch(t, "Forest")),
			wer:    "sitz-1", objekt: "sitz-1/land/0",
			wort: "eingefrorenen Abzug",
		},

		"eine Kreatur": {
			partie: alsKreatur(einLandtisch(t, "Forest")),
			wer:    "sitz-1", objekt: "sitz-1/land/0",
			wort: "ist kein Land",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := fall.partie.LandFuerMana(fall.wer, fall.objekt)

			if fall.wort == "" {
				if fehler != nil {
					t.Fatalf("das Land tappte nicht: %v", fehler)
				}

				return
			}

			if fehler == nil {
				t.Fatal("es kam kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %q", fehler)
			}
		})
	}
}

// mitFremdemObjektDavor legt ein zweites Objekt vor das Land.
func mitFremdemObjektDavor(t *testing.T) domain.Partie {
	t.Helper()

	partie := einLandtisch(t, "Forest")
	land := partie.Objekte[len(partie.Objekte)-1]

	partie.Objekte = append([]domain.Spielobjekt{{
		Kennung: "sitz-1/anderes/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "land",
	}}, land)

	return partie
}

// landAufDieHand schiebt das Land von der Tischzone auf die Hand.
func landAufDieHand(partie domain.Partie) domain.Partie {
	partie.Objekte[len(partie.Objekte)-1].Zone = domain.ZoneHand

	return partie
}

// ohneAbzug nimmt dem Sitz seinen eingefrorenen Abzug.
func ohneAbzug(partie domain.Partie) domain.Partie {
	partie.Sitze[0].Deck = domain.Partiedeck{Abdruck: "abcd"}

	return partie
}

// alsKreatur macht aus dem Wald eine Kreatur.
func alsKreatur(partie domain.Partie) domain.Partie {
	partie.Sitze[0].Deck.Karten[0].Seiten[0].Felder = map[string]string{
		"super type": "Creature", "sub type": "Bear",
	}

	return partie
}

// **Eine Manaart, die es nicht gibt, kommt nicht in den Vorrat** — CR 106.1b.
func TestEineUnbekannteArtKommtNichtInDenVorrat(t *testing.T) {
	t.Parallel()

	_, fehler := einePartie().ManaHinzu("sitz-1", "Q", 1)

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 106.1b") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Symbol, das CR 107 nicht kennt, ist unbezahlbar** — §8.
func TestEinUnbekanntesSymbolIstUnbezahlbar(t *testing.T) {
	t.Parallel()

	befunde := domain.Unbezahlbar(domain.ManakostenAus("[♄]"))

	if len(befunde) != 1 || !strings.Contains(befunde[0], "CR 107.4") {
		t.Errorf("die Befunde lauten %v", befunde)
	}
}

// **Ein Hybrid, dessen erste Hälfte fehlt, nimmt die zweite** — CR 107.4e.
//
// # Der Zweig, den dieser Lauf betritt
//
// **Der Sprung über eine Farbe, die gar nicht dasteht.** `[W/U]` gegen einen
// Vorrat aus einem **blauen** Mana: Weiß gibt es nicht, also wird der erste
// Weg übersprungen, ohne dass die Suche tiefer geht. Ohne diesen Lauf bliebe
// er unbetreten, und ein Vorrat ohne die erste Farbe hätte nie gezahlt.
func TestEinHybridSpringtUeberDieFehlendeHaelfte(t *testing.T) {
	t.Parallel()

	vorrat := domain.Manavorrat{}.Mehr(domain.ManaBlau, 1)

	if !vorrat.Bezahlbar(domain.ManakostenAus("[W/U]")) {
		t.Error("ein blaues Mana zahlte kein weiß-blaues Hybrid")
	}
}

// **Die erste Wahl kann die zweite verbauen** — CR 107.4e.
//
// # Der Zweig, den dieser Lauf betritt
//
// **Den Rücksprung.** `[W/U][W/B]` gegen ein Weißes und ein Blaues: nimmt
// das erste Symbol das **Weiße**, bleibt dem zweiten weder Weiß noch
// Schwarz — es geht nicht auf. Erst wenn die Suche **zurückgeht** und das
// erste Symbol das Blaue nehmen lässt, zahlt das zweite mit dem Weißen.
//
// **Ohne diesen Lauf bliebe die Stelle unbetreten**, und ein Weg, der nach
// dem ersten gescheiterten Versuch aufgäbe, wäre grün geblieben — die
// anderen Hybridläufe treffen immer schon mit dem ersten Weg (§10).
func TestDieErsteWahlKannDieZweiteVerbauen(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("[W/U][W/B]")

	vorrat := domain.Manavorrat{}.
		Mehr(domain.ManaWeiss, 1).Mehr(domain.ManaBlau, 1)

	danach, ging := vorrat.Bezahle(kosten)

	switch {
	case !ging:
		t.Fatal("die Suche ging nicht zurück — zwei Hybride blieben unbezahlt")
	case !danach.Leer():
		t.Errorf("es blieb Mana übrig: %v", danach)
	}

	// **Die Gegenprobe**: mit zwei Blauen geht es nicht — das zweite Symbol
	// kennt kein Blau.
	if zweiBlau := (domain.Manavorrat{}.Mehr(domain.ManaBlau, 2)); zweiBlau.
		Bezahlbar(kosten) {
		t.Error("zwei blaue Mana haben ein weiß-schwarzes Hybrid bezahlt")
	}
}
