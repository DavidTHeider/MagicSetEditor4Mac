// **Die Preise am Fenster** — Werkstück 3 von Phase 3.4
// (#1165).
//
// Vier Bindungen: holen, nachschlagen, Verlauf, Übersicht.
//
// # Der Vorbehalt reist mit
//
// Jede Antwort trägt **Quelle**, **Zeitpunkt** und **Alter**. Die Quelle sagt
// über ihre Preise ausdrücklich, dass sie für nichts einsteht (nachgelesen am
// 2026-08-27); eine Zahl ohne diese drei Angaben wäre hier eine Behauptung,
// die niemand deckt.

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Preisquelle ist, was die Preisseite braucht.
type Preisquelle interface {
	Holen(ctx context.Context, printing string) (
		[]domain.Preisstand, time.Duration, []domain.Befund, error)
	Neueste(ctx context.Context, printing string) ([]domain.Preisstand, error)
	Verlauf(ctx context.Context, printing string, reihe domain.Preisstand) (
		[]domain.Preisstand, error)
	Portfolio(ctx context.Context) (service.Portfolio, error)
}

// PreisstandDTO ist ein Preisstand, wie das Fenster ihn sieht.
type PreisstandDTO struct {
	Printing string `json:"printing"`

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441), als
	// Zeichenkette — leer, wo die Ablage keine kennt.
	DruckID string `json:"druckID"`

	// **Die sieben Teile der Reihe** (A3, #1442) — sie sagen, **welche**
	// Frage dieser Preis beantwortet.
	Quelle   string `json:"quelle"`
	Markt    string `json:"markt"`
	Art      string `json:"art"`
	Waehrung string `json:"waehrung"`
	Finish   string `json:"finish"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`

	// Lage sagt, was feststeht: `genannt`, `weggefallen` oder `ungeprueft`.
	Lage string `json:"lage"`

	// Wert ist der Preis in der Form, in der man ihn liest — `"4.17"`.
	//
	// **Leer, wo die Lage nicht `genannt` ist**: dort steht null, und sie
	// bedeutet nichts. Eine `"0.00"` sähe aus wie „umsonst".
	Wert string `json:"wert"`

	// Beobachtet ist der Tag, den die **Quelle** nennt, in RFC 3339 — leer,
	// wo sie keinen nennt (Scryfall nennt keinen).
	Beobachtet string `json:"beobachtet"`

	// Erhoben ist der Zeitpunkt in RFC 3339.
	//
	// **Als Zeichenkette und in UTC.** Ein `time.Time` über JSON ist beides
	// schon; was hier zählt, ist dass die Zone **dransteht** — eine Ortszeit
	// ohne sie liest sich auf einem anderen Rechner anders.
	Erhoben string `json:"erhoben"`

	// AlterStunden sagt, wie alt der Stand ist.
	//
	// **In Stunden und nicht in Nanosekunden**: `time.Duration` ist über JSON
	// eine achtzehnstellige Zahl, mit der im Fenster niemand rechnen will.
	AlterStunden int `json:"alterStunden"`
}

