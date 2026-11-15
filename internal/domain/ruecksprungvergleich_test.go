package domain_test

// **Der Vergleich vor der Rückkehr** — Werkstück O7 von Phase 5.2
// (#1431).
//
// Zwei Fragen, und sie zeigen in **verschiedene** Richtungen:
//
//  1. Was seit der Sicherung dazukam, **verschwindet** — der Datenverlust.
//  2. Was die Sicherung mehr hat als der Bestand, steht danach wieder da —
//     und **niemand weiß, ob es noch jemandem gehört**.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func stand(sets, karten, decks, exemplare int, sammlung ...domain.Sammlungsmenge) domain.Bestandsstand {
	return domain.Bestandsstand{
		AngelegtAm: time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC),
		Sets:       sets, Karten: karten, Decks: decks, Exemplare: exemplare,
		Sammlung: sammlung, SammlungGelesen: true,
	}
}

// **Der Verlust steht Gegenstand für Gegenstand da** — „es geht etwas
// verloren" ist keine Auskunft, mit der jemand entscheiden kann.
func TestDerVergleichNenntJedenGegenstandEinzeln(t *testing.T) {
	t.Parallel()

	lage := domain.Ruecksprungvergleich(
		stand(5, 300, 2, 40),
		stand(6, 312, 2, 38),
	)

	if len(lage.Zeilen) != 4 {
		t.Fatalf("es stehen %d Zeilen da: %+v", len(lage.Zeilen), lage.Zeilen)
	}

	verluste := map[string]int{}
	for _, zeile := range lage.Zeilen {
		verluste[zeile.Gegenstand] = zeile.Verlust()
	}

	for gegenstand, erwartet := range map[string]int{
		domain.GegenstandSets: 1, domain.GegenstandKarten: 12,
		domain.GegenstandDecks: 0, domain.GegenstandExemplare: 0,
	} {
		if verluste[gegenstand] != erwartet {
			t.Errorf("%s verliert %d, erwartet war %d", gegenstand, verluste[gegenstand], erwartet)
		}
	}

	if !lage.Datenstand.Equal(time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC)) {
		t.Errorf("der Datenstand ist %v", lage.Datenstand)
	}
}

// **Und die andere Richtung: was wieder dasteht, ist unklar.**
func TestWasDieSicherungMehrHatIstNachDerRueckkehrUnklar(t *testing.T) {
	t.Parallel()

	lage := domain.Ruecksprungvergleich(
		stand(1, 1, 0, 4,
			domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Zustand: "nm", Sprache: "en", Menge: 4, Name: "Bolt"},
			domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:xyz", Zustand: "nm", Sprache: "en", Menge: 2},
		),
		stand(1, 1, 0, 3,
			domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Zustand: "nm", Sprache: "en", Menge: 1},
			domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:xyz", Zustand: "nm", Sprache: "en", Menge: 5},
		),
	)

	if len(lage.Unklar) != 1 {
		t.Fatalf("es stehen %d unklare Mengen da: %+v", len(lage.Unklar), lage.Unklar)
	}

	unklar := lage.Unklar[0]

	if unklar.Schluessel != "scryfall:abc" || unklar.Menge != 3 {
		t.Errorf("unklar ist %+v, erwartet waren drei von scryfall:abc", unklar)
	}

	// **Der lesbare Bezug kommt aus der Sicherung** — sie ist die Seite, die
	// danach dasteht.
	if unklar.Name != "Bolt" {
		t.Errorf("der Name fehlt: %+v", unklar)
	}
}

// **Ein Gegenstand, den es jetzt gar nicht mehr gibt, ist ganz unklar.**
func TestEineZeileDieEsJetztNichtGibtIstGanzUnklar(t *testing.T) {
	t.Parallel()

	lage := domain.Ruecksprungvergleich(
		stand(0, 0, 0, 2,
			domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 2}),
		stand(0, 0, 0, 0),
	)

	if len(lage.Unklar) != 1 || lage.Unklar[0].Menge != 2 {
		t.Errorf("unklar ist %+v, erwartet waren zwei", lage.Unklar)
	}
}

