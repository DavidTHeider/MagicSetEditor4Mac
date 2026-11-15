package app

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau einer Massenänderung und das Anwenden mit Erwartung** —
// Werkstück U6 von Phase 5.1
// (#1417).
// Hier steht nur die Übersetzung (§2): Kennungen als Zeichenketten hinein,
// die Zeilen der Domäne hinaus, der Fehler übersetzt.

// FeldvorwertDTO ist der Altwert einer Karte, wie die Vorschau ihn gezeigt
// hat — und wie das Fenster ihn beim Anwenden zurückgibt.
type FeldvorwertDTO struct {
	KarteID string `json:"karteID"`
	Wert    string `json:"wert"`
	GabEs   bool   `json:"gabEs"`
}

// MassenvorschauzeileDTO ist eine Zeile der Vorschau.
type MassenvorschauzeileDTO struct {
	KarteID    string `json:"karteID"`
	Name       string `json:"name"`
	Alt        string `json:"alt"`
	GabAlt     bool   `json:"gabAlt"`
	Neu        string `json:"neu"`
	Ausschluss string `json:"ausschluss"`
}

// MassenvorschauAntwort ist die Vorschau samt Fehler.
//
// **`Zeilen` ist nie `null`** — eine leere Liste, damit die Seite nicht
// über `undefined` stolpert (§3, „zwischen den Prüfebenen").
type MassenvorschauAntwort struct {
	Feld   string                   `json:"feld"`
	Wert   string                   `json:"wert"`
	Zeilen []MassenvorschauzeileDTO `json:"zeilen"`
	Fehler *Fehler                  `json:"fehler"`
}

// FelderVorschau sagt, was ein Feldwert auf den genannten Karten hieße.
func (a *App) FelderVorschau(
	ctx context.Context, setID string, karteIDs []string, feld, wert string,
) MassenvorschauAntwort {
	ctx = a.imProfil(ctx)

	leer := MassenvorschauAntwort{Feld: feld, Wert: wert, Zeilen: []MassenvorschauzeileDTO{}}

	setkennung, kennungen, befund := massenkennungen(setID, karteIDs, feld)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	vorschau, err := a.massenschreiber.FelderVorschau(ctx, setkennung, kennungen, feld, wert)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	zeilen := make([]MassenvorschauzeileDTO, 0, len(vorschau.Zeilen))
	for _, zeile := range vorschau.Zeilen {
		zeilen = append(zeilen, MassenvorschauzeileDTO{
			KarteID:    strconv.FormatInt(zeile.KarteID, 10),
			Name:       zeile.Name,
			Alt:        zeile.Alt,
			GabAlt:     zeile.GabAlt,
			Neu:        zeile.Neu,
			Ausschluss: zeile.Ausschluss,
		})
	}

	return MassenvorschauAntwort{Feld: vorschau.Feld, Wert: vorschau.Wert, Zeilen: zeilen}
}

// FelderSetzenMitErwartung schreibt, was die Vorschau gezeigt hat — und
// nennt bei einem Konflikt die Karten, die inzwischen anders dastehen.
func (a *App) FelderSetzenMitErwartung(
	ctx context.Context, setID string, erwartet []FeldvorwertDTO, feld, wert string,
) MassenAntwort {
	ctx = a.imProfil(ctx)

	leer := MassenAntwort{Betroffen: []string{}}

	roh := make([]string, 0, len(erwartet))
	for _, vorwert := range erwartet {
		roh = append(roh, vorwert.KarteID)
	}

	setkennung, kennungen, befund := massenkennungen(setID, roh, feld)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	vorwerte := make([]domain.Feldvorwert, 0, len(erwartet))
	for i, vorwert := range erwartet {
		vorwerte = append(vorwerte, domain.Feldvorwert{
			KarteID: kennungen[i], Wert: vorwert.Wert, GabEs: vorwert.GabEs,
		})
	}

	vorgang, err := a.massenschreiber.FelderSetzenMitErwartung(ctx, setkennung, vorwerte, feld, wert)
	if err == nil {
		leer.Vorgang = vorgang

		return leer
	}

	leer.Fehler = Uebersetze(ctx, err)

	var massenfehler *domain.BulkError
	if errors.As(err, &massenfehler) {
		for _, kennung := range massenfehler.CardIDs {
			leer.Betroffen = append(leer.Betroffen, strconv.FormatInt(kennung, 10))
		}
	}

	return leer
}

// massenkennungen liest Set, Karten und Feldnamen — **alle, bevor etwas
// geschieht** (dieselbe Regel wie in [App.FelderSetzen]).
func massenkennungen(
	setID string, karteIDs []string, feld string,
) (setkennung int64, kartenkennungen []int64, befund *Fehler) {
	setkennung, setbefund := benannteKennungAus("setID", setID)
	if setbefund != nil {
		return 0, nil, setbefund
	}

	if strings.TrimSpace(feld) == "" {
		return 0, nil, Nutzerfehler(domain.NewFieldError(
			"feld", domain.SchluesselPflichtfeld, "Ein Feld braucht einen Namen."))
	}

	kennungen := make([]int64, 0, len(karteIDs))

	for _, roh := range karteIDs {
		kennung, befund := kennungAus(roh)
		if befund != nil {
			return 0, nil, befund
		}

		kennungen = append(kennungen, kennung)
	}

	return setkennung, kennungen, nil
}
