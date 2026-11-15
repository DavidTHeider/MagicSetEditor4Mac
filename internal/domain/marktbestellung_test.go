package domain_test

// **Wann eine fremde Bestellung den eigenen Bestand bewegen darf** —
// Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Warum das eine Regel und keine Gewohnheit ist
//
// **Weil ein Irrtum hier doppelt abbucht.** CardTrader Zero legt zu jedem
// Verkauf **zwei** Bestellungen an — eine im Zustand `hub_pending` und
// später eine gebündelte im Zustand `paid`. Wer beide für Verkäufe hielte,
// nähme dasselbe Stück zweimal aus dem Bestand.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Tabelle, so wie sie beim Anbieter steht.**
//
// **Nachgelesen am 2026-09-19** unter „Order Lifecycle" und nicht erschlossen
// (§16, Regel 1): der Anbieter schreibt sie als Tabelle hin, und genau diese
// drei Zeilen hat sie.
func TestWannEineBestellungDenBestandBewegt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		zustand domain.Bestellzustand
		ueberCT0,
		bewegt bool
	}{
		"beim Hub liegend, über CardTrader Zero": {
			domain.BestellungHubOffen, true, true,
		},
		"bezahlt, über CardTrader Zero": {
			domain.BestellungBezahlt, true, false,
		},
		"bezahlt, ohne CardTrader Zero": {
			domain.BestellungBezahlt, false, true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if domain.BewegtDenBestand(fall.zustand, fall.ueberCT0) != fall.bewegt {
				t.Errorf("erwartet %v", fall.bewegt)
			}
		})
	}
}

// **Jeder andere Zustand bewegt nichts.**
//
// **Die Gegenprobe**: eine Regel, die nur die drei Zeilen kennte und sonst
// „ja" sagte, wäre bei `canceled` eine Abbuchung für eine Bestellung, die es
// nicht mehr gibt. Was nicht in der Tabelle steht, bewegt nichts.
func TestJederAndereBestellzustandBewegtNichts(t *testing.T) {
	t.Parallel()

	for _, zustand := range domain.Bestellzustaende() {
		if zustand == domain.BestellungBezahlt ||
			zustand == domain.BestellungHubOffen {
			continue
		}

		for _, ueberCT0 := range []bool{false, true} {
			if domain.BewegtDenBestand(zustand, ueberCT0) {
				t.Errorf("%q bewegt den Bestand (Zero: %v)", zustand, ueberCT0)
			}
		}
	}
}

// **Acht Zustände, und sie heißen so, wie der Anbieter sie nennt.**
//
// **Nicht übersetzt**: was von außen kommt, kommt in seinen Worten, und eine
// deutsche Entsprechung machte aus einem Vergleich eine Zuordnung, die
// auseinanderlaufen kann (§9).
func TestEsGibtAchtBestellzustaende(t *testing.T) {
	t.Parallel()

	zustaende := domain.Bestellzustaende()

	if len(zustaende) != 8 {
		t.Fatalf("%d Zustände, erwartet 8: %v", len(zustaende), zustaende)
	}

	gesehen := map[domain.Bestellzustand]bool{}

	for _, zustand := range zustaende {
		if gesehen[zustand] {
			t.Errorf("%q steht zweimal", zustand)
		}

		gesehen[zustand] = true
	}

	for _, erwartet := range []domain.Bestellzustand{
		"paid", "sent", "arrived", "done", "hub_pending",
		"request_for_cancel", "canceled", "lost",
	} {
		if !gesehen[erwartet] {
			t.Errorf("%q fehlt", erwartet)
		}
	}
}

// **Ein unbekannter Zustand ist ein Befund und kein stilles Ja.**
//
// **Fremde Daten sind der Normalfall** (§8, §16): der Anbieter darf einen
// neunten Zustand einführen, und dann soll dieses Programm es sagen — nicht
// raten.
func TestEinUnbekannterBestellzustandIstEinBefund(t *testing.T) {
	t.Parallel()

	if domain.IstBestellzustand("verschollen") {
		t.Error("ein erfundener Zustand gilt als bekannt")
	}

	// **Und die Gegenprobe**: ein Wächter, der zu allem nein sagte, wäre bei
	// dieser Zeile ebenso grün — und sagte nichts (§10).
	for _, zustand := range domain.Bestellzustaende() {
		if !domain.IstBestellzustand(zustand) {
			t.Errorf("%q gilt als unbekannt", zustand)
		}
	}

	if domain.BewegtDenBestand("verschollen", false) {
		t.Error("ein erfundener Zustand bewegt den Bestand")
	}
}
