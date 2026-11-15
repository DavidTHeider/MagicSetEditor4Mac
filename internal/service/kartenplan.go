package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Stilquelle gibt den Stil eines Sets.
//
// **Als Schnittstelle und nicht als [Stilspeicher]**, weil hier gebraucht wird,
// was sie leistet, und nicht, wie sie es tut — dieselbe Richtung wie bei den
// Repositorien (§2).
type Stilquelle interface {
	StilFuerSet(ctx context.Context, setID int64) (Vorlage, []domain.Befund, bool)

	// StilFuerKarte gibt den Stil **dieser Karte** — ihren eigenen, sonst den
	// des Sets.
	//
	//	const StyleSheet& Set::stylesheetFor(const CardP& card) {
	//	  if (card && card->stylesheet) return *card->stylesheet;
	//	  else                          return *stylesheet;
	//	}
	//
	// `src/data/set.cpp`, Zeilen 90–93, Fassung f9d9356 vom 2021-07-06 (§15,
	// #395).
	StilFuerKarte(ctx context.Context, karte *domain.CardRow) (Vorlage, []domain.Befund, bool)
}

// MitStilen setzt die Naht zum Stilpaket.
//
// **Ohne sie bleibt es beim Behelfsplan.** Ein Planbauer ohne Stilquelle ist
// kein Fehler: die Tests der Bildausfuhr und des Bogens brauchen keinen Stil,
// und wer ihn hier erzwänge, machte jede Vorrichtung schwerer, ohne dass eine
// Zusage dazukäme.
func (p *Planbauer) MitStilen(quelle Stilquelle) *Planbauer {
	p.stile = quelle

	return p
}

// Kartenplan ist **der** Weg von einer Karte zu ihrem Zeichenplan.
//
// # Warum es genau einen gibt
//
// Vorschau, Bildausfuhr und Druckbogen riefen bis
// #206 jeder
// für sich `Behelfsplan`. Damit war alles, was #120, #173 und #186 gebaut
// haben, in der laufenden Anwendung nicht zu sehen — und zwei Wege zu derselben
// Karte laufen auseinander, sobald einer etwas lernt (#34).
//
// # Der Behelfsplan bleibt, aber er wird benannt
//
// Fällt der Stil aus — kein Paket, kein Paketverzeichnis, ein Stil ohne
// Kartengröße —, entsteht der Behelfsplan **samt Befund**. Eine Vorschau, die
// stillschweigend behilft, sieht aus wie eine, die den Stil nicht kann; und
// eine leere Seite wäre die schlechteste der drei Antworten (§8).
func (p *Planbauer) Kartenplan(ctx context.Context, karte *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	if p.stile == nil {
		plan, err := p.Behelfsplan(ctx, karte)

		return plan, befundliste("für die Vorschau ist kein Stilverzeichnis eingerichtet"), err
	}

	vorlage, befunde, da := p.stile.StilFuerKarte(ctx, karte)
	if !da {
		plan, err := p.Behelfsplan(ctx, karte)

		return plan, befunde, err
	}

	plan, planbefunde, err := p.Stilplan(ctx, karte, vorlage)
	if err != nil {
		// **Ein Abbruch ist etwas anderes als ein untauglicher Stil.** Wer ihn
		// in einen Befund umdeutete, ließe die Vorschau einen Plan zeichnen,
		// den niemand mehr sehen will (§7).
		if ctx.Err() != nil {
			return nil, nil, err
		}

		behelf, behelffehler := p.Behelfsplan(ctx, karte)

		return behelf, append(befunde, befundliste("der Stil trägt nicht: %v", err)...),
			behelffehler
	}

	return plan, append(befunde, planbefunde...), nil
}

// Kartenumgebung gibt die Umgebung, in der eine Karte gerechnet wird.
//
// **Für die Skriptkonsole** (#1070): sie soll denselben Kontext sehen wie der
// Rasterer — `card`, `set`, `stylesheet` und die Funktionen des Spiels. Eine
// zweite Umgebung daneben liefe auseinander, und dann beantwortete die
// Konsole Fragen über ein Programm, das es nicht gibt.
//
// **Ohne Stilpaket bleibt die Bibliothek** samt Karte: ein Ausdruck wie
// `1 + 1` braucht keine Vorlage.
func (p *Planbauer) Kartenumgebung(ctx context.Context, karte *domain.CardRow) (
	*msescript.Umgebung, []domain.Befund,
) {
	if p.stile == nil {
		return msescript.Bibliothek().Darueber(),
			befundliste("für die Konsole ist kein Stilverzeichnis eingerichtet")
	}

	vorlage, befunde, da := p.stile.StilFuerKarte(ctx, karte)
	if !da {
		return msescript.Bibliothek().Darueber(), befunde
	}

	umgebung, _, weitere := umgebungAus(ctx, karte, vorlage, p.schriften, p.bilder)

	return umgebung, append(befunde, weitere...)
}
