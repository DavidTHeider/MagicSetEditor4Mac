package app

// **Die Liste der Handelsvorgänge und ihre Entwürfe** — TRADE-01
// (#1715).
//
// Die Übersetzung zwischen Fläche und [service.Handelsverzeichnis]: Filter
// hinein, Zeilen hinaus, und „ich" ist die Peerkennung dieses Profils — wie
// beim Eröffnen.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// HandelsfilterDTO ist, was eine Fläche von der Liste sehen will.
//
// **Zeitpunkte als RFC 3339 oder als Tag** (`2026-09-30`): ein Tag heißt
// „von seinem Anfang" bei `seit` und „bis zu seinem Ende" bei `bis`, in UTC.
// Das Fenster schickt den Anfang des Tages in seiner Zeitzone als RFC 3339.
type HandelsfilterDTO struct {
	Partner string   `json:"partner"`
	Lagen   []string `json:"lagen"`
	Suche   string   `json:"suche"`
	Marke   string   `json:"marke"`
	Seit    string   `json:"seit"`
	Bis     string   `json:"bis"`
	Ordnung string   `json:"ordnung"`
	Archiv  bool     `json:"archiv"`
	Sprache string   `json:"sprache"`
}

// HandelspartnerDTO ist ein Gegenüber einer Zeile.
type HandelspartnerDTO struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
}

// HandelsvorgangzeileDTO ist ein Vorgang in der Liste.
type HandelsvorgangzeileDTO struct {
	Vorgang    HandelsvorgangDTO   `json:"vorgang"`
	Lage       string              `json:"lage"`
	Naechster  string              `json:"naechster"`
	Gegenueber []HandelspartnerDTO `json:"gegenueber"`
	Karten     []string            `json:"karten"`
	Marken     []string            `json:"marken"`
	Angelegt   string              `json:"angelegt"`
	Geaendert  string              `json:"geaendert"`
	Archiviert bool                `json:"archiviert"`

	// Unbestaetigt heißt: dieser Bestand ist verbucht, der des Gegenübers
	// noch nicht bestätigt (AK-04).
	Unbestaetigt bool `json:"unbestaetigt"`
}

// HandelsuebersichtAntwort ist die Liste — **auch eine Absage trägt eine**,
// eine leere.
type HandelsuebersichtAntwort struct {
	Zeilen []HandelsvorgangzeileDTO `json:"zeilen"`
	Fehler *Fehler                  `json:"fehler"`
}

