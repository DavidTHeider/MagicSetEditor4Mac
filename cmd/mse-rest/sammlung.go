// **Die Sammlung und die Preise für den Dienst auf der Rückschleife** —
// #1165.
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// sammlung reicht die fünf Züge an die Naht durch.
type sammlung struct{ db *sql.DB }

func (s sammlung) Sammlung(ctx context.Context) ([]rest.Exemplarsicht, error) {
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
) (fenster []rest.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	naht, fehler := app.SammlungsnahtUeber(s.db)
	if fehler != nil {
		return nil, domain.Fensterzahlen{}, fehler
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
	[]rest.Exemplarsicht, error,
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

func (s sammlung) ExemplarAnlegen(ctx context.Context, exemplar rest.Exemplarsicht) (
	int64, error,
) {
	naht, err := app.SammlungsnahtUeber(s.db)
	if err != nil {
		return 0, err
	}

	return naht.ExemplarAnlegen(ctx, alsExemplarDTO(exemplar))
}

func (s sammlung) ExemplarAendern(ctx context.Context, exemplar rest.Exemplarsicht) error {
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
func alsExemplarsichten(dtos []app.ExemplarDTO) []rest.Exemplarsicht {
	sichten := make([]rest.Exemplarsicht, 0, len(dtos))

	for _, dto := range dtos {
		sichten = append(sichten, rest.Exemplarsicht{
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
func alsExemplarDTO(sicht rest.Exemplarsicht) app.ExemplarDTO {
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

// preise reicht die vier Züge an die Naht durch.
// preise ist die Naht zu Preisen, Markt und KI.
//
// **Sie lebt so lange wie der Dienst**: `dienst.go` legt sie **einmal** beim
// Start an. Das ist der Grund für [preise.kianwendung] — siehe dort.
type preise struct {
	db *sql.DB

	// anwendung ist die **eine** KI-Anwendung dieses Dienstes.
	anwendung *app.App
}

// kianwendung gibt die Anwendung dieses Dienstes — und baut sie beim ersten
// Mal.
//
// # Warum sie gehalten wird
//
// **Weil [service.KIAuftragswerk] seine offenen Aufträge im Prozess hält**
// und nicht im Bestand. Eine neue Anwendung je Ruf hätte ein neues Werk, und
// eine Kennung aus `POST /ki/batch` wäre an `POST /ki/gruppe` nicht
// einlösbar — der Dienst gäbe Zettel aus, die er selbst nicht annimmt.
//
// **Faul gebaut**: eine Sitzung, die nie ein Modell fragt, baut sie nicht.
func (p *preise) kianwendung() *app.App {
	if p.anwendung == nil {
		p.anwendung = app.KIAnwendungUeber(p.db)
	}

	return p.anwendung
}

func (p *preise) PreiseHolen(ctx context.Context, printing string) (
	[]rest.Preissicht, []string, error,
) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return nil, nil, err
	}

	staende, befunde, holfehler := naht.PreiseHolen(ctx, printing)
	if holfehler != nil {
		return nil, befunde, holfehler
	}

	return alsPreissichten(staende), befunde, nil
}

func (p *preise) Preise(ctx context.Context, printing string) ([]rest.Preissicht, error) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	staende, lesefehler := naht.Preise(ctx, printing)
	if lesefehler != nil {
		return nil, lesefehler
	}

	return alsPreissichten(staende), nil
}

func (p *preise) Preisverlauf(ctx context.Context, reihe rest.Preisreihe) (
	[]rest.Preissicht, error,
) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	staende, lesefehler := naht.Preisverlauf(ctx, app.PreisreiheDTO{
		Printing: reihe.Printing, Quelle: reihe.Quelle, Markt: reihe.Markt,
		Art: reihe.Art, Waehrung: reihe.Waehrung, Finish: reihe.Finish,
		Zustand: reihe.Zustand, Sprache: reihe.Sprache,
	})
	if lesefehler != nil {
		return nil, lesefehler
	}

	return alsPreissichten(staende), nil
}

// Preisbulk holt eine ganze Preisdatei und legt ab, was darin steht.
func (p *preise) Preisbulk(ctx context.Context, quelle, umfang string) (
	rest.Bulkbilanzsicht, error,
) {
	antwort := app.PreisbulkAnwendungUeber(p.db).Preisbulk(ctx, quelle, umfang)
	if antwort.Fehler != nil {
		return rest.Bulkbilanzsicht{}, antwort.Fehler.AlsFehler()
	}

	return rest.Bulkbilanzsicht{
		Quelle: antwort.Quelle, Umfang: antwort.Umfang,
		Drucke: antwort.Drucke, Staende: antwort.Staende,
		Befunde: antwort.Befunde,
	}, nil
}

// Marktangebote gibt die Posten eines Marktplatzes zu einem Druck (A4, #1443).
func (p *preise) Marktangebote(ctx context.Context, wunsch rest.Marktwunsch) (
	rest.Marktsicht, error,
) {
	antwort := app.MarktAnwendungUeber(p.db).Marktangebote(ctx,
		wunsch.Quelle, wunsch.Printing, wunsch.Sprache, wunsch.Finish, wunsch.Zustaende)
	if antwort.Fehler != nil {
		return rest.Marktsicht{}, antwort.Fehler.AlsFehler()
	}

	angebote := make([]rest.Angebotssicht, 0, len(antwort.Angebote))
	for _, angebot := range antwort.Angebote {
		angebote = append(angebote, rest.Angebotssicht(angebot))
	}

	return rest.Marktsicht{
		Quelle: antwort.Quelle, Fassung: antwort.Fassung,
		Zugeordnet: antwort.Zugeordnet, Angebote: angebote,
		Befunde: antwort.Befunde,
	}, nil
}

// Anbieterstand nennt das Kontingent an einem Anbieterkonto (A5, #1444).
func (p *preise) Anbieterstand(ctx context.Context, dienst string) (
	rest.Anbietersicht, error,
) {
	antwort := app.AnbieterAnwendungUeber(p.db).Anbieterstand(ctx, dienst)
	if antwort.Fehler != nil {
		return rest.Anbietersicht{}, antwort.Fehler.AlsFehler()
	}

	return rest.Anbietersicht(antwort.AnbieterstandDTO), nil
}

// Fremdpreisverlauf holt den Verlauf eines Drucks bei einer Quelle (A5, #1444).
func (p *preise) Fremdpreisverlauf(ctx context.Context, wunsch rest.Verlaufsbezug) (
	rest.Verlaufssicht, error,
) {
	antwort := app.AnbieterAnwendungUeber(p.db).Fremdpreisverlauf(ctx,
		wunsch.Quelle, wunsch.Printing, wunsch.Finish, wunsch.Zustaende)
	if antwort.Fehler != nil {
		return rest.Verlaufssicht{}, antwort.Fehler.AlsFehler()
	}

	return rest.Verlaufssicht(antwort.FremdpreisbilanzDTO), nil
}
