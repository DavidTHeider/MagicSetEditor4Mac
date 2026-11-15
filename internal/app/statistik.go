package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Statistiker ist, was `app` für die Kennzahlen braucht.
type Statistiker interface {
	Kennzahlen(ctx context.Context, setID int64) (domain.SetStats, error)
	Belegung(ctx context.Context, setID int64) ([]domain.FieldCount, error)
	Verteilung(ctx context.Context, setID int64, feld string) ([]domain.ValueCount, error)
}

// ZahlDTO ist eine Zeile einer Belegung.
type ZahlDTO struct {
	Name   string `json:"name"`
	Karten int    `json:"karten"`
}

// WertzahlDTO ist eine Zeile der Verteilung — samt den Karten des Werts,
// aufsteigend. UXR-F6
// (#2013).
//
// **Ein eigener Typ und nicht ein Feld in [ZahlDTO].** Eine Zeile der
// Belegung ist ein Feld, und eine Karte steht in vielen davon: Feldzahl mal
// Kartenzahl Kennungen für eine Frage, die niemand stellt. Und ein Feld mit
// `omitempty` ließe die leere Liste verschwinden (`TestKeineLeereListe…`) —
// die Belegung sähe dann aus, als nennte sie ihre Karten und hätte keine.
type WertzahlDTO struct {
	Name   string `json:"name"`
	Karten int    `json:"karten"`

	// KartenIDs sind die Karten des Werts als Zeichenketten, wie jede Kennung
	// an dieser Naht. **Eine Liste und nie `null`**: ein Wert zählt mindestens
	// eine Karte, aber die Seite filtert damit, ohne es zu prüfen.
	KartenIDs []string `json:"kartenIDs"`
}

// verteilungszeile macht aus einem Wert der Verteilung eine Zeile.
func verteilungszeile(wert domain.ValueCount) WertzahlDTO {
	return WertzahlDTO{
		Name: wert.Value, Karten: wert.Cards, KartenIDs: kennungstexte(wert.CardIDs),
	}
}

// StatistikDTO sind die Kennzahlen eines Sets, wie das Frontend sie sieht.
type StatistikDTO struct {
	Karten int `json:"karten"`
	Felder int `json:"felder"`

	// Leere ist die Zahl der Karten ohne einen einzigen nichtleeren Wert.
	Leere int `json:"leere"`

	Belegung []ZahlDTO `json:"belegung"`
}

// StatistikAntwort trägt die Kennzahlen oder einen Fehler.
type StatistikAntwort struct {
	Statistik StatistikDTO `json:"statistik"`
	Fehler    *Fehler      `json:"fehler"`
}

// VerteilungAntwort trägt die Verteilung eines Felds oder einen Fehler.
type VerteilungAntwort struct {
	Verteilung []WertzahlDTO `json:"verteilung"`
	Fehler     *Fehler       `json:"fehler"`
}

// Statistik gibt die Kennzahlen eines Sets.
func (a *App) Statistik(ctx context.Context, setID string) StatistikAntwort {
	ctx = a.imProfil(ctx)

	leer := StatistikAntwort{Statistik: StatistikDTO{Belegung: []ZahlDTO{}}}

	kennung, befund := kennungAus(setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	zahlen, err := a.statistiker.Kennzahlen(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	belegung, err := a.statistiker.Belegung(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	statistik := StatistikDTO{
		Karten:   zahlen.Cards,
		Felder:   zahlen.Fields,
		Leere:    zahlen.Empty,
		Belegung: leer.Statistik.Belegung,
	}

	for _, feld := range belegung {
		statistik.Belegung = append(statistik.Belegung, ZahlDTO{
			Name: feld.Field, Karten: feld.Cards,
		})
	}

	return StatistikAntwort{Statistik: statistik}
}

// Verteilung sagt, wie viele Karten welchen Wert eines Felds tragen.
func (a *App) Verteilung(ctx context.Context, setID, feld string) VerteilungAntwort {
	ctx = a.imProfil(ctx)

	leer := VerteilungAntwort{Verteilung: []WertzahlDTO{}}

	kennung, befund := kennungAus(setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	werte, err := a.statistiker.Verteilung(ctx, kennung, feld)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	for _, wert := range werte {
		leer.Verteilung = append(leer.Verteilung, verteilungszeile(wert))
	}

	return leer
}

// MitStatistiker setzt die Naht zur Statistik.
func (a *App) MitStatistiker(statistiker Statistiker) { a.statistiker = statistiker }
