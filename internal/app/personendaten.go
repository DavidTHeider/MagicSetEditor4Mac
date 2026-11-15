// **Der Export personenbezogener Daten an allen vier Flächen** — Werkstück 8
// von Phase 4.12
// (#1351),
// Punkt 10 der Kommentierung vom 2026-09-05.
//
// # Warum die Auskunft aus `service` durchgereicht wird
//
// **Weil sie selbst schon das Ausgabeformat ist.** [service.Personendaten]
// trägt die Marken, unter denen die Datei geschrieben wird; sie hier noch
// einmal als DTO zu erklären hieße, dieselbe Feldliste an zwei Stellen zu
// führen — und die zweite liefe zurück, ohne dass es jemand sähe. Das ist
// dieselbe Überlegung, aus der die Zuordnung fremder Formate in `mtgjson` an
// **einer** Stelle steht (§2).
//
// # Und warum sie am Fenster als Datei herauskommt
//
// **Weil ein Export eine Datei ist.** Was in der Auskunft steht, sieht man auf
// der Seite ohnehin; wozu es sie gibt, ist der Fall, in dem jemand sie
// **mitnehmen** will. Ohne Fenster gibt es keinen Dialog, und dort kommt sie
// als JSON heraus — dieselbe Auskunft, ein anderer Weg.

package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// auskunftPacken schreibt die Auskunft als eingerücktes JSON.
//
// **Eine Naht und kein fester Ruf** — derselbe Grund wie bei [App.packen]:
// [service.Personendaten] trägt nur Zeichenketten, Zeiten, Wahrheitswerte
// und Tabellen davon, und daran scheitert `json.MarshalIndent` nicht. Der
// Fehlerzweig wäre ohne diese Naht eine **Behauptung** darüber, was im
// Ernstfall geschieht, und niemand prüfte sie nach (§3).
//
// **Als Paketvariable und nicht als Feld auf [App]**: er hängt an keiner
// Anwendung, und ein Feld dafür wäre eine Naht, die im Betrieb niemand setzt.
var auskunftPacken = func(auskunft any) ([]byte, error) {
	return json.MarshalIndent(auskunft, "", "  ")
}

// Personenauskunfterin ist die Naht zu `internal/service`.
type Personenauskunfterin interface {
	Personendaten(ctx context.Context, nutzerID int64) (service.Personendaten, error)
}

// PersonendatenAntwort trägt, was über einen Menschen gespeichert ist.
type PersonendatenAntwort struct {
	Daten  *service.Personendaten `json:"daten"`
	Fehler *Fehler                `json:"fehler"`
}

// MitPersonenauskunft setzt die Naht.
func (a *App) MitPersonenauskunft(auskunft Personenauskunfterin) {
	a.personenauskunft = auskunft
}

// Personendaten gibt heraus, was über einen Menschen gespeichert ist.
func (a *App) Personendaten(ctx context.Context, nutzerID string) PersonendatenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return PersonendatenAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return PersonendatenAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	daten, fehler := a.personenauskunft.Personendaten(ctx, kennung)
	if fehler != nil {
		return PersonendatenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return PersonendatenAntwort{Daten: &daten}
}

// PersonendatenSichern fragt nach einem Ziel und schreibt die Auskunft.
//
// **Der Dialog gehört hierher und nicht ins Frontend** — dieselbe Begründung
// wie überall sonst: ein Fenster bekäme einen Namen und keinen Pfad.
//
// **Ein Abbruch ist kein Fehler.** Wer den Dialog schließt, hat nichts
// gewählt.
func (a *App) PersonendatenSichern(ctx context.Context, nutzerID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	// **Vor dem Dialog gefragt** (ID-03): wer erst nach einer Datei fragt
	// und dann absagt, hat jemanden umsonst suchen lassen.
	if befund := a.nutzerVorab(ctx, nutzerID); befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	ziel, err := a.waehleDatei(ctx, Dateiauswahl{
		Titel:     DialogtexteFuer(a.sprache).PersonendatenSichern,
		Filter:    DialogtexteFuer(a.sprache).JSONFilter,
		Muster:    "*.json",
		Vorschlag: "meine-daten.json",
		Speichern: true,
	})
	if err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if ziel == "" {
		return SchreibAntwort{}
	}

	return a.PersonendatenSchreiben(ctx, nutzerID, ziel)
}

// PersonendatenSchreiben schreibt die Auskunft an einen benannten Ort.
//
// **Als eingerücktes JSON.** Sie wird von Menschen gelesen, nicht von einem
// Programm — eine Zeile mit zwanzig Feldern beantwortet die Frage nicht, für
// die es diese Auskunft gibt.
func (a *App) PersonendatenSchreiben(
	ctx context.Context, nutzerID, ziel string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if ziel == "" {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es wurde kein Ziel genannt", domain.ErrValidation))}
	}

	antwort := a.Personendaten(ctx, nutzerID)
	if antwort.Fehler != nil {
		return SchreibAntwort{Fehler: antwort.Fehler}
	}

	roh, fehler := auskunftPacken(antwort.Daten)
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"die Auskunft packen: %w", fehler))}
	}

	// **Modus 0600, auch über einer vorhandenen Datei** (SEC-20, #2133): eine
	// Auskunft über einen Menschen ist für ihn und für niemanden sonst.
	if err := service.SchreibeAtomarVertraulich(ziel, roh); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Auskunft ließ sich nicht schreiben: %s",
			domain.ErrValidation, err.Error()))}
	}

	return SchreibAntwort{}
}
