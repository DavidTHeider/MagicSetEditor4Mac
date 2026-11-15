package domain_test

// **Ein Objekt wird zur Kopie eines anderen** — CR 707, Werkstück M18
// (#1619).
//
// # Was hier kopiert wird und was nicht
//
// **CR 707.2**: die Kopie nimmt die **kopierbaren Werte** der Merkmale des
// Originals an. In diesem Programm sind das die **gedruckten** Werte — eine
// Karte hat keine anderen, solange kein Effekt sie ändert. Was ein Effekt
// daraus machte, ist `M10`
// (#1612);
// die Wahlen eines Zaubers auf dem Stapel (Modus, X) sind `M8`.
//
// **CR 707.3**: der Zustand der Kopie bleibt ihrer — getappt bleibt getappt,
// und die Marken liegen weiter darauf. Das ist der Unterschied zum
// Zonenwechsel nebenan.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Kopie nimmt die Merkmale des Originals an** — CR 707.2.
func TestEineKopieNimmtDieMerkmaleAn(t *testing.T) {
	t.Parallel()

	danach, fehler := eineGestaltenpartie(t).Kopieren(
		"sitz-1/baer/0", "sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	eigen, da := danach.Eigenschaften(objektMit(t, danach, "sitz-1/baer/0"))
	if !da {
		t.Fatal("die Kopie hat keine Eigenschaften")
	}

	if eigen.Name != "Wandler" {
		t.Errorf("die Kopie heißt %q", eigen.Name)
	}
}

// **Und ihr Zustand bleibt ihrer** — CR 707.3 mit CR 110.5.
func TestEineKopieBehaeltIhrenZustand(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[1].Zaehler = map[string]int{domain.MarkePlusEins: 2}

	danach, fehler := partie.Kopieren("sitz-1/baer/0", "sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	baer := objektMit(t, danach, "sitz-1/baer/0")

	// **Der Bär war getappt** — die Vorrichtung legt ihn so hin.
	if !baer.Getappt {
		t.Error("die Kopie ist enttappt worden")
	}

	if baer.Zaehler[domain.MarkePlusEins] != 2 {
		t.Errorf("auf der Kopie liegen %d Marken",
			baer.Zaehler[domain.MarkePlusEins])
	}

	// **Und die Marken gelten für die neuen Werte** — CR 613.4c: eine 1/1
	// mit zwei Marken ist eine 3/3.
	staerke, _ := danach.Werte(baer)
	if staerke.Zahl != 3 {
		t.Errorf("die Kopie hat Stärke %d", staerke.Zahl)
	}
}

// **Eine Kopie eines Spielsteins nimmt seine Felder** — CR 111.1 mit
// CR 707.2.
func TestEineKopieEinesSpielsteinsNimmtSeineFelder(t *testing.T) {
	t.Parallel()

	partie, fehler := eineGestaltenpartie(t).Spielstein(
		"sitz-1", "sitz-1/soldat/0", einSpielsteinfeld())
	if fehler != nil {
		t.Fatalf("spielstein: %v", fehler)
	}

	danach, fehler := partie.Kopieren("sitz-1/baer/0", "sitz-1/soldat/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	eigen, _ := danach.Eigenschaften(objektMit(t, danach, "sitz-1/baer/0"))
	if eigen.Name != "Soldat" {
		t.Errorf("die Kopie heißt %q", eigen.Name)
	}
}

// **Und die Felder der Vorlage bleiben ihr** — eine Tabelle, die beide
// teilten, machte aus zwei Objekten eines.
func TestEineKopieTeiltDieFelderNicht(t *testing.T) {
	t.Parallel()

	partie, fehler := eineGestaltenpartie(t).Spielstein(
		"sitz-1", "sitz-1/soldat/0", einSpielsteinfeld())
	if fehler != nil {
		t.Fatalf("spielstein: %v", fehler)
	}

	danach, fehler := partie.Kopieren("sitz-1/baer/0", "sitz-1/soldat/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	for stelle := range danach.Objekte {
		if danach.Objekte[stelle].Kennung == "sitz-1/baer/0" {
			danach.Objekte[stelle].Felder["power"] = "9"
		}
	}

	eigen, _ := danach.Eigenschaften(objektMit(t, danach, "sitz-1/soldat/0"))
	if eigen.Staerke.Zahl != 1 {
		t.Errorf("die Vorlage hat Stärke %d", eigen.Staerke.Zahl)
	}
}

// **Was keine Kopie ist.**
func TestWasKeineKopieIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		objekt, vorlage string
		erwartet        error
	}{
		"ein Objekt, das es nicht gibt": {
			objekt: "sitz-1/greif/0", vorlage: "sitz-1/baer/0",
			erwartet: domain.ErrNotFound,
		},
		"eine Vorlage, die es nicht gibt": {
			objekt: "sitz-1/baer/0", vorlage: "sitz-1/greif/0",
			erwartet: domain.ErrNotFound,
		},
		"sich selbst": {
			objekt: "sitz-1/baer/0", vorlage: "sitz-1/baer/0",
			erwartet: domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := eineGestaltenpartie(t).Kopieren(
				fall.objekt, fall.vorlage)
			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("kopieren: %v", fehler)
			}
		})
	}
}

