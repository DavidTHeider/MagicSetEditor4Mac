package service_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Booster zieht, was seine Posten nennen** —
// #1072.
//
// Der Aufbau ist der des echten Packs in klein: eine Gruppe mit zwei Posten,
// darunter zwei Blätter mit Karten.
func TestEinBoosterZiehtWasSeinePostenNennen(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher(
		[]msepaket.Pack{
			{Name: "booster", Auswahl: "all", Posten: []msepaket.Packposten{
				{Name: "common", Anzahl: 10, Gewicht: 1},
				{Name: "rare", Anzahl: 1, Gewicht: 1},
			}},
			{Name: "common", Auswahl: "no replace"},
			{Name: "rare", Auswahl: "no replace"},
		},
		map[string][]*domain.CardRow{
			"common": packkarten("c", 12),
			"rare":   packkarten("r", 4),
		})

	gezogen, fehler := zieher.Ziehen("booster", 7)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	if len(gezogen) != 11 {
		t.Fatalf("%d Karten, erwartet elf", len(gezogen))
	}

	seltene := 0

	for _, karte := range gezogen {
		if strings.HasPrefix(namenVon(karte), "r") {
			seltene++
		}
	}

	if seltene != 1 {
		t.Errorf("%d seltene, erwartet eine", seltene)
	}
}

// **Derselbe Same gibt dieselbe Ziehung.**
//
// Ohne diese Zusage ist kein Test möglich, und das steht so in
// #889.
func TestDerselbeSameGibtDieselbeZiehung(t *testing.T) {
	t.Parallel()

	bauen := func() *service.Packzieher {
		return service.NeuerPackzieher(
			[]msepaket.Pack{{Name: "blatt", Auswahl: "no replace"}},
			map[string][]*domain.CardRow{"blatt": packkarten("k", 20)})
	}

	erste, fehler := bauen().Ziehen("blatt", 42)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	zweite, fehler := bauen().Ziehen("blatt", 42)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	if namenreihe(erste) != namenreihe(zweite) {
		t.Errorf("zwei Läufe mit demselben Samen: %q gegen %q",
			namenreihe(erste), namenreihe(zweite))
	}

	// **Und ein anderer Same gibt eine andere.** Ohne diese Zeile bestünde
	// den Lauf darüber auch ein Zieher, der immer dasselbe gibt.
	andere, fehler := bauen().Ziehen("blatt", 43)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	if namenreihe(erste) == namenreihe(andere) {
		t.Error("zwei Samen gaben dieselbe Ziehung")
	}
}

// **Ein Kreis wird abgewiesen, nicht durchlaufen.**
//
// MSE hätte dort eine endlose Rekursion; hier ist es ein Befund mit dem Weg
// darin — dieselbe Haltung wie beim Einbinder.
func TestEinPackkreisWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "a", Auswahl: "all", Posten: []msepaket.Packposten{{Name: "b", Anzahl: 1, Gewicht: 1}}},
		{Name: "b", Auswahl: "all", Posten: []msepaket.Packposten{{Name: "a", Anzahl: 1, Gewicht: 1}}},
	}, nil)

	_, fehler := zieher.Ziehen("a", 1)
	if fehler == nil {
		t.Fatal("der Kreis lief durch")
	}

	if !errors.Is(fehler, domain.ErrValidation) || !strings.Contains(fehler.Error(), "Kreis") {
		t.Errorf("Fehler %v", fehler)
	}
}

// **Eine Packart, die es nicht gibt, ist ein Befund.**
func TestEineUnbekanntePackartIstEinBefund(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher(nil, nil)

	if _, fehler := zieher.Ziehen("gibtsnicht", 1); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v", fehler)
	}
}

// packkarten baut eine Reihe von Karten mit fortlaufendem Namen.
func packkarten(vorsilbe string, wieViele int) []*domain.CardRow {
	reihe := make([]*domain.CardRow, 0, wieViele)

	for nummer := range wieViele {
		reihe = append(reihe, &domain.CardRow{
			ID: int64(nummer),
			Values: []domain.Value{
				{Field: "name", Text: vorsilbe + strconv.Itoa(nummer)},
			},
		})
	}

	return reihe
}

// namenVon gibt den Namen einer Karte.
func namenVon(karte *domain.CardRow) string {
	for _, wert := range karte.Values {
		if wert.Field == "name" {
			return wert.Text
		}
	}

	return ""
}

// namenreihe gibt die Namen einer Ziehung als eine Zeichenkette.
func namenreihe(gezogen []*domain.CardRow) string {
	namen := make([]string, 0, len(gezogen))
	for _, karte := range gezogen {
		namen = append(namen, namenVon(karte))
	}

	return strings.Join(namen, ",")
}
