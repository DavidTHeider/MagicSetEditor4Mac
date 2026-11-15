package service_test

// **Was die Leitung nicht geht** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass die drei Abschlussschritte an der Leitung enden — ohne Schreibe.**
// Die Leitung bewegt den Zustand und sonst nichts; `verbuchen`
// bewegt den **Bestand**, `bestaetigen` zeichnet den **Beleg** gegen. Ginge
// einer davon hier durch, stünde ein Vorgang auf `lokalVerbucht`, ohne dass
// eine Karte den Besitzer gewechselt hätte — und das sähe von außen aus wie
// ein verbuchter Handel.
//
// **Die Gegenprobe steht daneben**: derselbe Vorgang, derselbe Stand, ein
// Schritt der Matrix, der **kein** Abschluss ist — er geht durch. Ohne sie
// sagte das Grün oben nur, dass irgendetwas abgewiesen wird.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// imZustand legt einen Vorgang in einem Zustand ab.
func imZustand(
	ablage *handelsprobe, zustand domain.Handelszustand,
) domain.Handelsangebot {
	angebot := einHandelsangebot()
	angebot.Zustand = zustand
	ablage.staende[angebot.Vorgang] = angebot

	return angebot
}

// **Kein Abschlussschritt geht über die Leitung.**
func TestDieLeitungGehtKeinenAbschlussschritt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		schritt domain.Handelsschritt
		von     domain.Handelszustand
	}{
		{domain.SchrittVerbuchen, domain.HandelUebergabeOffen},
		{domain.SchrittAbgleichen, domain.HandelLokalVerbucht},
		{domain.SchrittBestaetigen, domain.HandelAbgleichOffen},
	} {
		t.Run(string(fall.schritt), func(t *testing.T) {
			t.Parallel()

			leitung, ablage := eineHandelsleitung()
			angebot := imZustand(ablage, fall.von)
			vorher := ablage.geschrieben

			_, fehler := leitung.Schreiten(context.Background(),
				angebot.Vorgang, fall.schritt, "anna")
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", fall.schritt, fehler)
			}

			// **Die Absage nennt den Weg, der es kann.** Eine Absage, die nur
			// „nein" sagt, schickt jemanden suchen.
			if !strings.Contains(fehler.Error(), "Verbuchung") {
				t.Errorf("die Absage nennt die Verbuchung nicht: %v", fehler)
			}

			if ablage.geschrieben != vorher {
				t.Errorf("die Absage hat geschrieben: %d statt %d",
					ablage.geschrieben, vorher)
			}
		})
	}
}

// **Und ein Schritt, der keiner ist, geht durch** — die Gegenprobe.
func TestEinNichtabschlussSchrittGehtUeberDieLeitung(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	angebot := imZustand(ablage, domain.HandelVorbereitet)

	// **Bert und nicht Anna**: Anna gibt zwei Stück weg und hat nichts
	// festgehalten — seit der Hüllenprobe `nat04handel` (#1672) bestätigt sie
	// die Übergabe so nicht mehr (`handelsschrittzusagen_test.go`).
	gegangen, fehler := leitung.Schreiten(context.Background(),
		angebot.Vorgang, domain.SchrittUebergeben, "bert")
	if fehler != nil {
		t.Fatalf("Schreiten: %v", fehler)
	}

	// **Eine Stimme von zweien bewegt den Zustand nicht** — sie wird
	// eingetragen, und das ist die Wirkung, die hier zählt.
	if len(gegangen.Zustimmungen) != 1 {
		t.Errorf("eingetragen wurden %v", gegangen.Zustimmungen)
	}

	if ablage.geschrieben != 1 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}
