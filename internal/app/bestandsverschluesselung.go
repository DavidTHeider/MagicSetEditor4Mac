// **Die Verschlüsselung des Bestands am Fenster** — Werkstück 14 von Phase
// 4.12 (#1351),
// nach [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// **Keine fachliche Entscheidung hier** (§2): übersetzt wird zwischen Wails
// und dem Dienst, mehr nicht.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Verschluesselungsantwort sagt, wie der Bestand liegt.
//
// **Der Stand kommt auf jedem Weg zurück** — auch nach dem Umschalten. Sonst
// müsste die Seite ihn nachladen, und zwischen Antwort und Nachladen läge ein
// Stand, den niemand kennt (§7, Regel 3).
type Verschluesselungsantwort struct {
	// Verschluesselt sagt, ob die Datei eine Chiffre trägt.
	Verschluesselt bool `json:"verschluesselt"`

	Fehler *Fehler `json:"fehler"`
}

// Verschluesselungsstand sagt, ob der Bestand verschlüsselt auf der Platte
// liegt.
func (a *App) Verschluesselungsstand(ctx context.Context) Verschluesselungsantwort {
	ctx = a.imProfil(ctx)

	an, fehler := a.bestandsweg().Verschluesselungsstand()

	// **Dieselbe Form wie beim Umschalten**: Stand und Fehler reisen zusammen.
	// Ein eigener Ausgang für den Fehlerfall wäre ein Zweig mehr, der dasselbe
	// täte — `Uebersetze(nil)` gibt `nil`.
	return Verschluesselungsantwort{
		Verschluesselt: an,
		Fehler:         Uebersetze(ctx, fehler),
	}
}

// VerschluesselungUmschalten schaltet sie an oder aus.
//
// # Ein Ruf und nicht zwei
//
// **Weil es ein Schalter ist.** Zwei Bindungen — einschalten, ausschalten —
// hießen an der Oberfläche zwei Knöpfe, von denen immer einer nichts tut; und
// wer den falschen drückt, bekäme eine Abweisung statt einer Wirkung.
//
// Die Abweisung bleibt trotzdem: wer *einschalten* verlangt, während es schon
// an ist, bekommt sie. Das ist kein Widerspruch, sondern der Fall, in dem ein
// zweiter Ruf gegen einen Stand geht, den die Fläche noch nicht kennt (§7,
// Regel 4).
func (a *App) VerschluesselungUmschalten(
	ctx context.Context, an bool,
) Verschluesselungsantwort {
	ctx = a.imProfil(ctx)

	var stand bool

	fehler := a.imTausch(func() error {
		var tauschfehler error
		stand, tauschfehler = umschalten(ctx, a.bestandsweg(), an)

		return tauschfehler
	})

	// **Der Stand reist auch mit einem Fehler mit.** Ein Tausch kann zur
	// Hälfte stattgefunden haben, und dann ist „was liegt jetzt da?" die
	// Frage, an der die Seite hängt — sie soll nicht nachladen müssen (§7,
	// Regel 3).
	return Verschluesselungsantwort{
		Verschluesselt: stand,
		Fehler:         Uebersetze(ctx, fehler),
	}
}

// umschalten wählt die Richtung.
//
// **Hier und nicht in der Bindung**, weil MCP, Kommandozeile und Dienst
// dieselbe Wahl treffen — und drei Kopien einer Fallunterscheidung laufen
// auseinander.
func umschalten(
	ctx context.Context, weg *service.Bestandsweg, an bool,
) (bool, error) {
	if an {
		return weg.VerschluesselungEinschalten(ctx)
	}

	return weg.VerschluesselungAusschalten(ctx)
}

// Verschluesselungsstand sagt, ob der Bestand verschlüsselt auf der Platte
// liegt.
//
// **Die zwei Bindungen stehen hier und nicht in `bindungen.go`** — dort wäre
// die Datei über die Zeilengrenze gewachsen (§2), und der Schnitt folgt der
// Frage: *wie kommt die Verschlüsselung ans Fenster* ist eine eigene.
func (b *Bindungen) Verschluesselungsstand(ctx context.Context) Verschluesselungsantwort {
	return b.app(ctx).Verschluesselungsstand(b.ruf(ctx))
}

// VerschluesselungUmschalten schaltet sie an oder aus.
//
// **Ein Ruf und nicht zwei** — der Grund steht bei
// [App.VerschluesselungUmschalten].
func (b *Bindungen) VerschluesselungUmschalten(ctx context.Context, an bool) Verschluesselungsantwort {
	return b.app(ctx).VerschluesselungUmschalten(b.ruf(ctx), an)
}
