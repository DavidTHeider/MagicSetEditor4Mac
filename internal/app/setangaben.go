package app

// **Die Setangaben** — Titel, Setcode, Copyright, Symbol
// (#567).
//
// MSE hat dafür eine eigene Lasche neben „Cards" (`SetInfoPanel`,
// `src/gui/set/set_info_panel.cpp`, Fassung f9d9356 vom 2021-07-06):
//
//	void SetInfoEditor::onChangeSet() {
//	  setStyles(set->stylesheet, set->stylesheet->set_info_style);
//	  setData(set->data);
//	}
//
// `src/gui/control/native_look_editor.cpp`, Zeilen 187–190. Die **Felder**
// kommen aus dem Spiel, die **Werte** aus `set->data` — hier aus `set_values`.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setwertquelle ist, was `app` für die Setangaben braucht — die Naht zum
// Bestand.
type Setwertquelle interface {
	Setwerte(ctx context.Context, setID int64) ([]domain.Value, error)
	SetwertSetzenMitVorwert(ctx context.Context, setID int64, feld, wert string) (
		vorher string, hatte bool, fehler error,
	)

	// Setkopf gibt die fünf Formatangaben (#1325).
	Setkopf(ctx context.Context, setID int64) (domain.Setkopf, error)
}

// MitSetwerten setzt die Naht zu den Setangaben.
func (a *App) MitSetwerten(quelle Setwertquelle) { a.setwerte = quelle }

// SetangabenAntwort trägt Felder und Werte eines Sets.
//
// **Beides in einer Antwort**, anders als bei den Karten: dort ändern sich die
// Werte bei jedem Tastendruck und die Beschreibungen nie, also sind es zwei
// Aufrufe. Ein Set ist **ein** Gegenstand — es gibt keine Liste, durch die die
// 89 Beschreibungen mitreisen müssten.
type SetangabenAntwort struct {
	Felder []FeldbeschreibungDTO `json:"felder"`
	Werte  []WertDTO             `json:"werte"`

	// Kopf sind die fünf Formatangaben (#1325).
	//
	// **In derselben Antwort**, weil sie dieselbe Frage beantworten: was steht
	// über dieses Set in seiner Datei? Ein eigener Aufruf wäre ein zweiter Weg
	// zum selben Gegenstand.
	Kopf SetkopfDTO `json:"kopf"`

	Fehler *Fehler `json:"fehler"`
}

// SetkopfDTO sind die fünf Formatangaben für das Fenster.
//
// **Fünf benannte Felder und keine Liste**: jede hat eine eigene Bedeutung und
// eine eigene Aufschrift, und eine Liste zwänge das Fenster, die Namen zu
// deuten.
type SetkopfDTO struct {
	MSEFassung        string `json:"mseFassung"`
	Spielfassung      string `json:"spielfassung"`
	Stilfassung       string `json:"stilfassung"`
	Versionskontrolle string `json:"versionskontrolle"`
	Apprenticecode    string `json:"apprenticecode"`
}

// Setangaben gibt die Angabenfelder eines Sets samt ihren Werten.
func (a *App) Setangaben(ctx context.Context, setID string) SetangabenAntwort {
	ctx = a.imProfil(ctx)

	leer := SetangabenAntwort{Felder: []FeldbeschreibungDTO{}, Werte: []WertDTO{}}

	kennung, befund := kennungAus(setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	felder, err := a.stile.Setfelder(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	werte, err := a.setwerte.Setwerte(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	kopf, err := a.setwerte.Setkopf(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	return SetangabenAntwort{
		Felder: beschreibungenAus(felder, a.sprache),
		Werte:  setwerteAus(werte),
		Kopf: SetkopfDTO{
			MSEFassung:        kopf.MSEFassung,
			Spielfassung:      kopf.Spielfassung,
			Stilfassung:       kopf.Stilfassung,
			Versionskontrolle: kopf.Versionskontrolle,
			Apprenticecode:    kopf.Apprenticecode,
		},
	}
}

// setwerteAus macht aus den Werten des Bestands die des Fensters.
//
// **Eine leere Liste und kein `nil`**: über die Wails-Bindung stünde dort
// sonst `null`, und das Frontend prüfte auf zwei Dinge statt auf eines.
func setwerteAus(werte []domain.Value) []WertDTO {
	umgesetzt := make([]WertDTO, 0, len(werte))
	for i := range werte {
		umgesetzt = append(umgesetzt, WertDTO{Feld: werte[i].Field, Wert: werte[i].Text})
	}

	return umgesetzt
}

// SetwertSetzen schreibt eine Setangabe.
//
// **Der Feldname wird geprüft, der Wert nicht** — dieselbe Regel wie bei einem
// Kartenfeld: ein leerer Name ergäbe eine Zeile, die niemand wiederfindet; ein
// leerer Wert ist ein gültiger Wert.
func (a *App) SetwertSetzen(ctx context.Context, setID, feld, wert string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if befund := domain.PruefeFeldname(feld); befund != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, befund)}
	}

	_, _, err := a.setwerte.SetwertSetzenMitVorwert(ctx, kennung, feld, wert)

	return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
}
