package domain_test

// **Ein Land spielen** — Werkstück M4 der Regelmaschine
// (#1606).
//
// # Wogegen hier geprüft wird
//
// **CR 116.2a** (Fassung 2026-08-19): ein Land zu spielen ist eine
// **Sonderhandlung** — sie geht nicht über den Stapel, und sie geht nur, wenn
// jemand Priorität hat, der Stapel leer ist und es eine Hauptphase seines
// eigenen Zuges ist.
//
// **CR 305.2**: einmal je Zug. Der zweite Versuch wird **abgewiesen** und
// nicht still verschluckt — so steht es in der Abnahme des Issues.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineLandpartie gibt eine Partie mit zwei Wäldern auf der Hand, in der
// ersten Hauptphase.
func eineLandpartie(t *testing.T) domain.Partie {
	t.Helper()

	partie := einePartie().ZugBeginnen("sitz-1")

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "wald", Name: "Forest", Menge: 2,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Basic Land", "sub type": "Forest",
			}}},
		}, {
			Kennung: "blitz", Name: "Blitz", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Instant", "casting cost": "[R]",
			}}},
		}},
	}
	partie.Sitze = sitze

	for _, kennung := range []string{"sitz-1/wald/0", "sitz-1/wald/1"} {
		partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
			Kennung: kennung, Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneHand, Karte: "wald",
		})
	}

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/blitz/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneHand, Karte: "blitz",
	})

	return inDieHauptphase(t, partie)
}

// inDieHauptphase fährt den Zug bis zur ersten Hauptphase.
//
// **Über den Weg und nicht durch eine Zuweisung** (§3): wer den Abschnitt
// hinschriebe, prüfte die Vorrichtung statt des Zugs.
func inDieHauptphase(t *testing.T, partie domain.Partie) domain.Partie {
	t.Helper()

	for range 20 {
		if partie.Zug.Abschnitt.Phase == domain.PhaseHauptVorKampf {
			return partie
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen bei %+v: %v", partie.Zug.Abschnitt, fehler)
		}

		partie = danach
	}

	t.Fatal("die Hauptphase kam nicht")

	return partie
}

// **Ein Land kommt auf das Schlachtfeld, ohne den Stapel zu berühren** —
// CR 116.2a.
func TestEinLandKommtOhneStapelAufDenTisch(t *testing.T) {
	t.Parallel()

	danach, fehler := eineLandpartie(t).LandSpielen("sitz-1", "sitz-1/wald/0")
	if fehler != nil {
		t.Fatalf("land spielen: %v", fehler)
	}

	gespielt := objektMit(t, danach, "sitz-1/wald/0")

	switch {
	case gespielt.Zone != domain.ZoneTisch:
		t.Errorf("das Land liegt in %q", gespielt.Zone)
	case danach.StapelOben() != nil:
		t.Error("etwas liegt auf dem Stapel")
	case danach.Zug.Laender != 1:
		t.Errorf("gezählt sind %d Länder", danach.Zug.Laender)
	}
}

// **Ein zweites Land im selben Zug wird abgewiesen** — CR 305.2.
//
// **Das ist die Abnahme des Issues**: „Ein Land je Zug, und der zweite
// Versuch wird abgewiesen statt still verschluckt."
func TestEinZweitesLandImSelbenZugWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	partie, fehler := eineLandpartie(t).LandSpielen("sitz-1", "sitz-1/wald/0")
	if fehler != nil {
		t.Fatalf("erstes Land: %v", fehler)
	}

	danach, fehler := partie.LandSpielen("sitz-1", "sitz-1/wald/1")

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 305.2") {
		t.Fatalf("der Fehler lautet %v", fehler)
	}

	// **Und es bleibt auf der Hand** — kein halber Stand (CR 733).
	if zweites := objektMit(t, danach, "sitz-1/wald/1"); zweites.Zone !=
		domain.ZoneHand {
		t.Errorf("das zweite Land liegt in %q", zweites.Zone)
	}
}

// **Im nächsten Zug geht wieder eines** — CR 305.2a zählt je Zug.
//
// **Die Gegenprobe** (§10): ohne sie sagte der Lauf darüber nur, dass
// **irgendwann** abgewiesen wird — ein Zähler, der nie zurückgesetzt würde,
// käme genauso durch, und niemand spielte je ein zweites Land.
func TestImNaechstenZugGehtWiederEines(t *testing.T) {
	t.Parallel()

	partie, fehler := eineLandpartie(t).LandSpielen("sitz-1", "sitz-1/wald/0")
	if fehler != nil {
		t.Fatalf("erstes Land: %v", fehler)
	}

	vorher := partie.Zug.Nummer

	for range 40 {
		if partie.Zug.Nummer > vorher &&
			partie.Zug.Abschnitt.Phase == domain.PhaseHauptVorKampf {
			break
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen: %v", fehler)
		}

		partie = danach
	}

	if partie.Zug.Laender != 0 {
		t.Fatalf("der Zähler steht im neuen Zug bei %d", partie.Zug.Laender)
	}
}

