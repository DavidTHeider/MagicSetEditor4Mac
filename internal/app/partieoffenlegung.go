package app

// **Freigeben, zurücknehmen, vergessen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Die drei Wege, und warum es drei sind
//
// Das Issue verlangt „Persistentes Protokoll/Replay mit Sichtrechten, Consent
// für gegebenenfalls spätere Offenlegung und Aufbewahrungs-/Löschweg". Die
// **Sichtrechte** stehen schon: [App.PartieVerlauf] schneidet ab, was einem
// Leser nicht gehört. Hier kommen die beiden anderen dazu — die Einwilligung
// und der Löschweg.
//
// # Warum die Einwilligung `schreibend` ist und trotzdem nichts spielt
//
// **Weil sie den Bestand ändert.** Sie greift in keine Partie ein und erzeugt
// kein Ereignis; was sie ändert, ist, wer ein Protokoll lesen darf — und das
// ist eine Zusage über Daten und damit eine Änderung
// (`docs/SCHREIBWEGE.md`).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// OffenlegungDTO ist die Einwilligung eines Sitzes.
type OffenlegungDTO struct {
	// Von ist der Sitz, dessen private Ergänzungen es sind.
	Von string `json:"von"`

	// An sind die Sitze, die mitlesen dürfen — `*` heißt: jeder.
	An []string `json:"an"`
}

// PartieoffenlegungAntwort trägt alle Einwilligungen einer Partie.
type PartieoffenlegungAntwort struct {
	Offenlegungen []OffenlegungDTO `json:"offenlegungen"`

	// Nachladefehler sagt: offengelegt oder zurückgenommen **ist**, nur die
	// Liste danach war nicht zu lesen (§7, Regel 3; #1970) — dann ist
	// `offenlegungen` leer und `fehler` auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// PartieOffenlegungen sagt, wer wem etwas freigegeben hat.
func (a *App) PartieOffenlegungen(
	ctx context.Context, kennung string,
) PartieoffenlegungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	gaben, fehler := a.partienaht().Freigaben(ctx, kennung)
	if fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	return PartieoffenlegungAntwort{Offenlegungen: offenlegungenAus(gaben)}
}

// PartieOffenlegen trägt die Einwilligung eines Sitzes ein.
//
// **Die alte wird ersetzt und nicht ergänzt** — der Grund steht in
// `service`: wer eine Liste schickt, meint diese Liste.
func (a *App) PartieOffenlegen(
	ctx context.Context, kennung, von string, an []string,
) PartieoffenlegungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	fehler := a.partienaht().Offenlegen(ctx, kennung,
		domain.Offenlegung{Von: von, An: an})
	if fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	return a.offenlegungenNachDemSchreiben(ctx, kennung)
}

// PartieOffenlegungZuruecknehmen schließt die Tür wieder.
//
// **Und holt nichts ein.** Was hinausging, ist draußen — das Fenster sagt es
// mit, und diese Naht behauptet es nicht anders.
func (a *App) PartieOffenlegungZuruecknehmen(
	ctx context.Context, kennung, von string,
) PartieoffenlegungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	if fehler := a.partienaht().Zuruecknehmen(ctx, kennung, von); fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	return a.offenlegungenNachDemSchreiben(ctx, kennung)
}

// PartieVergessen löscht das Protokoll einer beendeten Partie.
func (a *App) PartieVergessen(
	ctx context.Context, kennung string,
) PartieoffenlegungAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	if fehler := a.partienaht().Vergessen(ctx, kennung); fehler != nil {
		return leereOffenlegung(ctx, fehler)
	}

	return PartieoffenlegungAntwort{Offenlegungen: []OffenlegungDTO{}}
}

// offenlegungenNachDemSchreiben gibt die Liste nach einem **geglückten**
// Schritt.
//
// **Geschrieben ist geschrieben** (§7, Regel 3;
// #1970):
// scheitert das Lesen danach, steht sein Fehler in
// [PartieoffenlegungAntwort.Nachladefehler] und nicht in `fehler`.
func (a *App) offenlegungenNachDemSchreiben(
	ctx context.Context, kennung string,
) PartieoffenlegungAntwort {
	antwort := a.PartieOffenlegungen(ctx, kennung)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// leereOffenlegung gibt eine Antwort mit leerer Liste und übersetztem Fehler.
//
// **Eine leere Liste und kein `nil`** — dieselbe Zusage auf dem Fehlerweg wie
// auf dem Erfolgsweg. Genau hier ist sie in Phase 5.4 zweimal gebrochen.
func leereOffenlegung(
	ctx context.Context, fehler error,
) PartieoffenlegungAntwort {
	return PartieoffenlegungAntwort{
		Offenlegungen: []OffenlegungDTO{},
		Fehler:        Uebersetze(ctx, fehler),
	}
}

// offenlegungenAus bildet die Einwilligungen ab.
func offenlegungenAus(gaben []domain.Offenlegung) []OffenlegungDTO {
	uebersetzt := make([]OffenlegungDTO, 0, len(gaben))

	for _, gabe := range gaben {
		an := gabe.An
		if an == nil {
			an = []string{}
		}

		uebersetzt = append(uebersetzt, OffenlegungDTO{Von: gabe.Von, An: an})
	}

	return uebersetzt
}
