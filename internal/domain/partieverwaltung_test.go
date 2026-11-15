package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineVerwaltbarePartie gibt eine laufende Partie mit einem Objekt.
func eineVerwaltbarePartie() domain.Partie {
	return domain.Partie{
		ID:      "p1",
		Fassung: 7,
		Spiel:   "magic",
		Stand:   domain.StandLaeuft,
		Wirt:    "sitz1",
		Sitze:   []domain.Sitz{{Kennung: "sitz1", Name: "Anna", Leben: 12}},
		Objekte: []domain.Spielobjekt{{Kennung: "k1", Zone: domain.ZoneTisch}},
		Zug:     domain.Zugstand{Nummer: 4},
	}
}

func TestLoeschbarNachDerRegel(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name         string
		stand        domain.Partiestand
		offenlegunge int
		erwartet     bool
	}{
		{
			// **Eine laufende Partie ohne Offenlegung ist ein Entwurf.**
			name: "laufend und unbesehen", stand: domain.StandLaeuft,
			erwartet: true,
		},
		{
			// **Eine pausierte auch** — sie ist nicht beendet.
			name: "pausiert", stand: domain.StandPausiert, erwartet: true,
		},
		{
			// **Eine beendete bleibt** — AK-05, letzter Satz:
			// „Abgeschlossene gemeinsame Partien bleiben nachvollziehbar."
			// Was mit ihr geschieht, entscheidet die Datenpflege nach ihrer
			// Frist und nicht ein Knopf.
			name: "beendet", stand: domain.StandBeendet, erwartet: false,
		},
		{
			// **Und eine, für die jemand eine Einwilligung gegeben hat,
			// ebenso**: sie ist kein Entwurf mehr, sondern eine Zusage.
			name: "mit Offenlegung", stand: domain.StandLaeuft,
			offenlegunge: 1, erwartet: false,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			partie := eineVerwaltbarePartie()
			partie.Stand = fall.stand

			if domain.Loeschbar(partie, fall.offenlegunge) != fall.erwartet {
				t.Errorf("löschbar ist %t, erwartet %t",
					!fall.erwartet, fall.erwartet)
			}
		})
	}
}

func TestKopiereGibtEinenUebungszustand(t *testing.T) {
	t.Parallel()

	vorlage := eineVerwaltbarePartie()
	vorlage.Ausgang = domain.Partieausgang{Sieger: "sitz1"}
	vorlage.Stand = domain.StandBeendet

	kopie := domain.Kopiere(vorlage, "p2")

	// **Eine neue Partie und keine Fortsetzung.**
	if kopie.ID != "p2" || kopie.Fassung != 1 {
		t.Errorf("Kennung %q, Fassung %d", kopie.ID, kopie.Fassung)
	}

	// **Sie läuft** — auch aus einer beendeten heraus: das ist der Sinn eines
	// Übungszustands.
	if kopie.Stand != domain.StandLaeuft {
		t.Errorf("Stand %q", kopie.Stand)
	}

	// **Und sie ist nicht entschieden.**
	if kopie.Ausgang.Sieger != "" || kopie.Ausgang.Unentschieden {
		t.Errorf("Ausgang %+v", kopie.Ausgang)
	}

	// **Die Stellung bleibt** — Karten, Sitze, Leben, Zug.
	if len(kopie.Objekte) != 1 || kopie.Objekte[0].Kennung != "k1" {
		t.Errorf("Objekte %+v", kopie.Objekte)
	}

	if len(kopie.Sitze) != 1 || kopie.Sitze[0].Leben != 12 {
		t.Errorf("Sitze %+v", kopie.Sitze)
	}

	if kopie.Zug.Nummer != 4 {
		t.Errorf("Zug %d", kopie.Zug.Nummer)
	}
}

// **Die Kopie hängt nicht an der Vorlage.** Ein geteiltes Slice machte aus
// zwei Partien eine — wer in der Kopie zöge, änderte das Original.
func TestKopiereTeiltKeineListe(t *testing.T) {
	t.Parallel()

	vorlage := eineVerwaltbarePartie()
	kopie := domain.Kopiere(vorlage, "p2")

	kopie.Objekte[0].Zone = domain.ZoneFriedhof
	kopie.Sitze[0].Leben = 1

	if vorlage.Objekte[0].Zone != domain.ZoneTisch {
		t.Error("die Vorlage hat ihr Objekt verloren")
	}

	if vorlage.Sitze[0].Leben != 12 {
		t.Error("die Vorlage hat ihren Lebensstand verloren")
	}
}
