package domain_test

// **Die Liste der eigenen Partien** —
// #1600.
//
// # Wogegen hier geprüft wird
//
// **Dass eine Liste eine Liste ist und keine Sicht.** Wer sie liest, will
// wissen, **ob** es eine Partie gibt und welche — nicht, was darin liegt. Ein
// Eintrag, der Objekte trüge, wäre eine zweite Projektion neben
// [domain.Partie.Sicht], und sie müsste dieselben Zusagen halten (G1).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Zeile nennt, woran man eine Partie wiedererkennt.**
func TestEineZeileNenntWoranManDiePartieErkennt(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.ID = "p1"
	partie.Fassung = 7
	partie.Stand = domain.StandPausiert
	partie.Zug = domain.Zugstand{Nummer: 3, Aktiv: "sitz-2"}

	zeile := partie.Zeile()

	switch {
	case zeile.ID != "p1":
		t.Errorf("die Kennung ist %q", zeile.ID)
	case zeile.Spiel != "magic":
		t.Errorf("das Spiel ist %q", zeile.Spiel)
	case zeile.Stand != domain.StandPausiert:
		t.Errorf("der Stand ist %q", zeile.Stand)
	case zeile.Fassung != 7:
		t.Errorf("die Fassung ist %d", zeile.Fassung)
	case zeile.Zug != 3:
		t.Errorf("der Zug ist %d", zeile.Zug)
	case len(zeile.Sitze) != 2 || zeile.Sitze[0].Name != "David":
		t.Errorf("die Sitze sind %v", zeile.Sitze)
	}
}

// **Und sie trägt kein einziges Objekt.**
//
// **Das ist die Zusage, an der hier am meisten hängt** (G1): die Bibliothek
// ist auch für ihren Eigner verdeckt, und eine Liste, die Objekte mitgäbe,
// umginge die Projektion, statt sie zu benutzen.
//
// **Geprüft über die Gestalt und nicht über einen Blick**: der Verbund hat
// die Felder, die er hat, und ein neues Feld mit Kartennamen fiele hier auf.
func TestEineZeileTraegtKeineObjekte(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Objekte = []domain.Spielobjekt{{
		Kennung: "sitz-1/k1/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneBibliothek, Karte: "wald", Name: "Wald",
	}}

	zeile := partie.Zeile()

	for _, wort := range []string{"wald", "Wald", "sitz-1/k1/0"} {
		for _, sitz := range zeile.Sitze {
			if sitz.Name == wort || sitz.Kennung == wort {
				t.Errorf("die Zeile nennt %q", wort)
			}
		}
	}

	// **Die Kartenzahl steht dabei** — sie verrät nichts über die Karten und
	// sagt einem Menschen, wie weit die Partie ist.
	if zeile.Objekte != 1 {
		t.Errorf("die Zeile zählt %d Objekte", zeile.Objekte)
	}
}

// **Eine Partie ohne Sitze gibt eine leere Liste und kein `nil`.**
func TestEineZeileOhneSitzeGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Sitze = nil

	if sitze := partie.Zeile().Sitze; sitze == nil || len(sitze) != 0 {
		t.Errorf("die Sitze sind %#v", sitze)
	}
}

// **Eine Zeile kennt kein Profil, und ihr eigener Sitz bleibt leer.**
//
// **Weil `Zeile` nur die Partie sieht** — welches Profil fragt, weiß sie nicht,
// und ein Sitz, der ohne Frage „der eigene“ hieße, wäre geraten. Befüllt wird
// er dort, wo das Profil bekannt ist (UXR-F16,
// #2031).
func TestEineZeileKenntKeinProfilUndNenntKeinenEigenenSitz(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Sitze[0].ProfilID = 7

	zeile := partie.Zeile()

	if zeile.EigenerSitz != "" || zeile.Dran {
		t.Errorf("die Zeile nennt %q und dran=%v", zeile.EigenerSitz, zeile.Dran)
	}
}

// **Ein Profil sitzt dort, wo ein Sitz an es gebunden ist — und sonst nirgends.**
//
// # Was hier die Regel ist
//
// **Die Bindung und nichts sonst** (G8,
// #1465):
// ein Sitz ohne Profil gehört niemandem, auch nicht dem, der die Partie
// angelegt hat. Die Null heißt „keines“ und darf deshalb nie einen
// ungebundenen Sitz finden — beide tragen sie.
func TestEinProfilSitztDortWoEinSitzAnEsGebundenIst(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		bindung [2]int64
		profil  int64
		sitzt   []string
	}{
		"ein gebundener Sitz":             {[2]int64{0, 7}, 7, []string{"sitz-2"}},
		"ein anderes Profil":              {[2]int64{0, 7}, 9, []string{}},
		"die Null sitzt nirgends":         {[2]int64{0, 7}, 0, []string{}},
		"zwei Sitze in Tischreihenfolge":  {[2]int64{7, 7}, 7, []string{"sitz-1", "sitz-2"}},
		"ohne jede Bindung ist keiner da": {[2]int64{0, 0}, 7, []string{}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := einePartie()
			partie.Sitze[0].ProfilID = fall.bindung[0]
			partie.Sitze[1].ProfilID = fall.bindung[1]

			sitzt := partie.SitzeVonProfil(fall.profil)

			if sitzt == nil || len(sitzt) != len(fall.sitzt) {
				t.Fatalf("es sitzen %#v, erwartet %#v", sitzt, fall.sitzt)
			}

			for stelle, kennung := range fall.sitzt {
				if sitzt[stelle] != kennung {
					t.Errorf("an Stelle %d sitzt %q, erwartet %q",
						stelle, sitzt[stelle], kennung)
				}
			}
		})
	}
}
