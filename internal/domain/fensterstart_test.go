// **Wie das Fenster aufgeht** — UX-14.AK-01
// (#1686).
package domain_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Beim ersten Start ist das Fenster maximiert** — das Kriterium sagt es
// wörtlich, und es sagt ebenso wörtlich: *„nicht automatisch im
// macOS-Vollbildmodus."* Ein Vollbild nimmt die Menüleiste mit, und eine
// Anwendung, die sich beim ersten Start den Bildschirm nimmt, hat nicht
// gefragt.
func TestDerFensterstartIstVorgemerktMaximiert(t *testing.T) {
	t.Parallel()

	if !slices.Contains(domain.Einstellungsreihe(), domain.EinstellungFensterstart) {
		t.Fatalf("%q steht nicht in der Einstellungsreihe",
			domain.EinstellungFensterstart)
	}

	if vorgabe := domain.EinstellungsVorgabe(domain.EinstellungFensterstart); vorgabe != domain.FensterstartMaximiert {
		t.Errorf("die Vorgabe ist %q, erwartet %q",
			vorgabe, domain.FensterstartMaximiert)
	}
}

// **Drei Werte und kein Vollbild.** Das Kriterium zählt sie auf: Normalgröße,
// Maximierung, Wiederherstellung des letzten Zustands. Ein vierter Wert
// „vollbild" stünde da als Einladung zu dem, was AK-01 ausschließt.
func TestDerFensterstartKenntDreiWerte(t *testing.T) {
	t.Parallel()

	erwartet := []string{
		domain.FensterstartMaximiert,
		domain.FensterstartNormal,
		domain.FensterstartZuletzt,
	}

	werte := domain.ErlaubteEinstellungswerte(domain.EinstellungFensterstart)

	if !slices.Equal(werte, erwartet) {
		t.Errorf("die Werte sind %v, erwartet %v", werte, erwartet)
	}
}

// **Er ist eine Frage dieser Maschine und nicht dieses Profils.** Ein
// Bildschirm hat eine Größe; wer sein Profil auf einen anderen Rechner
// mitnimmt, nimmt keine Fenstergröße mit — dieselbe Überlegung wie bei der
// Seitenleiste.
func TestDerFensterstartGehoertDemGeraet(t *testing.T) {
	t.Parallel()

	ebene, kennt := domain.EbeneVon(domain.EinstellungFensterstart)

	if !kennt {
		t.Fatalf("%q hat keine Ebene", domain.EinstellungFensterstart)
	}

	if ebene != domain.EbeneGeraet {
		t.Errorf("der Fensterstart liegt auf der Ebene %q, erwartet %q",
			ebene, domain.EbeneGeraet)
	}
}

// **Und eine Rücksetzung des Layouts nimmt ihn mit.** Wer „das Fenster sieht
// falsch aus" sagt, meint auch seine Größe.
func TestDerFensterstartStehtBeiDenLayouteinstellungen(t *testing.T) {
	t.Parallel()

	for _, schluessel := range domain.Layouteinstellungen() {
		if schluessel == domain.EinstellungFensterstart {
			return
		}
	}

	t.Errorf("%q fehlt in den Layouteinstellungen", domain.EinstellungFensterstart)
}

// **Ein Fenstermaß ist eine Zahl in Grenzen** — und beide Fehlerwege sind
// betreten, sonst stünde die Regel da und niemand wüsste, ob sie greift.
func TestPruefeFenstermass(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wert    string
		erlaubt bool
	}{
		"eine gewöhnliche Breite":     {"1600", true},
		"das Mindestmaß selbst":       {"700", true},
		"das Höchstmaß selbst":        {"8192", true},
		"zu klein":                    {"20", false},
		"zu groß":                     {"99999", false},
		"keine Zahl":                  {"breit", false},
		"leer":                        {"", false},
		"eine Zahl mit Nachkommateil": {"1600.5", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeFenstermass(fall.wert)

			if fall.erlaubt && fehler != nil {
				t.Errorf("%q wurde abgewiesen: %v", fall.wert, fehler)
			}

			if !fall.erlaubt && fehler == nil {
				t.Errorf("%q ging durch", fall.wert)
			}
		})
	}
}
