package domain_test

// **Die Commandersteuer und der Commanderschaden** — CR 903.8, CR 903.10a,
// Werkstück M19
// (#1620).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was eine Commanderpartie *ist*** — vierzig Leben, ein
// benannter Commander, der Weg zurück in die Kommandozone. Hier steht, was
// sie **kostet** und was sie **zählt**, und `commander_test.go` stand mit
// beidem bei 522 Zeilen (§2).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Commander aus der Kommandozone kostet beim zweiten Mal mehr** —
// CR 903.8.
func TestEinCommanderAusDerKommandozoneKostetMehr(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)
	partie.Objekte[0].Zone = domain.ZoneKommando
	partie.Zug = domain.Zugstand{
		Nummer: 1, Aktiv: "sitz-1", Prioritaet: "sitz-1",
	}

	// **Drei Mana reichen beim ersten Mal** — `2R`.
	partie.Sitze[0].Manavorrat = domain.Manavorrat{
		domain.ManaRot: 1, domain.ManaFarblos: 2,
	}

	erst, fehler := partie.Wirken("sitz-1", "sitz-1/held/0", nil)
	if fehler != nil {
		t.Fatalf("erstes wirken: %v", fehler)
	}

	if steuer := erst.Sitze[0].Commandersteuer; steuer != 1 {
		t.Errorf("die Steuer steht bei %d", steuer)
	}

	// **Beim zweiten Mal kostet er fünf** — drei plus zwei (CR 903.8).
	zweit := erst
	zweit.Objekte[0].Zone = domain.ZoneKommando
	zweit.Sitze[0].Manavorrat = domain.Manavorrat{
		domain.ManaRot: 1, domain.ManaFarblos: 3,
	}

	if _, fehler := zweit.Wirken("sitz-1", "sitz-1/held/0", nil); fehler == nil {
		t.Error("vier Mana haben für den zweiten Ruf gereicht")
	}

	zweit.Sitze[0].Manavorrat = domain.Manavorrat{
		domain.ManaRot: 1, domain.ManaFarblos: 4,
	}

	if _, fehler := zweit.Wirken("sitz-1", "sitz-1/held/0", nil); fehler != nil {
		t.Errorf("fünf Mana haben nicht gereicht: %v", fehler)
	}
}

// **Aus der Hand kostet er nichts dazu** — CR 903.8 nennt die Kommandozone.
func TestAusDerHandKostetDerCommanderNichtsDazu(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)
	partie.Objekte[0].Zone = domain.ZoneHand
	partie.Zug = domain.Zugstand{
		Nummer: 1, Aktiv: "sitz-1", Prioritaet: "sitz-1",
	}
	partie.Sitze[0].Commandersteuer = 3
	partie.Sitze[0].Manavorrat = domain.Manavorrat{
		domain.ManaRot: 1, domain.ManaFarblos: 2,
	}

	danach, fehler := partie.Wirken("sitz-1", "sitz-1/held/0", nil)
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Und die Steuer steigt nicht** — sie zählt Rufe aus der
	// Kommandozone.
	if steuer := danach.Sitze[0].Commandersteuer; steuer != 3 {
		t.Errorf("die Steuer steht bei %d", steuer)
	}
}

