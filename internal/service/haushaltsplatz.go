package service

import (
	"context"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Platz, den eine schwere Arbeit belegt** — Werkstück O5 von Phase 5.2
// (#1429).
//
// # Ein Haushalt je Prozess, ersetzbar je Dienst
//
// Die Grenzen gelten der **Maschine**: zwei Rasterungen zugleich, gleich ob
// das Fenster, die Kommandozeile oder MCP sie anstößt. Deshalb gibt es einen
// Haushalt je Prozess — [Maschinenhaushalt] — und jeder Dienst, der schwere
// Arbeit tut, belegt dort seinen Platz, bevor er beginnt.
//
// **Und trotzdem eingesetzt, nicht erzeugt** (§16): ein Dienst nimmt über
// `MitHaushalt` einen eigenen, und ein Test wartet nie auf einen fremden
// Platz. Ohne Einsatz gilt der der Maschine — das ist der Betrieb.

var (
	maschinenhaushalt       *Haushalt
	maschinenhaushaltEinmal sync.Once
)

// Maschinenhaushalt ist der Haushalt dieses Prozesses, mit den Vorgabegrenzen.
func Maschinenhaushalt() *Haushalt {
	maschinenhaushaltEinmal.Do(func() {
		maschinenhaushalt = NeuerHaushalt(domain.Vorgabegrenzen())
	})

	return maschinenhaushalt
}

// haushaltswahl ist, was ein Dienst einbettet, der Plätze braucht.
type haushaltswahl struct {
	haushalt *Haushalt
}

// MitHaushalt setzt den Haushalt — für Tests und für einen Prozess, der
// andere Grenzen will.
func (w *haushaltswahl) MitHaushalt(haushalt *Haushalt) {
	w.haushalt = haushalt
}

// platz belegt einen Platz der Klasse — oder gibt den Fehler des Kontexts.
func (w *haushaltswahl) platz(ctx context.Context, klasse domain.Ressourcenklasse) (func(), error) {
	haushalt := w.haushalt
	if haushalt == nil {
		haushalt = Maschinenhaushalt()
	}

	return haushalt.Belegen(ctx, klasse)
}
