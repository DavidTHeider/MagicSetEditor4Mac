package service

// **Der Ausschnitt einer Sammlung** — Werkstück
// #1500.
//
// # Warum das hier steht und nicht an der Naht
//
// Es sind **zwei** Schritte über **zwei** Abfragen: ordnen, dann holen. Das
// ist eine fachliche Regel über mehr als eine Entität (§2) — an der Naht
// stünde sie in `internal/app`, und die soll nicht entscheiden, was eine Sicht
// bedeutet.
//
// # Und warum eine eigene Ablage
//
// **Weil eine Schnittstelle dorthin gehört, wo sie gebraucht wird** (§2).
// `Sammlungsablage` trägt sechs Methoden und wird von sechs Anwendungsfällen
// erfüllt; sie um zwei zu erweitern, die nur dieser eine braucht, machte jedes
// Doppel im Projekt um zwei Methoden länger.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fensterablage ist, was ein Fenster über eine Sammlung braucht.
type Fensterablage interface {
	// Fenster liest die Ordnung und die Zeilen des Ausschnitts in **einem**
	// Stand; `waehle` entscheidet dazwischen, welche Kennungen gebraucht
	// werden — und in welcher Reihenfolge.
	//
	// **Ein Ruf und nicht zwei** (Befund 13 des Zweitblicks über Phase 5.2):
	// zwei Abfragen ohne gemeinsames Geschäft beschreiben womöglich zwei
	// Stände, und die Antwort wäre dann eine Sammlung, die es nie gab.
	Fenster(ctx context.Context, waehle func(ordnung []*domain.Exemplar) []int64) (
		[]*domain.Exemplar, error)
}

// Sammlungsfenster ist der Anwendungsfall „einen Ausschnitt meiner Sammlung
// sehen".
type Sammlungsfenster struct {
	ablage Fensterablage
}

// NeuesSammlungsfenster baut es über einer Ablage.
func NeuesSammlungsfenster(ablage Fensterablage) *Sammlungsfenster {
	return &Sammlungsfenster{ablage: ablage}
}

// Sicht gibt den Ausschnitt [ab, ab+anzahl) einer Sicht — samt beider Zahlen.
//
// **`gesamt` zählt die Sicht, `insgesamt` die Sammlung.** Ohne die zweite
// stünde in der Fußzeile „111 von 111", wo jemand zehntausend Karten besitzt,
// und die Suche sähe aus, als hätte sie nichts weggelassen.
//
// **Eine Anzahl von null heißt „alles"** — die Gruppenansicht fasst über die
// ganze Liste zusammen, und ein Ausschnitt gäbe Gruppen, die es nicht gibt.
func (f *Sammlungsfenster) Sicht(
	ctx context.Context, sicht domain.Sammlungssicht, ab, anzahl int,
) (fenster []*domain.Exemplar, zahlen domain.Fensterzahlen, fehler error) {
	if pruefehler := sicht.Pruefe(); pruefehler != nil {
		return nil, domain.Fensterzahlen{}, pruefehler
	}

	volle, fehler := f.ablage.Fenster(ctx, func(alle []*domain.Exemplar) []int64 {
		zahlen.Insgesamt = len(alle)

		// **Im selben Lesegeschäft wie der Ausschnitt** (U15, #2160): eine
		// Summe aus einem zweiten Ruf gehörte zu einem anderen Stand.
		for _, eines := range alle {
			zahlen.Stueck += eines.Menge
		}

		ausschnitt, gesehen := domain.Ausschnitt(sicht.Gesehen(alle), ab, anzahl)
		zahlen.Gesamt = gesehen

		kennungen := make([]int64, len(ausschnitt))
		for stelle, eines := range ausschnitt {
			kennungen[stelle] = eines.ID
		}

		return kennungen
	})
	if fehler != nil {
		return nil, domain.Fensterzahlen{}, fehler
	}

	return volle, zahlen, nil
}
