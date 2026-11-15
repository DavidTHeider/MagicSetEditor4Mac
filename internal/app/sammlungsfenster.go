package app

// **Nur holen, was gezeichnet wird** — Befund aus U8, Werkstück
// #1500.
//
// # Der Befund, gemessen
//
// **Bei 10.000 Exemplaren steht das Gitter nach 209 bis 371 ms** (drei Läufe,
// 1280 × 800, 2026-09-09) — das Ziel aus `docs/UX-ARCHITEKTUR.md` sind 200 ms.
// Was bleibt, ist **nicht das Zeichnen**: die Fensterliste hält 24 Zeilen im
// DOM. Es ist die Naht — `Sammlung()` schickt alle Exemplare in einem Ruf,
// rund 3 MB JSON.
//
// # Und warum die Sicht mitgeht
//
// **Weil ein Ausschnitt ohne sie der falsche wäre.** Wer nach Menge sortiert
// und die ersten zweihundert holt, bekommt die zweihundert größten Stapel —
// und nicht die ersten zweihundert nach Kennung, die dann jemand im Fenster
// sortiert. Filtern und Ordnen gehören dorthin, wo die ganze Liste liegt.
//
// **Der Ruf ohne Ausschnitt bleibt** ([Sammlungsantwort]): das Issue verlangt
// es ausdrücklich („erweiternd"), und die Gruppenansicht braucht ihn.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungsfensterantwort ist ein Ausschnitt samt der Gesamtzahl.
type Sammlungsfensterantwort struct {
	Exemplare []ExemplarDTO `json:"exemplare"`

	// Gesamt ist, wie viele Zeilen die **Sicht** hat — nicht, wie viele die
	// Sammlung hat.
	//
	// **Beides steht da**, weil beides gebraucht wird: `aria-rowcount` nennt
	// die Zeilen der Liste, und „N von M" nennt daneben, wie viele es
	// überhaupt sind.
	Gesamt int `json:"gesamt"`

	// Insgesamt ist die Zahl der Exemplare **ohne** Sicht.
	Insgesamt int `json:"insgesamt"`

	// Stueck ist die Summe der Mengen **ohne** Sicht — U15 (#2160): die
	// Startseite braucht sie und holte dafür die ganze Liste.
	Stueck int `json:"stueck"`

	Fehler *Fehler `json:"fehler"`
}

// Sammlungsausschnitt ist der Dienst hinter dieser Bindung.
//
// **Die Schnittstelle steht hier, wo sie gebraucht wird** (§2) — erfüllt wird
// sie von `service.Sammlungsfenster`.
type Sammlungsausschnitt interface {
	Sicht(ctx context.Context, sicht domain.Sammlungssicht, ab, anzahl int) (
		fenster []*domain.Exemplar, zahlen domain.Fensterzahlen, fehler error)
}

// MitSammlungsfenster setzt den Dienst hinter [App.Sammlungsfenster].
func (a *App) MitSammlungsfenster(dienst Sammlungsausschnitt) {
	a.sammlungsfenster = dienst
}

// Sammlungsfenster gibt den Ausschnitt einer Sicht.
//
// **`anzahl` von null heißt „alles"** — die Gruppenansicht braucht die ganze
// Liste, und ein zweiter Weg dafür wären zwei Wahrheiten über dieselbe Sicht.
func (a *App) Sammlungsfenster(
	ctx context.Context, suche, ausfuehrung, sortierung string, ab, anzahl int,
) Sammlungsfensterantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungsfenster == nil {
		return Sammlungsfensterantwort{
			Exemplare: []ExemplarDTO{}, Fehler: Uebersetze(ctx, keineSammlung()),
		}
	}

	fenster, zahlen, fehler := a.sammlungsfenster.Sicht(ctx,
		domain.Sammlungssicht{
			Suche:       suche,
			Ausfuehrung: domain.Ausfuehrung(ausfuehrung),
			Sortierung:  domain.Sortierung(sortierung),

			// **Die Ordnung folgt der Sprache des Fensters** — LOC-03.AK-01
			// (#1689);
			// ohne sie ordnet der Kern nach Deutsch.
			Sprache: a.sprache,
		}, ab, anzahl)
	if fehler != nil {
		return Sammlungsfensterantwort{
			Exemplare: []ExemplarDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	return Sammlungsfensterantwort{
		Exemplare: zuExemplarDTOs(fenster),
		Gesamt:    zahlen.Gesamt,
		Insgesamt: zahlen.Insgesamt,
		Stueck:    zahlen.Stueck,
	}
}

// Sammlungsfenster ist die Fensterbindung.
func (b *Bindungen) Sammlungsfenster(
	ctx context.Context,
	suche, ausfuehrung, sortierung string, ab, anzahl int,
) Sammlungsfensterantwort {
	return b.app(ctx).Sammlungsfenster(
		b.ruf(ctx), suche, ausfuehrung, sortierung, ab, anzahl)
}
