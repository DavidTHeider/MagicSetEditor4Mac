package domain_test

// **Ein Haufen ist eine Anordnung und keine Zone** —
// #1756,
// PLAY-03.AK-03 und AK-04
// (#1711).
//
// # Was CR 400.5 erlaubt
//
// In einer Zone, deren Reihenfolge nichts bedeutet, darf ein Spieler seine
// Objekte anordnen, wie er will. **Genau das ist ein Haufen** — sieben
// Wälder übereinander sind keine Unterzone des Schlachtfelds, sondern eine
// Geste am Tisch.
//
// # Und was AK-04 verlangt
//
// *„Ein zusammengefasster Stapel darf individuelle Ziele oder getappte
// Zustände nicht verschleiern."*
//
// **Das wird hier nicht versprochen, sondern durch die Bedingung
// ausgeschlossen**: ein Haufen entsteht nur aus Objekten, die in **jedem
// gezeichneten Zustand** gleich sind. Dann kann er nichts verschleiern,
// statt dass eine Prüfung es hinterher behauptet (§10).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiWaelder gibt eine Partie mit zwei gleichen Ländern auf dem Tisch.
func zweiWaelder() domain.Partie {
	partie := einePartie()

	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "w1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: "wald",
		},
		{
			Kennung: "w2", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: "wald",
		},
	}

	return partie
}

// haufenVon gibt den Haufen eines Objekts.
func haufenVon(t *testing.T, partie domain.Partie, kennung string) string {
	t.Helper()

	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt.Haufen
		}
	}

	t.Fatalf("%q liegt nicht am Tisch", kennung)

	return ""
}

func TestZweiGleicheKartenBildenEinenHaufen(t *testing.T) {
	danach, fehler := zweiWaelder().Haeufe("w1", "w2")
	if fehler != nil {
		t.Fatalf("häufen: %v", fehler)
	}

	// **Der Haufen trägt die Kennung dessen, auf den gelegt wurde.** Eine
	// eigene Kennung wäre eine zweite Größe, die niemand vergibt.
	if haufenVon(t, danach, "w1") != "w2" {
		t.Errorf("w1 liegt im Haufen %q", haufenVon(t, danach, "w1"))
	}

	if haufenVon(t, danach, "w2") != "w2" {
		t.Errorf("w2 liegt im Haufen %q", haufenVon(t, danach, "w2"))
	}
}

// **Ein Dritter kommt auf denselben Haufen und nicht auf einen zweiten.**
func TestEinDritterKommtAufDenselbenHaufen(t *testing.T) {
	partie := zweiWaelder()
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "w3", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
		Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: "wald",
	})

	danach, fehler := partie.Haeufe("w1", "w2")
	if fehler != nil {
		t.Fatalf("häufen: %v", fehler)
	}

	danach, fehler = danach.Haeufe("w3", "w1")
	if fehler != nil {
		t.Fatalf("den dritten häufen: %v", fehler)
	}

	if haufenVon(t, danach, "w3") != "w2" {
		t.Errorf("w3 liegt im Haufen %q statt in w2",
			haufenVon(t, danach, "w3"))
	}
}

func TestEnthaeufenNimmtEinesHeraus(t *testing.T) {
	danach, fehler := zweiWaelder().Haeufe("w1", "w2")
	if fehler != nil {
		t.Fatalf("häufen: %v", fehler)
	}

	danach, fehler = danach.Enthaeufe("w1")
	if fehler != nil {
		t.Fatalf("enthäufen: %v", fehler)
	}

	if haufenVon(t, danach, "w1") != "" {
		t.Errorf("w1 liegt noch im Haufen %q", haufenVon(t, danach, "w1"))
	}

	// **Der Rest bleibt liegen.** Wer eine Karte herausnimmt, löst nicht den
	// Haufen auf — die anderen liegen weiter beieinander.
	if haufenVon(t, danach, "w2") != "w2" {
		t.Errorf("w2 wurde mit herausgenommen")
	}
}

// **Die Bedingung, die AK-04 einlöst.** Jeder Fall hier ist ein Zustand, den
// ein Haufen verschleiern könnte — und deshalb entsteht er gar nicht erst.
func TestEinHaufenEntstehtNurAusGleichemZustand(t *testing.T) {
	faelle := []struct {
		name   string
		aendre func(*domain.Spielobjekt)
	}{
		{"einer ist getappt", func(o *domain.Spielobjekt) {
			o.Getappt = true
		}},
		{"einer ist verdeckt", func(o *domain.Spielobjekt) {
			o.Verdeckt = true
		}},
		{"einer trägt Marken", func(o *domain.Spielobjekt) {
			o.Zaehler = map[string]int{"plusEinsPlusEins": 1}
		}},
		{"einer hängt an etwas", func(o *domain.Spielobjekt) {
			o.HaengtAn = "w2"
		}},
		{"einer liegt woanders", func(o *domain.Spielobjekt) {
			o.Zone = domain.ZoneHand
		}},
		{"einer gehört einem anderen", func(o *domain.Spielobjekt) {
			o.Beherrscher = "sitz-2"
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			partie := zweiWaelder()
			fall.aendre(&partie.Objekte[0])

			_, fehler := partie.Haeufe("w1", "w2")
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("der Haufen entstand trotzdem: %v", fehler)
			}
		})
	}
}

// **Und in einer geordneten Zone gar nicht** — CR 400.5 gibt die freie
// Anordnung nur dort, wo die Reihenfolge nichts bedeutet. Ein Haufen in der
// Bibliothek wäre eine Anordnung über einer, die schon gilt.
func TestInEinerGeordnetenZoneGibtEsKeinenHaufen(t *testing.T) {
	partie := zweiWaelder()
	partie.Objekte[0].Zone = domain.ZoneBibliothek
	partie.Objekte[1].Zone = domain.ZoneBibliothek

	_, fehler := partie.Haeufe("w1", "w2")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Haufen entstand in der Bibliothek: %v", fehler)
	}
}

// **Auf sich selbst legt niemand etwas.**
func TestEinObjektHaeuftSichNichtAufSichSelbst(t *testing.T) {
	_, fehler := zweiWaelder().Haeufe("w1", "w1")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("w1 liegt auf sich selbst: %v", fehler)
	}
}

// **Ein Objekt, das es nicht gibt, ist ein Befund und kein Haufen** (§8).
func TestEinFremdesObjektLaesstSichNichtHaeufen(t *testing.T) {
	for _, fall := range []struct{ wen, worauf string }{
		{"gibtesnicht", "w2"},
		{"w1", "gibtesnicht"},
	} {
		_, fehler := zweiWaelder().Haeufe(fall.wen, fall.worauf)
		if fehler == nil {
			t.Errorf("%q auf %q ging durch", fall.wen, fall.worauf)
		}
	}

	if _, fehler := zweiWaelder().Enthaeufe("gibtesnicht"); fehler == nil {
		t.Error("ein fremdes Objekt ließ sich enthäufen")
	}
}

// **Was nicht in einem Haufen liegt, lässt sich nicht herausnehmen.** Ein
// stilles Nichts wäre eine Antwort, die aussieht wie ein Erfolg.
func TestWasInKeinemHaufenLiegtLaesstSichNichtEnthaeufen(t *testing.T) {
	_, fehler := zweiWaelder().Enthaeufe("w1")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("w1 ließ sich aus keinem Haufen nehmen: %v", fehler)
	}
}
