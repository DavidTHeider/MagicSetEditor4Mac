package app

// **Am Tisch reden und zeigen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Zwei Bindungen für beide Seiten des Tisches
//
// **Wirt und Gast rufen dasselbe.** An welchem Ende dieses Programm sitzt,
// weiß es aus seinem eigenen Stand; eine Oberfläche, die es wissen müsste,
// entschiede bei jedem Wort neu, welche Bindung sie nimmt — und irgendwann
// entschiede sie falsch (§16).
//
// # Warum das eigene Wort nicht zurückkommt
//
// **Weil es schon dasteht.** Der Rundruf geht an alle **außer** dem Sprecher;
// sein Fenster zeigt den Satz beim Absenden. Ein Echo käme als zweiter Satz an
// und sähe aus wie ein doppeltes Wort.
//
// # Und warum hier kein Vorgangsriegel steht
//
// **Weil es keinen Datensatz gibt** (§7, Regel 4). Zwei Sätze zugleich sind
// kein Konflikt, sondern ein Gespräch; ein Zeiger, der auf den vorigen wartet,
// wäre ein Zeiger von gestern. Was hier trotzdem gilt, ist die Reihenfolge auf
// der Leitung — und die hält [service.Partiesitzung] mit ihrem Schloss.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PartienebenbeiAntwort sagt, ob es hinausging.
//
// **Und sonst nichts.** Was aus einem Satz wird, ist, dass ihn jemand liest;
// eine Antwort mit Inhalt behauptete mehr, als diese Naht weiß.
type PartienebenbeiAntwort struct {
	Fehler *Fehler `json:"fehler"`
}

// PartieSagen schickt einen Satz an den Tisch.
func (a *App) PartieSagen(
	ctx context.Context, text string,
) PartienebenbeiAntwort {
	ctx = a.imProfil(ctx)

	// **`externSenden`** — der Satz verlässt diesen Rechner. Dieselbe Klasse
	// wie `PartieWuenschen` (N7, #1568).
	if fehler := a.darf(ctx, domain.WirkungExternSenden); fehler != nil {
		return PartienebenbeiAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.amTisch(ctx,
		func(sitz string) error {
			return a.partie.runde.Sage(
				ctx, domain.Chat{Von: sitz, Text: text})
		},
		func(gast *service.Partiegast) error {
			return gast.Sage(
				ctx, domain.Chat{Von: gastplatzhalter, Text: text})
		})
}

// PartieZeigen schickt einen Zeiger, einen Griff, ein Schweben oder einen Ping.
//
// **`objekt` darf leer sein** — ein Zeiger auf freier Fläche zeigt auf keine
// Karte, und das ist kein Fehler.
//
// **Die Stelle ist eine Weltkoordinate in Millimetern**, kein Bildpunkt: zwei
// Fenster sind verschieden groß, und beide Enden rechnen die Tischwelt aus
// derselben Sicht ([domain.Praesenz]).
func (a *App) PartieZeigen(
	ctx context.Context, art, objekt string, x, y float64,
) PartienebenbeiAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungExternSenden); fehler != nil {
		return PartienebenbeiAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	gezeigt := func(sitz string) domain.Praesenz {
		return domain.Praesenz{
			Von: sitz, Art: domain.Praesenzart(art),
			Objekt: objekt, X: x, Y: y,
		}
	}

	return a.amTisch(ctx,
		func(sitz string) error {
			// **Den Stand vergibt die Runde**, denn er muss über den Abriss
			// einer Gastsitzung hinweg steigen.
			return a.partie.runde.Zeige(ctx, gezeigt(sitz))
		},
		func(gast *service.Partiegast) error {
			eigene := gezeigt(gastplatzhalter)
			eigene.Stand = a.partie.beitritt.zeigerstand.Add(1)

			return gast.Zeige(ctx, eigene)
		})
}

// amTisch führt aus, was für diese Seite des Tisches gilt.
//
// **Genau eine der beiden trifft zu.** Wer freigegeben hat, ist Wirt; wer
// beigetreten ist, ist Gast. Beides zugleich gibt es nicht — und keines von
// beidem ist kein Netzfehler, sondern ein Konflikt mit dem Stand (§8).
func (a *App) amTisch(
	ctx context.Context,
	alsWirt func(sitz string) error,
	alsGast func(gast *service.Partiegast) error,
) PartienebenbeiAntwort {
	if offen, wirt := a.offeneFreigabe(); offen != nil {
		return nebenbeiAntwort(ctx, alsWirt(wirt))
	}

	gast, griff, fehler := a.gastAmTisch(ctx)
	if fehler != nil {
		return nebenbeiAntwort(ctx, fehler)
	}

	if fehler := alsGast(gast); fehler != nil {
		return nebenbeiAntwort(ctx, fehler)
	}

	// **Gesagt ist noch nicht gehört.** [service.Sitzungsgriff.Sende] legt in
	// die Warteschlange; hinaus geht es erst hier. Genau dieser Schritt hat
	// in diesem Werkstück schon einmal gefehlt, und der Tisch blieb still.
	return nebenbeiAntwort(ctx, griff.Leere(ctx))
}

// nebenbeiAntwort übersetzt einen Fehler — oder gibt die leere Antwort.
func nebenbeiAntwort(ctx context.Context, fehler error) PartienebenbeiAntwort {
	if fehler == nil {
		return PartienebenbeiAntwort{}
	}

	return PartienebenbeiAntwort{Fehler: Uebersetze(ctx, fehler)}
}
