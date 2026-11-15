package app

// **Die Formatnaht** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Ein Format ist **Daten** (Vermerk 0061); diese Schicht reicht sie durch und
// entscheidet nichts.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Formatnaht ist, was das Fenster zur Formatpflege braucht.
type Formatnaht interface {
	Bauarten() []service.Bauartvorschlag
	AusBauart(ctx context.Context, bauart domain.Bauart,
		name, spiel, scryfallname string) (domain.Format, error)
	Anlegen(ctx context.Context, format domain.Format) (domain.Format, error)
	Aendern(ctx context.Context, format domain.Format) (domain.Format, error)
	Format(ctx context.Context, kennung int64) (domain.Format, error)
	Formate(ctx context.Context, spiel string) ([]domain.Format, error)
	Loeschen(ctx context.Context, kennung int64) error
}

// FormatDTO ist ein Format für das Fenster.
type FormatDTO struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Name  string `json:"name"`
	Spiel string `json:"spiel"`

	// **Null heißt: keine Grenze** — bei `sideboardgroesse` heißt es: kein
	// Sideboard. Der Unterschied steht in `domain.Format`.
	Mindestgroesse    int `json:"mindestgroesse"`
	Hoechstgroesse    int `json:"hoechstgroesse"`
	Sideboardgroesse  int `json:"sideboardgroesse"`
	HoechstzahlGleich int `json:"hoechstzahlGleich"`

	// SideboardOffen heißt: ein Sideboard **ohne** Obergrenze.
	//
	// **Es ist nicht dasselbe wie `sideboardgroesse: 0`.** Die Null sagt
	// „kein Sideboard"; CR 100.4b legt bei limitiertem Spiel die übrigen
	// Karten des Pools ins Sideboard, ohne eine Zahl zu nennen. Eine Zahl,
	// die zwei Dinge sagt, sagt am Ende keines.
	//
	// **Nachgetragen am 2026-09-01.** Das Feld gab es seit der
	// Korrekturrunde in `domain`, aber an keiner Fläche — und damit nirgends,
	// wo jemand es sehen konnte (Befund 1 des Nachlaufs über Phase 4.6).
	SideboardOffen bool `json:"sideboardOffen"`

	// Scryfallname ist leer, wenn Scryfall dieses Format nicht kennt — dann
	// bleibt jede Karte `ungeprüft`.
	Scryfallname string `json:"scryfallname"`
}

// BauartDTO ist eine Bauart samt den Zahlen, die sie mitbringt.
type BauartDTO struct {
	Bauart string    `json:"bauart"`
	Format FormatDTO `json:"format"`
}

// FormateAntwort trägt die Übersicht.
type FormateAntwort struct {
	Formate []FormatDTO `json:"formate"`
	Fehler  *Fehler     `json:"fehler"`
}

// FormatAntwort trägt ein einzelnes Format.
type FormatAntwort struct {
	Format *FormatDTO `json:"format"`
	Fehler *Fehler    `json:"fehler"`
}

// BauartenAntwort trägt die Vorschläge.
type BauartenAntwort struct {
	Bauarten []BauartDTO `json:"bauarten"`
	Fehler   *Fehler     `json:"fehler"`
}

// MitFormatnaht setzt die Naht — für Tests und den Composition Root.
func (a *App) MitFormatnaht(naht Formatnaht) { a.formate = naht }

// Bauarten gibt die drei Bauarten samt ihren nachgelesenen Zahlen.
//
// **Ohne Fehler, und das ist kein Versehen**: die Liste steht in `domain` und
// hängt an keinem Bestand.
func (a *App) Bauarten() BauartenAntwort {
	vorschlaege := a.formate.Bauarten()

	antwort := BauartenAntwort{Bauarten: make([]BauartDTO, 0, len(vorschlaege))}
	for _, vorschlag := range vorschlaege {
		antwort.Bauarten = append(antwort.Bauarten, BauartDTO{
			Bauart: string(vorschlag.Bauart),
			Format: alsFormat(vorschlag.Format),
		})
	}

	return antwort
}

