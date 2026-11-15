package domain_test

// **Welche Nachricht zu welchem Stand gehört** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass aus jedem Zustand, der die Gegenseite angeht, genau eine Nachricht
// folgt — und aus den drei übrigen keine.** Bis zu diesem Issue waren die elf
// Nachrichtenarten eine Aufzählung ohne Verbraucher: `T5` prüfte sie beim
// **Empfangen**, gesendet hat sie niemand.
//
// # Warum drei Zustände keine Nachricht ergeben
//
//   - **Entwurf** hat die Gegenseite nie gesehen.
//   - **Abgelaufen** ist eine Beobachtung dieser Seite an einer Uhr, die
//     drüben anders geht.
//   - **Klärung** ist der Fall, den zwei Menschen lösen — eine Nachricht
//     darüber ersetzte das Gespräch nicht und sähe aus, als täte sie es.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJederZustandWaehltSeineNachricht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		zustand domain.Handelszustand
		art     domain.Handelsnachrichtenart
	}{
		{domain.HandelAngeboten, domain.NachrichtVorschlag},
		{domain.HandelVerhandlung, domain.NachrichtGegenvorschlag},
		{domain.HandelAngenommen, domain.NachrichtAnnahme},
		{domain.HandelVorbereitet, domain.NachrichtVorbereitung},
		{domain.HandelUebergabeOffen, domain.NachrichtUebergabe},
		{domain.HandelLokalVerbucht, domain.NachrichtCommitstand},
		{domain.HandelAbgleichOffen, domain.NachrichtCommitstand},
		{domain.HandelBeidseitigVerbucht, domain.NachrichtBeleg},
		{domain.HandelAbgelehnt, domain.NachrichtAblehnung},
		{domain.HandelAbgebrochen, domain.NachrichtAbbruch},
	} {
		t.Run(string(fall.zustand), func(t *testing.T) {
			t.Parallel()

			art, hinaus := domain.NachrichtenartFuer(fall.zustand)
			if !hinaus {
				t.Fatalf("aus %q geht nichts hinaus", fall.zustand)
			}

			if art != fall.art {
				t.Errorf("aus %q wurde %q statt %q", fall.zustand, art, fall.art)
			}
		})
	}
}

// **Und aus dreien geht nichts hinaus.**
func TestAusDreiZustaendenGehtKeineNachrichtHinaus(t *testing.T) {
	t.Parallel()

	for _, zustand := range []domain.Handelszustand{
		domain.HandelEntwurf, domain.HandelAbgelaufen, domain.HandelKlaerung,
	} {
		if art, hinaus := domain.NachrichtenartFuer(zustand); hinaus {
			t.Errorf("aus %q ginge %q hinaus", zustand, art)
		}
	}
}

// **Jeder der dreizehn Zustände ist entschieden** — die Gegenprobe gegen eine
// Tabelle, die still zu kurz wird.
//
// **Über [domain.Handelszustaende] und nicht über eine Liste daneben**: ein
// vierzehnter Zustand soll diesen Lauf rot machen, statt stumm durchzufallen.
func TestJederHandelszustandIstEntschieden(t *testing.T) {
	t.Parallel()

	stumm := map[domain.Handelszustand]bool{
		domain.HandelEntwurf: true, domain.HandelAbgelaufen: true,
		domain.HandelKlaerung: true,
	}

	for _, zustand := range domain.Handelszustaende() {
		art, hinaus := domain.NachrichtenartFuer(zustand)

		switch {
		case stumm[zustand] && hinaus:
			t.Errorf("%q sollte stumm sein, gibt aber %q", zustand, art)

		case !stumm[zustand] && !hinaus:
			t.Errorf("%q gibt keine Nachricht", zustand)

		case hinaus && !domain.IstHandelsnachrichtenart(art):
			t.Errorf("%q gibt die unbekannte Art %q", zustand, art)
		}
	}
}
