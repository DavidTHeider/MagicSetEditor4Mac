package app

// Das **Zufallspack** — einen Booster aus dem offenen Set ziehen
// (#1072,
// Werkstück 4 von Phase W).
//
// # Was MSE hier tut
//
// `src/gui/set/random_pack_panel.cpp` (Fassung f9d9356 vom 2021-07-06): links
// die Packarten mit einer Anzahl je Art, rechts die gezogenen Karten. Der
// Same steht dort als „seed" in einem Feld — dieselbe Zahl gibt dieselbe
// Ziehung.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PackartDTO ist eine Packart, wie die Oberfläche sie anbietet.
type PackartDTO struct {
	// Name ist der Name aus dem Paket — **unübersetzt** (§9).
	Name string `json:"name"`

	// Waehlbar sagt, ob MSE sie zum Ziehen anbietet. Im echten Pack sind es
	// vier von dreißig; die übrigen sind Bausteine.
	Waehlbar bool `json:"waehlbar"`

	// Karten ist die Zahl der Karten, aus denen sie **erreichbar** zieht.
	Karten int `json:"karten"`
}

// PackAntwort trägt die Packarten oder einen Fehler.
type PackAntwort struct {
	Arten []PackartDTO `json:"arten"`

	// Befunde sind die Packarten, deren Filter sich nicht ausrechnen ließ.
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// PackziehungAntwort trägt die gezogenen Karten.
type PackziehungAntwort struct {
	// Karten stehen in ihrer **Ziehreihenfolge** — so wie MSE sie legt.
	Karten []KarteDTO `json:"karten"`

	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// Packquelle ist die Naht zum Zufallspack.
type Packquelle interface {
	Angebot(ctx context.Context, setID int64) ([]service.Packangebot, []domain.Befund, error)
	Ziehen(ctx context.Context, setID int64, art string, same uint64) (
		[]*domain.CardRow, []domain.Befund, error,
	)
}

// MitPack setzt die Naht.
func (a *App) MitPack(quelle Packquelle) { a.pack = quelle }

// Packarten gibt die Packarten des Spiels hinter einem Set.
func (a *App) Packarten(ctx context.Context, setID string) PackAntwort {
	ctx = a.imProfil(ctx)

	leer := PackAntwort{Arten: []PackartDTO{}, Befunde: []string{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.pack == nil {
		leer.Fehler = keinPack()

		return leer
	}

	arten, befunde, err := a.pack.Angebot(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	umgesetzt := make([]PackartDTO, 0, len(arten))
	for _, art := range arten {
		umgesetzt = append(umgesetzt,
			PackartDTO{Name: art.Name, Waehlbar: art.Waehlbar, Karten: art.Karten})
	}

	return PackAntwort{Arten: umgesetzt, Befunde: befundtexte(befunde)}
}

// PackZiehen zieht ein Pack und gibt die Karten in ihrer Ziehreihenfolge.
//
// **Der Same kommt von außen**, und das ist die Zusage: dieselbe Zahl gibt
// dieselbe Ziehung. Ohne sie könnte niemand ein Ergebnis noch einmal
// herstellen — und kein Test wäre möglich (#889).
//
// **Als Text und nicht als Zahl**: über die Wails-Bindung käme ein `uint64`
// als JavaScript-Zahl an, und die verliert oberhalb von 2⁵³ die letzten
// Stellen. Ein Same, der beim Hinweg schon ein anderer ist, wäre keine
// Wiederholbarkeit.
func (a *App) PackZiehen(ctx context.Context, setID, art, same string) PackziehungAntwort {
	ctx = a.imProfil(ctx)

	leer := PackziehungAntwort{Karten: []KarteDTO{}, Befunde: []string{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.pack == nil {
		leer.Fehler = keinPack()

		return leer
	}

	samen, samenfehler := samenAus(same)
	if samenfehler != nil {
		leer.Fehler = samenfehler

		return leer
	}

	karten, befunde, err := a.pack.Ziehen(ctx, kennung, art, samen)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	return PackziehungAntwort{
		Karten: kartenDTO(karten), Befunde: befundtexte(befunde),
	}
}

// keinPack meldet die fehlende Naht.
func keinPack() *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: das Zufallspack ist nicht eingerichtet", domain.ErrValidation))
}

// samenAus liest den Samen aus seinem Text.
//
// **Als Text und nicht als Zahl** — der Grund steht an [App.PackZiehen]. Ein
// leerer Text heißt null: das ist ein gültiger Same, und wer keinen angibt,
// bekommt eine Ziehung, die sich wiederholen lässt.
func samenAus(text string) (uint64, *Fehler) {
	if text == "" {
		return 0, nil
	}

	samen, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return 0, Nutzerfehler(domain.NewFieldError("same", domain.SchluesselValidierung,
			"Der Same %q ist keine ganze Zahl ab null.", text))
	}

	return samen, nil
}
