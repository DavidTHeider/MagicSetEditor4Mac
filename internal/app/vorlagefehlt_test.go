package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stillebildquelle ist eine Bildquelle, die nie gefragt wird: der Weg
// scheitert vorher an der fehlenden Vorlage. Sie steht da, damit er überhaupt
// bis dorthin kommt.
type stillebildquelle struct{}

func (stillebildquelle) Holen(context.Context, string, string) (
	*service.Bilduebernahme, error,
) {
	return nil, errors.New("gefragt wurde gar nicht")
}

// behelfsplanbauer gibt, was der echte Planbauer ohne Vorlage gibt: einen
// Ersatzplan **ohne Feldflächen** und einen Befund, der sagt, warum.
type behelfsplanbauer struct {
	ohneWortlisten
	ohneWahlbilder
}

func (behelfsplanbauer) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{Breite: 375, Hoehe: 523},
		[]domain.Befund{{Meldung: "das Spielpaket magic ist nicht eingebunden"}},
		nil
}

func (behelfsplanbauer) Kartenpalette(context.Context, *domain.CardRow, string) (
	[]service.Paletteneintrag, bool, []domain.Befund, error,
) {
	return nil, false, nil, nil
}

// **Ein fehlendes Vorlagenpaket sagt, dass es fehlt** —
// #1055.
//
// # Der Befund
//
// Wer ein Scryfall-Bild in eine Karte setzen will, deren Spielpaket nicht
// installiert ist, bekam an **jeder** Fläche:
//
//	das Bild holen: Validierung fehlgeschlagen:
//	Es ist ein unerwarteter Fehler aufgetreten. (unbekannt)
//
// **Gleich an allen drei Flächen, und überall unbrauchbar.** Der Fall ist der
// Normalfall für ein frisch eingerichtetes System: ein Set von Scryfall
// einlesen geht ohne Vorlage, ein Bild einsetzen nicht. Wer die Meldung las,
// wusste nicht, dass ihm ein Paket fehlt — und schon gar nicht, welches.
//
// # Warum der Schlüssel und nicht nur der Satz
//
// Go kennt die Sprache der Sitzung nicht (§9). Ein deutscher Satz aus Go
// stünde in einer englischen Oberfläche auf Deutsch; ein **Schlüssel** daneben
// lässt sich übersetzen. Ohne ihn fällt jede Fläche auf „unbekannt" zurück,
// und genau das war der Befund.
func TestEineFehlendeVorlageSagtEsBeimBild(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{},
		karteprobe{karte: &domain.CardRow{ID: 1, SetID: 7}},
		&behelfsplanbauer{}, render.NeuerRasterer(ohneBilder{}))
	anwendung.MitScryfallbildern(stillebildquelle{})

	antwort := anwendung.ScryfallBild(
		probekontext(t), "1", "image", "Lightning Bolt", "normal")
	if antwort.Fehler == nil {
		t.Fatal("ohne Vorlage kam kein Fehler")
	}

	if antwort.Fehler.Schluessel == domain.SchluesselUnbekannt {
		t.Errorf("der Schlüssel ist %q — damit fällt jede Fläche auf den "+
			"Rückfallsatz zurück", antwort.Fehler.Schluessel)
	}

	if strings.Contains(antwort.Fehler.Meldung, "unerwarteter Fehler") {
		t.Errorf("die Meldung sagt nichts: %q", antwort.Fehler.Meldung)
	}

	// **Und sie nennt die Vorlage**, sonst weiß niemand, was zu tun ist.
	if !strings.Contains(strings.ToLower(antwort.Fehler.Meldung), "vorlage") {
		t.Errorf("die Meldung nennt die Vorlage nicht: %q", antwort.Fehler.Meldung)
	}
}
