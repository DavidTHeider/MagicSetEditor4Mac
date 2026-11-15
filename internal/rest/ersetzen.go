package rest

// **Suchen und Ersetzen über die Rückschleife** — Werkstück 2 von Phase 3.7
// (#1168).

import (
	"context"
	"net/http"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fremdersetzung ist, was der Dienst vom Ersetzen braucht.
type Fremdersetzung interface {
	ErsetzungVorschau(ctx context.Context, setID int64, frage domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error)
	ErsetzungAusfuehren(ctx context.Context, setID int64, frage domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error)
}

// ersetzungVorschau sagt, was sich ändern würde — und schreibt nichts.
//
// # Warum POST und nicht GET
//
// **Weil die Frage in den Rumpf gehört.** Sie trägt fünf Angaben, darunter
// zwei freie Texte; als Abfrage wären sie eine Adresse, die in jedem Protokoll
// steht — und ein gesuchter Kartentext gehört niemandem außer dem Bestand.
// Geändert wird trotzdem nichts; das steht in der Antwort und nicht in der
// Methode. Dieselbe Begründung wie bei `POST /sets/{setID}/ablage/kopieren`.
func (d *Dienst) ersetzungVorschau(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.ersetze(schreiben, anfrage, d.naehte.Ersetzung.ErsetzungVorschau)
}

// ersetzungAusfuehren ersetzt und gibt dieselbe Übersicht zurück.
func (d *Dienst) ersetzungAusfuehren(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.ersetze(schreiben, anfrage, d.naehte.Ersetzung.ErsetzungAusfuehren)
}

// ersetze ist der gemeinsame Rumpf beider Wege.
func (d *Dienst) ersetze(
	schreiben http.ResponseWriter, anfrage *http.Request,
	arbeit func(context.Context, int64, domain.Ersetzungsfrage) (
		domain.Ersetzungsvorschau, error),
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Text               string `json:"text"`
		Ersatz             string `json:"ersatz"`
		Feld               string `json:"feld"`
		GrossKleinBeachten bool   `json:"grossKleinBeachten"`
		GanzesWort         bool   `json:"ganzesWort"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	frage := domain.Ersetzungsfrage{
		Text:               rumpf.Text,
		Ersatz:             rumpf.Ersatz,
		Feld:               rumpf.Feld,
		GrossKleinBeachten: rumpf.GrossKleinBeachten,
		GanzesWort:         rumpf.GanzesWort,
	}

	if befunde := frage.Validate(); len(befunde) > 0 {
		fehlerhaft(schreiben, befunde[0])

		return
	}

	schau, fehler := arbeit(anfrage.Context(), setID, frage)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"stellen": ersetzungsstellen(schau.Stellen),
		"karten":  schau.Karten,
		"treffer": schau.Treffer,
		"weitere": schau.Weitere,
	})
}

// ersetzungsstellen macht aus den Stellen eine Liste — **immer eine**.
func ersetzungsstellen(stellen []domain.Ersetzungsstelle) []map[string]any {
	liste := make([]map[string]any, 0, len(stellen))

	for _, stelle := range stellen {
		liste = append(liste, map[string]any{
			"karteID": stelle.KarteID,
			"nummer":  stelle.Nummer + 1,
			"feld":    stelle.Feld,
			"vorher":  stelle.Vorher,
			"nachher": stelle.Nachher,
			"treffer": stelle.Treffer,
		})
	}

	return liste
}
