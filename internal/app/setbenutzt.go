// **Zuletzt benutzt** —
// #891,
// Werkstück 3 von Phase Y.
//
// # Warum das eine eigene Bindung ist
//
// „Zuletzt benutzt" ist eine Aussage über das **Fenster**: jemand hat ein Set
// aufgeschlagen. Kein Lesen sagt das — `Karten` und `Setangaben` werden auch
// gerufen, wenn niemand hinsieht, und ein Leseweg, der nebenbei schreibt, ist
// die Sorte Überraschung, die man später sucht.
//
// # Warum sie nicht an den Flächen ohne Fenster steht
//
// §16 verlangt entweder die Fähigkeit an allen vier Flächen oder einen Grund.
// Der Grund: **es gibt dort nichts zu merken.** MSE2 führt die Liste in seinen
// *Einstellungen* (`Settings::addRecentFile`) — sie ist eine Eigenschaft der
// Bedienung, nicht des Bestands. Eine Kommandozeile, die ein Set anfasst, hat
// kein Menü, in dem es danach oben stünde.

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setbenutzung merkt sich, dass ein Set aufgeschlagen wurde.
//
// **Eine Naht und kein Aufruf in `internal/storage`** (§2). Erfüllt wird sie
// von `storage.SetSpeicher`, verdrahtet in `app.go`.
type Setbenutzung interface {
	SetBenutzt(ctx context.Context, setID int64) error
}

// MitSetbenutzung setzt die Naht.
func (a *App) MitSetbenutzung(benutzung Setbenutzung) { a.benutzung = benutzung }

// SetBenutzt merkt sich, dass ein Set gerade aufgeschlagen wurde.
//
// **Und setzt die Menüleiste neu.** Das Untermenü „Zuletzt benutzt" steht in
// der Ablage und führt die neun jüngsten Sets; ohne diesen Schritt zeigte es
// bis zum nächsten Start die Reihenfolge von vorhin.
func (a *App) SetBenutzt(ctx context.Context, setID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.benutzung == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Benutzung eines Sets ist nicht eingerichtet",
			domain.ErrValidation))}
	}

	if err := a.benutzung.SetBenutzt(ctx, kennung); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	a.arbeitVermerken(ctx, domain.ArbeitsartSet, setID, domain.VorgangGeoeffnet)

	a.menueNeu(ctx)

	return SchreibAntwort{}
}

// menueNeu setzt die Leiste mit der heutigen Liste der jüngsten Sets.
//
// **Ohne Wache auf das Fenster.** `SpracheSetzen` setzte die Leiste seit jeher
// unbedingt; die Naht `menueSetzen` ist der Ort, an dem entschieden wird, was
// ohne Fenster geschieht — nicht diese Zeile.
//
// **Mit dem Stand, den das Fenster zuletzt gemeldet hat** (UXR-F14, #2029): die
// Leiste entsteht neu, und ohne ihn stünden Seitenleiste, Navigator und
// Inspector danach wieder gegraut da. **Gelesen wird der Bestand vor der
// Sperre** und die Liste für den nächsten Stand gemerkt (`menuenaht.go`).
func (a *App) menueNeu(ctx context.Context) {
	zuletzt := a.zuletztBenutzte(ctx)

	a.menuesperre.Lock()
	defer a.menuesperre.Unlock()

	a.zuletztliste = zuletzt
	a.menueAufbauen()
}

// zuletztBenutzte gibt die neun jüngsten Sets.
//
// **Neun, wie MSE2** (`Settings::max_recent_sets`, `src/data/settings.hpp`,
// Fassung f9d9356 vom 2021-07-06, Zeile 139).
//
// **Ein Fehler beim Lesen gibt eine leere Liste.** Die Menüleiste ist
// Bestenfalls-Arbeit; ein Bestand, der sich nicht lesen lässt, meldet sich an
// jeder anderen Stelle lauter, und ein Programm, das deswegen keine Leiste
// bekäme, wäre unbedienbar.
func (a *App) zuletztBenutzte(ctx context.Context) []Zuletzteintrag {
	if a.bestand == nil {
		return nil
	}

	koepfe, err := a.leser().Setkoepfe(ctx)
	if err != nil {
		return nil
	}

	eintraege := make([]Zuletzteintrag, 0, hoechstensZuletzt)

	for _, kopf := range koepfe {
		if len(eintraege) == hoechstensZuletzt {
			break
		}

		eintraege = append(eintraege, Zuletzteintrag{
			ID:   kennungText(kopf.ID),
			Name: kopf.Name,
		})
	}

	return eintraege
}
