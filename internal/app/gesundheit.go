package app

// **Die Gesundheit: lebt er, kann er arbeiten** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Was hier nicht entschieden wird
//
// **Nichts.** Diese Datei übersetzt zwischen `service.Betriebsschau` und den
// Flächen (§2). Ob ein fremder Ausfall die Bereitschaft nimmt, steht in
// `domain.Betriebszustand` — an **einer** Stelle, damit vier Flächen dieselbe
// Antwort geben.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Gesundheitsnaht ist, was die Fassade von der Schau braucht.
type Gesundheitsnaht interface {
	Zustand(ctx context.Context) domain.Betriebszustand
}

// FremdzustandDTO ist ein fremder Dienst für eine Fläche.
type FremdzustandDTO struct {
	Dienst     string `json:"dienst"`
	Erreichbar bool   `json:"erreichbar"`
	Grund      string `json:"grund"`
}

// GesundheitAntwort trägt den Betriebszustand.
//
// **`bereit` steht als eigenes Feld da, obwohl es sich rechnen lässt.** Eine
// Fläche soll die Regel nicht nachbauen müssen — und vier Nachbauten liefen
// auseinander (§16).
type GesundheitAntwort struct {
	Lebt   bool              `json:"lebt"`
	Bereit bool              `json:"bereit"`
	Grund  string            `json:"grund"`
	Fremde []FremdzustandDTO `json:"fremde"`
	Fehler *Fehler           `json:"fehler"`
}

// MitGesundheit hängt die Schau ein.
func (a *App) MitGesundheit(naht Gesundheitsnaht) { a.gesundheit = naht }

// Gesundheit sagt, ob dieser Prozess lebt und arbeiten kann.
//
// **Ohne Naht lebt er und ist nicht bereit** — und das ist kein Fehler,
// sondern die Wahrheit: ein Prozess ohne Bestand kann nicht arbeiten. Ein
// Befund hier machte aus einem gültigen Zustand eine Störung.
// **`imProfil` steht hier, obwohl dieser Weg kein Profil braucht** — der
// Wächter über die Bindungen verlangt die Zeile von jeder, und sein Kopf sagt
// warum: eine Aufzählung der Ausnahmen wäre eine zweite Liste, die veralten
// kann, und der Preis für eine Zeile zu viel ist null.
func (a *App) Gesundheit(ctx context.Context) GesundheitAntwort {
	ctx = a.imProfil(ctx)

	if a.gesundheit == nil {
		return GesundheitAntwort{
			Lebt:   true,
			Grund:  string(domain.GrundBestandZu),
			Fremde: []FremdzustandDTO{},
		}
	}

	zustand := a.gesundheit.Zustand(ctx)

	antwort := GesundheitAntwort{
		Lebt:   zustand.Lebt,
		Bereit: zustand.Bereit(),
		Grund:  string(zustand.Grund),
		Fremde: make([]FremdzustandDTO, 0, len(zustand.Fremde)),
	}

	for _, fremd := range zustand.Fremde {
		antwort.Fremde = append(antwort.Fremde, FremdzustandDTO{
			Dienst:     string(fremd.Dienst),
			Erreichbar: fremd.Erreichbar,
			Grund:      fremd.Grund,
		})
	}

	return antwort
}

// Gesundheit ist die Fensterbindung.
func (b *Bindungen) Gesundheit(ctx context.Context) GesundheitAntwort {
	return b.app(ctx).Gesundheit(b.ruf(ctx))
}
