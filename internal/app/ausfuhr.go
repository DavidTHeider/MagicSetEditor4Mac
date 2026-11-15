package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Exporteur schreibt ein Set hinaus — die Naht zu `internal/service`.
type Exporteur interface {
	MTGJSON(ctx context.Context, setID int64, spiel, pfad string) (*service.Ausfuhr, error)
}

// AusfuhrAntwort ist der Ausgang eines Exports.
type AusfuhrAntwort struct {
	Pfad   string `json:"pfad"`
	Karten int    `json:"karten"`

	// Befunde sind die Stellen, an denen etwas **nicht abbildbar** war.
	//
	// MTG JSON beschreibt Karten, nicht ihre Darstellung; was MSE führt, hat
	// dort nicht überall einen Platz. Was verlorengeht, steht hier — nicht in
	// einem Kommentar im Code (#142).
	Befunde []string `json:"befunde"`

	// Abgebrochen sagt, dass niemand einen Ort gewählt hat.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// MTGJSONSchreiben schreibt ein Set als MTG JSON an einen Pfad.
func (a *App) MTGJSONSchreiben(ctx context.Context, setID, spiel, pfad string) AusfuhrAntwort {
	ctx = a.imProfil(ctx)

	leer := AusfuhrAntwort{Befunde: []string{}}

	if pfad == "" {
		leer.Abgebrochen = true

		return leer
	}

	kennung, err := strconv.ParseInt(setID, 10, 64)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	ausfuhr, err := a.exporteur.MTGJSON(ctx, kennung, spiel, pfad)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	for _, befund := range ausfuhr.Befunde {
		leer.Befunde = append(leer.Befunde, befund.String())
	}

	a.arbeitVermerken(ctx, domain.ArbeitsartSet, setID, domain.VorgangAusgegeben)

	leer.Pfad = ausfuhr.Pfad
	leer.Karten = ausfuhr.Karten

	return leer
}

// MitExporteur setzt die Naht zur Ausfuhr.
func (a *App) MitExporteur(exporteur Exporteur) { a.exporteur = exporteur }
