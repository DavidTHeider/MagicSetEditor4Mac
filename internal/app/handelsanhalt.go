package app

// **Der Preisanhalt eines Tauschs an der Naht** — TRADE-02b
// (#1951).
//
// **Zwei Wege hinein, eine Rechnung**: das Fenster fragt nach den Zeilen, die
// gerade im Entwurf stehen — sie sind noch nirgends gespeichert —, die
// übrigen Flächen nach einem Vorgang und seiner letzten Revision.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// AnhaltspreisDTO ist ein Betrag in einer Währung.
type AnhaltspreisDTO struct {
	Waehrung string `json:"waehrung"`

	// Wert steht in der kleinsten Einheit der Währung.
	Wert   int64  `json:"wert"`
	Quelle string `json:"quelle"`

	// Erhoben ist RFC 3339 — bei einer Summe der älteste Bestandteil.
	Erhoben string `json:"erhoben"`
}

// AnhaltszeileDTO ist der Preis einer Position; leer heißt „kein Preis".
type AnhaltszeileDTO struct {
	Preise []AnhaltspreisDTO `json:"preise"`

	// Druck ist die Karte der Position — P10 (#2378), **erweiternd** (§16):
	// wer das Feld nicht liest, bekommt dieselben Preise wie vorher.
	Druck HandelsdruckDTO `json:"druck"`
}

// HandelsdruckDTO ist, was eine Handelszeile über ihre Karte sagt; leer heißt
// unbekannt.
type HandelsdruckDTO struct {
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`

	// Bild ist die Adresse eines kleinen Kartenbilds — nur, was Scryfall
	// schon einmal beantwortet hat; geholt wird dafür nichts.
	Bild string `json:"bild"`

	// Eigen heißt: der Druck hat keine Scryfall-Kennung und nie ein Bild.
	Eigen bool `json:"eigen"`
}

// AnhaltsseiteDTO ist, was eine Partei gibt.
type AnhaltsseiteDTO struct {
	Von       string            `json:"von"`
	Summen    []AnhaltspreisDTO `json:"summen"`
	OhnePreis int               `json:"ohnePreis"`
}

// AnhaltsdifferenzDTO ist erhalten weniger gegeben.
type AnhaltsdifferenzDTO struct {
	Waehrung     string `json:"waehrung"`
	Wert         int64  `json:"wert"`
	Vollstaendig bool   `json:"vollstaendig"`
}

// PreisanhaltDTO ist der ganze Anhalt.
type PreisanhaltDTO struct {
	Zeilen        []AnhaltszeileDTO     `json:"zeilen"`
	Seiten        []AnhaltsseiteDTO     `json:"seiten"`
	Differenzen   []AnhaltsdifferenzDTO `json:"differenzen"`
	Hauptwaehrung string                `json:"hauptwaehrung"`
}

// PreisanhaltAntwort ist der Anhalt samt seinem Fehler.
type PreisanhaltAntwort struct {
	Anhalt PreisanhaltDTO `json:"anhalt"`
	Fehler *Fehler        `json:"fehler"`
}

// HandelPreisanhalt rechnet den Anhalt über Zeilen eines Entwurfs — **rein
// lesend**; die Preise gehen nirgends hin (AK-03).
func (a *App) HandelPreisanhalt(
	ctx context.Context, positionen []HandelspositionDTO,
) PreisanhaltAntwort {
	ctx = a.imProfil(ctx)

	return a.preisanhalt(ctx, alsPositionen(positionen))
}

// HandelPreisanhaltVon rechnet den Anhalt über die letzte Revision eines
// Vorgangs — der Weg für Kommandozeile, Dienst und MCP.
func (a *App) HandelPreisanhaltVon(ctx context.Context, vorgang string) PreisanhaltAntwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.handelsleitung().Stand(ctx, vorgang)
	if fehler != nil {
		return preisanhaltAbsage(ctx, fehler)
	}

	return a.preisanhalt(ctx, stand.Positionen)
}

// preisanhalt liest Vorgabe und Preise und rechnet.
func (a *App) preisanhalt(
	ctx context.Context, positionen []domain.Handelsposition,
) PreisanhaltAntwort {
	einstellungen, fehler := a.einstellungswerte(ctx)
	if fehler != nil {
		return preisanhaltAbsage(ctx, fehler)
	}

	ich := a.eigeneKennung(ctx)

	anhalt, fehler := service.Handelsanhalt(ctx, a.preisleser(),
		positionen, ich, einstellungen[domain.VorgabeWaehrung])
	if fehler != nil {
		return preisanhaltAbsage(ctx, fehler)
	}

	drucke, fehler := service.Handelsdrucke(ctx,
		a.handelsdruckleser(), a.druckvorrat(), positionen, ich)
	if fehler != nil {
		return preisanhaltAbsage(ctx, fehler)
	}

	return PreisanhaltAntwort{Anhalt: alsPreisanhaltDTO(anhalt, drucke)}
}

// preisanhaltAbsage gibt leere Listen und nicht `null` (§8).
func preisanhaltAbsage(ctx context.Context, fehler error) PreisanhaltAntwort {
	return PreisanhaltAntwort{
		Anhalt: PreisanhaltDTO{
			Zeilen: []AnhaltszeileDTO{}, Seiten: []AnhaltsseiteDTO{},
			Differenzen: []AnhaltsdifferenzDTO{},
		},
		Fehler: Uebersetze(ctx, fehler),
	}
}

// alsPreisanhaltDTO gibt dem Anhalt die Gestalt der Naht; `drucke` stehen in
// der Reihenfolge der Zeilen.
func alsPreisanhaltDTO(anhalt service.Preisanhalt, drucke []service.Handelsdruck) PreisanhaltDTO {
	dto := PreisanhaltDTO{
		Zeilen:        make([]AnhaltszeileDTO, 0, len(anhalt.Zeilen)),
		Seiten:        make([]AnhaltsseiteDTO, 0, len(anhalt.Seiten)),
		Differenzen:   make([]AnhaltsdifferenzDTO, 0, len(anhalt.Differenzen)),
		Hauptwaehrung: anhalt.Hauptwaehrung,
	}

	for stelle, zeile := range anhalt.Zeilen {
		dto.Zeilen = append(dto.Zeilen, AnhaltszeileDTO{
			Preise: alsAnhaltspreise(zeile.Preise), Druck: HandelsdruckDTO(drucke[stelle]),
		})
	}

	for _, seite := range anhalt.Seiten {
		dto.Seiten = append(dto.Seiten, AnhaltsseiteDTO{
			Von: seite.Von, Summen: alsAnhaltspreise(seite.Summen), OhnePreis: seite.OhnePreis,
		})
	}

	for _, differenz := range anhalt.Differenzen {
		dto.Differenzen = append(dto.Differenzen, AnhaltsdifferenzDTO(differenz))
	}

	return dto
}

// alsAnhaltspreise gibt Beträgen die Gestalt der Naht.
func alsAnhaltspreise(preise []service.Anhaltspreis) []AnhaltspreisDTO {
	dtos := make([]AnhaltspreisDTO, 0, len(preise))
	for _, preis := range preise {
		dtos = append(dtos, AnhaltspreisDTO{
			Waehrung: preis.Waehrung, Wert: preis.Wert,
			Quelle: string(preis.Quelle), Erhoben: alsZeitpunkt(preis.Erhoben),
		})
	}

	return dtos
}
