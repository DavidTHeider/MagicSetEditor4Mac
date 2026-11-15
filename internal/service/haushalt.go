package service

import (
	"context"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Haushalt vergibt die Plätze je Ressourcenklasse — Werkstück O5 von Phase 5.2
// (#1429).
//
// # Service-nah, damit keine Fläche daran vorbeikommt
//
// Bis O5 gab es Grenzen nur an REST — eine Drossel je Anfrage. Wer über MCP
// oder die Kommandozeile zehn Rasterungen zugleich anstieß, tat es. Der
// Haushalt steht hier, wo die Arbeit beginnt: Fenster, Kommandozeile, MCP und
// REST belegen denselben Platz, und wer ihn nicht bekommt, wartet — oder
// gibt auf, wenn sein Kontext endet.
//
// # Ein Kanal je Klasse
//
// Ein gepufferter Kanal ist der Semaphor der Sprache: `Belegen` schickt, bis
// der Puffer voll ist, und wartet sonst mit dem Kontext; `freigeben` nimmt
// wieder heraus. **Freigeben ist idempotent** — ein zweiter Ruf tut nichts,
// sonst schüfe ein `defer` neben einem ausdrücklichen Ruf einen Platz aus
// dem Nichts.
type Haushalt struct {
	grenzen domain.Haushaltsgrenzen
	plaetze map[domain.Ressourcenklasse]chan struct{}
}

// NeuerHaushalt baut die Plätze aus den Grenzen — eine Klasse ohne Grenze hat
// keinen Kanal und ist frei.
func NeuerHaushalt(grenzen domain.Haushaltsgrenzen) *Haushalt {
	plaetze := make(map[domain.Ressourcenklasse]chan struct{}, len(grenzen.Parallel))

	for klasse, anzahl := range grenzen.Parallel {
		if anzahl > 0 {
			plaetze[klasse] = make(chan struct{}, anzahl)
		}
	}

	return &Haushalt{grenzen: grenzen, plaetze: plaetze}
}

// Grenzen sagt, was gilt — für die Auskunft an den Flächen.
func (h *Haushalt) Grenzen() domain.Haushaltsgrenzen {
	return h.grenzen
}

// Versuchen nimmt einen Platz, wenn einer frei ist — ohne zu warten.
func (h *Haushalt) Versuchen(klasse domain.Ressourcenklasse) (func(), bool) {
	kanal, begrenzt := h.plaetze[klasse]
	if !begrenzt {
		return func() {}, true
	}

	select {
	case kanal <- struct{}{}:
		return freigeber(kanal), true
	default:
		return nil, false
	}
}

// Belegen nimmt einen Platz oder wartet, bis einer frei wird — oder bis der
// Kontext endet; dann ist die Antwort sein Fehler, kein eigener.
func (h *Haushalt) Belegen(ctx context.Context, klasse domain.Ressourcenklasse) (func(), error) {
	kanal, begrenzt := h.plaetze[klasse]
	if !begrenzt {
		return func() {}, nil
	}

	select {
	case kanal <- struct{}{}:
		return freigeber(kanal), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// freigeber gibt den Platz genau einmal zurück.
func freigeber(kanal chan struct{}) func() {
	var einmal sync.Once

	return func() {
		einmal.Do(func() { <-kanal })
	}
}
