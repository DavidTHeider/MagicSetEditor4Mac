package app

// **Der Plan vor einer Paketänderung, Ersetzen und zurück** — DOM-01.AK-02
// und AK-03 (#1696).
//
// Die Anwendung reicht durch; gerechnet wird in `service` (Paketplan,
// Ersetzen, Zurueck). **Ein Weg für alle Pläne**, mit dem Vorgang als
// Angabe: vier Wege, die sich nur im Namen unterschieden, wären vier Stellen
// für dieselbe Übersetzung.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PaketplanDTO ist ein Plan, wie das Fenster ihn zeigt.
type PaketplanDTO struct {
	Vorgang string `json:"vorgang"`
	Paket   string `json:"paket"`

	// Quelle ist der Ort, aus dem eingebunden oder ersetzt würde — leer beim
	// Entfernen und beim Zurück.
	Quelle string `json:"quelle"`

	Von       string             `json:"von"`
	Zu        string             `json:"zu"`
	Geaendert bool               `json:"geaendert"`
	Moeglich  bool               `json:"moeglich"`
	Sets      []string           `json:"sets"`
	Konflikte []AbhaengigkeitDTO `json:"konflikte"`
}

// PaketplanAntwort trägt einen Plan — oder einen Abbruch.
type PaketplanAntwort struct {
	Plan        *PaketplanDTO `json:"plan"`
	Abgebrochen bool          `json:"abgebrochen"`
	Fehler      *Fehler       `json:"fehler"`
}

// PaketPlan plant einen Vorgang — `entfernen` und `zurueck` an einem
// vorhandenen Paket, `einbinden` aus einer Quelle (ein Ersetzen, wenn es das
// Paket schon gibt).
func (a *App) PaketPlan(ctx context.Context, vorgang, art, name, quelle string) PaketplanAntwort {
	ctx = a.imProfil(ctx)

	plan, err := planen(ctx, a.pakete, vorgang, art, name, quelle)
	if err != nil {
		return PaketplanAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return PaketplanAntwort{Plan: &plan}
}

// planen wählt den Plan zum Vorgang und übersetzt ihn — **eine Stelle** für
// das Fenster und für [Paketweg.Plan].
func planen(
	ctx context.Context, fuehrer Paketfuehrer, vorgang, art, name, quelle string,
) (PaketplanDTO, error) {
	var (
		plan service.Paketplan
		err  error
	)

	switch service.Paketvorgang(vorgang) {
	case service.PaketvorgangEntfernen:
		plan, err = fuehrer.PlanEntfernen(ctx, art, name)
	case service.PaketvorgangEinbinden:
		plan, err = fuehrer.PlanEinbinden(ctx, quelle)
	case service.PaketvorgangZurueck:
		plan, err = fuehrer.PlanZurueck(ctx, art, name)
	default:
		err = domain.NewFieldError("vorgang", domain.SchluesselValidierung,
			"»%s« ist kein Vorgang; geplant werden entfernen, einbinden und zurueck.", vorgang)
	}

	if err != nil {
		return PaketplanDTO{}, err
	}

	dto := PaketplanDTO{
		Vorgang: string(plan.Vorgang), Paket: plan.Paket, Von: plan.Von, Zu: plan.Zu,
		Geaendert: plan.Geaendert, Moeglich: plan.Moeglich,
		Sets: append([]string{}, plan.Sets...), Konflikte: alsAbhaengigkeiten(plan.Konflikte),
	}

	if plan.Vorgang == service.PaketvorgangEinbinden || plan.Vorgang == service.PaketvorgangErsetzen {
		dto.Quelle = quelle
	}

	return dto, nil
}

// PaketErsetzen legt eine neue Fassung ein — mit `bestaetigt`, wenn das
// vorhandene Paket lokal geändert ist.
func (a *App) PaketErsetzen(ctx context.Context, quelle string, bestaetigt bool) PaketAntwort {
	ctx = a.imProfil(ctx)

	return a.eingefuehrt(ctx)(a.pakete.Ersetzen(ctx, quelle, bestaetigt))
}

// PaketZurueck holt die vorige Fassung zurück.
func (a *App) PaketZurueck(ctx context.Context, art, name string) PaketAntwort {
	ctx = a.imProfil(ctx)

	return a.eingefuehrt(ctx)(a.pakete.Zurueck(ctx, art, name))
}

// eingefuehrt macht aus einem Paket samt Fehler eine Antwort — und vermerkt
// die Arbeit daran wie beim Einbinden: eine Fassung kam herein.
func (a *App) eingefuehrt(ctx context.Context) func(domain.Paket, error) PaketAntwort {
	return func(paket domain.Paket, err error) PaketAntwort {
		if err != nil {
			return PaketAntwort{Fehler: Uebersetze(ctx, err)}
		}

		gelesen := a.mitUebersicht(ctx, paket)
		a.arbeitVermerken(ctx, domain.ArbeitsartPaket, gelesen.Name, domain.VorgangEingefuehrt)

		return PaketAntwort{Paket: &gelesen}
	}
}
