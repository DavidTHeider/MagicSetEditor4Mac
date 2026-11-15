package domain_test

// **Die Commandervariante** — Werkstück M19 der Regelmaschine
// (#1620).
//
// # Vier Regeln, und jede hat eine Nummer
//
//   - **Vierzig Leben** — CR 903.7 (nicht CR 103.4: das ist der Grundfall
//     mit zwanzig).
//   - **Die Steuer** — CR 903.8: zwei generisches Mana mehr je vorherigem
//     Wirken aus der Kommandozone.
//   - **Einundzwanzig Schaden** — CR 903.10a, und zwar **Kampfschaden**
//     desselben Commanders über die ganze Partie.
//   - **Der Weg zurück** — CR 903.9a: der Eigner **darf** ihn aus Friedhof
//     oder Exil in die Kommandozone legen. Eine Wahl, wie CR 704.5j.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Commanderpartie fängt mit vierzig Leben an** — CR 903.7.
func TestEineCommanderpartieFaengtMitVierzigLebenAn(t *testing.T) {
	t.Parallel()

	profil := domain.Commanderprofil("magic")

	if profil.Startleben != 40 {
		t.Errorf("das Commanderprofil nennt %d Leben", profil.Startleben)
	}

	partie := domain.Partie{
		Spiel: "magic", Profil: profil,
		Regelwerk: domain.Regelwerk{
			Kennung: "commander", Fassung: domain.Commanderfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Wirt: "sitz-1",
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "David"},
			{Kennung: "sitz-2", Name: "Gegenüber"},
		},
	}.MitStartleben()

	for _, sitz := range partie.Sitze {
		if sitz.Leben != 40 {
			t.Errorf("%q fängt mit %d Leben an", sitz.Kennung, sitz.Leben)
		}
	}
}

// **Die Steuer steigt mit jedem Wirken aus der Kommandozone** — CR 903.8.
func TestDieCommandersteuerSteigt(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		vorher   int
		erwartet int
	}{
		"das erste Mal":  {vorher: 0, erwartet: 0},
		"das zweite Mal": {vorher: 1, erwartet: 2},
		"das dritte Mal": {vorher: 2, erwartet: 4},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if steuer := domain.Commandersteuer(
				fall.vorher); steuer != fall.erwartet {
				t.Errorf("die Steuer ist %d", steuer)
			}
		})
	}
}

// **Einundzwanzig Kampfschaden desselben Commanders beenden die Partie** —
// CR 903.10a.
func TestEinundzwanzigCommanderschadenBeendenDiePartie(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)
	partie.Sitze[1].Commanderschaden = map[string]int{"sitz-1/held/0": 21}

	danach := partie.Zustandspruefung()

	if len(danach.Ausgang.Ausgeschieden) == 0 {
		t.Fatalf("die Partie läuft weiter: %+v", danach.Ausgang)
	}

	if danach.Ausgang.Sieger != "sitz-1" {
		t.Errorf("gewonnen hat %q", danach.Ausgang.Sieger)
	}
}

// **Zwanzig nicht** — die Gegenprobe.
//
// **Und zwei verschiedene Commander mit je zwanzig auch nicht** (CR 903.10a
// sagt „desselben"): wer sie zusammenzählte, ließe eine Partie enden, die
// weiterläuft.
func TestZwanzigCommanderschadenBeendenNichts(t *testing.T) {
	t.Parallel()

	faelle := map[string]map[string]int{
		"zwanzig von einem": {"sitz-1/held/0": 20},
		"zwanzig von zwei":  {"sitz-1/held/0": 20, "sitz-1/held/1": 20},
	}

	for name, schaden := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := eineCommanderpartie(t)
			partie.Sitze[1].Commanderschaden = schaden

			danach := partie.Zustandspruefung()
			if len(danach.Ausgang.Ausgeschieden) != 0 {
				t.Errorf("die Partie ist zu Ende: %+v", danach.Ausgang)
			}
		})
	}
}

// **Und ohne Commanderprofil zählt gar nichts** — die Regel gehört der
// Variante (CR 903), nicht dem Grundspiel.
func TestOhneCommanderprofilZaehltDerSchadenNicht(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Sitze[1].Commanderschaden = map[string]int{"sitz-1/held/0": 21}

	if danach := partie.Zustandspruefung(); len(
		danach.Ausgang.Ausgeschieden) != 0 {
		t.Errorf("die Partie ist zu Ende: %+v", danach.Ausgang)
	}
}