// **Profile werden nicht vermischt.** Zwei Profile, derselbe Druck: die
// Rückkehr betrifft nur das, dessen Menge sich unterscheidet.
func TestZweiProfileWerdenNichtVermischt(t *testing.T) {
	t.Parallel()

	lage := domain.Ruecksprungvergleich(
		stand(0, 0, 0, 6,
			domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 3},
			domain.Sammlungsmenge{Profil: 2, Schluessel: "scryfall:abc", Menge: 3},
		),
		stand(0, 0, 0, 4,
			domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 1},
			domain.Sammlungsmenge{Profil: 2, Schluessel: "scryfall:abc", Menge: 3},
		),
	)

	if len(lage.Unklar) != 1 {
		t.Fatalf("es stehen %d unklare Mengen da: %+v", len(lage.Unklar), lage.Unklar)
	}

	if lage.Unklar[0].Profil != 1 || lage.Unklar[0].Menge != 2 {
		t.Errorf("unklar ist %+v, erwartet waren zwei im Profil 1", lage.Unklar[0])
	}
}

// **Eine Rückkehr auf denselben Stand lässt nichts unklar** — und das ist der
// Fall, den eine Sperre, die immer sperrt, nicht von den anderen
// unterscheidet (§10).
func TestEineRueckkehrAufDenselbenStandLaesstNichtsUnklar(t *testing.T) {
	t.Parallel()

	gleich := stand(3, 20, 1, 5,
		domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 5})

	lage := domain.Ruecksprungvergleich(gleich, gleich)

	if len(lage.Unklar) != 0 {
		t.Errorf("es blieb etwas unklar: %+v", lage.Unklar)
	}

	for _, zeile := range lage.Zeilen {
		if zeile.Verlust() != 0 {
			t.Errorf("%s verliert %d", zeile.Gegenstand, zeile.Verlust())
		}
	}
}

// **Die unklaren Mengen stehen in fester Ordnung** — sonst zeigte die
// Oberfläche bei jedem Aufruf eine andere Reihe, und niemand fände etwas
// wieder.
//
// **Fünf Fälle, weil die Ordnung fünf Stufen hat.** Ein Lauf, der nur den
// Druck unterscheidet, ließe die vier darunter ungeprüft — und dann steht in
// der Oberfläche Foil vor nicht-Foil oder umgekehrt, je nach Tageslaune der
// Sortierung (§10).
func TestDieUnklarenMengenStehenInFesterOrdnung(t *testing.T) {
	t.Parallel()

	sicherung := stand(0, 0, 0, 9)
	for _, schluessel := range []string{"scryfall:c", "scryfall:a", "scryfall:b"} {
		sicherung.Sammlung = append(sicherung.Sammlung,
			domain.Sammlungsmenge{Profil: 1, Schluessel: schluessel, Menge: 3})
	}

	lage := domain.Ruecksprungvergleich(sicherung, stand(0, 0, 0, 0))

	reihe := make([]string, 0, len(lage.Unklar))
	for _, unklar := range lage.Unklar {
		reihe = append(reihe, unklar.Schluessel)
	}

	if len(reihe) != 3 || reihe[0] != "scryfall:a" || reihe[2] != "scryfall:c" {
		t.Errorf("die Reihe ist %v", reihe)
	}
}

