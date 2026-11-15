// **Die Decklisteneinfuhr für alle vier Flächen** — Werkstück `T8` von Phase
// 5.6 (#1473).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Warum sie neben [Listeneinfuhrnaht] steht und nicht darin
//
// **Weil sie etwas anderes schreibt.** Jene legt Exemplare in die Sammlung —
// Besitz; diese legt Plätze in ein Deck. Das Issue trennt beides
// ausdrücklich: „Bestehende Listeneinfuhr schreibt Sammlung und darf nicht
// direkt als Deckimport verwendet werden."
//
// **Geteilt wird der Weg zur Deutung**: derselbe Scryfall-Klient über
// dieselbe eine Stelle, und damit dieselben Drosseln.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Deckvorschauzeile ist eine Zeile, wie sie ins Deck käme.
type Deckvorschauzeile struct {
	Bereich string `json:"bereich"`
	Menge   int    `json:"menge"`

	// Name, Setcode und Nummer sind, was dastand — nicht, was gefunden wurde.
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`

	// Printing ist der Druck — leer, solange die Zeile nicht eindeutig ist.
	Printing string `json:"printing"`

	// Treffer sind die Drucke zur Wahl.
	//
	// **Ohne `omitempty`**: eine leere Liste muss als `[]` ankommen und nicht
	// als `null` — die Oberfläche zählt sie (§3, „zwischen den Prüfebenen").
	Treffer []Listentreffer `json:"treffer"`
	Drucke  int             `json:"drucke"`

	Befund string `json:"befund"`

	// Uebernehmbar sagt, ob diese Zeile ohne Rückfrage ins Deck kann.
	//
	// **Mitgeschickt und nicht drüben gerechnet** — dieselbe Überlegung wie
	// bei [Listendeutungszeile.Eindeutig]: vier Flächen, die dieselbe Regel
	// je selbst rechnen, rechnen sie irgendwann verschieden.
	Uebernehmbar bool `json:"uebernehmbar"`
}

// Deckvorschau ist, was aus einer Liste würde — ohne dass etwas geschieht.
type Deckvorschau struct {
	DeckID int64 `json:"deckID"`

	// Stand ist der Abdruck des Decks, aus dem diese Vorschau entstand.
	//
	// **Er geht bei der Bestätigung zurück** (§7, Regel 2): wer dazwischen
	// eine Karte hineinlegt, bekommt einen Konflikt statt einer Deckliste
	// über einer anderen.
	Stand string `json:"stand"`

	Zeilen  []Deckvorschauzeile `json:"zeilen"`
	Befunde []string            `json:"befunde"`

	// Hauptdeck und Sideboard sind die Stückzahlen, die hinzukämen.
	Hauptdeck int `json:"hauptdeck"`
	Sideboard int `json:"sideboard"`
}

// Deckeinfuhrzeile ist eine Zeile, für die entschieden ist.
type Deckeinfuhrzeile struct {
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Bereich  string `json:"bereich"`
}

// Deckeinfuhrbericht sagt, was gelegt wurde.
type Deckeinfuhrbericht struct {
	Gelegt int `json:"gelegt"`
	Stueck int `json:"stueck"`

	// Stand ist der Abdruck **nach** dem Legen.
	Stand string `json:"stand"`
}

// Deckeinfuhrnaht ist die Einfuhr, wie eine Fläche sie ruft.
type Deckeinfuhrnaht struct{ einfuhr *service.Deckeinfuhr }

// DeckeinfuhrUeber baut sie über einer Datenbank.
//
// **Derselbe Klient wie überall, über die eine Stelle** — er bringt die
// Drosseln und den Dienstschalter mit (§16, Regel 4).
func DeckeinfuhrUeber(db *sql.DB) (*Deckeinfuhrnaht, error) {
	klient, err := Scryfallklient(NeueEinstellungswahl(
		storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)))
	if err != nil {
		return nil, err
	}

	deutung := service.NeueListeneinfuhr(klient,
		storage.NeuerSammlungsspeicher(db)).MitSammelsuche(klient)

	return &Deckeinfuhrnaht{einfuhr: service.NeueDeckeinfuhr(
		deutung, storage.NeueDeckablage(db))}, nil
}

// Vorschau liest die Liste und zeigt, was ins Deck käme.
func (d *Deckeinfuhrnaht) Vorschau(
	ctx context.Context, deckID int64, text string,
) (*Deckvorschau, error) {
	gedeutet, err := d.einfuhr.Vorschau(ctx, deckID, text)
	if err != nil {
		return nil, err
	}

	vorschau := &Deckvorschau{
		DeckID: gedeutet.DeckID, Stand: gedeutet.Stand,
		Zeilen:    make([]Deckvorschauzeile, 0, len(gedeutet.Zeilen)),
		Befunde:   befundtexte(gedeutet.Befunde),
		Hauptdeck: gedeutet.Hauptdeck, Sideboard: gedeutet.Sideboard,
	}

	for _, zeile := range gedeutet.Zeilen {
		vorschau.Zeilen = append(vorschau.Zeilen, alsVorschauzeile(zeile))
	}

	return vorschau, nil
}

// alsVorschauzeile macht aus einer Zeile ihre Gestalt für die Flächen.
func alsVorschauzeile(zeile service.Deckvorschauzeile) Deckvorschauzeile {
	treffer := make([]Listentreffer, 0, len(zeile.Treffer))
	for _, gefunden := range zeile.Treffer {
		treffer = append(treffer, alsListentreffer(gefunden))
	}

	return Deckvorschauzeile{
		Bereich: string(zeile.Bereich), Menge: zeile.Menge,
		Name: zeile.Name, Setcode: zeile.Setcode, Nummer: zeile.Nummer,
		Printing: zeile.Printing, Treffer: treffer, Drucke: zeile.Drucke,
		Befund: zeile.Befund, Uebernehmbar: zeile.Uebernehmbar,
	}
}

// Uebernehmen legt die bestätigten Zeilen ins Deck.
func (d *Deckeinfuhrnaht) Uebernehmen(
	ctx context.Context, deckID int64, stand string, zeilen []Deckeinfuhrzeile,
) (*Deckeinfuhrbericht, error) {
	bestaetigt := make([]service.Deckzeile, 0, len(zeilen))

	for _, zeile := range zeilen {
		bestaetigt = append(bestaetigt, service.Deckzeile{
			Printing: zeile.Printing, Menge: zeile.Menge,
			Bereich: domain.Deckbereich(zeile.Bereich),
		})
	}

	bericht, err := d.einfuhr.Uebernehmen(ctx, deckID, stand, bestaetigt)
	if err != nil {
		return nil, err
	}

	return &Deckeinfuhrbericht{
		Gelegt: bericht.Gelegt, Stueck: bericht.Stueck, Stand: bericht.Stand,
	}, nil
}
