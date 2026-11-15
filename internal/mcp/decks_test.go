package mcp_test

// **Decks und Formate über MCP** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Die fachliche Zusage prüft `internal/flaechen`** — dort läuft derselbe
// Ablauf über alle drei Flächen. Hier steht, dass jedes Werkzeug seine
// Angaben **weiterreicht** und seinen Fehler nicht verschluckt.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Es merkt sich jede Angabe** — das ist der ganze Zweck: geprüft wird, dass
// sie ankommt, nicht was danach geschieht.
func TestJedesDeckwerkzeugReichtSeineAngabenDurch(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		werkzeug  string
		argumente string
		erwartet  string
		pruefe    func(*testing.T, *deckprobe)
	}{
		{"lesen.decks", `{"spiel":"magic"}`, "Decks", func(t *testing.T, p *deckprobe) {
			if p.spiel != "magic" {
				t.Errorf("Spiel %q", p.spiel)
			}
		}},
		{"lesen.deck", `{"deckID":7}`, "Deck", func(t *testing.T, p *deckprobe) {
			if p.deckID != 7 {
				t.Errorf("Deck %d", p.deckID)
			}
		}},
		{"lesen.deck.pruefung", `{"deckID":7}`, "DeckPruefen", nil},
		{"lesen.deck.bestand", `{"deckID":7}`, "DeckAbgleichen", nil},
		{"lesen.deck.zuordnungen", `{"deckID":7}`, "DeckZuordnungen", nil},
		{"lesen.deck.fehlkarten", `{"deckID":7}`, "DeckFehlkarten", nil},
		{"lesen.deck.liste", `{"deckID":7}`, "Deckliste", nil},
		{"lesen.exemplar.decks", `{"exemplarID":7}`, "ExemplarDecks", func(t *testing.T, p *deckprobe) {
			if p.exemplarID != 7 {
				t.Errorf("Exemplar %d", p.exemplarID)
			}
		}},
		{
			"schreiben.deck.zuordnen", `{"deckID":7,"platzID":3,"exemplarID":5,"menge":2}`,
			"DeckZuordnen", func(t *testing.T, p *deckprobe) {
				if p.platzID != 3 || p.exemplarID != 5 || p.menge != 2 {
					t.Errorf("%+v", p)
				}
			},
		},
		{
			"schreiben.deck.zuordnung.loesen", `{"deckID":7,"zuordnungID":4}`,
			"DeckZuordnungLoesen", func(t *testing.T, p *deckprobe) {
				if p.zuordnungID != 4 {
					t.Errorf("Zuordnung %d", p.zuordnungID)
				}
			},
		},
		{"lesen.formatbauarten", `{}`, "Bauarten", nil},
		{"lesen.formate", `{"spiel":"magic"}`, "Formate", nil},
		{
			"schreiben.deck.anlegen", `{"name":"Aggro","spiel":"magic","notiz":"N"}`,
			"DeckAnlegen", func(t *testing.T, p *deckprobe) {
				if p.name != "Aggro" || p.notiz != "N" {
					t.Errorf("%+v", p)
				}
			},
		},
		{
			"schreiben.deck.umbenennen",
			`{"deckID":7,"fassung":2,"name":"Neu","notiz":""}`,
			"DeckUmbenennen", func(t *testing.T, p *deckprobe) {
				if p.fassung != 2 {
					t.Errorf("Fassung %d", p.fassung)
				}
			},
		},
		{"schreiben.deck.loeschen", `{"deckID":7}`, "DeckLoeschen", nil},
		{
			"schreiben.deck.fehlkarten.wuenschen", `{"deckID":7}`,
			"DeckFehlkartenWuenschen", func(t *testing.T, p *deckprobe) {
				if p.deckID != 7 {
					t.Errorf("gewünscht für Deck %d", p.deckID)
				}
			},
		},
		{
			"schreiben.deck.ausset", `{"setID":4,"name":"Werkstatt","menge":2}`,
			"DeckAusSet", func(t *testing.T, p *deckprobe) {
				if p.setID != 4 || p.name != "Werkstatt" || p.menge != 2 {
					t.Errorf("aus Set %d als %q mit %d", p.setID, p.name, p.menge)
				}
			},
		},
		{
			"schreiben.deck.kopieren", `{"deckID":7,"name":"Aggro II"}`,
			"DeckKopieren", func(t *testing.T, p *deckprobe) {
				if p.deckID != 7 || p.name != "Aggro II" {
					t.Errorf("kopiert %d als %q", p.deckID, p.name)
				}
			},
		},
		{
			"schreiben.deckplatz.legen",
			`{"deckID":7,"karteID":9,"printing":"","menge":4,"bereich":"sideboard"}`,
			"DeckplatzLegen", func(t *testing.T, p *deckprobe) {
				if p.karteID != 9 || p.menge != 4 || p.bereich != "sideboard" {
					t.Errorf("%+v", p)
				}
			},
		},
		{
			"schreiben.deckplatz.nehmen", `{"deckID":7,"platzID":3}`,
			"DeckplatzNehmen", func(t *testing.T, p *deckprobe) {
				if p.platzID != 3 {
					t.Errorf("Platz %d", p.platzID)
				}
			},
		},
		{
			// **Der Bereichswechsel** — Befund 7 des Zweitblicks über Phase 5.1.
			"schreiben.deckplatz.verschieben",
			`{"deckID":7,"platzID":3,"bereich":"sideboard"}`,
			"DeckplatzVerschieben", func(t *testing.T, p *deckprobe) {
				if p.platzID != 3 || p.bereich != "sideboard" {
					t.Errorf("%+v", p)
				}
			},
		},
		{
			"schreiben.deck.format", `{"deckID":7,"formatID":5}`,
			"DeckFormatWaehlen", func(t *testing.T, p *deckprobe) {
				if p.formatID != 5 {
					t.Errorf("Format %d", p.formatID)
				}
			},
		},
		{
			"schreiben.format.ausbauart",
			`{"bauart":"commander","name":"C","spiel":"magic","scryfallname":"commander"}`,
			"FormatAusBauart", func(t *testing.T, p *deckprobe) {
				if p.bauart != "commander" || p.scryfall != "commander" {
					t.Errorf("%+v", p)
				}
			},
		},
		{
			"schreiben.format.setzen", `{"id":5,"name":"S","spiel":"magic"}`,
			"FormatSetzen", func(t *testing.T, p *deckprobe) {
				if p.angaben["name"] != "S" {
					t.Errorf("%+v", p.angaben)
				}
			},
		},
		{
			"schreiben.format.loeschen", `{"formatID":5}`,
			"FormatLoeschen", func(t *testing.T, p *deckprobe) {
				if p.formatID != 5 {
					t.Errorf("Format %d", p.formatID)
				}
			},
		},
	} {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &deckprobe{ergebnis: map[string]any{"da": true}}

			roh, fehler := rufeDeckwerkzeug(t, probe, fall.werkzeug, fall.argumente)
			if fehler != nil {
				t.Fatalf("%v", fehler)
			}

			if probe.gerufen != fall.erwartet {
				t.Errorf("gerufen wurde %q, erwartet %q", probe.gerufen, fall.erwartet)
			}

			if roh == "" {
				t.Error("die Antwort ist leer")
			}

			if fall.pruefe != nil {
				fall.pruefe(t, probe)
			}
		})
	}
}

