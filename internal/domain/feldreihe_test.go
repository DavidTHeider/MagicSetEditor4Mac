package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Feldreihen eines Spiels** — Werkstück 3 von Phase 3.6
// (#1167).

// probespiel baut ein Spiel aus Feldnamen.
func probespiel(namen ...string) domain.Game {
	felder := make([]domain.Field, 0, len(namen))
	for _, name := range namen {
		felder = append(felder, domain.Field{Name: name})
	}

	return domain.Game{Name: "probe", Fields: felder}
}

// **Ein Feld mit numeriertem Geschwister bildet eine Reihe.**
func TestEinNumeriertesGeschwisterBildetEineReihe(t *testing.T) {
	t.Parallel()

	reihen := probespiel("name", "name 2", "name 3", "rarity").Feldreihen()

	if len(reihen) != 1 {
		t.Fatalf("es sind %d Reihen: %+v", len(reihen), reihen)
	}

	if reihen[0].Grundfeld != "name" {
		t.Errorf("Grundfeld %q", reihen[0].Grundfeld)
	}

	if len(reihen[0].Stellen) != 2 {
		t.Fatalf("Stellen: %+v", reihen[0].Stellen)
	}

	if reihen[0].Stellen[0].Feld != "name 2" || reihen[0].Stellen[0].Nummer != 2 {
		t.Errorf("erste Stelle: %+v", reihen[0].Stellen[0])
	}
}

// **Eine Zahl ohne Grundfeld ist keine Reihe.**
//
// # Warum das die entscheidende Regel ist
//
// `magic.mse-game` führt `loyalty cost 1` bis `loyalty cost 16` — und **kein**
// `loyalty cost`. Das sind die Zeilen eines Planeswalkers, nicht die Seiten
// einer Karte. Wer allein an der Zahl träfe, machte aus einem Planeswalker
// eine sechzehnseitige Karte.
//
// Nachgezählt an `magic.mse-game` am 2026-08-27, über den Leser dieses
// Projekts: 249 Kartenfelder, **43 Reihen** — und **40 Felder mit Zahl ohne
// Grundfeld**.
func TestEineZahlOhneGrundfeldIstKeineReihe(t *testing.T) {
	t.Parallel()

	reihen := probespiel("loyalty cost 1", "loyalty cost 2", "name").Feldreihen()

	if len(reihen) != 0 {
		t.Errorf("es kamen %+v", reihen)
	}
}

// **Die Reihen stehen in der Reihenfolge des Spielpakets**, wie alles andere
// auch — eine alphabetische wäre eine zweite Ordnung neben der, die MSE
// benutzt.
func TestDieReihenStehenInDerOrdnungDesSpiels(t *testing.T) {
	t.Parallel()

	reihen := probespiel("type", "type 2", "name", "name 2").Feldreihen()

	if len(reihen) != 2 || reihen[0].Grundfeld != "type" {
		t.Errorf("Reihenfolge: %+v", reihen)
	}
}

// **Und die Stellen einer Reihe stehen aufsteigend**, auch wenn das Paket sie
// anders aufzählt.
func TestDieStellenStehenAufsteigend(t *testing.T) {
	t.Parallel()

	reihen := probespiel("name", "name 3", "name 2").Feldreihen()

	if len(reihen) != 1 {
		t.Fatalf("Reihen: %+v", reihen)
	}

	if reihen[0].Stellen[0].Nummer != 2 || reihen[0].Stellen[1].Nummer != 3 {
		t.Errorf("Stellen: %+v", reihen[0].Stellen)
	}
}

// **Eine Zahl mitten im Namen zählt nicht** — nur die am Ende, hinter einem
// Leerzeichen.
func TestNurDieZahlAmEndeZaehlt(t *testing.T) {
	t.Parallel()

	for name, spiel := range map[string]domain.Game{
		"ohne Leerzeichen": probespiel("name", "name2"),
		"mitten drin":      probespiel("name", "name 2 lang"),
		"führende Null":    probespiel("name", "name 02"),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if reihen := spiel.Feldreihen(); len(reihen) != 0 {
				t.Errorf("es kamen %+v", reihen)
			}
		})
	}
}

// **Die Stelle 1 gibt es nicht.**
//
// Das Grundfeld **ist** die erste Stelle; ein `name 1` daneben wäre eine
// zweite erste, und niemand könnte sagen, welche gilt. In `magic.mse-game`
// steht kein einziges solches Paar.
func TestDieStelleEinsGibtEsNicht(t *testing.T) {
	t.Parallel()

	reihen := probespiel("name", "name 1", "name 2").Feldreihen()

	if len(reihen) != 1 {
		t.Fatalf("Reihen: %+v", reihen)
	}

	for _, stelle := range reihen[0].Stellen {
		if stelle.Nummer == 1 {
			t.Errorf("die Stelle 1 steht in der Reihe: %+v", reihen[0].Stellen)
		}
	}
}
