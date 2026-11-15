package service

// **Die Post für Kartenbilder** — P9 und P9b
// (#2377),
// Vermerk 0198. Wer ein Bild nicht selbst zeichnen kann, fragt den, dem das
// Deck gehört — der Gast den Wirt, der Wirt einen Gast. **Beide warten auf
// demselben Weg**: Wunsch hinaus, Antwort über die Leitung zurück, zugestellt
// dem, der genau sie erfragt hat.
//
// # Warum eine späte Antwort kein Fehler ist
//
// **Weil sie der Leitung nichts tut.** Wer nicht mehr wartet, hat die Frist
// verstreichen lassen; eine Antwort danach legte sonst den Tisch nieder.

import (
	"context"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bildpost stellt Bildantworten denen zu, die auf sie warten.
//
// **Eine je Sitzung und nicht je Gast**: das Fenster schickt über einen Gast,
// den es je Ruf baut, empfangen wird in der Schleife über einen anderen. Die
// Post teilen beide ([Partiegast.MitBildpost]); der Nullwert ist bereit.
type Bildpost struct {
	schloss sync.Mutex
	wartend map[domain.Bildwunsch][]chan domain.Bildantwort
}

// erwarte trägt einen Wartenden ein — **bevor** der Wunsch hinausgeht, damit
// keine schnelle Antwort ins Leere fällt.
func (p *Bildpost) erwarte(wunsch domain.Bildwunsch) chan domain.Bildantwort {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	if p.wartend == nil {
		p.wartend = map[domain.Bildwunsch][]chan domain.Bildantwort{}
	}

	kanal := make(chan domain.Bildantwort, 1)
	p.wartend[wunsch] = append(p.wartend[wunsch], kanal)

	return kanal
}

// vergiss trägt einen Wartenden aus, der nicht mehr wartet.
func (p *Bildpost) vergiss(wunsch domain.Bildwunsch, kanal chan domain.Bildantwort) {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	uebrig := p.wartend[wunsch][:0]
	for _, eines := range p.wartend[wunsch] {
		if eines != kanal {
			uebrig = append(uebrig, eines)
		}
	}

	if len(uebrig) == 0 {
		delete(p.wartend, wunsch)

		return
	}

	p.wartend[wunsch] = uebrig
}

// liefere stellt eine Antwort jedem zu, der auf ihren Wunsch wartet.
func (p *Bildpost) liefere(antwort domain.Bildantwort) {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	for _, kanal := range p.wartend[antwort.Wunsch] {
		kanal <- antwort
	}

	delete(p.wartend, antwort.Wunsch)
}

// frage schickt einen Wunsch hinaus und wartet auf die Antwort zu ihm.
//
// **Eingetragen wird vor dem Hinausgeben**, damit keine schnelle Antwort ins
// Leere fällt; `hinaus` muss den Wunsch **sofort** hinausgeben — die Sitzung
// legt nur in die Schlange, und ein Wunsch, der dort wartet, während sein
// Absender auf die Antwort wartet, käme nie an. `wer` nennt im Fehler, wer
// nicht geantwortet hat.
func (p *Bildpost) frage(
	ctx context.Context, wunsch domain.Bildwunsch,
	hinaus func(context.Context, []byte) error, wer string,
) (domain.Bildantwort, error) {
	kanal := p.erwarte(wunsch)

	inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // drei Zeichenketten klagen nicht
		domain.KontrolleBildwunsch, wunsch)

	if fehler := hinaus(ctx, inhalt); fehler != nil {
		p.vergiss(wunsch, kanal)

		return domain.Bildantwort{}, fehler
	}

	select {
	case antwort := <-kanal:
		return antwort, nil
	case <-ctx.Done():
		p.vergiss(wunsch, kanal)

		return domain.Bildantwort{}, fmt.Errorf(
			"%w: %s hat in der Frist kein Bild geschickt",
			domain.ErrConflict, wer)
	}
}