// **Und jedes gibt den Fehler seiner Naht weiter.**
//
// **Das Protokoll trennt Ergebnis und Fehlschlag**: ein Ergebnis mit einem
// Fehler darin läse sich wie ein Erfolg.
func TestJedesDeckwerkzeugGibtSeinenFehlerWeiter(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Naht schweigt")

	for _, werkzeug := range []struct{ name, argumente string }{
		{"lesen.decks", `{}`},
		{"lesen.deck", `{"deckID":7}`},
		{"lesen.deck.pruefung", `{"deckID":7}`},
		{"lesen.deck.bestand", `{"deckID":7}`},
		{"lesen.formatbauarten", `{}`},
		{"lesen.formate", `{}`},
		{"schreiben.deck.anlegen", `{"name":"A","spiel":"magic"}`},
		{"schreiben.deck.umbenennen", `{"deckID":7,"fassung":1,"name":"A"}`},
		{"schreiben.deck.loeschen", `{"deckID":7}`},
		{"schreiben.deck.kopieren", `{"deckID":7,"name":"A"}`},
		{"schreiben.deck.fehlkarten.wuenschen", `{"deckID":7}`},
		{"schreiben.deck.ausset", `{"setID":4,"menge":1}`},
		{"schreiben.deckplatz.legen", `{"deckID":7,"menge":1,"bereich":"hauptdeck"}`},
		{"schreiben.deckplatz.nehmen", `{"deckID":7,"platzID":3}`},
		{"schreiben.deckplatz.verschieben", `{"deckID":7,"platzID":3,"bereich":"sideboard"}`},
		{"schreiben.deck.format", `{"deckID":7}`},
		{"schreiben.format.ausbauart", `{"bauart":"x","name":"A","spiel":"magic"}`},
		{"schreiben.format.setzen", `{"name":"A","spiel":"magic"}`},
		{"schreiben.format.loeschen", `{"formatID":5}`},

		// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
		{"lesen.deck.zuordnungen", `{"deckID":7}`},
		{"lesen.deck.fehlkarten", `{"deckID":7}`},
		{"lesen.exemplar.decks", `{"exemplarID":5}`},
		{"schreiben.deck.zuordnen", `{"deckID":7,"platzID":3,"exemplarID":5,"menge":1}`},
		{"schreiben.deck.zuordnung.loesen", `{"deckID":7,"zuordnungID":9}`},
	} {
		t.Run(werkzeug.name, func(t *testing.T) {
			t.Parallel()

			probe := &deckprobe{fehler: kaputt}

			if _, fehler := rufeDeckwerkzeug(
				t, probe, werkzeug.name, werkzeug.argumente); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Ein unlesbares Argument ist ein Fehler und kein Ruf.**
func TestEinUnlesbaresDeckargument(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range []string{
		"lesen.decks", "lesen.deck", "lesen.deck.pruefung", "lesen.deck.bestand",
		"lesen.formate", "schreiben.deck.anlegen", "schreiben.deck.umbenennen",
		"schreiben.deck.loeschen", "schreiben.deckplatz.legen",
		"schreiben.deckplatz.nehmen", "schreiben.deckplatz.verschieben",
		"schreiben.deck.format", "schreiben.deck.kopieren",
		"schreiben.deck.fehlkarten.wuenschen", "schreiben.deck.ausset",
		"schreiben.format.ausbauart", "schreiben.format.setzen",
		"schreiben.format.loeschen",
	} {
		t.Run(werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &deckprobe{}

			_, fehler := rufeDeckwerkzeug(t, probe, werkzeug, `kein JSON`)
			if fehler == nil {
				t.Error("der Unsinn kam durch")
			}

			if probe.gerufen != "" {
				t.Errorf("gerufen wurde trotzdem %q", probe.gerufen)
			}
		})
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
//
// **Und das ist kein Fehler**: ein Prozess ohne Decks bietet sie gar nicht
// erst an, statt sie mit einem Absturz anzubieten.
func TestOhneDecknahtGibtEsKeineDeckwerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.Contains(werkzeug.Name, "deck") ||
			strings.Contains(werkzeug.Name, "format") {
			t.Errorf("das Werkzeug %q steht ohne Naht da", werkzeug.Name)
		}
	}
}

// **Ein kaputtes Argument ist ein Fehler und kein leeres Ergebnis.**
//
// **Die drei Werkzeuge mit eigenem `Ruf`** (U9, #1420): die übrigen teilen
// sich `deckruf`, und dessen Zweig ist längst betreten. Ohne diesen Lauf
// bliebe die Stelle unbetreten, an der ein Modell etwas schickt, das kein
// JSON ist.
func TestEinKaputtesArgumentDerZuordnungIstEinFehler(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"lesen.exemplar.decks",
		"schreiben.deck.zuordnen",
		"schreiben.deck.zuordnung.loesen",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeDeckwerkzeug(t, &deckprobe{}, name, `{"deckID":`); fehler == nil {
				t.Error("das kaputte Argument kam durch")
			}
		})
	}
}
