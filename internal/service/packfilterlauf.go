package service

// **Die Packfilter über alle Karten, verteilt** —
// #2466.
//
// Der erste Aufruf von „Zufallspack“ brauchte mit *Blue* (3.693 Karten)
// 80 s an der Kommandozeile; nach dem Satz (siehe [ohneSatz]) blieben 44 s
// Auswertung der Kartenskripte, Karte für Karte auf **einem** Kern. Die
// Problemprüfung verteilt dieselbe Arbeit seit PERF02 auf Arbeiter; hier
// geschieht es ebenso.
//
// # Was gleich bleibt
//
// **Die Reihenfolge.** Jede Karte bekommt ihr Urteil an ihrer Stelle, und
// zusammengeführt wird in der Reihenfolge des Sets — ein Pack zieht aus der
// Liste, und eine andere Reihenfolge änderte, was derselbe Samen zieht. **Und
// der Grund eines Filters, der scheitert**: der der letzten Karte, an der er
// scheiterte, wie in der Schleife davor.

import (
	"context"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// MitArbeitern setzt, wie viele Karten zugleich ausgewertet werden — **die
// Wurzel weiß es**, wie bei der Problemprüfung. Weniger als einer ist einer.
func (p *Packbereich) MitArbeitern(n int) *Packbereich {
	p.arbeiter = max(n, 1)

	return p
}

// filterurteil ist, was die Filter über eine Karte sagen — je Art in der
// Reihenfolge der Filter.
type filterurteil struct {
	trifft []bool
	grund  []string
}

// filterlauf wertet jeden Filter an jeder Karte aus.
//
// **Ein Bauer je Arbeiter**: er merkt sich je Karte, was ein Schlüsselwort
// bewirkt hat, und zwei Arbeiter an einem räumten einander den Vermerk.
func (p *Packbereich) filterlauf(
	ctx context.Context, karten []*domain.CardRow, mitFilter []msepaket.Pack,
	zerteilt map[string]*msescript.Knoten, vorlage Vorlage,
) ([]filterurteil, error) {
	urteile := make([]filterurteil, len(karten))
	arbeit := make(chan int)

	var gruppe sync.WaitGroup

	for range max(p.arbeiter, 1) {
		gruppe.Go(func() {
			bauer := p.bauerJeStil(ctx, vorlage)

			for stelle := range arbeit {
				if ctx.Err() == nil {
					urteile[stelle] = urteilen(ctx, bauer, karten[stelle], mitFilter, zerteilt)
				}
			}
		})
	}

	// **Die Arbeiter leeren den Kanal bis zum Ende** — derselbe Grund wie in
	// der Problemprüfung: ein Senden kann nicht hängen bleiben.
	for stelle := range karten {
		if ctx.Err() != nil {
			break
		}

		arbeit <- stelle
	}

	close(arbeit)
	gruppe.Wait()

	return urteile, ctx.Err()
}

// urteilen wertet die Filter an einer Karte aus — **ohne Satz** (#2466): ein
// Filter liest Felder, keine Maße.
func urteilen(
	ctx context.Context, bauer *stilbauer, karte *domain.CardRow,
	mitFilter []msepaket.Pack, zerteilt map[string]*msescript.Knoten,
) filterurteil {
	urteil := filterurteil{
		trifft: make([]bool, len(mitFilter)),
		grund:  make([]string, len(mitFilter)),
	}

	umgebung, _, _ := bauer.fuer(ohneSatz(ctx), karte)

	for stelle, art := range mitFilter {
		knoten, brauchbar := zerteilt[art.Name]
		if !brauchbar {
			continue
		}

		wert, fehler := msescript.Auswerten(ctx, knoten, umgebung)
		if fehler != nil {
			urteil.grund[stelle] = fehler.Error()

			continue
		}

		trifft, gut := wert.AlsWahrheit()
		urteil.trifft[stelle] = gut && trifft
	}

	return urteil
}
