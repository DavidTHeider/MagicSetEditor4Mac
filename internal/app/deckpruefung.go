package app

// **Die Prüf- und Abgleichnaht** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei neben `deck.go`.** Der Schnitt folgt der Frage: dort
// steht, wie ein Deck entsteht und was darin liegt; hier, was man darüber
// erfährt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Deckpruefnaht prüft ein Deck gegen sein Format.
type Deckpruefnaht interface {
	Pruefe(ctx context.Context, deckID int64) (
		domain.Deckurteil, []domain.Befund, error)
}

// Deckabgleichnaht gleicht ein Deck gegen die Sammlung ab.
type Deckabgleichnaht interface {
	Fuer(ctx context.Context, deckID int64) (service.Habenstand, error)
}

// KartenurteilDTO ist die Antwort für einen Platz.
type KartenurteilDTO struct {
	PlatzID     string `json:"platzID"`
	Bezeichnung string `json:"bezeichnung"`
	Menge       int    `json:"menge"`
	Bereich     string `json:"bereich"`

	// Zulaessigkeit ist eines von Scryfalls vier Wörtern — `legal`,
	// `not_legal`, `restricted`, `banned` — oder `ungeprüft`.
	//
	// **Nicht übersetzt.** Eine Übersetzung machte aus einer Auskunft eine
	// Auslegung; das Fenster hat für die fünf einen Katalogtext.
	Zulaessigkeit string `json:"zulaessigkeit"`

	Grund  string `json:"grund"`
	Quelle string `json:"quelle"`
}

// GroessenbefundDTO ist ein Verstoß gegen eine Zahl des Formats.
type GroessenbefundDTO struct {
	Feld    string `json:"feld"`
	Meldung string `json:"meldung"`
}

// DeckurteilDTO ist das ganze Ergebnis einer Formatprüfung.
type DeckurteilDTO struct {
	Formatname string `json:"formatname"`

	Hauptdeck int `json:"hauptdeck"`
	Sideboard int `json:"sideboard"`

	Urteile  []KartenurteilDTO   `json:"urteile"`
	Groessen []GroessenbefundDTO `json:"groessen"`

	// Zulaessig sagt, ob nichts dagegen steht. **`ungeprüft` steht nicht
	// dagegen** — das ist die Zusage aus Vermerk 0061.
	Zulaessig  bool `json:"zulaessig"`
	Ungeprueft int  `json:"ungeprueft"`
}

// DeckurteilAntwort trägt das Urteil samt den Befunden des Wegs dorthin.
type DeckurteilAntwort struct {
	Urteil *DeckurteilDTO `json:"urteil"`

	// Befunde sagen, was unterwegs schiefging — eine Gegenstelle, die schwieg,
	// ein Wort, das niemand kennt. **Kein Fehler**: die Prüfung lief weiter.
	//
	// **Als Sätze und nicht als Verbund**, wie überall in diesem Paket: was
	// das Fenster zeigt, ist die Meldung.
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// HabenzeileDTO ist die Antwort für einen Druck im Deck.
type HabenzeileDTO struct {
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441);
	// Vorlaeufig sagt, dass er keinen Scryfall-Alias hat.
	DruckID    string `json:"druckID"`
	Vorlaeufig bool   `json:"vorlaeufig"`

	Bezeichnung string `json:"bezeichnung"`
	Gebraucht   int    `json:"gebraucht"`
	Vorhanden   int    `json:"vorhanden"`
	Fehlend     int    `json:"fehlend"`
	Grund       string `json:"grund"`
}

// HabenstandAntwort trägt den Abgleich gegen die Sammlung.
type HabenstandAntwort struct {
	Zeilen []HabenzeileDTO `json:"zeilen"`

	// Fehlend und Unbekannt stehen **getrennt**: „vier fehlen" und „über vier
	// ist nichts bekannt" sind zwei Auskünfte.
	Fehlend   int `json:"fehlend"`
	Unbekannt int `json:"unbekannt"`

	Fehler *Fehler `json:"fehler"`
}

