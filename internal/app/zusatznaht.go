package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Zusatzweg liest und setzt die Zusatzfelder einer Karte — für
// Kommandozeile, MCP und REST
// (#2298).
//
// **Derselbe Dienst wie im Fenster** ([App.Kartenzusatzfelder]) und dieselbe
// Abbildung ([zusatzfelderAus]); zwei Fassungen liefen auseinander, sobald
// eine von beiden ein Feld dazubekäme.
type Zusatzweg struct{ pflege Zusatzwertpflege }

// ZusatznahtUeber baut den Weg über einem Bestand — mit dem Stil **der
// Karte** und dem Schreibweg jedes anderen Felds, samt Urheber.
func ZusatznahtUeber(db *sql.DB, urheber domain.Urheber) (*Zusatzweg, error) {
	pakete, paketfehler := Paketverzeichnis()
	bilder, bildfehler := Bilderverzeichnis()

	if fehler := errors.Join(paketfehler, bildfehler); fehler != nil {
		return nil, fehler
	}

	leser := storage.NeuerSetLeser(db)
	stile := service.NeuerStilspeicher(leser, paketwurzelnFuer(pakete)...).
		MitBildablage(service.NeueBildablage(bilder)).MitStilwerten(leser)

	return &Zusatzweg{pflege: service.NeueZusatzwertpflege(
		leser, stile, storage.NeuerKartenschreiber(db).Als(urheber))}, nil
}

// Felder gibt die Zusatzfelder einer Karte, die sich von Hand setzen lassen.
func (w *Zusatzweg) Felder(ctx context.Context, setID, karteID int64) ([]ZusatzfeldDTO, error) {
	gelesen, err := w.pflege.Felder(ctx, setID, karteID)
	if err != nil {
		return nil, err
	}

	return zusatzfelderAus(gelesen, ""), nil
}

// ZusatzvorwertDTO ist, was vor dem Setzen galt — der Block `extra data` und
// darin das Feld ([service.Zusatzvorwert]).
type ZusatzvorwertDTO struct {
	Block       string
	BlockDa     bool
	Wert        string
	Gespeichert bool
}

// Setzen schreibt den Wert eines Zusatzfelds und gibt, was vorher galt — aus
// demselben Stand, auf den geschrieben wurde.
func (w *Zusatzweg) Setzen(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (ZusatzvorwertDTO, error) {
	vorher, err := w.pflege.Setzen(ctx, setID, karteID, feld, wert)

	return ZusatzvorwertDTO(vorher), err
}
