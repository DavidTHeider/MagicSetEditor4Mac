// **Die Kartenliste in drei Gestalten** —
// #891,
// Werkstück 5 von Phase Y.
//
// Bis zum 2026-08-25 gab es eine: [App.Karten] mit **jedem** Feldwert **jeder**
// Karte. An `Black.mse-set` mit 3.788 Karten sind das 4.761.182 Bytes JSON und
// 35,4 ms je Aufruf (`BenchmarkKartenliste`) — für eine Liste, die einen Namen
// je Zeile zeigt, und für **eine** Karte, die offen ist.
//
// Seitdem holt jede Fläche, was sie braucht:
//
//   - [App.Kartenzeilen] — Name und Feldzahl je Karte (245.641 Bytes, 10,0 ms)
//   - [App.Karte] — **eine** Karte mit allem
//   - [App.Kartenfeldwerte] — **ein** Feld über das ganze Set, für die
//     Massenänderung
//
// [App.Karten] bleibt: die Statistik liest jedes Feld jeder Karte, und für sie
// ist die große Antwort die richtige.

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KartenzeileDTO ist eine Karte, wie die **Liste** sie zeigt (#891).
//
// **Ohne Feldwerte.** `Black.mse-set` hat 3.788 Karten; sie mit jedem Wert zu
// laden waren 4.761.182 Bytes JSON je Aufruf und 35,4 ms allein auf der
// Go-Seite — gemessen am 2026-08-25 (`BenchmarkKartenliste`). Die Liste
// braucht davon einen Namen je Karte.
type KartenzeileDTO struct {
	ID     string `json:"id"`
	Nummer int    `json:"nummer"`

	// Name ist der Wert des Feldes `name` — leer, wenn es keinen gibt.
	Name string `json:"name"`

	// Felder ist die **Zahl** der Feldwerte, nicht die Werte.
	//
	// Die Liste zeigt „ohne Felder" an einer Karte, die keine hat: der
	// Unterschied zwischen „heißt nicht" und „ist leer".
	Felder int `json:"felder"`
}

// KartenzeilenAntwort ist die Liste — **nie `null`** (§3).
type KartenzeilenAntwort struct {
	Zeilen []KartenzeileDTO `json:"zeilen"`
	Fehler *Fehler          `json:"fehler"`
}

// Einzelkartenleser gibt eine einzelne Karte (#891).
type Einzelkartenleser interface {
	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)
}

// MitEinzelkarte setzt die Naht.
func (a *App) MitEinzelkarte(leser Einzelkartenleser) { a.einzelkarte = leser }

// einzelleser gibt die Naht — oder eine, die sich meldet.
func (a *App) einzelleser() Einzelkartenleser {
	if a.einzelkarte != nil {
		return a.einzelkarte
	}

	return ohneEinzelnaht{}
}

// ohneEinzelnaht meldet, dass niemand verdrahtet hat.
type ohneEinzelnaht struct{}

func (ohneEinzelnaht) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return nil, fmt.Errorf("%w: der Kartenleser ist nicht eingerichtet",
		domain.ErrValidation)
}

// Kartenzeilenleser gibt die Karten eines Sets als Liste (#891).
//
// **Eine eigene Naht und nicht ein fünfter Weg in [Bestandsleser].** Der wird
// von einem Dutzend Doubles erfüllt; eine Methode mehr dort hieße, jedes davon
// um eine Antwort zu erweitern, die sein Fall nicht braucht.
type Kartenzeilenleser interface {
	Kartenzeilen(ctx context.Context, setID int64) ([]domain.CardLine, error)
}

// MitKartenzeilen setzt die Naht.
func (a *App) MitKartenzeilen(leser Kartenzeilenleser) { a.zeilen = leser }

// zeilenleser gibt die Naht — oder eine, die sich meldet.
func (a *App) zeilenleser() Kartenzeilenleser {
	if a.zeilen != nil {
		return a.zeilen
	}

	return ohneZeilennaht{}
}

// ohneZeilennaht meldet, dass niemand verdrahtet hat.
type ohneZeilennaht struct{}

func (ohneZeilennaht) Kartenzeilen(context.Context, int64) ([]domain.CardLine, error) {
	return nil, fmt.Errorf("%w: die Kartenliste ist nicht eingerichtet",
		domain.ErrValidation)
}