// Formate gibt die Formate eines Spiels — ein leeres Spiel heißt: alle.
func (a *App) Formate(ctx context.Context, spiel string) FormateAntwort {
	ctx = a.imProfil(ctx)

	formate, err := a.formate.Formate(ctx, spiel)
	if err != nil {
		return FormateAntwort{Formate: []FormatDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	liste := make([]FormatDTO, 0, len(formate))
	for _, format := range formate {
		liste = append(liste, alsFormat(format))
	}

	return FormateAntwort{Formate: liste}
}

// FormatAusBauart legt ein Format mit den Zahlen einer Bauart an.
func (a *App) FormatAusBauart(
	ctx context.Context, bauart, name, spiel, scryfallname string,
) FormatAntwort {
	ctx = a.imProfil(ctx)

	format, err := a.formate.AusBauart(ctx,
		domain.Bauart(bauart), name, spiel, scryfallname)
	if err != nil {
		return FormatAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einFormat(format)
}

// FormatSetzen legt ein Format an oder schreibt es fort.
//
// **Ein Weg für beides**, und die Kennung entscheidet: leer heißt anlegen.
// Zwei Bindungen dafür wären zwei Fassungen derselben Umrechnung.
func (a *App) FormatSetzen(ctx context.Context, gewuenscht FormatDTO) FormatAntwort {
	ctx = a.imProfil(ctx)

	format, befund := ausFormat(gewuenscht)
	if befund != nil {
		return FormatAntwort{Fehler: befund}
	}

	if format.ID == 0 {
		neu, err := a.formate.Anlegen(ctx, format)
		if err != nil {
			return FormatAntwort{Fehler: Uebersetze(ctx, err)}
		}

		return einFormat(neu)
	}

	geaendert, err := a.formate.Aendern(ctx, format)
	if err != nil {
		return FormatAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einFormat(geaendert)
}

// FormatLoeschen nimmt ein Format heraus — **und lässt die Decks stehen**.
func (a *App) FormatLoeschen(ctx context.Context, formatID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("formatID", formatID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.formate.Loeschen(ctx, kennung))}
}

// einFormat packt ein Format in seine Antwort.
func einFormat(format domain.Format) FormatAntwort {
	dto := alsFormat(format)

	return FormatAntwort{Format: &dto}
}

// alsFormat macht aus einem Format sein DTO.
func alsFormat(format domain.Format) FormatDTO {
	return FormatDTO{
		ID:                kennungtext(format.ID),
		Fassung:           kennungtext(format.Fassung),
		Name:              format.Name,
		Spiel:             format.Spiel,
		Mindestgroesse:    format.Mindestgroesse,
		Hoechstgroesse:    format.Hoechstgroesse,
		Sideboardgroesse:  format.Sideboardgroesse,
		SideboardOffen:    format.SideboardOffen,
		HoechstzahlGleich: format.HoechstzahlGleich,
		Scryfallname:      format.Scryfallname,
	}
}

// ausFormat liest ein DTO zurück in die Domäne.
//
// **Eine leere Kennung ist die Null und kein Fehler** — sie heißt „anlegen".
func ausFormat(dto FormatDTO) (domain.Format, *Fehler) {
	format := domain.Format{
		Name:              dto.Name,
		Spiel:             dto.Spiel,
		Mindestgroesse:    dto.Mindestgroesse,
		Hoechstgroesse:    dto.Hoechstgroesse,
		Sideboardgroesse:  dto.Sideboardgroesse,
		SideboardOffen:    dto.SideboardOffen,
		HoechstzahlGleich: dto.HoechstzahlGleich,
		Scryfallname:      dto.Scryfallname,
	}

	if dto.ID != "" {
		kennung, befund := benannteKennungAus("formatID", dto.ID)
		if befund != nil {
			return domain.Format{}, befund
		}

		format.ID = kennung
	}

	if dto.Fassung != "" {
		fassung, befund := benannteKennungAus("fassung", dto.Fassung)
		if befund != nil {
			return domain.Format{}, befund
		}

		format.Fassung = fassung
	}

	return format, nil
}