// MitDeckpruefung setzt die Prüfnaht — für Tests und den Composition Root.
func (a *App) MitDeckpruefung(pruefung Deckpruefnaht) { a.deckpruefung = pruefung }

// MitDeckabgleich setzt die Abgleichnaht.
//
// **Getrennt von der Prüfnaht**, und das ist der Unterschied: der Abgleich
// liest die **eigene** Sammlung und steht immer bereit; die Prüfung braucht
// Scryfall und fehlt, wenn es keinen Zugang gibt.
func (a *App) MitDeckabgleich(abgleich Deckabgleichnaht) { a.deckabgleich = abgleich }

// DeckPruefen beurteilt ein Deck gegen sein gewähltes Format.
func (a *App) DeckPruefen(ctx context.Context, deckID string) DeckurteilAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DeckurteilAntwort{Befunde: []string{}, Fehler: befund}
	}

	urteil, befunde, err := a.deckpruefung.Pruefe(ctx, kennung)
	if err != nil {
		return DeckurteilAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	dto := alsDeckurteil(urteil)

	return DeckurteilAntwort{Urteil: &dto, Befunde: alsMeldungen(befunde)}
}

// DeckAbgleichen sagt, was von einem Deck in der Sammlung liegt.
func (a *App) DeckAbgleichen(ctx context.Context, deckID string) HabenstandAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return HabenstandAntwort{Zeilen: []HabenzeileDTO{}, Fehler: befund}
	}

	stand, err := a.deckabgleich.Fuer(ctx, kennung)
	if err != nil {
		return HabenstandAntwort{Zeilen: []HabenzeileDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	antwort := HabenstandAntwort{
		Zeilen:    make([]HabenzeileDTO, 0, len(stand.Zeilen)),
		Fehlend:   stand.Fehlend,
		Unbekannt: stand.Unbekannt,
	}

	for _, zeile := range stand.Zeilen {
		antwort.Zeilen = append(antwort.Zeilen, alsHabenzeile(zeile))
	}

	return antwort
}

// alsHabenzeile bildet eine Zeile des Abgleichs ab.
//
// **Herausgezogen, weil es einen zweiten Aufrufer gibt** (U9, #1420): die
// Zuordnung zeigt dieselbe Zeile mit drei Zahlen mehr.
func alsHabenzeile(zeile service.Habenzeile) HabenzeileDTO {
	return HabenzeileDTO{
		Printing:    zeile.Printing,
		DruckID:     kennungtext(zeile.DruckID),
		Vorlaeufig:  zeile.Vorlaeufig,
		Bezeichnung: zeile.Bezeichnung,
		Gebraucht:   zeile.Gebraucht,
		Vorhanden:   zeile.Vorhanden,
		Fehlend:     zeile.Fehlend,
		Grund:       zeile.Grund,
	}
}

// alsDeckurteil macht aus einem Urteil sein DTO.
func alsDeckurteil(urteil domain.Deckurteil) DeckurteilDTO {
	dto := DeckurteilDTO{
		Formatname: urteil.Formatname,
		Hauptdeck:  urteil.Hauptdeck,
		Sideboard:  urteil.Sideboard,
		Urteile:    make([]KartenurteilDTO, 0, len(urteil.Urteile)),
		Groessen:   make([]GroessenbefundDTO, 0, len(urteil.Groessen)),
		Zulaessig:  urteil.Zulaessig(),
		Ungeprueft: urteil.Ungeprueft(),
	}

	for _, eines := range urteil.Urteile {
		dto.Urteile = append(dto.Urteile, KartenurteilDTO{
			PlatzID:       zahltext(eines.PlatzID),
			Bezeichnung:   eines.Bezeichnung,
			Menge:         eines.Menge,
			Bereich:       string(eines.Bereich),
			Zulaessigkeit: string(eines.Zulaessigkeit),
			Grund:         eines.Grund,
			Quelle:        eines.Quelle,
		})
	}

	for _, befund := range urteil.Groessen {
		dto.Groessen = append(dto.Groessen, GroessenbefundDTO{
			Feld: befund.Feld, Meldung: befund.Meldung,
		})
	}

	return dto
}
