// **Was ein Profil vorgeben kann** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// # Warum das Angebot hier steht und nicht im Dienst
//
// **Weil es fünf verschiedene Quellen zusammenführt**: Formate und Decks aus
// dem Bestand, Spiel- und Stilpakete aus dem Paketkatalog, Lagerorte aus der
// Sammlung. Jede davon hat ihren eigenen Dienst, und ein sechster, der die
// fünf kennt, wäre einer, der nichts entscheidet.
//
// **`internal/app` darf das** — es ist die Schicht, die zusammenführt (§2). Was
// es **nicht** tut, ist urteilen: ob eine Wahl gilt, entscheidet
// `service.Vorgabenverwaltung` gegen genau diese Liste.

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Vorgabenquellen sind die fünf Nähte, aus denen ein Angebot entsteht.
//
// **Eine Naht je Frage und keine gemeinsame.** Wer sie zusammenlegte, hätte
// eine Schnittstelle, die fünf Dienste zugleich verlangt — und ein Test, der
// eine Art prüft, müsste vier andere belegen.
type Vorgabenquellen struct {
	// Formate gibt die Formate **aller** Spiele — ein leeres Spiel heißt: alle.
	Formate func(ctx context.Context, spiel string) ([]domain.Format, error)

	// Decks gibt die Decks aller Spiele.
	Decks func(ctx context.Context, spiel string) ([]domain.Deck, error)

	// Pakete gibt den Katalog dieses Rechners.
	Pakete func(ctx context.Context) ([]domain.Paket, error)

	// Lagerorte gibt die Orte, an denen etwas liegt.
	Lagerorte func(ctx context.Context) ([]service.Lagerortzeile, error)
}

// Vorgabenangebot beantwortet, was es für eine Art gibt.
type Vorgabenangebot struct{ quellen Vorgabenquellen }

// NeuesVorgabenangebot baut es.
func NeuesVorgabenangebot(quellen Vorgabenquellen) *Vorgabenangebot {
	return &Vorgabenangebot{quellen: quellen}
}

// Angebot gibt die Werte, die ein Profil bei dieser Art wählen kann.
func (a *Vorgabenangebot) Angebot(ctx context.Context, art domain.Vorgabeart) (
	[]service.Vorgabewahl, error,
) {
	switch art {
	case domain.VorgabeFormat:
		return a.formate(ctx)
	case domain.VorgabeSpiel:
		return a.pakete(ctx, "mse-game")
	case domain.VorgabeVorlage:
		return a.pakete(ctx, "mse-style")
	case domain.VorgabeDeck:
		return a.decks(ctx)
	case domain.VorgabeLagerort:
		return a.lagerorte(ctx)
	}

	// **Eine Art, die es nicht gibt, ist ein leeres Angebot und kein Fehler.**
	// Wer sie setzen will, wird von [domain.Profilvorgabe.Validate] abgewiesen
	// — mit einer Meldung, die den Namen nennt. Hier noch einmal zu urteilen
	// hieße, dieselbe Frage an zwei Stellen zu beantworten.
	return []service.Vorgabewahl{}, nil
}

// formate gibt jedes Format samt seinem Spiel.
//
// **Mit dem Spiel in der Beschriftung**, weil zwei Spiele denselben Formatnamen
// führen dürfen: „Standard" allein sagte dann nicht, welches gemeint ist.
func (a *Vorgabenangebot) formate(ctx context.Context) (
	[]service.Vorgabewahl, error,
) {
	gefunden, fehler := a.quellen.Formate(ctx, "")
	if fehler != nil {
		return nil, fehler
	}

	wahlen := make([]service.Vorgabewahl, 0, len(gefunden))
	for _, eines := range gefunden {
		wahlen = append(wahlen, service.Vorgabewahl{
			Wert:         zahltext(eines.ID),
			Beschriftung: fmt.Sprintf("%s (%s)", eines.Name, eines.Spiel),
		})
	}

	return wahlen, nil
}

// decks gibt jedes Deck samt seinem Spiel — dieselbe Überlegung.
func (a *Vorgabenangebot) decks(ctx context.Context) (
	[]service.Vorgabewahl, error,
) {
	gefunden, fehler := a.quellen.Decks(ctx, "")
	if fehler != nil {
		return nil, fehler
	}

	wahlen := make([]service.Vorgabewahl, 0, len(gefunden))
	for _, eines := range gefunden {
		wahlen = append(wahlen, service.Vorgabewahl{
			Wert:         zahltext(eines.ID),
			Beschriftung: fmt.Sprintf("%s (%s)", eines.Name, eines.Spiel),
		})
	}

	return wahlen, nil
}

// pakete gibt die Pakete einer Art.
//
// **Der Name ist der Wert und nicht die Kennung**: ein Paket hat im Bestand
// keine — es liegt auf der Platte, und sein Name ist, womit `SetNeu` es
// anspricht.
func (a *Vorgabenangebot) pakete(ctx context.Context, art string) (
	[]service.Vorgabewahl, error,
) {
	katalog, fehler := a.quellen.Pakete(ctx)
	if fehler != nil {
		return nil, fehler
	}

	wahlen := []service.Vorgabewahl{}

	for _, paket := range katalog {
		if paket.Art != art {
			continue
		}

		wahlen = append(wahlen, service.Vorgabewahl{
			Wert: paket.Name, Beschriftung: beschriftet(paket),
		})
	}

	return wahlen, nil
}

// beschriftet gibt den Kurznamen eines Pakets — oder seinen Namen.
//
// **Denselben Rückfall wie MSE**: `Packaged::validate` nimmt den Dateinamen
// ohne Endung, wenn `short name:` fehlt.
func beschriftet(paket domain.Paket) string {
	if kurz := strings.TrimSpace(paket.Kurzname); kurz != "" {
		return kurz
	}

	return paket.Name
}

// lagerorte gibt die Orte, an denen etwas liegt.
//
// **Ohne den leeren.** „Ohne Angabe" ist kein Ort, den man zuordnen kann; er
// steht in der Sammlungssicht als Sammelposten und nicht als Wahl.
func (a *Vorgabenangebot) lagerorte(ctx context.Context) (
	[]service.Vorgabewahl, error,
) {
	gefunden, fehler := a.quellen.Lagerorte(ctx)
	if fehler != nil {
		return nil, fehler
	}

	wahlen := []service.Vorgabewahl{}

	for _, zeile := range gefunden {
		if strings.TrimSpace(zeile.Lagerort) == "" {
			continue
		}

		wahlen = append(wahlen, service.Vorgabewahl{
			Wert: zeile.Lagerort, Beschriftung: zeile.Lagerort,
		})
	}

	return wahlen, nil
}
