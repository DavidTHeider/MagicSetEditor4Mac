package app

// **Decks und Formate für die drei Flächen ohne Fenster** — Werkstück 7 von
// Phase 4.6
// (#1293).
//
// # Warum das hier steht und nicht in jeder Fläche
//
// **Damit alle vier dieselbe Anwendung benutzen** (§16). Ein zweiter Deckbau
// in `cmd/mse` wäre eine zweite Fachlogik, und zwei Fachlogiken laufen
// auseinander. Dieselbe Form wie `ablagenahtflaechen.go` und
// `sammlungfuerflaechen.go`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Deckwege bündelt, was Kommandozeile, MCP und REST für Decks brauchen.
//
// **Ein Verbund und keine vier Rückgabewerte**: die vier Nähte gehören
// zusammen, und ein Aufrufer, der drei davon vergisst, hat eine Fläche mit
// Lücken.
type Deckwege struct {
	Bau      *service.Deckbau
	Formate  *service.Formatpflege
	Pruefung *service.Formatpruefung
	Abgleich *service.Habenabgleich

	// Zuordnung ist die Naht von U9 (#1420).
	Zuordnung *service.Deckzuordnungspflege

	// Wunsch setzt, was fehlt, auf die Wunschliste (DOM-03.AK-02, #1698).
	Wunsch *service.Fehlkartenwunsch

	// Setdeck legt aus einem Set ein Deck an (FMT-04, #1693).
	Setdeck *service.Setdeck
}

// DeckwegeUeber baut sie über einem Bestand.
//
// **Die Prüfung braucht Scryfall.** Fehlt der Zugang, kommt der Fehler hier
// heraus — anders als beim Fenster, das ohne ihn trotzdem aufgeht. Eine
// Kommandozeile, die stillschweigend nicht prüft, wäre schlimmer als eine,
// die es sagt.
func DeckwegeUeber(db *sql.DB) (Deckwege, error) {
	var (
		decks     = storage.NeueDeckablage(db)
		formate   = storage.NeueFormatablage(db)
		verweise  = storage.NeueKartenverweise(db)
		setzeiger = storage.NeuerSetLeser(db)
	)

	// **Die Namensnähte gehören hierhin und nicht nur an `deckstart.go`**
	// (#1376). Es sind zwei Composition Roots (§2), und beim ersten Anlauf
	// war nur die des Fensters verdrahtet: der Dienst war grün, das Fenster
	// zeigte Namen — und Kommandozeile, MCP und REST weiter UUIDs (§16).
	sammlung := storage.NeuerSammlungsspeicher(db)
	abgleich := service.NeuerHabenabgleich(decks, sammlung, verweise)
	bau := service.NeuerDeckbau(decks, setzeiger, formate).MitKartennamen(setzeiger)

	wege := Deckwege{
		Bau:      bau,
		Formate:  service.NeueFormatpflege(formate),
		Abgleich: abgleich,

		// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420), **samt der
		// Naht zum Handeln** seit `T2` (#1467, 0120): was ein Handelsvorgang
		// festhält, ist hier nicht mehr null.
		Zuordnung: service.NeueDeckzuordnungspflege(
			decks, storage.NeueDeckzuordnungen(db), sammlung, abgleich).
			MitReservierungen(storage.NeueReservierungsablage(db)),

		Wunsch: service.NeuerFehlkartenwunsch(abgleich, bau, sammlung),

		Setdeck: service.NeuesSetdeck(setzeiger, decks),
	}

	// **Wer den Bestand hat, baut die Dienstwahl** — die Schalter stehen in
	// den Einstellungen, und die liegen im Bestand (Phase 4.8).
	klient, fehler := Scryfallklient(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)))
	if fehler != nil {
		return wege, fehler
	}

	// **Mit der Naht zu den Feldern eigener Karten**
	// (#1622):
	// die Typzeile eines selbstgebauten Standardlands steht im Bestand, und
	// bis zum 2026-09-17 wurde sie **nur** bei Scryfall gesucht. Zwanzig
	// eigene Gebirge waren damit zwanzig gleiche Karten (CR 100.2a).
	wege.Pruefung = service.NeueFormatpruefung(
		decks, formate, klient, verweise).MitKartenfeldern(setzeiger)
	wege.Bau.MitDrucknamen(klient)

	return wege, nil
}

// Deckflaeche ist die Anwendung selbst, für eine Fläche ohne Fenster.
//
// **Dieselben DTOs wie am Fenster.** Wer für die Kommandozeile eine zweite
// Gestalt bauen wollte, hätte zwei Antworten auf dieselbe Frage.
type Deckflaeche struct{ app *App }

// DeckflaecheUeber baut sie über einem Bestand.
func DeckflaecheUeber(db *sql.DB) (*Deckflaeche, error) {
	wege, fehler := DeckwegeUeber(db)

	anwendung := NeueApp("")
	anwendung.MitDecknaht(wege.Bau)
	anwendung.MitFormatnaht(wege.Formate)
	anwendung.MitDeckabgleich(wege.Abgleich)
	anwendung.MitDeckzuordnung(wege.Zuordnung)
	anwendung.MitFehlkartenwunsch(wege.Wunsch)
	anwendung.MitSetdeck(wege.Setdeck)

	// **Auch wenn Scryfall fehlt, steht der Rest bereit.** Der Fehler geht
	// mit hinaus; der Aufrufer entscheidet, ob er ihn meldet oder ohne
	// Prüfung weitermacht.
	if wege.Pruefung != nil {
		anwendung.MitDeckpruefung(wege.Pruefung)
	}

	return &Deckflaeche{app: anwendung}, fehler
}