// **Und die Ordnung greift auf jeder ihrer fünf Stufen.**
func TestDieOrdnungGreiftAufJederStufe(t *testing.T) {
	t.Parallel()

	// **Absichtlich verkehrt herum eingetragen** — käme die Reihe so heraus,
	// wie sie hineinging, sagte der Lauf nichts über die Sortierung.
	sicherung := stand(0, 0, 0, 12,
		domain.Sammlungsmenge{Profil: 2, Schluessel: "a", Zustand: "nm", Sprache: "de", Menge: 1},
		domain.Sammlungsmenge{Profil: 1, Schluessel: "b", Zustand: "nm", Sprache: "de", Menge: 1},
		domain.Sammlungsmenge{Profil: 1, Schluessel: "a", Zustand: "sp", Sprache: "de", Menge: 1},
		domain.Sammlungsmenge{Profil: 1, Schluessel: "a", Zustand: "nm", Sprache: "en", Menge: 1},
		domain.Sammlungsmenge{Profil: 1, Schluessel: "a", Zustand: "nm", Sprache: "de", Finish: domain.FinishFoil, Menge: 1},
		domain.Sammlungsmenge{Profil: 1, Schluessel: "a", Zustand: "nm", Sprache: "de", Menge: 1},
	)

	lage := domain.Ruecksprungvergleich(sicherung, stand(0, 0, 0, 0))

	// **`foil` steht vor `nonfoil`, seit die Ausführung ein Wort ist** (T2,
	// #1467). Vorher war sie ein Wahrheitswert, und `false` kam vor `true`.
	// **Die Zusage ist die Bestimmtheit und nicht die Reihenfolge** — welche
	// von zweien vorn steht, ist gleichgültig, solange es immer dieselbe ist.
	erwartet := []domain.Sammlungsmenge{
		{Profil: 1, Schluessel: "a", Zustand: "nm", Sprache: "de", Finish: domain.FinishFoil, Menge: 1},
		{Profil: 1, Schluessel: "a", Zustand: "nm", Sprache: "de", Menge: 1},
		{Profil: 1, Schluessel: "a", Zustand: "nm", Sprache: "en", Menge: 1},
		{Profil: 1, Schluessel: "a", Zustand: "sp", Sprache: "de", Menge: 1},
		{Profil: 1, Schluessel: "b", Zustand: "nm", Sprache: "de", Menge: 1},
		{Profil: 2, Schluessel: "a", Zustand: "nm", Sprache: "de", Menge: 1},
	}

	if len(lage.Unklar) != len(erwartet) {
		t.Fatalf("es stehen %d Mengen da, erwartet waren %d", len(lage.Unklar), len(erwartet))
	}

	for stelle, soll := range erwartet {
		if lage.Unklar[stelle] != soll {
			t.Errorf("an Stelle %d steht %+v, erwartet war %+v", stelle, lage.Unklar[stelle], soll)
		}
	}
}

// **Sagt eine der beiden Seiten nichts über ihre Sammlung, wird von Hand
// nachgesehen** — und nicht „nichts unklar" gemeldet.
//
// Das ist die Zusage aus dem Issue: bei fehlenden Nachweisen bleibt die
// Unsicherheit sichtbar. Eine leere Liste sähe hier aus wie ein Freibrief.
func TestOhneSammlungsauskunftWirdVonHandNachgesehen(t *testing.T) {
	t.Parallel()

	stumm := stand(1, 1, 0, 0)
	stumm.SammlungGelesen = false

	for name, lage := range map[string]domain.Ruecksprunglage{
		"die Sicherung schweigt": domain.Ruecksprungvergleich(stumm, stand(1, 1, 0, 0)),
		"der Bestand schweigt":   domain.Ruecksprungvergleich(stand(1, 1, 0, 0), stumm),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !lage.HandpruefungNoetig {
				t.Error("die Lage verlangt keine Handprüfung")
			}
		})
	}

	// **Die Gegenprobe**: sprechen beide, ist keine nötig — sonst stünde der
	// Satz immer da und sagte nichts (§10).
	if domain.Ruecksprungvergleich(stand(1, 1, 0, 0), stand(1, 1, 0, 0)).HandpruefungNoetig {
		t.Error("zwei sprechende Seiten verlangen eine Handprüfung")
	}
}