// **Der Eigner ändert sich nicht** — CR 108.3: er ist, wer mit der Karte
// angefangen hat, und das Kopieren ändert daran nichts.
//
// **Warum das zählt**: er entscheidet, in wessen Friedhof die Kopie beim
// Sterben geht (CR 404.1).
func TestEineKopieBehaeltIhrenEigner(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[1].Eigner = "sitz-2"
	partie.Objekte[1].Karte = ""
	partie.Objekte[1].Felder = map[string]string{"name": "Fremd"}

	danach, fehler := partie.Kopieren("sitz-1/wandler/0", "sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if eigner := objektMit(t, danach,
		"sitz-1/wandler/0").Eigner; eigner != "sitz-1" {
		t.Errorf("die Kopie gehört jetzt %q", eigner)
	}
}

// **Und eine Vorlage ohne bekannte Felder wird abgewiesen** — §8: ein Verweis
// ins Leere ist ein Befund und kein Absturz.
func TestEineVorlageOhneFelderWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[1].Karte = "gibtesnicht"

	_, fehler := partie.Kopieren("sitz-1/wandler/0", "sitz-1/baer/0")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("eine Vorlage ohne Felder: %v", fehler)
	}
}

// **Eine Kopie endet mit der Zone** — CR 707.4 mit CR 400.7.
func TestEineKopieEndetMitDerZone(t *testing.T) {
	t.Parallel()

	kopiert, fehler := eineGestaltenpartie(t).Kopieren(
		"sitz-1/baer/0", "sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	danach, fehler := kopiert.Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	eigen, _ := danach.Eigenschaften(objektMit(t, danach, "sitz-1/baer/0"))
	if eigen.Name != "Bär" {
		t.Errorf("im Friedhof liegt %q", eigen.Name)
	}
}

// **Eine Vorlage einer fremden Seite fällt auf die Vorderseite zurück** —
// CR 712.8a, und derselbe Rückfall wie bei den Eigenschaften (§8).
func TestEineVorlageMitUnbekannterSeiteNimmtDieVorderseite(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[1].Seite = 9

	danach, fehler := partie.Kopieren("sitz-1/wandler/0", "sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	eigen, _ := danach.Eigenschaften(objektMit(t, danach, "sitz-1/wandler/0"))
	if eigen.Name != "Bär" {
		t.Errorf("kopiert wurde %q", eigen.Name)
	}
}

// **Und eine Karte ohne Felder gibt keine Tabelle** — eine leere wäre eine
// Aussage, und es gibt keine.
func TestEineVorlageOhneFelderGibtKeineTabelle(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Sitze[0].Deck.Karten[1].Seiten = []domain.Abzugsseite{{}}

	danach, fehler := partie.Kopieren("sitz-1/wandler/0", "sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if felder := objektMit(t, danach,
		"sitz-1/wandler/0").Felder; felder != nil {
		t.Errorf("die Kopie trägt %v", felder)
	}
}

// **Und eine Vorlage eines anderen Sitzes wird dort gesucht.**
func TestEineVorlageEinesAnderenSitzesWirdDortGesucht(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Sitze[1].Deck = partie.Sitze[0].Deck
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "baer",
	})

	danach, fehler := partie.Kopieren("sitz-1/wandler/0", "sitz-2/baer/0")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	eigen, _ := danach.Eigenschaften(objektMit(t, danach, "sitz-1/wandler/0"))
	if eigen.Name != "Bär" {
		t.Errorf("kopiert wurde %q", eigen.Name)
	}
}
