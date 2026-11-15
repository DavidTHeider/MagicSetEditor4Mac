package app

// **Turniere an der Naht** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Warum jede Kennung Text ist
//
// **Dieselbe Regel wie überall an dieser Naht**: JavaScript kennt nur
// `float64`, und eine Kennung jenseits von 2^53 käme gerundet an. Ein Turnier
// erreicht diese Größe nie — die Regel gilt trotzdem, weil eine Ausnahme
// erklärt werden müsste und die Erklärung teurer wäre als die Zeile.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Turnierleitung ist, was diese Naht vom Dienst braucht.
type Turnierleitung interface {
	Anlegen(ctx context.Context, turnier domain.Turnier) (domain.Turnier, error)
	Turnier(ctx context.Context, kennung int64) (domain.Turnier, error)
	Turniere(ctx context.Context, spiel string) ([]domain.Turnier, error)
	Anmelden(ctx context.Context, turnierID int64, teilnehmer domain.Teilnehmer) (
		domain.Teilnehmer, error)
	Aussteigen(ctx context.Context, turnierID, teilnehmerID int64,
		ausgestiegen bool) error
	Rangliste(ctx context.Context, turnierID int64) ([]domain.Rang, error)
	Paarungsvorschlag(ctx context.Context, turnierID int64) ([]domain.Paarung, error)
	RundeLegen(ctx context.Context, turnierID int64) (domain.Runde, error)
	ErgebnisSetzen(ctx context.Context, turnierID, fassung, paarungID int64,
		stand *domain.Ergebnis) error
	Loeschen(ctx context.Context, kennung int64) error
	Kopieren(ctx context.Context, quelle int64, name string, mitTeilnehmenden bool) (
		domain.Turnier, error)
	Korrekturfolgen(ctx context.Context, turnierID, paarungID int64,
		stand *domain.Ergebnis) (domain.Korrekturfolgen, error)
	Abschliessen(ctx context.Context, turnierID, fassung int64) error
	Wiedereroeffnen(ctx context.Context, turnierID, fassung int64) error
	Protokoll(ctx context.Context, turnierID int64) ([]domain.Protokolleintrag, error)

	// Ausgabe, Einlesen und Bericht — DOM-04.AK-03 (#1699), Vermerk 0140.
	Ausgeben(ctx context.Context, turnierID int64) ([]byte, error)
	Einlesen(ctx context.Context, roh []byte) (domain.Turnier, []domain.Einfuhrbefund, error)
	Bericht(ctx context.Context, turnierID int64, texte domain.Berichttexte) (string, error)
}

// TurnierDTO ist ein Turnier für die Oberfläche.
type TurnierDTO struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Name  string `json:"name"`
	Spiel string `json:"spiel"`
	Notiz string `json:"notiz"`

	// Einzelspiel heißt: ein Match besteht aus **einem** Spiel.
	//
	// **Es entscheidet die Anzeige, nicht die Rechnung** (MTR §3.1): Spiel-
	// und Gegnerspielquote sagen dann nichts, weil sie Kopien der Matchzahlen
	// sind. Gerechnet werden sie trotzdem.
	Einzelspiel bool `json:"einzelspiel"`

	// FormatID ist das Format, nach dem gespielt wird — **leer, wenn keines
	// gewählt ist** (DOM-04.AK-01, #1699).
	FormatID string `json:"formatID"`

	// Abgeschlossen ist das Datum des Abschlusses — **leer heißt offen**
	// (DOM-04.AK-04, #1699, Vermerk 0139).
	Abgeschlossen string `json:"abgeschlossen"`

	Teilnehmer []TeilnehmerDTO `json:"teilnehmer"`
	Runden     []RundeDTO      `json:"runden"`

	// EmpfohleneRunden ist die Rundenzahl aus Appendix E der *Magic
	// Tournament Rules* — **null heißt: keine Empfehlung**.
	//
	// **Ein Vorschlag und keine Vorschrift**: die Staffel gilt für Premier
	// tournaments und sonst „at the Tournament Organizer's discretion". Wer
	// weiterspielen will, kann.
	EmpfohleneRunden int `json:"empfohleneRunden"`
}

// TeilnehmerDTO ist ein Teilnehmer samt seiner eingereichten Liste.
type TeilnehmerDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// DeckID ist die **Herkunft** der Liste, nicht ihre Quelle — leer, wenn
	// keines hinterlegt wurde.
	DeckID string `json:"deckID"`

	// Deckabzug ist die Liste, wie sie beim Anmelden aussah.
	Deckabzug []AbzugsplatzDTO `json:"deckabzug"`

	Ausgestiegen bool `json:"ausgestiegen"`
	Stelle       int  `json:"stelle"`
}

// AbzugsplatzDTO ist eine Zeile der eingereichten Liste.
//
// **Text und keine Verweise** (Vermerk 0062, Entscheidung 3): sie überlebt
// jedes Löschen.
type AbzugsplatzDTO struct {
	ID          string `json:"id"`
	Bezeichnung string `json:"bezeichnung"`
	Printing    string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441), als
	// Zeichenkette — leer bei einer eigenen Karte.
	DruckID string `json:"druckID"`

	Menge   int    `json:"menge"`
	Bereich string `json:"bereich"`
}

// RundeDTO ist eine Runde samt ihren Paarungen.
type RundeDTO struct {
	ID        string       `json:"id"`
	Nummer    int          `json:"nummer"`
	Paarungen []PaarungDTO `json:"paarungen"`
}

// PaarungDTO ist ein Match — oder ein Bye.
type PaarungDTO struct {
	ID string `json:"id"`

	A string `json:"a"`

	// B ist **leer bei einem Bye**.
	B string `json:"b"`

	// Ergebnis ist der Spielstand — **nil, solange nicht gespielt ist**.
	Ergebnis *ErgebnisDTO `json:"ergebnis"`
}

// ErgebnisDTO ist ein Spielstand, kein Sieger.
type ErgebnisDTO struct {
	SpieleA       int `json:"spieleA"`
	SpieleB       int `json:"spieleB"`
	Unentschieden int `json:"unentschieden"`
}

// RangDTO ist eine Zeile der Rangliste.
//
// **Die Quoten kommen gerundet heraus, gerechnet wird ungerundet** (Vermerk
// 0062, Entscheidung 6). Vier Nachkommastellen, weil zwei Teilnehmer, die
// sich in der vierten unterscheiden, nicht gleichauf sind.
type RangDTO struct {
	TeilnehmerID string `json:"teilnehmerID"`
	Name         string `json:"name"`
	Platz        int    `json:"platz"`

	Siege         int `json:"siege"`
	Niederlagen   int `json:"niederlagen"`
	Unentschieden int `json:"unentschieden"`
	Byes          int `json:"byes"`

	Matchpunkte int `json:"matchpunkte"`

	Matchquote       float64 `json:"matchquote"`
	Gegnermatchquote float64 `json:"gegnermatchquote"`
	Spielquote       float64 `json:"spielquote"`
	Gegnerspielquote float64 `json:"gegnerspielquote"`

	Ausgestiegen bool `json:"ausgestiegen"`
}

// MitMatchnaht setzt die Matchleitung — für Läufe ohne Bestand (§3).
func (a *App) MitMatchnaht(leitung Matchleitung) *App {
	a.matchleitung = leitung

	return a
}

// matchnaht gibt die eingesetzte Leitung.
func (a *App) matchnaht() Matchleitung { return a.matchleitung }