// Preisantwort trägt die Preisstände eines Drucks.
type Preisantwort struct {
	Staende []PreisstandDTO `json:"staende"`

	// Befunde sind die Preise, die nicht zu deuten waren (§8).
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// PortfoliozeileDTO ist die Summe einer Währung.
type PortfoliozeileDTO struct {
	Waehrung string `json:"waehrung"`

	// Wert ist die Summe in der Form, in der man sie liest.
	Wert string `json:"wert"`

	Exemplare      int    `json:"exemplare"`
	AeltesterPreis string `json:"aeltesterPreis"`
	AlterStunden   int    `json:"alterStunden"`
	Quelle         string `json:"quelle"`
}

// Portfolioantwort trägt die Übersicht.
type Portfolioantwort struct {
	Zeilen []PortfoliozeileDTO `json:"zeilen"`

	// OhnePreis ist die Zahl der Exemplare, für die kein Preis vorliegt.
	OhnePreis int `json:"ohnePreis"`

	Fehler *Fehler `json:"fehler"`
}

// MitPreisen setzt die Naht.
func (a *App) MitPreisen(quelle Preisquelle) { a.preise = quelle }

// PreiseHolen fragt die Quelle nach den Preisen eines Drucks.
//
// **Auf Anweisung und nicht von allein** — ein Programm, das beim Öffnen
// Netzverkehr für dreihundert Karten auslöst, überrascht seine Nutzenden.
func (a *App) PreiseHolen(ctx context.Context, printing string) Preisantwort {
	ctx = a.imProfil(ctx)

	if a.preise == nil {
		return leerePreisantwort(Uebersetze(ctx, keinePreise()))
	}

	staende, alter, befunde, err := a.preise.Holen(ctx, printing)
	if err != nil {
		antwort := leerePreisantwort(Uebersetze(ctx, err))
		antwort.Befunde = alsMeldungen(befunde)

		return antwort
	}

	return Preisantwort{
		Staende: zuPreisstandDTOs(staende, alter),
		Befunde: alsMeldungen(befunde),
	}
}

// Preise gibt die abgelegten Stände eines Drucks — ohne zu fragen.
func (a *App) Preise(ctx context.Context, printing string) Preisantwort {
	ctx = a.imProfil(ctx)

	if a.preise == nil {
		return leerePreisantwort(Uebersetze(ctx, keinePreise()))
	}

	staende, err := a.preise.Neueste(ctx, printing)
	if err != nil {
		return leerePreisantwort(Uebersetze(ctx, err))
	}

	return Preisantwort{
		Staende: zuPreisstandDTOs(staende, alterSeit(a.jetzt(), staende)),
		Befunde: []string{},
	}
}

// Preisverlauf gibt eine Reihe in der Zeit.
//
// **Die Reihe wird genannt und nicht erraten** (A3, #1442): mit MTGJSON neben
// Scryfall wäre eine fest eingesetzte Quelle die Antwort auf eine Frage, die
// niemand gestellt hat.
func (a *App) Preisverlauf(ctx context.Context, reihe PreisreiheDTO) Preisantwort {
	ctx = a.imProfil(ctx)

	if a.preise == nil {
		return leerePreisantwort(Uebersetze(ctx, keinePreise()))
	}

	staende, err := a.preise.Verlauf(ctx, reihe.Printing, reihe.alsReihe())
	if err != nil {
		return leerePreisantwort(Uebersetze(ctx, err))
	}

	return Preisantwort{
		Staende: zuPreisstandDTOs(staende, alterSeit(a.jetzt(), staende)),
		Befunde: []string{},
	}
}

// Portfolio gibt die Übersicht über die eigene Sammlung.
func (a *App) Portfolio(ctx context.Context) Portfolioantwort {
	ctx = a.imProfil(ctx)

	if a.preise == nil {
		return Portfolioantwort{
			Zeilen: []PortfoliozeileDTO{}, Fehler: Uebersetze(ctx, keinePreise()),
		}
	}

	uebersicht, err := a.preise.Portfolio(ctx)
	if err != nil {
		return Portfolioantwort{Zeilen: []PortfoliozeileDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return Portfolioantwort{
		Zeilen:    zuPortfoliozeilen(uebersicht.Zeilen, a.jetzt()),
		OhnePreis: uebersicht.OhnePreis,
	}
}

// leerePreisantwort gibt eine Antwort mit leeren Listen und einem Fehler.
//
// **Leere Listen und kein `nil`** (§3): sonst käme im Fenster `undefined` an,
// wo eine Liste erwartet wird.
func leerePreisantwort(fehler *Fehler) Preisantwort {
	return Preisantwort{
		Staende: []PreisstandDTO{}, Befunde: []string{}, Fehler: fehler,
	}
}

// keinePreise meldet die fehlende Naht.
func keinePreise() error {
	return fmt.Errorf("%w: die Preise sind nicht eingerichtet", domain.ErrValidation)
}