// **Was kein Land ist, wird nicht gespielt** — CR 116.2a verweist auf CR 305.
func TestEinBlitzIstKeinLand(t *testing.T) {
	t.Parallel()

	_, fehler := eineLandpartie(t).LandSpielen("sitz-1", "sitz-1/blitz/0")

	if fehler == nil || !strings.Contains(fehler.Error(), "kein Land") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Außerhalb der Hauptphase geht es nicht** — CR 116.2a.
func TestAusserhalbDerHauptphaseGehtKeinLand(t *testing.T) {
	t.Parallel()

	partie := einePartie().ZugBeginnen("sitz-1")

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "wald", Name: "Forest", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Basic Land", "sub type": "Forest",
			}}},
		}},
	}
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/wald/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneHand, Karte: "wald",
	})

	// **Der Versorgungsschritt vergibt Priorität** (CR 503) und ist keine
	// Hauptphase — genau der Fall, den CR 116.2a ausschließt.
	if partie.Zug.Abschnitt.Phase != domain.PhaseAnfang {
		t.Fatalf("der Zug steht bei %+v", partie.Zug.Abschnitt)
	}

	_, fehler := partie.LandSpielen("sitz-1", "sitz-1/wald/0")

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 116.2a") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Jeder Weg, auf dem ein Land nicht gespielt wird, sagt warum** — §8.
func TestJederFehlerwegDesLandspielensSagtWarum(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		bauen  func(t *testing.T) domain.Partie
		wer    string
		objekt string
		wort   string
	}{
		"ohne Priorität": {
			bauen: eineLandpartie, wer: "sitz-2", objekt: "sitz-1/wald/0",
			wort: "keine Priorität",
		},

		"nicht am Zug": {
			bauen: fremderZug, wer: "sitz-2", objekt: "sitz-1/wald/0",
			wort: "nicht der Zug",
		},

		"bei vollem Stapel": {
			bauen: mitEtwasAufDemStapel, wer: "sitz-1",
			objekt: "sitz-1/wald/0",
			wort:   "leerem Stapel",
		},

		"eine fremde Karte": {
			bauen: mitFremdemLand, wer: "sitz-1", objekt: "sitz-2/wald/0",
			wort: "beherrscht das Objekt",
		},

		"eine Karte ohne Abzug": {
			bauen: ohneLandabzug, wer: "sitz-1", objekt: "sitz-1/wald/0",
			wort: "eingefrorenen Abzug",
		},

		"ein Objekt, das es nicht gibt": {
			bauen: eineLandpartie, wer: "sitz-1", objekt: "gibtsnicht",
			wort: "gibt es in dieser Partie nicht",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := fall.bauen(t).LandSpielen(fall.wer, fall.objekt)

			if fehler == nil {
				t.Fatal("es kam kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %q", fehler)
			}
		})
	}
}

// fremderZug gibt dem zweiten Sitz Priorität im Zug des ersten.
//
// **Ein Zustand, den CR 117 im Spiel erlaubt**: wer nicht am Zug ist, kann
// sehr wohl Priorität haben — er darf nur kein Land spielen (CR 116.2a).
func fremderZug(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineLandpartie(t)
	partie.Zug.Prioritaet = "sitz-2"

	return partie
}

// mitEtwasAufDemStapel legt den Blitz auf den Stapel.
func mitEtwasAufDemStapel(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineLandpartie(t)

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Kennung == "sitz-1/blitz/0" {
			objekte[stelle].Zone = domain.ZoneStapel
		}
	}

	partie.Objekte = objekte

	return partie
}

// mitFremdemLand legt ein Land an den zweiten Sitz.
func mitFremdemLand(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineLandpartie(t)

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/wald/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneHand, Karte: "wald",
	})

	return partie
}

// ohneLandabzug nimmt dem Sitz seinen eingefrorenen Abzug.
func ohneLandabzug(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineLandpartie(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{Abdruck: "abcd"}
	partie.Sitze = sitze

	return partie
}

// **In einem Sandkasten gilt CR 116.2a nicht** — es gibt keine Hauptphase.
//
// **§15, Punkt 4**: eine Partie aus G1 bis G8 hat keinen Zugbegriff, und sie
// soll weiterlaufen, wie sie angefangen hat. Ein Land dort abzuweisen, weil
// „es keine Hauptphase ist", wäre eine Regel, die aus dem Nichts käme.
func TestImSandkastenGehtEinLandJederzeit(t *testing.T) {
	t.Parallel()

	partie := einePartie()

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "wald", Name: "Forest", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Basic Land", "sub type": "Forest",
			}}},
		}},
	}
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/wald/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneHand, Karte: "wald",
	})

	if partie.Zug.Nummer != 0 {
		t.Fatalf("die Partie hat einen Zug: %d", partie.Zug.Nummer)
	}

	danach, fehler := partie.LandSpielen("sitz-1", "sitz-1/wald/0")
	if fehler != nil {
		t.Fatalf("im Sandkasten: %v", fehler)
	}

	if gespielt := objektMit(t, danach,
		"sitz-1/wald/0"); gespielt.Zone != domain.ZoneTisch {
		t.Errorf("das Land liegt in %q", gespielt.Zone)
	}
}
