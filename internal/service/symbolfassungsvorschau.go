package service

// **Ein Symbol der Bibliothek in den Fassungen eines Sets** — DOM-02.AK-02
// (#1697):
// die Größen- und Farbvorschau.
//
// # Über denselben Bildknoten wie auf der Karte
//
// **Jede Fassung wird ein `symbol_variation`-Knoten** — gebaut von
// [fassungsbild] und [alsZeichenbaum], denselben Wegen, die eine Karte
// nimmt. Die Farben werden damit **einmal** gedeutet, und der Rasterer wählt
// den Filter nach `fill type:` wie auf jeder Karte. Eine zweite Deutung hier
// liefe mit jener auseinander.
//
// **Gezeichnet wird eine Ebene je Plan**, so groß wie die Vorschau; die
// Größen entstehen beim Rastern über die Vergrößerung.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Symbolfassungsquelle nennt die Fassungen des Setsymbols eines Sets.
type Symbolfassungsquelle interface {
	Symbolfassungen(ctx context.Context, setID int64) ([]msepaket.Symbolfassung, error)
}

// Fassungsplan ist eine Fassung und der Plan, der das Symbol in ihr zeichnet.
type Fassungsplan struct {
	Name string
	Plan *render.Plan
}

// MitFassungen setzt die Quelle der Fassungen.
func (b *Symbolbibliothek) MitFassungen(quelle Symbolfassungsquelle) *Symbolbibliothek {
	b.fassungen = quelle

	return b
}

// MitPruefer setzt den Riegel vor jede Datei. Ohne ihn wird nicht gefragt —
// die Tests der Wege selbst brauchen keinen.
func (b *Symbolbibliothek) MitPruefer(pruefer Pruefer) *Symbolbibliothek {
	b.pruefer = pruefer

	return b
}

// Fassungsplaene gibt je Fassung des Setsymbols einen Plan — **jede einmal**,
// die erste gewinnt.
func (b *Symbolbibliothek) Fassungsplaene(
	ctx context.Context, id, setID int64, groesse float64,
) ([]Fassungsplan, error) {
	zeichnung, err := b.Zeichnung(ctx, id)
	if err != nil {
		return nil, err
	}

	plaene := []Fassungsplan{}

	if b.fassungen == nil {
		return plaene, nil
	}

	fassungen, err := b.fassungen.Symbolfassungen(ctx, setID)
	if err != nil {
		return nil, err
	}

	gesehen := map[string]bool{}

	for _, fassung := range fassungen {
		if fassung.Feld != setsymbolfeld || gesehen[fassung.Name] {
			continue
		}

		gesehen[fassung.Name] = true
		plaene = append(plaene, Fassungsplan{Name: fassung.Name, Plan: fassungsplan(zeichnung, fassung, groesse)})
	}

	return plaene, nil
}

// bibliothekszeichnung ist der Verweis, unter dem der Bildknoten die
// Zeichnung sucht — sie kommt aus der Bibliothek und nicht aus einem Paket.
const bibliothekszeichnung = "bibliothek"

// fassungsplan baut den Plan einer Fassung.
func fassungsplan(zeichnung msepaket.Symbolzeichnung, fassung msepaket.Symbolfassung, groesse float64) *render.Plan {
	baum, _ := fassungsbild(bibliothekszeichnung, fassung).AlsBild()

	return &render.Plan{
		Breite: groesse, Hoehe: groesse,
		Ebenen: []render.Ebene{{
			Art:     render.ArtSymbol,
			Rahmen:  render.Rechteck{Breite: groesse, Hoehe: groesse},
			Bild:    alsZeichenbaum(baum, "", festeZeichnung{zeichnung}),
			Deckung: 1,
		}},
	}
}

// festeZeichnung antwortet jedem Verweis mit derselben Zeichnung.
type festeZeichnung struct{ zeichnung msepaket.Symbolzeichnung }

func (f festeZeichnung) Symbolgestalt(string) msepaket.Symbolzeichnung { return f.zeichnung }
