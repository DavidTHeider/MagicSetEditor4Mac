package app

// Die **Schlüsselwörter eines Sets** — sehen und pflegen
// (#1062).
//
// # Warum die des Spiels mitkommen
//
// Weil die Frage „welche gibt es hier?" beide meint. `magic.mse-game` führt
// **334**; ein Set führt selten mehr als eine Handvoll. Wer nur die eigenen
// sähe, wüsste nicht, welche schon da sind — und schriebe ein zweites
// „Fliegend".
//
// MSE zeigt sie in **einer** Liste und die festen grau
// (`KeywordList::getItems` und `OnGetItemAttr`,
// `src/gui/control/keyword_list.cpp`, Zeilen 153–165 und 225–235, Fassung
// f9d9356 vom 2021-07-06).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SchluesselwortDTO ist ein Schlüsselwort für die Oberfläche.
type SchluesselwortDTO struct {
	// ID ist die Kennung — **leer bei einem des Spiels**: das gehört keinem
	// Set und lässt sich nicht ändern.
	ID string `json:"id"`

	// Die fünf Felder heißen wie in der Datei, **unübersetzt** (§9).
	Wort       string `json:"wort"`
	Muster     string `json:"muster"`
	Erinnerung string `json:"erinnerung"`
	Regeln     string `json:"regeln"`
	Modus      string `json:"modus"`

	// Fest sagt, dass es aus dem Spiel kommt und unveränderlich ist.
	Fest bool `json:"fest"`

	// Gueltig sagt, dass sein Muster greift. Ein ungültiges wird von MSE
	// abgeschaltet — „that would make MSE hang".
	Gueltig bool `json:"gueltig"`
}

// SchluesselwoerterAntwort trägt die Liste oder einen Fehler.
type SchluesselwoerterAntwort struct {
	// Woerter ist **eine leere Liste und kein `nil`**: über die Wails-Bindung
	// stünde dort sonst `null`.
	Woerter []SchluesselwortDTO `json:"woerter"`

	// Modi sind die Namen der Modi des Spiels — die Auswahl je Schlüsselwort.
	Modi []string `json:"modi"`

	Fehler *Fehler `json:"fehler"`
}

// Schluesselwoerter gibt die des Sets und die des Spiels.
func (a *App) Schluesselwoerter(ctx context.Context, setID string) SchluesselwoerterAntwort {
	ctx = a.imProfil(ctx)

	leer := SchluesselwoerterAntwort{Woerter: []SchluesselwortDTO{}, Modi: []string{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.schluesselwoerter == nil {
		leer.Fehler = nichtEingerichtet()

		return leer
	}

	liste, err := a.schluesselwoerter.Liste(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	return SchluesselwoerterAntwort{
		Woerter: schluesselwortDTO(liste.Woerter),
		Modi:    liste.Modi,
	}
}

// SchluesselwortAnlegen hängt ein leeres ans Ende und gibt seine Kennung.
func (a *App) SchluesselwortAnlegen(ctx context.Context, setID string) SchluesselwortAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchluesselwortAntwort{Fehler: befund}
	}

	if a.schluesselwoerter == nil {
		return SchluesselwortAntwort{Fehler: nichtEingerichtet()}
	}

	neu, err := a.schluesselwoerter.Anlegen(ctx, kennung)
	if err != nil {
		return SchluesselwortAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SchluesselwortAntwort{WortID: kennungText(neu)}
}

// SchluesselwortAntwort trägt die Kennung eines neu angelegten Schlüsselworts.
type SchluesselwortAntwort struct {
	WortID string  `json:"wortID"`
	Fehler *Fehler `json:"fehler"`
}

// SchluesselwortSetzen schreibt ein Feld.
//
// **Nur die des Sets.** Ein Schlüsselwort des Spiels trägt keine Kennung; der
// Weg dorthin ist damit gar nicht erst gangbar, und die Oberfläche sagt es
// zusätzlich.
func (a *App) SchluesselwortSetzen(
	ctx context.Context, setID, wortID, feld, wert string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, wortkennung, befund := beideKennungen(setID, wortID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.schluesselwoerter == nil {
		return SchreibAntwort{Fehler: nichtEingerichtet()}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx,
		a.schluesselwoerter.Setzen(ctx, setkennung, wortkennung, feld, wert))}
}

// SchluesselwortLoeschen nimmt eines weg.
func (a *App) SchluesselwortLoeschen(ctx context.Context, setID, wortID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, wortkennung, befund := beideKennungen(setID, wortID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.schluesselwoerter == nil {
		return SchreibAntwort{Fehler: nichtEingerichtet()}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx,
		a.schluesselwoerter.Loeschen(ctx, setkennung, wortkennung))}
}

// nichtEingerichtet meldet die fehlende Naht.
//
// **Ein eigener Fall und kein Absturz.** Vor `MitSchluesselwoertern` gibt es
// sie nicht — im Betrieb setzt der Composition Root sie, in einem Test nicht
// zwingend. Ein Nilzeiger wäre dort ein Absturz statt einer Meldung.
func nichtEingerichtet() *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: die Schlüsselwörter sind nicht eingerichtet", domain.ErrValidation))
}

// beideKennungen liest Set und Schlüsselwort — **beide benannt** (#808).
func beideKennungen(setID, wortID string) (setkennung, wortkennung int64, befund *Fehler) {
	setkennung, befund = benannteKennungAus("setID", setID)
	if befund != nil {
		return 0, 0, befund
	}

	wortkennung, befund = benannteKennungAus("wortID", wortID)
	if befund != nil {
		return 0, 0, befund
	}

	return setkennung, wortkennung, nil
}

// schluesselwortDTO setzt die Liste um.
func schluesselwortDTO(woerter []service.Schluesselwortsicht) []SchluesselwortDTO {
	umgesetzt := make([]SchluesselwortDTO, 0, len(woerter))

	for _, wort := range woerter {
		kennung := ""
		if wort.ID != 0 {
			kennung = kennungText(wort.ID)
		}

		umgesetzt = append(umgesetzt, SchluesselwortDTO{
			ID: kennung, Wort: wort.Wort, Muster: wort.Muster,
			Erinnerung: wort.Erinnerung, Regeln: wort.Regeln, Modus: wort.Modus,
			Fest: wort.Fest, Gueltig: wort.Gueltig,
		})
	}

	return umgesetzt
}

// Schluesselwortquelle ist die Naht zu den Schlüsselwörtern eines Sets.
//
// **Hier deklariert und nicht in `service`**: `internal/app` nennt, was es
// braucht (§2), und Gos strukturelle Typisierung besorgt die Zuordnung.
type Schluesselwortquelle interface {
	Liste(ctx context.Context, setID int64) (service.Schluesselwortliste, error)
	Anlegen(ctx context.Context, setID int64) (int64, error)
	Setzen(ctx context.Context, setID, wortID int64, feld, wert string) error
	Loeschen(ctx context.Context, setID, wortID int64) error

	// Parameterhilfen gibt beide Menüs zu einem Schlüsselwort
	// (#1342).
	Parameterhilfen(ctx context.Context, setID, wortID int64) (
		service.Parameterhilfen, error)
}

// MitSchluesselwoertern setzt die Naht.
func (a *App) MitSchluesselwoertern(quelle Schluesselwortquelle) {
	a.schluesselwoerter = quelle
}