// HandelUebersicht gibt die Vorgänge dieses Profils.
func (a *App) HandelUebersicht(ctx context.Context, wunsch HandelsfilterDTO) HandelsuebersichtAntwort {
	ctx = a.imProfil(ctx)

	absage := func(fehler error) HandelsuebersichtAntwort {
		return HandelsuebersichtAntwort{Zeilen: []HandelsvorgangzeileDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	filter, fehler := alsHandelsfilter(wunsch)
	if fehler != nil {
		return absage(fehler)
	}

	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return absage(fehler)
	}

	zeilen, fehler := a.handelsverzeichnis().Liste(ctx, string(ich.Kennung()), filter)
	if fehler != nil {
		return absage(fehler)
	}

	dtos := make([]HandelsvorgangzeileDTO, 0, len(zeilen))
	for _, zeile := range zeilen {
		dtos = append(dtos, alsZeilenDTO(zeile, string(ich.Kennung())))
	}

	return HandelsuebersichtAntwort{Zeilen: dtos}
}

// alsHandelsfilter übersetzt den Filter einer Fläche.
func alsHandelsfilter(wunsch HandelsfilterDTO) (domain.Handelsfilter, error) {
	filter := domain.Handelsfilter{
		Partner: wunsch.Partner, Suche: wunsch.Suche, Marke: wunsch.Marke,
		Ordnung: domain.Handelsordnung(wunsch.Ordnung), Archiv: wunsch.Archiv, Sprache: wunsch.Sprache,
	}

	for _, lage := range wunsch.Lagen {
		filter.Lagen = append(filter.Lagen, domain.Handelslage(lage))
	}

	var fehler error

	if filter.Seit, fehler = tagOderZeitpunkt("seit", wunsch.Seit, false); fehler != nil {
		return domain.Handelsfilter{}, fehler
	}

	if filter.Bis, fehler = tagOderZeitpunkt("bis", wunsch.Bis, true); fehler != nil {
		return domain.Handelsfilter{}, fehler
	}

	return filter, nil
}

// tagOderZeitpunkt liest einen Zeitpunkt oder einen Tag — ein Tag als `bis`
// reicht bis zu seinem Ende.
func tagOderZeitpunkt(feld, roh string, ende bool) (time.Time, error) {
	if roh == "" {
		return time.Time{}, nil
	}

	if zeitpunkt, fehler := time.Parse(time.RFC3339, roh); fehler == nil {
		return zeitpunkt, nil
	}

	tag, fehler := time.Parse(time.DateOnly, roh)
	if fehler != nil {
		return time.Time{}, domain.NewFieldError(feld, domain.SchluesselValidierung,
			"%q ist kein Zeitpunkt — erwartet ist ein Tag wie 2026-09-30 oder ein Zeitpunkt nach RFC 3339.", roh)
	}

	if ende {
		return tag.AddDate(0, 0, 1).Add(-time.Nanosecond), nil
	}

	return tag, nil
}

// alsZeilenDTO macht aus einer Zeile ihre Gestalt für die Flächen.
func alsZeilenDTO(zeile domain.Handelsvorgangzeile, ich string) HandelsvorgangzeileDTO {
	partner := make([]HandelspartnerDTO, 0, len(zeile.Gegenueber))
	for _, gegenueber := range zeile.Gegenueber {
		partner = append(partner, HandelspartnerDTO{Kennung: gegenueber.Kennung, Name: gegenueber.Name})
	}

	return HandelsvorgangzeileDTO{
		Vorgang: alsHandelsvorgangDTO(zeile.Angebot, ich), Lage: string(zeile.Lage),
		Naechster: string(zeile.Naechster), Gegenueber: partner,
		Karten: alsListe(zeile.Karten), Marken: alsListe(zeile.Marken),
		Angelegt: zeitAlsText(zeile.Angelegt), Geaendert: zeitAlsText(zeile.Geaendert),
		Archiviert: zeile.Archiviert, Unbestaetigt: zeile.Angebot.Unbestaetigt(),
	}
}

// zeitAlsText gibt einen Zeitpunkt nach RFC 3339 — eine Nullzeit als leer.
func zeitAlsText(zeitpunkt time.Time) string {
	if zeitpunkt.IsZero() {
		return ""
	}

	return zeitpunkt.UTC().Format(time.RFC3339)
}

// HandelEntwurfAendern schreibt einen Entwurf neu.
func (a *App) HandelEntwurfAendern(ctx context.Context, wunsch Handelswunsch) HandelsvorgangAntwort {
	ctx = a.imProfil(ctx)

	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	mir := string(ich.Kennung())

	gespeichert, fehler := a.handelsverzeichnis().EntwurfAendern(ctx, domain.Handelsangebot{
		Vorgang: wunsch.Vorgang, Art: domain.Handelsart(wunsch.Art),
		Parteien: append([]string{mir}, wunsch.Gegenueber...), Positionen: alsPositionen(wunsch.Positionen),
		BewertungGeteilt: wunsch.BewertungGeteilt,
	})
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	return HandelsvorgangAntwort{Vorgang: alsHandelsvorgangDTO(gespeichert, a.eigeneKennung(ctx))}
}

// HandelKopieren legt aus einem Vorgang einen neuen Entwurf an.
func (a *App) HandelKopieren(ctx context.Context, vorgang string) HandelsvorgangAntwort {
	ctx = a.imProfil(ctx)

	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	kopie, fehler := a.handelsverzeichnis().Kopieren(ctx, string(ich.Kennung()), vorgang)
	if fehler != nil {
		return handelsabsage(ctx, fehler)
	}

	return HandelsvorgangAntwort{Vorgang: alsHandelsvorgangDTO(kopie, a.eigeneKennung(ctx))}
}

// HandelEntwurfLoeschen löscht einen Entwurf.
func (a *App) HandelEntwurfLoeschen(ctx context.Context, vorgang string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.handelsverzeichnis().EntwurfLoeschen(ctx, vorgang))}
}

// HandelArchivieren legt einen beendeten Vorgang ins Archiv oder holt ihn
// heraus.
func (a *App) HandelArchivieren(ctx context.Context, vorgang string, an bool) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.handelsverzeichnis().Archivieren(ctx, vorgang, an))}
}
