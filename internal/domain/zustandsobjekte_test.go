package domain_test

// **Die zustandsbasierten Aktionen, die ein Objekt kosten** — Werkstück M5
// der Regelmaschine
// (#1607).
//
// # Wogegen hier geprüft wird
//
//   - CR 704.5d — ein Spielstein außerhalb des Schlachtfelds hört auf zu
//     existieren.
//   - CR 704.5f — Widerstandskraft null oder weniger: in den Friedhof.
//   - CR 704.5g — markierter Schaden ab der Widerstandskraft: zerstört.
//   - CR 704.5i — ein Planeswalker mit Loyalität null.
//   - CR 704.5v/w — eine Kampfkarte mit Verteidigung null.
//   - CR 704.5m/n/p — was an nichts hängt, und was gar nicht hängen darf.
//   - CR 704.3 — sie **wiederholen sich**, bis keine mehr anliegt.
//
// **In den Friedhof des Eigners**, sagt jede dieser Regeln — nicht des
// Beherrschers.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineTischpartie gibt eine Partie mit einer Kreatur auf dem Tisch.
func eineTischpartie(t *testing.T, felder map[string]string) domain.Partie {
	t.Helper()

	partie := eineKartenpartie()
	partie.Sitze[0].Deck.Karten[0].Seiten[0].Felder = felder

	return partie
}

// **Ein Spielstein außerhalb des Schlachtfelds hört auf zu existieren** —
// CR 704.5d.
func TestEinSpielsteinAusserhalbDesTischesVerschwindet(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "stein", Art: domain.ObjektSpielstein, Name: "Goblin",
		Beherrscher: "sitz-1", Eigner: "sitz-1", Zone: domain.ZoneFriedhof,
	})

	danach := partie.Zustandspruefung()

	for _, objekt := range danach.Objekte {
		if objekt.Kennung == "stein" {
			t.Fatalf("der Spielstein liegt noch in %q", objekt.Zone)
		}
	}
}

// **Auf dem Tisch bleibt er** — die Gegenprobe.
func TestEinSpielsteinAufDemTischBleibt(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "stein", Art: domain.ObjektSpielstein, Name: "Goblin",
		Beherrscher: "sitz-1", Eigner: "sitz-1", Zone: domain.ZoneTisch,
	})

	if len(partie.Zustandspruefung().Objekte) != len(partie.Objekte) {
		t.Fatal("der Spielstein auf dem Tisch ist verschwunden")
	}
}

// **Widerstandskraft null oder weniger** — CR 704.5f: in den Friedhof des
// **Eigners**, und Regeneration ersetzt das nicht.
func TestEineKreaturOhneWiderstandskraftGehtInDenFriedhof(t *testing.T) {
	t.Parallel()

	partie := eineTischpartie(t, map[string]string{
		"name": "Winzling", "super type": "Creature", "sub type": "Goblin",
		"power": "1", "toughness": "1",
	})
	partie.Objekte[0].Beherrscher = "sitz-2"
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkeMinusEins: 1}

	danach := partie.Zustandspruefung()

	objekt := objektMit(t, danach, partie.Objekte[0].Kennung)

	switch {
	case objekt.Zone != domain.ZoneFriedhof:
		t.Errorf("die Kreatur liegt in %q", objekt.Zone)
	case objekt.Beherrscher != "sitz-1":
		t.Errorf("im Friedhof beherrscht sie %q — es ist der des Eigners",
			objekt.Beherrscher)
	}
}

// **Tödlicher Schaden** — CR 704.5g: markierter Schaden ab der
// Widerstandskraft zerstört die Kreatur.
func TestEineKreaturMitToedlichemSchadenWirdZerstoert(t *testing.T) {
	t.Parallel()

	partie := eineTischpartie(t, map[string]string{
		"name": "Bär", "super type": "Creature", "sub type": "Bear",
		"power": "2", "toughness": "2",
	})
	partie.Objekte[0].Schaden = 2

	danach := partie.Zustandspruefung()

	if objektMit(t, danach, partie.Objekte[0].Kennung).Zone != domain.ZoneFriedhof {
		t.Fatal("die Kreatur steht noch auf dem Tisch")
	}
}

// **Ein Schaden weniger tut nichts** — die Gegenprobe zu CR 704.5g.
func TestEinSchadenUnterDerWiderstandskraftTutNichts(t *testing.T) {
	t.Parallel()

	partie := eineTischpartie(t, map[string]string{
		"name": "Bär", "super type": "Creature", "sub type": "Bear",
		"power": "2", "toughness": "2",
	})
	partie.Objekte[0].Schaden = 1

	if objektMit(t, partie.Zustandspruefung(), partie.Objekte[0].Kennung).
		Zone != domain.ZoneTisch {
		t.Fatal("die Kreatur ist gestorben")
	}
}

// **Und eine Karte ohne Widerstandskraft stirbt an keinem Schaden** —
// CR 208.3 mit CR 704.5g: eine Nichtkreatur hat keine.
func TestEinLandStirbtAnKeinemSchaden(t *testing.T) {
	t.Parallel()

	partie := eineTischpartie(t, map[string]string{
		"name": "Berg", "super type": "Basic Land", "sub type": "Mountain",
	})
	partie.Objekte[0].Schaden = 99

	if objektMit(t, partie.Zustandspruefung(), partie.Objekte[0].Kennung).
		Zone != domain.ZoneTisch {
		t.Fatal("das Land ist gestorben")
	}
}

// **Ein Planeswalker mit Loyalität null** — CR 704.5i.
func TestEinPlaneswalkerOhneLoyalitaetGehtInDenFriedhof(t *testing.T) {
	t.Parallel()

	partie := eineTischpartie(t, map[string]string{
		"name": "Jace", "super type": "Legendary Planeswalker",
		"sub type": "Jace", "loyalty": "3",
	})
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkeLoyalitaet: 0}

	if objektMit(t, partie.Zustandspruefung(), partie.Objekte[0].Kennung).
		Zone != domain.ZoneFriedhof {
		t.Fatal("der Planeswalker steht noch auf dem Tisch")
	}
}

// **Ohne Loyalitätsmarken geschieht nichts** — CR 306.5b legt sie beim
// Hineinkommen an, und das ist ein Ersetzungseffekt (`M11`). Solange keine
// dastehen, weiß niemand, wie viele es sein sollten.
func TestEinPlaneswalkerOhneMarkenBleibtStehen(t *testing.T) {
	t.Parallel()

	partie := eineTischpartie(t, map[string]string{
		"name": "Jace", "super type": "Legendary Planeswalker",
		"sub type": "Jace", "loyalty": "3",
	})

	if objektMit(t, partie.Zustandspruefung(), partie.Objekte[0].Kennung).
		Zone != domain.ZoneTisch {
		t.Fatal("der Planeswalker ist gestorben")
	}
}

// **Eine Kampfkarte mit Verteidigung null** — CR 704.5w.
func TestEineKampfkarteOhneVerteidigungGehtInDenFriedhof(t *testing.T) {
	t.Parallel()

	partie := eineTischpartie(t, map[string]string{
		"name": "Belagerung", "super type": "Battle", "sub type": "Siege",
		"defense": "4",
	})
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkeVerteidigung: 0}

	if objektMit(t, partie.Zustandspruefung(), partie.Objekte[0].Kennung).
		Zone != domain.ZoneFriedhof {
		t.Fatal("die Kampfkarte steht noch auf dem Tisch")
	}
}
