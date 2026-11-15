// **Suchen und Ersetzen über ein Set** — Werkstück 2 von Phase 3.7
// (#1168).

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ersetzungsquelle ist, was das Fenster vom Ersetzen braucht.
type Ersetzungsquelle interface {
	Vorschau(ctx context.Context, setID int64, frage domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error)
	Ausfuehren(ctx context.Context, setID int64, frage domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error)
}

// MitErsetzung setzt den Anwendungsfall — für Tests und den Composition Root.
func (a *App) MitErsetzung(quelle Ersetzungsquelle) { a.ersetzung = quelle }

// ErsetzungsstelleDTO ist ein Feldwert, der sich ändern würde.
type ErsetzungsstelleDTO struct {
	KarteID string `json:"karteID"`

	// Nummer ist die Stelle der Karte im Set — **eins-basiert**, wie die
	// Kartenliste sie zeigt.
	Nummer int `json:"nummer"`

	Feld string `json:"feld"`

	// Vorher und Nachher stehen **ohne Auszeichnungen** da: ein `<i>` in einer
	// Vorschau sieht wie ein Fehler aus.
	Vorher  string `json:"vorher"`
	Nachher string `json:"nachher"`

	Treffer int `json:"treffer"`
}

// ErsetzungsAntwort ist, was eine Ersetzung ändern würde — oder geändert hat.
type ErsetzungsAntwort struct {
	// Stellen ist **begrenzt** (`domain.GrenzeErsetzungsvorschau`); Karten und
	// Treffer zählen alles.
	Stellen []ErsetzungsstelleDTO `json:"stellen"`

	Karten  int  `json:"karten"`
	Treffer int  `json:"treffer"`
	Weitere bool `json:"weitere"`

	// Stand ist der Fingerabdruck der betroffenen Stellen (#1241). Das Fenster
	// gibt ihn beim Bestätigen zurück; gilt er nicht mehr, wird abgewiesen.
	Stand string `json:"stand"`

	Fehler *Fehler `json:"fehler"`
}

// ErsetzungVorschau sagt, was sich ändern würde — und schreibt nichts.
func (a *App) ErsetzungVorschau(
	ctx context.Context, setID string, frage ErsetzungsfrageDTO,
) ErsetzungsAntwort {
	ctx = a.imProfil(ctx)

	return a.ersetze(ctx, setID, frage, func(
		quelle Ersetzungsquelle, kennung int64, gestellt domain.Ersetzungsfrage,
	) (domain.Ersetzungsvorschau, error) {
		return quelle.Vorschau(ctx, kennung, gestellt)
	})
}

// ErsetzungAusfuehren ersetzt und gibt dieselbe Übersicht zurück.
func (a *App) ErsetzungAusfuehren(
	ctx context.Context, setID string, frage ErsetzungsfrageDTO,
) ErsetzungsAntwort {
	ctx = a.imProfil(ctx)

	return a.ersetze(ctx, setID, frage, func(
		quelle Ersetzungsquelle, kennung int64, gestellt domain.Ersetzungsfrage,
	) (domain.Ersetzungsvorschau, error) {
		return quelle.Ausfuehren(ctx, kennung, gestellt)
	})
}

// ErsetzungsfrageDTO ist die Frage, wie das Fenster sie stellt.
//
// **Ein eigener Verbund und keine sechs Stellen**: sechs Werte hintereinander
// sagen am Aufrufort nichts mehr, und JavaScript hat keine benannten
// Argumente. Dieselbe Begründung wie bei `rahmen/tastatur.ts`.
type ErsetzungsfrageDTO struct {
	Text               string `json:"text"`
	Ersatz             string `json:"ersatz"`
	Feld               string `json:"feld"`
	GrossKleinBeachten bool   `json:"grossKleinBeachten"`
	GanzesWort         bool   `json:"ganzesWort"`

	// Stand ist der Fingerabdruck, den die Vorschau mitgab —
	// #1241.
	//
	// **Leer heißt: ohne Prüfung.** Warum, steht bei
	// [domain.Ersetzungsfrage.Stand].
	Stand string `json:"stand"`
}

// ersetze ist der gemeinsame Rumpf beider Wege.
func (a *App) ersetze(
	ctx context.Context, setID string, frage ErsetzungsfrageDTO,
	arbeit func(Ersetzungsquelle, int64, domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error),
) ErsetzungsAntwort {
	leer := ErsetzungsAntwort{Stellen: []ErsetzungsstelleDTO{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.ersetzung == nil {
		leer.Fehler = Uebersetze(ctx, fmt.Errorf(
			"%w: das Ersetzen ist nicht eingerichtet", domain.ErrValidation))

		return leer
	}

	schau, err := arbeit(a.ersetzung, kennung, domain.Ersetzungsfrage{
		Text:               frage.Text,
		Ersatz:             frage.Ersatz,
		Feld:               frage.Feld,
		GrossKleinBeachten: frage.GrossKleinBeachten,
		GanzesWort:         frage.GanzesWort,
		Stand:              frage.Stand,
	})
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	return ErsetzungsAntwort{
		Stellen: stellenDTO(schau.Stellen),
		Karten:  schau.Karten,
		Treffer: schau.Treffer,
		Weitere: schau.Weitere,
		Stand:   schau.Stand,
	}
}

// stellenDTO übersetzt die Stellen — **immer eine Liste**, nie `nil`.
//
// **Weil `null` und `[]` im Fenster verschieden sind** (§3, „zwischen den
// Prüfebenen"): eine leere Liste, die als `null` ankommt, kostet jeden Leser
// einen Zweig.
func stellenDTO(stellen []domain.Ersetzungsstelle) []ErsetzungsstelleDTO {
	umgesetzt := make([]ErsetzungsstelleDTO, 0, len(stellen))

	for _, stelle := range stellen {
		umgesetzt = append(umgesetzt, ErsetzungsstelleDTO{
			KarteID: strconv.FormatInt(stelle.KarteID, 10),
			// **Eins-basiert**, wie die Liste zählt (#666).
			Nummer:  stelle.Nummer + 1,
			Feld:    stelle.Feld,
			Vorher:  stelle.Vorher,
			Nachher: stelle.Nachher,
			Treffer: stelle.Treffer,
		})
	}

	return umgesetzt
}
