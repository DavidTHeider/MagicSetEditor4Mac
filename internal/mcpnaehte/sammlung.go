// **Die Sammlung für den MCP-Server** —
// #1165.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// sammlung reicht die fünf Züge an die Naht durch.
type sammlung struct{ db *sql.DB }

func (s sammlung) Sammlung(ctx context.Context) ([]mcp.Exemplarsicht, error) {
	naht, err := app.SammlungsnahtUeber(s.db)
	if err != nil {
		return nil, err
	}

	dtos, lesefehler := naht.Sammlung(ctx)
	if lesefehler != nil {
		return nil, lesefehler
	}

	return alsExemplarsichten(dtos), nil
}

// Sammlungsfenster gibt den Ausschnitt einer Sicht — Werkstück
// #1500.
func (s sammlung) Sammlungsfenster(
	ctx context.Context, suche, ausfuehrung, sortierung string, ab, anzahl int,
) (fenster []mcp.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	naht, err := app.SammlungsnahtUeber(s.db)
	if err != nil {
		return nil, domain.Fensterzahlen{}, err
	}

	dtos, zahlen, lesefehler := naht.Sammlungsfenster(ctx,
		domain.Sammlungssicht{
			Suche:       suche,
			Ausfuehrung: domain.Ausfuehrung(ausfuehrung),
			Sortierung:  domain.Sortierung(sortierung),
		}, ab, anzahl)
	if lesefehler != nil {
		return nil, domain.Fensterzahlen{}, lesefehler
	}

	return alsExemplarsichten(dtos), zahlen, nil
}

func (s sammlung) SammlungFuer(ctx context.Context, printing string) (
	[]mcp.Exemplarsicht, error,
) {
	naht, err := app.SammlungsnahtUeber(s.db)
	if err != nil {
		return nil, err
	}

	dtos, lesefehler := naht.SammlungFuer(ctx, printing)
	if lesefehler != nil {
		return nil, lesefehler
	}

	return alsExemplarsichten(dtos), nil
}

func (s sammlung) ExemplarAnlegen(ctx context.Context, exemplar mcp.Exemplarsicht) (
	int64, error,
) {
	naht, err := app.SammlungsnahtUeber(s.db)
	if err != nil {
		return 0, err
	}

	return naht.ExemplarAnlegen(ctx, alsExemplarDTO(exemplar))
}

func (s sammlung) ExemplarAendern(ctx context.Context, exemplar mcp.Exemplarsicht) error {
	naht, err := app.SammlungsnahtUeber(s.db)
	if err != nil {
		return err
	}

	return naht.ExemplarAendern(ctx, alsExemplarDTO(exemplar))
}

func (s sammlung) ExemplarLoeschen(ctx context.Context, id, fassung int64) error {
	naht, err := app.SammlungsnahtUeber(s.db)
	if err != nil {
		return err
	}

	return naht.ExemplarLoeschen(ctx, id, fassung)
}

// alsExemplarsichten übersetzt die DTOs in die Sicht dieser Fläche.
//
// **Feld für Feld und nicht über JSON.** Ein Umweg über `Marshal` sähe kürzer
// aus und verlöre beim ersten Feld, dessen Marken auseinanderlaufen, genau
// dieses Feld — still.
func alsExemplarsichten(dtos []app.ExemplarDTO) []mcp.Exemplarsicht {
	sichten := make([]mcp.Exemplarsicht, 0, len(dtos))

	for _, dto := range dtos {
		sichten = append(sichten, mcp.Exemplarsicht{
			ID:              dto.ID,
			Fassung:         dto.Fassung,
			Printing:        dto.Printing,
			DruckID:         dto.DruckID,
			Vorlaeufig:      dto.Vorlaeufig,
			Menge:           dto.Menge,
			Zustand:         dto.Zustand,
			Sprache:         dto.Sprache,
			Foil:            dto.Finish == string(domain.FinishFoil),
			Finish:          dto.Finish,
			Erwerbsdatum:    dto.Erwerbsdatum,
			Erwerbspreis:    dto.Erwerbspreis,
			Erwerbswaehrung: dto.Erwerbswaehrung,
			Notiz:           dto.Notiz,
			Name:            dto.Name,
			Setcode:         dto.Setcode,
			Nummer:          dto.Nummer,
			Lagerort:        dto.Lagerort,
			Wunsch:          dto.Wunsch,
			Wunschrang:      dto.Wunschrang,
			Wunschart:       dto.Wunschart,
			Bezeichnung:     dto.Bezeichnung,
		})
	}

	return sichten
}

// alsExemplarDTO übersetzt zurück.
func alsExemplarDTO(sicht mcp.Exemplarsicht) app.ExemplarDTO {
	return app.ExemplarDTO{
		ID:              sicht.ID,
		Fassung:         sicht.Fassung,
		Printing:        sicht.Printing,
		DruckID:         sicht.DruckID,
		Menge:           sicht.Menge,
		Zustand:         sicht.Zustand,
		Sprache:         sicht.Sprache,
		Finish:          app.AusfuehrungAus(sicht.Finish, sicht.Foil),
		Erwerbsdatum:    sicht.Erwerbsdatum,
		Erwerbspreis:    sicht.Erwerbspreis,
		Erwerbswaehrung: sicht.Erwerbswaehrung,
		Notiz:           sicht.Notiz,
		Name:            sicht.Name,
		Setcode:         sicht.Setcode,
		Nummer:          sicht.Nummer,
		Lagerort:        sicht.Lagerort,
		Wunsch:          sicht.Wunsch,
		Wunschrang:      sicht.Wunschrang,
		Wunschart:       sicht.Wunschart,
	}
}
