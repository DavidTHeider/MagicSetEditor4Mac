package domain_test

// **Wie viele benannte Wege es gibt, und was die zwei aus M10 anlegen** —
// Werkstücke M8 und M10 von Phase 5.5
// (#1610,
// #1612).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, *was* eine Wirkung tut** — je Art ein Fall, und dazu die
// Gegenproben aus CR 101.2. Hier steht, **wie viele es sind** und was die
// zwei aus `M10` von den neun aus `M8` unterscheidet: sie legen einen
// **Dauereffekt** an statt einer einmaligen Änderung (CR 611.1 gegen
// CR 610.1). `spielwirkung_test.go` stand mit beidem bei 412 Zeilen (§2).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Und die Liste der Arten ist die Zusage** — ein Wächter zählt sie.
//
// **Neun seit `M8`, elf seit `M10`, dreizehn seit `M11`**: `werte` und
// `faehigkeit` legen einen **Dauereffekt** an statt einer einmaligen
// Änderung (CR 611.1 gegen CR 610.1), `stattdessen` und `verhindern` ändern,
// **was geschieht** (CR 614.1, CR 615.1).
func TestEsGibtDreizehnBenannteWirkungen(t *testing.T) {
	t.Parallel()

	if anzahl := len(domain.Wirkungsarten()); anzahl != 13 {
		t.Errorf("%d Arten: %v", anzahl, domain.Wirkungsarten())
	}
}

// **Die zwei aus M11 legen einen Ersetzungseffekt an** — CR 614.1.
func TestZweiWirkungenLegenEineErsetzungAn(t *testing.T) {
	t.Parallel()

	partie := eineWirkungspartie(t)

	for _, wirkung := range []domain.Spielwirkung{
		{Art: domain.WirkungStattdessen, Objekt: "sitz-1/baer/0"},
		{Art: domain.WirkungVerhindern, Objekt: "sitz-1/baer/0", Menge: 2},
	} {
		danach, ging := partie.Anwenden(wirkung)
		if !ging {
			t.Fatalf("%+v ging nicht", wirkung)
		}

		partie = danach
	}

	if anzahl := len(partie.Ersetzungen); anzahl != 2 {
		t.Fatalf("%d Ersetzungen: %+v", anzahl, partie.Ersetzungen)
	}

	// **Beide gelten bis zum Zugende** — dieselbe Frist wie beim
	// Dauereffekt, und aus demselben Grund.
	for _, effekt := range partie.Ersetzungen {
		if effekt.Dauer != domain.DauerZugende {
			t.Errorf("%+v gilt unbestimmt", effekt)
		}
	}

	// **Und `verhindern` trägt seine Zahl** — CR 615.10.
	if partie.Ersetzungen[1].Menge != 2 {
		t.Errorf("%+v", partie.Ersetzungen[1])
	}
}

// **Die zwei aus M10 legen einen Dauereffekt an** — CR 611.1.
func TestZweiWirkungenLegenEinenDauereffektAn(t *testing.T) {
	t.Parallel()

	partie := eineWirkungspartie(t)

	danach, ging := partie.Anwenden(domain.Spielwirkung{
		Art: domain.WirkungWerte, Objekt: "sitz-1/baer/0", Menge: 2,
	})
	if !ging {
		t.Fatal("der Effekt ging nicht")
	}

	staerke, _ := danach.Werte(objektVon(t, danach, "sitz-1/baer/0"))
	if staerke.Zahl != 4 {
		t.Errorf("Stärke %d", staerke.Zahl)
	}

	danach, ging = danach.Anwenden(domain.Spielwirkung{
		Art: domain.WirkungFaehigkeit, Objekt: "sitz-1/baer/0",
		Marke: domain.SchluesselFlug,
	})
	if !ging {
		t.Fatal("die Fähigkeit ging nicht")
	}

	if !danach.HatFaehigkeit("sitz-1/baer/0", domain.SchluesselFlug) {
		t.Error("der Bär fliegt nicht")
	}

	// **Beide gelten bis zum Zugende** — CR 611.2b in der einzigen Gestalt,
	// die ein Skript stellen kann.
	for _, effekt := range danach.Dauereffekte {
		if effekt.Dauer != domain.DauerZugende {
			t.Errorf("%+v gilt unbestimmt", effekt)
		}
	}
}

// **Und ein Schlüsselwort, das dieses Programm nicht führt, geht nicht** —
// CR 101.2: eine Fähigkeit, die niemand prüft, wäre eine Zusage ohne
// Wirkung (§10).
func TestEinUnbekanntesSchluesselwortWirdNichtVerliehen(t *testing.T) {
	t.Parallel()

	faelle := []domain.Spielwirkung{
		{
			Art: domain.WirkungFaehigkeit, Objekt: "sitz-1/baer/0",
			Marke: "ward",
		},
		{Art: domain.WirkungWerte, Objekt: "weg", Menge: 1},
	}

	for _, wirkung := range faelle {
		if _, ging := eineWirkungspartie(t).Anwenden(wirkung); ging {
			t.Errorf("%+v ging", wirkung)
		}
	}
}

// **Ein Zauber führt beim Verrechnen aus, was an ihm steht** — CR 608.2c.
//
// # Warum das hier steht und nicht nur im Service
//
// **Weil es die Zusage der Domäne ist.** Dass ein Skript Wirkungen
// aufschreibt, misst `internal/service`; dass ein Objekt auf dem Stapel sie
// beim Verrechnen **ausführt**, entscheidet [Partie.Verrechnen] — und ohne
// diesen Lauf stünde die Schleife darin unbetreten da (§10).
func TestEinZauberFuehrtBeimVerrechnenAus(t *testing.T) {
	t.Parallel()

	partie := eineWirkungspartie(t)

	partie.Objekte = append(append([]domain.Spielobjekt{}, partie.Objekte...),
		domain.Spielobjekt{
			Kennung: "sitz-1/blitz/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneStapel, Karte: "baer",
			Wirkungen: []domain.Spielwirkung{
				{Art: domain.WirkungLeben, Wer: "sitz-2", Menge: -3},
				{Art: domain.WirkungZiehen, Wer: "sitz-1", Menge: 1},
			},
		})

	danach := partie.Verrechnen()

	if leben := lebenVon(t, danach, "sitz-2"); leben != 17 {
		t.Errorf("sitz-2 hat %d Leben", leben)
	}

	if hand := len(handVon(danach, "sitz-1")); hand != 1 {
		t.Errorf("%d Karten auf der Hand", hand)
	}
}