// **Kampfschaden eines Commanders wird je Commander gezählt** — CR 903.10a.
func TestCommanderschadenWirdJeCommanderGezaehlt(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)
	partie.Zug = domain.Zugstand{
		Nummer: 3, Aktiv: "sitz-1", Prioritaet: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker,
		},
		Kampf: domain.Kampfstand{
			Erklaert: true, Geblockt: true,
			Angriffe: []domain.Angriff{
				{Kreatur: "sitz-1/held/0", Ziel: "sitz-2"},
				{Kreatur: "sitz-1/baer/0", Ziel: "sitz-2"},
			},
		},
	}
	partie.Objekte[0].Seit = 1
	partie.Objekte[1].Seit = 1

	// **Über den Zug und nicht über einen Befehl** — der Kampfschaden ist
	// eine zugbasierte Handlung (CR 703.4m) und geschieht, wenn der
	// Schritt beginnt. Beide passen, und der Schritt kommt.
	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach, fehler = danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	zweiter := danach.SitzVon("sitz-2")
	if zweiter == nil {
		t.Fatal("den zweiten Sitz gibt es nicht")
	}

	// **Der Held ist der Commander und teilt drei aus**, der Bär zwei —
	// und nur die drei zählen (CR 903.10a).
	if schaden := zweiter.Commanderschaden["sitz-1/held/0"]; schaden != 3 {
		t.Errorf("der Commanderschaden steht bei %d", schaden)
	}

	if _, dabei := zweiter.Commanderschaden["sitz-1/baer/0"]; dabei {
		t.Error("der Bär zählt als Commander mit")
	}

	// **Und das Leben sinkt um beides** — CR 903.10a zählt zusätzlich und
	// nicht anstelle.
	if zweiter.Leben != 35 {
		t.Errorf("der zweite Sitz hat %d Leben", zweiter.Leben)
	}
}

// **Ein Spielstein ist kein Commander** — CR 903.3 nennt eine **Karte**.
func TestEinSpielsteinIstKeinCommander(t *testing.T) {
	t.Parallel()

	partie, fehler := eineCommanderpartie(t).Spielstein(
		"sitz-1", "sitz-1/soldat/0", map[string]string{"name": "Soldat"})
	if fehler != nil {
		t.Fatalf("spielstein: %v", fehler)
	}

	geschoben, fehler := partie.Exilieren("sitz-1/soldat/0")
	if fehler != nil {
		t.Fatalf("exilieren: %v", fehler)
	}

	_, fehler = geschoben.Heimholen("sitz-1", "sitz-1/soldat/0")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Spielstein als Commander: %v", fehler)
	}
}

// **Und ein Objekt eines Sitzes, den es nicht gibt, ebenso wenig** — §8.
func TestEinObjektOhneSitzIstKeinCommander(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)
	partie.Objekte[0].Eigner = "weg"
	partie.Objekte[0].Zone = domain.ZoneFriedhof

	_, fehler := partie.Heimholen("weg", "sitz-1/held/0")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Objekt ohne Sitz: %v", fehler)
	}
}

// **Ein Commander allein trifft und wird gezählt** — CR 903.10a.
func TestEinCommanderAlleinWirdGezaehlt(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)
	partie.Zug = domain.Zugstand{
		Nummer: 3, Aktiv: "sitz-1", Prioritaet: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker,
		},
		Kampf: domain.Kampfstand{
			Erklaert: true, Geblockt: true,
			Angriffe: []domain.Angriff{
				{Kreatur: "sitz-1/held/0", Ziel: "sitz-2"},
			},
		},
	}
	partie.Objekte[0].Seit = 1

	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach, fehler = danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	if zweiter := danach.SitzVon("sitz-2"); zweiter == nil ||
		zweiter.Commanderschaden["sitz-1/held/0"] != 3 {
		t.Errorf("der Commanderschaden kam nicht an: %+v", zweiter)
	}
}

// **Was sonst in der Kommandozone liegt, zahlt keine Steuer** — CR 903.8
// nennt den Commander.
func TestNurDerCommanderZahltDieSteuer(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)
	partie.Objekte[1].Zone = domain.ZoneKommando
	partie.Zug = domain.Zugstand{
		Nummer: 1, Aktiv: "sitz-1", Prioritaet: "sitz-1",
	}
	partie.Sitze[0].Commandersteuer = 5
	partie.Sitze[0].Manavorrat = domain.Manavorrat{}

	// **Der Bär kostet nichts** — und die Steuer von fünf würde zehn
	// generisches Mana verlangen, wenn sie für ihn gälte.
	danach, fehler := partie.Wirken("sitz-1", "sitz-1/baer/0", nil)
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	if steuer := danach.Sitze[0].Commandersteuer; steuer != 5 {
		t.Errorf("die Steuer steht bei %d", steuer)
	}
}