// **Der Commander darf zurück in die Kommandozone** — CR 903.9a.
func TestEinCommanderDarfZurueckInDieKommandozone(t *testing.T) {
	t.Parallel()

	partie := eineCommanderpartie(t)

	gestorben, fehler := partie.Zerstoeren("sitz-1/held/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	danach, fehler := gestorben.Heimholen("sitz-1", "sitz-1/held/0")
	if fehler != nil {
		t.Fatalf("heimholen: %v", fehler)
	}

	if zone := objektMit(t, danach,
		"sitz-1/held/0").Zone; zone != domain.ZoneKommando {
		t.Errorf("der Commander liegt in %q", zone)
	}
}

// **Was kein Heimholen ist.**
func TestWasKeinHeimholenIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		wer, objekt string
		vorher      func(*domain.Partie)
		erwartet    error
	}{
		"ein Objekt, das es nicht gibt": {
			wer: "sitz-1", objekt: "sitz-1/greif/0",
			erwartet: domain.ErrNotFound,
		},
		"was kein Commander ist": {
			wer: "sitz-1", objekt: "sitz-1/baer/0",
			vorher: func(p *domain.Partie) {
				p.Objekte[1].Zone = domain.ZoneFriedhof
			},
			erwartet: domain.ErrValidation,
		},
		"was noch auf dem Tisch liegt": {
			wer: "sitz-1", objekt: "sitz-1/held/0",
			erwartet: domain.ErrValidation,
		},
		"wem er nicht gehört": {
			wer: "sitz-2", objekt: "sitz-1/held/0",
			vorher: func(p *domain.Partie) {
				p.Objekte[0].Zone = domain.ZoneFriedhof
			},
			erwartet: domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := eineCommanderpartie(t)
			if fall.vorher != nil {
				fall.vorher(&partie)
			}

			_, fehler := partie.Heimholen(fall.wer, fall.objekt)
			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("heimholen: %v", fehler)
			}
		})
	}
}

// eineCommanderpartie gibt eine Partie nach den Regeln von CR 903.
//
// **Zwei Sitze, jeder mit einem Commander** — CR 903.3 nennt ihn eine
// legendäre Kreaturenkarte; welche es ist, sagt der Sitz.
func eineCommanderpartie(t *testing.T) domain.Partie {
	t.Helper()

	deck := domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{
			{
				Kennung: "held", Name: "Held", Menge: 1,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"name": "Held", "super type": "Legendary Creature",
					"casting cost": "2R", "power": "3", "toughness": "3",
				}}},
			},
			{
				Kennung: "baer", Name: "Bär", Menge: 1,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"name": "Bär", "super type": "Creature",
					"power": "2", "toughness": "2",
				}}},
			},
		},
	}

	partie := domain.Partie{
		Spiel:  "magic",
		Profil: domain.Commanderprofil("magic"),
		Regelwerk: domain.Regelwerk{
			Kennung: "commander", Fassung: domain.Commanderfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Wirt: "sitz-1",
		Sitze: []domain.Sitz{
			{
				Kennung: "sitz-1", Name: "David", Deck: deck,
				Commander: "held",
			},
			{
				Kennung: "sitz-2", Name: "Gegenüber", Deck: deck,
				Commander: "held",
			},
		},
		Objekte: []domain.Spielobjekt{
			{
				Kennung: "sitz-1/held/0", Art: domain.ObjektKarte,
				Beherrscher: "sitz-1", Eigner: "sitz-1",
				Zone: domain.ZoneTisch, Karte: "held",
			},
			{
				Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
				Beherrscher: "sitz-1", Eigner: "sitz-1",
				Zone: domain.ZoneTisch, Karte: "baer",
			},
		},
	}.MitStartleben()

	return partie
}

// **Im Commanderprofil ist der Bereich keine Lücke** — und im Grundprofil
// ist er eine.
//
// **Das ist die Aussage des Profils**: eine Partie ohne Commandervariante
// sagt über sie nichts, und `unbekannt` ist die ehrliche Antwort (G4).
func TestDerCommanderbereichStehtNurImCommanderprofil(t *testing.T) {
	t.Parallel()

	commander := domain.Commanderprofil("magic").Matrix
	if grad := commander.Grad(
		domain.BereichCommander); grad != domain.UnterstuetzungAutomatisch {
		t.Errorf("im Commanderprofil steht %q", grad)
	}

	grund := domain.Grundprofil("magic").Matrix
	if grad := grund.Grad(
		domain.BereichCommander); grad != domain.UnterstuetzungManuell {
		t.Errorf("im Grundprofil steht %q", grad)
	}

	// **Und es nennt seine Grenzen** — ein automatischer Bereich ohne sie
	// behauptet Vollständigkeit (§10).
	eintrag, da := commander.Eintrag(domain.BereichCommander)
	if !da || len(eintrag.Grenzen) == 0 {
		t.Errorf("der Commanderbereich nennt keine Grenze: %+v", eintrag)
	}
}

// **Welches Profil eine Kennung gibt** — CR 903 gegen den Grundfall.
func TestWelchesProfilEineKennungGibt(t *testing.T) {
	t.Parallel()

	faelle := map[string]int{
		domain.Commanderkennung: domain.Commanderleben,
		"sandkasten":            20,
		"":                      20,
		"erfunden":              20,
	}

	for kennung, leben := range faelle {
		t.Run(kennung, func(t *testing.T) {
			t.Parallel()

			// **Was niemand kennt, wird zum Sandkasten** — ein Profil, das
			// niemand kennt, setzt nichts durch, und das sagt die Matrix.
			if gegeben := domain.ProfilZu(
				kennung, "magic").Startleben; gegeben != leben {
				t.Errorf("%q gibt %d Leben", kennung, gegeben)
			}
		})
	}
}
