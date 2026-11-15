package service_test

// **Eine ungeteilte Bewertung bleibt hier** — TRADE-03.AK-03
// (#1717).
//
// Sie ist eine Anzeige und kein Vertragsinhalt, solange sie nicht geteilt
// ist (`BewertungGeteilt`), und sie steht dann auch nicht im Abdruck. Mit der
// Nachricht ging sie trotzdem hinaus — was man für eine Karte zu zahlen
// bereit wäre, erfuhr das Gegenüber so nebenbei.

import (
	"context"
	"testing"
)

func TestEineUngeteilteBewertungGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	for _, geteilt := range []bool{false, true} {
		stand := derHeutigeStand()
		stand.BewertungGeteilt = geteilt
		stand.Positionen[0].Bewertung = 1995

		versand, ausgang, _ := einVersand(t, stand)

		if _, fehler := versand.Schicken(context.Background(), stand.Vorgang, dasGegenueber); fehler != nil {
			t.Fatalf("Schicken: %v", fehler)
		}

		gesendet := ausgang.zeilen[0].Nachricht

		erwartet := int64(0)
		if geteilt {
			erwartet = 1995
		}

		if bewertung := gesendet.Angebot.Positionen[0].Bewertung; bewertung != erwartet {
			t.Errorf("geteilt %v: hinaus ging %d", geteilt, bewertung)
		}

		if gesendet.Umschlag.Abdruck != gesendet.Angebot.Abdruck() {
			t.Errorf("geteilt %v: der Abdruck passt nicht mehr zum Angebot", geteilt)
		}
	}
}