func (d *Deckflaeche) Decks(ctx context.Context, spiel string) DecksAntwort {
	return d.app.Decks(ctx, spiel)
}

func (d *Deckflaeche) Deck(ctx context.Context, deckID string) DeckAntwort {
	return d.app.Deck(ctx, deckID)
}

func (d *Deckflaeche) DeckAnlegen(ctx context.Context, name, spiel, notiz string) DeckAntwort {
	return d.app.DeckAnlegen(ctx, name, spiel, notiz)
}

func (d *Deckflaeche) DeckUmbenennen(
	ctx context.Context, deckID, fassung, name, notiz string,
) DeckAntwort {
	return d.app.DeckUmbenennen(ctx, deckID, fassung, name, notiz)
}

func (d *Deckflaeche) DeckFehlkartenWuenschen(ctx context.Context, deckID string) FehlkartenwunschAntwort {
	return d.app.DeckFehlkartenWuenschen(ctx, deckID)
}

func (d *Deckflaeche) DeckKopieren(ctx context.Context, deckID, name string) DeckAntwort {
	return d.app.DeckKopieren(ctx, deckID, name)
}

func (d *Deckflaeche) DeckLoeschen(ctx context.Context, deckID string) SchreibAntwort {
	return d.app.DeckLoeschen(ctx, deckID)
}

func (d *Deckflaeche) DeckplatzLegen(
	ctx context.Context, deckID, karteID, printing string, menge int, bereich string,
) DeckplatzAntwort {
	return d.app.DeckplatzLegen(ctx, deckID, karteID, printing, menge, bereich)
}

func (d *Deckflaeche) DeckplatzNehmen(ctx context.Context, deckID, platzID string) SchreibAntwort {
	return d.app.DeckplatzNehmen(ctx, deckID, platzID)
}

// DeckplatzVerschieben legt einen Platz in den anderen Bereich.
func (d *Deckflaeche) DeckplatzVerschieben(
	ctx context.Context, deckID, platzID, bereich string,
) DeckplatzAntwort {
	return d.app.DeckplatzVerschieben(ctx, deckID, platzID, bereich)
}

func (d *Deckflaeche) DeckFormatWaehlen(ctx context.Context, deckID, formatID string) DeckAntwort {
	return d.app.DeckFormatWaehlen(ctx, deckID, formatID)
}

// DeckPruefen beurteilt ein Deck gegen sein Format.
//
// **Ohne Scryfall gibt es hier einen Fehler und kein Schweigen** — die Naht
// ist dann nicht gesetzt.
func (d *Deckflaeche) DeckPruefen(ctx context.Context, deckID string) DeckurteilAntwort {
	if d.app.deckpruefung == nil {
		return DeckurteilAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx,
			domain.NewFieldError("deckID", domain.SchluesselValidierung,
				"Die Formatprüfung braucht einen Zugang zu Scryfall, und der "+
					"ließ sich nicht einrichten."))}
	}

	return d.app.DeckPruefen(ctx, deckID)
}

// DeckZuordnungen gibt den Bedarf samt den Zuordnungen (U9, #1420).
func (d *Deckflaeche) DeckZuordnungen(ctx context.Context, deckID string) ZuordnungsstandAntwort {
	return d.app.DeckZuordnungen(ctx, deckID)
}

// DeckZuordnen legt Exemplare in einen Platz.
func (d *Deckflaeche) DeckZuordnen(
	ctx context.Context, deckID, platzID, exemplarID string, menge int,
) ZuordnungAntwort {
	return d.app.DeckZuordnen(ctx, deckID, platzID, exemplarID, menge)
}

// DeckZuordnungLoesen nimmt eine Zuordnung heraus.
func (d *Deckflaeche) DeckZuordnungLoesen(ctx context.Context, deckID, zuordnungID string) SchreibAntwort {
	return d.app.DeckZuordnungLoesen(ctx, deckID, zuordnungID)
}

// DeckFehlkarten gibt, was zu besorgen wäre.
func (d *Deckflaeche) DeckFehlkarten(ctx context.Context, deckID string) FehlkartenAntwort {
	return d.app.DeckFehlkarten(ctx, deckID)
}

// ExemplarDecks sagt, welche Decks ein Exemplar benutzen.
func (d *Deckflaeche) ExemplarDecks(ctx context.Context, exemplarID string) DeckbezuegeAntwort {
	return d.app.ExemplarDecks(ctx, exemplarID)
}

func (d *Deckflaeche) DeckAbgleichen(ctx context.Context, deckID string) HabenstandAntwort {
	return d.app.DeckAbgleichen(ctx, deckID)
}

func (d *Deckflaeche) Bauarten() BauartenAntwort { return d.app.Bauarten() }

func (d *Deckflaeche) Formate(ctx context.Context, spiel string) FormateAntwort {
	return d.app.Formate(ctx, spiel)
}

func (d *Deckflaeche) FormatAusBauart(
	ctx context.Context, bauart, name, spiel, scryfallname string,
) FormatAntwort {
	return d.app.FormatAusBauart(ctx, bauart, name, spiel, scryfallname)
}

func (d *Deckflaeche) FormatSetzen(ctx context.Context, format FormatDTO) FormatAntwort {
	return d.app.FormatSetzen(ctx, format)
}

func (d *Deckflaeche) FormatLoeschen(ctx context.Context, formatID string) SchreibAntwort {
	return d.app.FormatLoeschen(ctx, formatID)
}

// DecklisteAusgeben gibt ein Deck als Liste (NET-04, #1708).
func (d *Deckflaeche) DecklisteAusgeben(ctx context.Context, deckID string) DecklisteAntwort {
	return d.app.DecklisteAusgeben(ctx, deckID)
}
