package domain_test

// **Die Wege des Kampfschadens, die keinen Fehler geben** — Werkstück M6 der
// Regelmaschine
// (#1608).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was abgewiesen wird**; hier steht, **was still geschieht**
// — und `kampffehlerwege_test.go` stand mit beidem bei 523 Zeilen (§2).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die stillen Wege des Kampfschadens** — §8 und §10.
//
// # Warum sie einen Lauf haben
//
// **Weil keiner von ihnen einen Fehler gibt.** Eine Kreatur ohne lesbare
// Stärke teilt **nichts** zu (CR 510.1a), ein Blocker, den es nicht mehr
// gibt, ebenso (CR 510.1c) — und beides sieht von außen aus wie „es ist
// nichts passiert". Ein Zweig, den kein Lauf betritt, ist ein Versprechen,
// das niemand hält (§10).
func TestDieStillenWegeDesKampfschadens(t *testing.T) {
	t.Parallel()

	t.Run("eine Kreatur ohne Stärke teilt nichts zu", func(t *testing.T) {
		t.Parallel()

		partie := ohneStaerke(t)

		vorher := partie.Sitze[1].Leben

		danach := bisZumSchritt(t, partie, domain.SchrittKampfschaden)

		if danach.Sitze[1].Leben != vorher {
			t.Errorf("sitz-2 steht bei %d statt %d",
				danach.Sitze[1].Leben, vorher)
		}
	})

	t.Run("ein Blocker, der weg ist, teilt nichts zu", func(t *testing.T) {
		t.Parallel()

		partie := mitBlock(t)

		// **Der Bär ist nach dem Schaden tot** — und er hat trotzdem
		// ausgeteilt (CR 510.2). Der Riese trägt zwei.
		if riese := objektMit(t, partie,
			"sitz-1/riese/0"); riese.Schaden != 2 {
			t.Errorf("der Riese trägt %d Schaden", riese.Schaden)
		}
	})

	t.Run("ein Stern als Stärke teilt nichts zu", func(t *testing.T) {
		t.Parallel()

		partie := mitSternstaerke(t)

		vorher := partie.Sitze[1].Leben

		danach := bisZumSchritt(t, partie, domain.SchrittKampfschaden)

		if danach.Sitze[1].Leben != vorher {
			t.Errorf("sitz-2 steht bei %d statt %d — ein Stern hat "+
				"Schaden ausgeteilt", danach.Sitze[1].Leben, vorher)
		}
	})

	t.Run("ein Planeswalker als Ziel nimmt den Schaden", func(t *testing.T) {
		t.Parallel()

		partie := gegenEinenPlaneswalker(t)

		danach := bisZumSchritt(t, partie, domain.SchrittKampfschaden)

		if walker := objektMit(t, danach,
			"sitz-2/walker/0"); walker.Schaden != 3 {
			t.Errorf("der Planeswalker trägt %d Schaden", walker.Schaden)
		}
	})
}

// ohneStaerke gibt einen Angriff mit einer Kreatur ohne Stärkefeld.
func ohneStaerke(t *testing.T) domain.Partie {
	t.Helper()

	return angriffMitFeldern(t, map[string]string{
		"super type": "Creature", "sub type": "Giant", "toughness": "3",
	})
}

// mitSternstaerke gibt einen Angriff mit `*` als Stärke — CR 208.2.
func mitSternstaerke(t *testing.T) domain.Partie {
	t.Helper()

	return angriffMitFeldern(t, map[string]string{
		"super type": "Creature", "sub type": "Giant",
		"power": "*", "toughness": "3",
	})
}

// angriffMitFeldern baut einen Angriff mit diesen Kartenfeldern.
func angriffMitFeldern(t *testing.T, felder map[string]string) domain.Partie {
	t.Helper()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	karten := make([]domain.Abzugskarte, len(sitze[0].Deck.Karten))
	copy(karten, sitze[0].Deck.Karten)
	karten[0].Seiten = []domain.Abzugsseite{{Felder: felder}}
	sitze[0].Deck.Karten = karten
	partie.Sitze = sitze

	partie, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	partie, fehler = bisZumSchritt(t, partie, domain.SchrittBlocker).
		BlockerErklaeren("sitz-2", nil)
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	return partie
}

// gegenEinenPlaneswalker greift einen Planeswalker an.
func gegenEinenPlaneswalker(t *testing.T) domain.Partie {
	t.Helper()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[1].Deck.Karten = append(sitze[1].Deck.Karten, domain.Abzugskarte{
		Kennung: "walker", Name: "Wanderer", Menge: 1,
		Seiten: []domain.Abzugsseite{{Felder: map[string]string{
			"super type": "Legendary Planeswalker", "sub type": "Jace",
			"loyalty": "4",
		}}},
	})
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/walker/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "walker",
	})

	partie, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{
			Kreatur: "sitz-1/riese/0", Ziel: "sitz-2/walker/0",
		}})
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	partie, fehler = bisZumSchritt(t, partie, domain.SchrittBlocker).
		BlockerErklaeren("sitz-2", nil)
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	return partie
}
