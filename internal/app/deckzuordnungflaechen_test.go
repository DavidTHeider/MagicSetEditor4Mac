package app_test

// **Die Zuordnung über die Flächen ohne Fenster** — Werkstück U9 von
// Phase 5.1 (#1420).
//
// **Eine eigene Datei**, damit `deckflaeche_test.go` und
// `deckwerkstatt_test.go` unter der Zeilengrenze aus §2 bleiben. Geprüft
// wird, dass beide Wurzeln die Naht verdrahten — nicht die Fachregel: die
// steht in `service`.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der ganze Zuordnungsweg über die Fläche der Kommandozeile.**
func TestDerZuordnungswegUeberDieFlaeche(t *testing.T) {
	flaeche := deckflaeche(t)

	deck := flaeche.DeckAnlegen(probekontext(t), "Aggro", "magic", "")
	if deck.Fehler != nil {
		t.Fatalf("anlegen: %v", deck.Fehler)
	}

	platz := flaeche.DeckplatzLegen(probekontext(t), deck.Deck.ID, "",
		"a1b2c3d4-0000-4000-8000-00000000000a", 4, "hauptdeck")
	if platz.Fehler != nil {
		t.Fatalf("legen: %v", platz.Fehler)
	}

	stand := flaeche.DeckZuordnungen(probekontext(t), deck.Deck.ID)
	if stand.Fehler != nil || len(stand.Zeilen) != 1 {
		t.Fatalf("Stand: %+v", stand)
	}

	// **Ohne Exemplar geht keine Zuordnung** — die Sammlung ist leer.
	if fehler := flaeche.DeckZuordnen(probekontext(t), deck.Deck.ID, platz.Platz.ID, "1", 1).Fehler; fehler == nil {
		t.Error("die Zuordnung ohne Exemplar kam durch")
	}

	if fehlkarten := flaeche.DeckFehlkarten(probekontext(t), deck.Deck.ID); fehlkarten.Fehler != nil ||
		fehlkarten.Liste == "" {
		t.Errorf("Fehlkarten: %+v", fehlkarten)
	}

	// **Ein Exemplar, das es nicht gibt, ist ein Nein** — seit #1815 wie eines
	// aus einem fremden Profil; bis dahin kam eine leere Liste.
	if decks := flaeche.ExemplarDecks(probekontext(t), "1"); decks.Fehler == nil ||
		decks.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Rückweg: %+v", decks)
	}

	// **Eine Zuordnung, die es nicht gibt, ist ein Nein.**
	if fehler := flaeche.DeckZuordnungLoesen(probekontext(t), deck.Deck.ID, "9").Fehler; fehler == nil {
		t.Error("das Lösen ins Leere kam durch")
	}
}

// **Und derselbe Weg über die Werkstatt, die MCP und REST bedienen.**
func TestDerZuordnungswegUeberDieWerkstatt(t *testing.T) {
	bestand := frischerBestand(t)
	werkstatt := app.NeueDeckwerkstatt(bestand)

	sammlung, sammlungsfehler := app.SammlungsnahtUeber(bestand)
	if sammlungsfehler != nil {
		t.Fatalf("die Sammlung: %v", sammlungsfehler)
	}

	exemplarID, anlegefehler := sammlung.ExemplarAnlegen(probekontext(t), app.ExemplarDTO{
		Printing: "a1b2c3d4-0000-4000-8000-00000000000a", Menge: 1,
		Zustand: "nearMint", Sprache: "en",
	})
	if anlegefehler != nil {
		t.Fatalf("ein Exemplar: %v", anlegefehler)
	}

	deck, fehler := werkstatt.DeckAnlegen(probekontext(t), "Aggro", "magic", "")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	deckID := kennungAusWerkstatt(t, deck)

	platz, legefehler := werkstatt.DeckplatzLegen(probekontext(t), deckID, 0,
		"a1b2c3d4-0000-4000-8000-00000000000a", 4, "hauptdeck")
	if legefehler != nil {
		t.Fatalf("legen: %v", legefehler)
	}

	mussWerkstatt(t)(werkstatt.DeckZuordnungen(probekontext(t), deckID))
	mussWerkstatt(t)(werkstatt.DeckFehlkarten(probekontext(t), deckID))
	mussWerkstatt(t)(werkstatt.ExemplarDecks(probekontext(t), exemplarID))

	// **Zuordnen, wieder lösen** — der Weg, den nur ein echtes Exemplar geht.
	gesetzt, zuordnungsfehler := werkstatt.DeckZuordnen(
		probekontext(t), deckID, kennungAusWerkstatt(t, platz), exemplarID, 1)
	if zuordnungsfehler != nil {
		t.Fatalf("zuordnen: %v", zuordnungsfehler)
	}

	if err := werkstatt.DeckZuordnungLoesen(
		probekontext(t), deckID, kennungAusWerkstatt(t, gesetzt)); err != nil {
		t.Errorf("lösen: %v", err)
	}

	if _, err := werkstatt.DeckZuordnen(
		probekontext(t), deckID, kennungAusWerkstatt(t, platz), 9999, 1); err == nil {
		t.Error("die Zuordnung eines fehlenden Exemplars kam durch")
	}

	if err := werkstatt.DeckZuordnungLoesen(probekontext(t), deckID, 9); err == nil {
		t.Error("das Lösen ins Leere kam durch")
	}

	// **Und eine Kennung, die keine ist, kommt gar nicht erst hinein** — die
	// Werkstatt reicht Zahlen, die Naht rechnet in Zeichenketten.
	if _, err := werkstatt.DeckZuordnungen(probekontext(t), 0); err == nil {
		t.Error("die Null als Deck kam durch")
	}
}
