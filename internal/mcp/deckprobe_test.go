package mcp_test

// **Die Probe für die Deckwerkzeuge** — die Vorrichtung zu `decks_test.go`.
//
// **Eine eigene Datei**, seit die fünf Wege der Zuordnung (U9, #1420) sie
// über die Zeilengrenze aus §2 gehoben haben. Der Schnitt folgt der Frage:
// dort steht, **was** die Werkzeuge tun, hier, **womit** sie geprüft werden.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// deckprobe ist ein Doppel für `mcp.Decks`.
type deckprobe struct {
	ergebnis any
	fehler   error

	gerufen string
	deckID  int64
	fassung int64
	platzID int64

	// exemplarID und zuordnungID gehören zur Zuordnung (U9, #1420).
	exemplarID  int64
	zuordnungID int64

	formatID int64
	karteID  int64
	setID    int64
	menge    int
	spiel    string
	name     string
	notiz    string
	printing string
	bereich  string
	bauart   string
	scryfall string
	angaben  map[string]any
}

func (d *deckprobe) antwort(name string) (any, error) {
	d.gerufen = name

	return d.ergebnis, d.fehler
}

func (d *deckprobe) Decks(_ context.Context, spiel string) (any, error) {
	d.spiel = spiel

	return d.antwort("Decks")
}

func (d *deckprobe) Deck(_ context.Context, deckID int64) (any, error) {
	d.deckID = deckID

	return d.antwort("Deck")
}

func (d *deckprobe) DeckAnlegen(_ context.Context, name, spiel, notiz string) (
	any, error,
) {
	d.name, d.spiel, d.notiz = name, spiel, notiz

	return d.antwort("DeckAnlegen")
}

func (d *deckprobe) DeckUmbenennen(
	_ context.Context, deckID, fassung int64, name, notiz string,
) (any, error) {
	d.deckID, d.fassung, d.name, d.notiz = deckID, fassung, name, notiz

	return d.antwort("DeckUmbenennen")
}

func (d *deckprobe) DeckFehlkartenWuenschen(_ context.Context, deckID int64) (any, error) {
	d.deckID = deckID

	return d.antwort("DeckFehlkartenWuenschen")
}

func (d *deckprobe) DeckKopieren(_ context.Context, deckID int64, name string) (any, error) {
	d.deckID, d.name = deckID, name

	return d.antwort("DeckKopieren")
}

func (d *deckprobe) DeckAusSet(_ context.Context, setID int64, name string, menge int) (any, error) {
	d.setID, d.name, d.menge = setID, name, menge

	return d.antwort("DeckAusSet")
}

func (d *deckprobe) DeckLoeschen(_ context.Context, deckID int64) error {
	d.deckID = deckID
	d.gerufen = "DeckLoeschen"

	return d.fehler
}

func (d *deckprobe) DeckplatzLegen(
	_ context.Context, deckID, karteID int64, printing string, menge int, bereich string,
) (any, error) {
	d.deckID, d.karteID, d.printing = deckID, karteID, printing
	d.menge, d.bereich = menge, bereich

	return d.antwort("DeckplatzLegen")
}

func (d *deckprobe) DeckplatzNehmen(_ context.Context, deckID, platzID int64) error {
	d.deckID, d.platzID = deckID, platzID
	d.gerufen = "DeckplatzNehmen"

	return d.fehler
}

func (d *deckprobe) DeckplatzVerschieben(
	_ context.Context, deckID, platzID int64, bereich string,
) (any, error) {
	d.deckID, d.platzID = deckID, platzID
	d.bereich = bereich

	return d.antwort("DeckplatzVerschieben")
}

func (d *deckprobe) DeckFormatWaehlen(_ context.Context, deckID, formatID int64) (
	any, error,
) {
	d.deckID, d.formatID = deckID, formatID

	return d.antwort("DeckFormatWaehlen")
}

func (d *deckprobe) DeckPruefen(_ context.Context, deckID int64) (any, error) {
	d.deckID = deckID

	return d.antwort("DeckPruefen")
}

func (d *deckprobe) DeckAbgleichen(_ context.Context, deckID int64) (any, error) {
	d.deckID = deckID

	return d.antwort("DeckAbgleichen")
}

func (d *deckprobe) DeckZuordnungen(_ context.Context, deckID int64) (any, error) {
	d.deckID = deckID

	return d.antwort("DeckZuordnungen")
}

func (d *deckprobe) DeckZuordnen(
	_ context.Context, deckID, platzID, exemplarID int64, menge int,
) (any, error) {
	d.deckID, d.platzID, d.exemplarID, d.menge = deckID, platzID, exemplarID, menge

	return d.antwort("DeckZuordnen")
}

func (d *deckprobe) DeckZuordnungLoesen(_ context.Context, deckID, zuordnungID int64) error {
	d.deckID, d.zuordnungID = deckID, zuordnungID

	_, fehler := d.antwort("DeckZuordnungLoesen")

	return fehler
}

func (d *deckprobe) DeckFehlkarten(_ context.Context, deckID int64) (any, error) {
	d.deckID = deckID

	return d.antwort("DeckFehlkarten")
}

func (d *deckprobe) Deckliste(_ context.Context, deckID int64) (any, error) {
	d.deckID = deckID

	return d.antwort("Deckliste")
}

func (d *deckprobe) ExemplarDecks(_ context.Context, exemplarID int64) (any, error) {
	d.exemplarID = exemplarID

	return d.antwort("ExemplarDecks")
}

func (d *deckprobe) Bauarten(_ context.Context) (any, error) {
	return d.antwort("Bauarten")
}

func (d *deckprobe) Formate(_ context.Context, spiel string) (any, error) {
	d.spiel = spiel

	return d.antwort("Formate")
}

func (d *deckprobe) FormatAusBauart(
	_ context.Context, bauart, name, spiel, scryfallname string,
) (any, error) {
	d.bauart, d.name, d.spiel, d.scryfall = bauart, name, spiel, scryfallname

	return d.antwort("FormatAusBauart")
}

func (d *deckprobe) FormatSetzen(_ context.Context, format map[string]any) (
	any, error,
) {
	d.angaben = format

	return d.antwort("FormatSetzen")
}

func (d *deckprobe) FormatLoeschen(_ context.Context, formatID int64) error {
	d.formatID = formatID
	d.gerufen = "FormatLoeschen"

	return d.fehler
}

// rufeDeckwerkzeug sucht ein Werkzeug und ruft es.
func rufeDeckwerkzeug(
	t *testing.T, probe *deckprobe, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Decks: probe}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}