// Kartenzeilen gibt die Karten eines Sets als **Liste** — ohne ihre Feldwerte
// (#891).
//
// **Der Weg, den das Fenster beim Öffnen geht.** Was eine einzelne Karte
// trägt, holt [App.Karte], wenn jemand sie aufschlägt.
func (a *App) Kartenzeilen(ctx context.Context, setID string) KartenzeilenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return KartenzeilenAntwort{Zeilen: []KartenzeileDTO{}, Fehler: befund}
	}

	gelesen, err := a.zeilenleser().Kartenzeilen(ctx, kennung)
	if err != nil {
		return KartenzeilenAntwort{Zeilen: []KartenzeileDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	zeilen := make([]KartenzeileDTO, 0, len(gelesen))
	for _, zeile := range gelesen {
		zeilen = append(zeilen, KartenzeileDTO{
			ID:     strconv.FormatInt(zeile.ID, 10),
			Nummer: zeile.Number,
			Name:   zeile.Name,
			Felder: zeile.Fields,
		})
	}

	return KartenzeilenAntwort{Zeilen: zeilen}
}

// EinzelkarteAntwort trägt eine einzelne Karte (#891).
type EinzelkarteAntwort struct {
	Karte  KarteDTO `json:"karte"`
	Fehler *Fehler  `json:"fehler"`
}

// Karte gibt **eine** Karte samt ihren Feldwerten (#891).
//
// **Das Gegenstück zu [App.Kartenzeilen].** Wer eine Karte aufschlägt, braucht
// alles von ihr — und von keiner anderen.
func (a *App) Karte(ctx context.Context, karteID string) EinzelkarteAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("karteID", karteID)
	if befund != nil {
		return EinzelkarteAntwort{Karte: leereKarte(), Fehler: befund}
	}

	gelesen, err := a.einzelleser().KarteLesen(ctx, kennung)
	if err != nil {
		return EinzelkarteAntwort{Karte: leereKarte(), Fehler: Uebersetze(ctx, err)}
	}

	umgesetzt := kartenDTO([]*domain.CardRow{gelesen})

	return EinzelkarteAntwort{Karte: umgesetzt[0]}
}

// leereKarte ist die Antwort, wenn keine dasteht — **mit leerer Werteliste**.
//
// Ein `null` dort käme im Fenster als `undefined` an, und die erste Stelle,
// die darüber liefe, bräche (§3).
func leereKarte() KarteDTO {
	return KarteDTO{Werte: []WertDTO{}}
}

// Kartenfeldwerteleser gibt **ein** Feld über ein ganzes Set (#891).
//
// **Eine eigene Naht**, aus demselben Grund wie [Kartenzeilenleser]: eine
// Methode mehr in [Bestandsleser] hieße, jedes seiner Doubles um eine Antwort
// zu erweitern, die sein Fall nicht braucht.
type Kartenfeldwerteleser interface {
	Kartenfeldwerte(ctx context.Context, setID int64, feld string) (
		[]domain.CardField, error,
	)
}

// MitKartenfeldwerten setzt die Naht.
func (a *App) MitKartenfeldwerten(leser Kartenfeldwerteleser) {
	a.feldwerte = leser
}

// feldwerteleser gibt die Naht — oder eine, die sich meldet.
func (a *App) feldwerteleser() Kartenfeldwerteleser {
	if a.feldwerte != nil {
		return a.feldwerte
	}

	return ohneFeldwertnaht{}
}

// ohneFeldwertnaht meldet, dass niemand verdrahtet hat.
type ohneFeldwertnaht struct{}

func (ohneFeldwertnaht) Kartenfeldwerte(context.Context, int64, string) (
	[]domain.CardField, error,
) {
	return nil, fmt.Errorf("%w: die Feldwerte sind nicht eingerichtet",
		domain.ErrValidation)
}

// KartenfeldwertDTO ist der Wert **eines** Feldes **einer** Karte (#891).
type KartenfeldwertDTO struct {
	ID   string `json:"id"`
	Wert string `json:"wert"`
}

// KartenfeldwerteAntwort ist die Liste — **nie `null`** (§3).
type KartenfeldwerteAntwort struct {
	Werte  []KartenfeldwertDTO `json:"werte"`
	Fehler *Fehler             `json:"fehler"`
}

// Kartenfeldwerte gibt den Wert eines Feldes über alle Karten eines Sets.
//
// **Für die Massenänderung.** Sie zeigt neben jeder angehakten Karte den Wert,
// um den es geht; ohne ihn wäre eine Rücknahme unsichtbar — sie stellte etwas
// wieder her, was niemand gesehen hat. Bis zum 2026-08-25 kam dieser Wert mit
// [App.Karten], also mit jedem Feld jeder Karte.
//
// **Nur Karten, die das Feld haben, stehen darin.** Eine Zeile mit leerem Wert
// sagte, das Feld sei leer, und das ist etwas anderes als „gibt es nicht".
func (a *App) Kartenfeldwerte(ctx context.Context, setID, feld string) KartenfeldwerteAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return KartenfeldwerteAntwort{Werte: []KartenfeldwertDTO{}, Fehler: befund}
	}

	gelesen, err := a.feldwerteleser().Kartenfeldwerte(ctx, kennung, feld)
	if err != nil {
		return KartenfeldwerteAntwort{
			Werte: []KartenfeldwertDTO{}, Fehler: Uebersetze(ctx, err),
		}
	}

	werte := make([]KartenfeldwertDTO, 0, len(gelesen))
	for _, wert := range gelesen {
		werte = append(werte, KartenfeldwertDTO{
			ID:   strconv.FormatInt(wert.ID, 10),
			Wert: wert.Text,
		})
	}

	return KartenfeldwerteAntwort{Werte: werte}
}
