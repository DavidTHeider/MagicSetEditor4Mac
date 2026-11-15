package app

// **Die Decknaht** — Werkstück 7 von Phase 4.6
// (#1293).
//
// # Was diese Schicht tut, und was nicht
//
// Sie übersetzt zwischen Wails und dem Deckbau: Kennungen als Text hinein,
// DTOs hinaus, Fehler in die Gestalt, die alle vier Flächen kennen. **Eine
// fachliche Entscheidung steht hier nicht** (§2) — welches Format zu welchem
// Spiel gehört, entscheidet `service.Deckbau`.
//
// # Warum Kennungen als Text reisen
//
// Über die Wails-Bindung käme ein `int64` als JavaScript-Zahl an, und die
// verliert oberhalb von 2⁵³ die letzten Stellen. Dieselbe Regel wie beim
// Packsamen.

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Decknaht ist, was das Fenster zum Deckbau braucht.
type Decknaht interface {
	Anlegen(ctx context.Context, deck domain.Deck) (domain.Deck, error)
	Umbenennen(ctx context.Context, deck domain.Deck) (domain.Deck, error)
	Deck(ctx context.Context, kennung int64) (domain.Deck, error)
	Decks(ctx context.Context, spiel string) ([]domain.Deck, error)
	PlatzLegen(ctx context.Context, deckID int64, platz domain.Deckplatz) (
		domain.Deckplatz, error)
	PlatzNehmen(ctx context.Context, deckID, platzID int64) error

	// PlatzVerschieben legt einen Platz in den anderen Bereich — in einem Zug
	// (Befund 7 des Zweitblicks über Phase 5.1).
	PlatzVerschieben(ctx context.Context, deckID, platzID int64, bereich domain.Deckbereich) (
		domain.Deckplatz, error)
	FormatWaehlen(ctx context.Context, deckID, formatID int64) (domain.Deck, error)
	Loeschen(ctx context.Context, kennung int64) error

	// Kopieren legt eine Kopie unter neuem Namen an (DOM-03.AK-02, #1698).
	Kopieren(ctx context.Context, quelle int64, name string) (domain.Deck, error)
}

// DeckplatzDTO ist ein Eintrag eines Decks für das Fenster.
type DeckplatzDTO struct {
	ID string `json:"id"`

	// KarteID nennt eine eigene Karte, Printing einen echten Druck —
	// **genau eines von beiden ist gefüllt**.
	KarteID  string `json:"karteID"`
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441) — als
	// Zeichenkette wie jede Kennung, leer für eine eigene Karte. Vorlaeufig
	// sagt, dass der Druck keinen Alias hat.
	DruckID    string `json:"druckID"`
	Vorlaeufig bool   `json:"vorlaeufig"`

	// Typzeile und Manawert tragen die Auswertung — Werkstück U10 von
	// Phase 5.1 (#1421). Beides kommt mit dem Namen aus derselben Antwort;
	// leer und null heißen „nicht zu haben" (§16, Regel 5).
	Typzeile string `json:"typzeile"`
	Manawert int    `json:"manawert"`

	Menge   int    `json:"menge"`
	Bereich string `json:"bereich"`

	// Bezeichnung ist der **Name** der Karte
	// (#1376).
	//
	// **Leer heißt: nicht zu haben.** Der Name eines Drucks kommt von
	// Scryfall; ohne Netz bleibt er aus (§16, Regel 5), und das Fenster zeigt
	// dann die Kennung.
	Bezeichnung string `json:"bezeichnung"`

	// Manakosten sind die Kosten der Karte — für die Kurve eines Decks.
	//
	// **Leer heißt: nicht ausgefüllt.** Wer daraus eine Kurve rechnet, lässt
	// die leeren weg.
	Manakosten string `json:"manakosten"`
}

// DeckDTO ist ein Deck für das Fenster.
type DeckDTO struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Name  string `json:"name"`
	Spiel string `json:"spiel"`
	Notiz string `json:"notiz"`

	// FormatID ist leer, wenn kein Format gewählt ist — **nicht "0"**: eine
	// Null sähe aus wie eine Kennung.
	FormatID string `json:"formatID"`

	// Plaetze ist bei der Übersicht leer und beim einzelnen Deck gefüllt.
	//
	// **Eine leere Liste und kein `null`** — die Zusage aller Antworten
	// dieses Pakets.
	Plaetze []DeckplatzDTO `json:"plaetze"`

	Hauptdeck int `json:"hauptdeck"`
	Sideboard int `json:"sideboard"`

	// Herkunft sagt, woher dieses Deck kam — Werkstück 12 (#1351).
	//
	// **Beim einzelnen Deck gefüllt, in der Übersicht leer**: dieselbe
	// Zusage wie bei `Plaetze`, und aus demselben Grund — die Übersicht
	// beantwortet „welche Decks gibt es".
	Herkunft HerkunftDTO `json:"herkunft"`
}

// DecksAntwort trägt die Übersicht.
type DecksAntwort struct {
	Decks  []DeckDTO `json:"decks"`
	Fehler *Fehler   `json:"fehler"`
}

