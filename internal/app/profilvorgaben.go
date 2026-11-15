// **Profilvorgaben am Fenster und an den Flächen** — Werkstück 15 von
// Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// # Warum das Angebot mitkommt
//
// **Weil die Fläche daraus ihre Auswahl baut** — und weil derselbe Dienst
// danach urteilt. Ein Aufruf, der nur die gesetzten Werte gäbe, zwänge jede
// Fläche, die möglichen selbst zu holen; die Liste stünde dann an zwei Stellen,
// und die zweite liefe zurück (§7, Regel 2).

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Vorgabenfuehrer ist die Naht zur Vorgabenverwaltung.
type Vorgabenfuehrer interface {
	Stand(ctx context.Context, profilID int64) ([]service.Vorgabenstand, error)
	Setzen(ctx context.Context, profilID int64, vorgabe domain.Profilvorgabe) error
}

// MitVorgaben setzt die Naht.
func (a *App) MitVorgaben(fuehrer Vorgabenfuehrer) { a.vorgaben = fuehrer }

// VorgabenAntwort trägt alle fünf Arten.
type VorgabenAntwort struct {
	Vorgaben []service.Vorgabenstand `json:"vorgaben"`

	// Nachladefehler sagt: gesetzt **ist**, nur der Stand danach war nicht
	// zu lesen (§7, Regel 3; #1970) — dann ist `vorgaben` leer und
	// `fehler` auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// VorgabeangabenDTO ist, was eine Fläche an einer Art setzen darf.
type VorgabeangabenDTO struct {
	ProfilID string `json:"profilID"`

	// Art ist `format`, `spiel`, `vorlage`, `deck` oder `lagerort`.
	Art string `json:"art"`

	// Werte sind die gewählten — **eine leere Liste räumt ab**.
	Werte []string `json:"werte"`
}

// Profilvorgaben gibt die Vorgaben eines Profils samt ihrem Angebot.
func (a *App) Profilvorgaben(ctx context.Context, profilID string) VorgabenAntwort {
	ctx = a.imProfil(ctx)

	leer := []service.Vorgabenstand{}

	kennung, befund := benannteKennungAus("profilID", profilID)
	if befund != nil {
		return VorgabenAntwort{Vorgaben: leer, Fehler: befund}
	}

	// **Im gezeigten Profil liest, wer darin steht; an jedem anderen braucht
	// es die Rolle** (#1817).
	if fehler := a.lesbarAm(ctx, kennung); fehler != nil {
		return VorgabenAntwort{Vorgaben: leer, Fehler: Uebersetze(ctx, fehler)}
	}

	staende, fehler := a.vorgaben.Stand(ctx, kennung)
	if fehler != nil {
		return VorgabenAntwort{Vorgaben: leer, Fehler: Uebersetze(ctx, fehler)}
	}

	return VorgabenAntwort{Vorgaben: staende}
}

// ProfilvorgabeSetzen ersetzt die Werte **einer** Art.
//
// **Es setzt und verschmilzt nicht** — dieselbe Regel wie bei den Angaben
// eines Profils: was nicht mitkommt, steht danach nicht mehr da. Eine leere
// Liste ist deshalb keine Auslassung, sondern die Aussage „keine Vorgabe".
func (a *App) ProfilvorgabeSetzen(
	ctx context.Context, angaben VorgabeangabenDTO,
) VorgabenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("profilID", angaben.ProfilID)
	if befund != nil {
		return VorgabenAntwort{Vorgaben: []service.Vorgabenstand{}, Fehler: befund}
	}

	// **Die Rolle am genannten Profil** (#1817) — `query_only` gehört zum
	// gezeigten, und genannt werden kann ein anderes.
	if fehler := a.zumSchreibenAm(ctx, kennung); fehler != nil {
		return VorgabenAntwort{
			Vorgaben: []service.Vorgabenstand{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	werte := angaben.Werte
	if werte == nil {
		werte = []string{}
	}

	if fehler := a.vorgaben.Setzen(ctx, kennung, domain.Profilvorgabe{
		Art: domain.Vorgabeart(angaben.Art), Werte: werte,
	}); fehler != nil {
		return VorgabenAntwort{
			Vorgaben: []service.Vorgabenstand{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	// **Der ganze Stand zurück und nicht nur die eine Art** (§7, Regel 2): die
	// Seite zeigt fünf nebeneinander, und wer nur eine bekäme, müsste danach
	// noch einmal lesen.
	//
	// **Geschrieben ist geschrieben** (§7, Regel 3; #1970): scheitert dieses
	// Lesen — es fragt das Angebot **aller** Arten, das Setzen nur das
	// seiner —, steht sein Fehler daneben und nicht in `fehler`.
	antwort := a.Profilvorgaben(ctx, angaben.ProfilID)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}