// DeckAntwort trägt ein einzelnes Deck.
type DeckAntwort struct {
	Deck *DeckDTO `json:"deck"`

	// Nachladefehler sagt: angelegt **ist**, nur das Nachlesen danach
	// scheiterte (§7, Regel 3; #1970) — dann trägt `deck` das Deck, wie es
	// angelegt wurde, und `fehler` ist leer.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// MitDecknaht setzt die Naht — für Tests und den Composition Root.
func (a *App) MitDecknaht(naht Decknaht) { a.decks = naht }

// Decks gibt die Decks eines Spiels — ein leeres Spiel heißt: alle.
func (a *App) Decks(ctx context.Context, spiel string) DecksAntwort {
	ctx = a.imProfil(ctx)

	decks, err := a.decks.Decks(ctx, spiel)
	if err != nil {
		return DecksAntwort{Decks: []DeckDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return DecksAntwort{Decks: alsDeckliste(decks)}
}

// Deck gibt ein Deck samt seinen Plätzen.
func (a *App) Deck(ctx context.Context, deckID string) DeckAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DeckAntwort{Fehler: befund}
	}

	deck, err := a.decks.Deck(ctx, kennung)
	if err != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einDeck(deck)
}

// DeckAnlegen legt ein leeres Deck an.
func (a *App) DeckAnlegen(ctx context.Context, name, spiel, notiz string) DeckAntwort {
	ctx = a.imProfil(ctx)

	deck, err := a.decks.Anlegen(ctx, domain.Deck{
		Name: name, Spiel: spiel, Notiz: notiz,
	})
	if err != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einDeck(deck)
}

// DeckUmbenennen schreibt Name und Notiz fort — **gegen die Fassung** (§7,
// Regel 2).
func (a *App) DeckUmbenennen(
	ctx context.Context, deckID, fassung, name, notiz string,
) DeckAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DeckAntwort{Fehler: befund}
	}

	stand, standbefund := benannteKennungAus("fassung", fassung)
	if standbefund != nil {
		return DeckAntwort{Fehler: standbefund}
	}

	// **Das Format reist mit**, weil `Aendern` den ganzen Datensatz schreibt.
	// Ohne diesen Ruf verlöre ein Umbenennen die Formatwahl — und niemand
	// sähe es, bis jemand prüft.
	vorher, lesefehler := a.decks.Deck(ctx, kennung)
	if lesefehler != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, lesefehler)}
	}

	vorher.Fassung = stand
	vorher.Name = name
	vorher.Notiz = notiz

	deck, err := a.decks.Umbenennen(ctx, vorher)
	if err != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einDeck(deck)
}

// DeckKopieren legt eine Kopie eines Decks unter neuem Namen an — DOM-03.AK-02
// (#1698).
func (a *App) DeckKopieren(ctx context.Context, deckID, name string) DeckAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DeckAntwort{Fehler: befund}
	}

	deck, err := a.decks.Kopieren(ctx, kennung, name)
	if err != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einDeck(deck)
}

// DeckLoeschen nimmt ein Deck samt seinen Plätzen heraus.
func (a *App) DeckLoeschen(ctx context.Context, deckID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.decks.Loeschen(ctx, kennung))}
}

// einDeck packt ein Deck in seine Antwort.
func einDeck(deck domain.Deck) DeckAntwort {
	dto := alsDeck(deck)

	return DeckAntwort{Deck: &dto}
}

// alsDeckliste macht aus Decks ihre DTOs — **immer eine Liste, nie `null`**.
func alsDeckliste(decks []domain.Deck) []DeckDTO {
	liste := make([]DeckDTO, 0, len(decks))
	for _, deck := range decks {
		liste = append(liste, alsDeck(deck))
	}

	return liste
}

// alsDeck macht aus einem Deck sein DTO.
func alsDeck(deck domain.Deck) DeckDTO {
	dto := DeckDTO{
		ID:        zahltext(deck.ID),
		Fassung:   zahltext(deck.Fassung),
		Name:      deck.Name,
		Spiel:     deck.Spiel,
		Notiz:     deck.Notiz,
		FormatID:  kennungtext(deck.FormatID),
		Plaetze:   make([]DeckplatzDTO, 0, len(deck.Plaetze)),
		Hauptdeck: deck.Hauptdeck,
		Sideboard: deck.Sideboard,
		Herkunft:  alsHerkunft(deck.Herkunft),
	}

	for _, platz := range deck.Plaetze {
		dto.Plaetze = append(dto.Plaetze, alsDeckplatz(platz))
	}

	return dto
}

// alsDeckplatz macht aus einem Platz sein DTO.
//
// **Einmal für das Deck und für den frisch gelegten Platz**: beim frisch
// gelegten stehen Name und Kosten noch nicht da — `PlatzLegen` schreibt, und
// der Name kommt beim Lesen. Das Fenster lädt danach ohnehin neu (§7, Regel 3).
func alsDeckplatz(platz domain.Deckplatz) DeckplatzDTO {
	return DeckplatzDTO{
		ID:         zahltext(platz.ID),
		KarteID:    kennungtext(platz.KarteID),
		Printing:   platz.Printing,
		DruckID:    kennungtext(platz.DruckID),
		Vorlaeufig: platz.Vorlaeufig(),
		Menge:      platz.Menge,
		Bereich:    string(platz.Bereich),

		// **Name und Kosten, wenn es sie gibt** (#1376), dazu Typzeile und
		// Manawert (U10, #1421).
		Bezeichnung: platz.Bezeichnung,
		Manakosten:  platz.Manakosten,
		Typzeile:    platz.Typzeile,
		Manawert:    platz.Manawert,
	}
}

// zahltext gibt eine Zahl als Text.
//
// **Kennungen reisen als Text**: über die Wails-Bindung käme ein `int64` als
// JavaScript-Zahl an, und die verliert oberhalb von 2⁵³ die letzten Stellen.
func zahltext(zahl int64) string {
	return strconv.FormatInt(zahl, 10)
}

// kennungtext gibt eine Kennung als Text — **leer für die Null**.
//
// **Nicht `"0"`**: eine Null sähe aus wie eine Kennung, und das Fenster
// müsste sie an jeder Stelle als Sonderfall lesen.
func kennungtext(kennung int64) string {
	if kennung == 0 {
		return ""
	}

	return zahltext(kennung)
}
