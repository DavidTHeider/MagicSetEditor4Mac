package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Randfälle des Bildwerts** —
// #660.
//
// `exportsetsymbol_test.go` daneben prüft den Weg: Fassung, Zeichnung,
// Merkliste, Ablage. Hier stehen die Abzweigungen — was keine Karte und kein
// Bild ist, was eine fehlende Fähigkeit sagt, und was mit einer Angabe
// geschieht, die keine Zahl ist.

// **Die Randfälle des Bildwerts.**
//
// Sie stehen zusammen, weil sie dasselbe belegen: dass jede Verzweigung
// erreichbar ist. Ohne sie hielte die Abdeckung von 100 % eine Zusage, die
// niemand geprüft hat (§3).
func TestDieRandfaelleDesBildwerts(t *testing.T) {
	t.Parallel()

	// **Etwas, das weder Karte noch Bild ist.** Eine Zahl hat keine Marke und
	// gibt kein Bild her.
	t.Run("eine Zahl ist kein Bild", func(t *testing.T) {
		t.Parallel()

		_, _, err := mitSetsymbol(&beigabenmitschrift{}, &baumrastererprobe{},
			&symbolgestaltprobe{}).
			Fuehre(t.Context(), vorlageMitFassung(`write_image_file(12, file: "x.png")`),
				eineKarte(), nil, nil)
		if err == nil || !strings.Contains(err.Error(), "weder eine Karte noch ein Bild") {
			t.Fatalf("Fehler = %v", err)
		}
	})

	// **Eine Zeichenkette ist ein Dateibild** — MSEs `ScriptString::toImage`.
	t.Run("eine Zeichenkette ist ein Dateibild", func(t *testing.T) {
		t.Parallel()

		rasterer := &baumrastererprobe{}

		if _, _, err := mitSetsymbol(&beigabenmitschrift{}, rasterer, &symbolgestaltprobe{}).
			Fuehre(t.Context(),
				vorlageMitFassung(`write_image_file("bild.png", file: "x.png")`),
				eineKarte(), nil, nil); err != nil {
			t.Fatalf("ausführen: %v", err)
		}

		if len(rasterer.baeume) != 1 || rasterer.baeume[0].Pfad != "/bild.png" {
			t.Errorf("gerastert wurde %+v", rasterer.baeume)
		}
	})

	// **Ohne Symbolquelle kann die Vorlage nur Karten schreiben.**
	t.Run("ohne Symbolquelle nur Karten", func(t *testing.T) {
		t.Parallel()

		_, _, err := probelauf(service.Ausfuhrmittel{
			Beigaben:    &beigabenmitschrift{},
			Bilder:      stummerBildschreiber{},
			Bildziel:    t.TempDir(),
			Verzeichnis: "ausfuhr-files",
		}).Fuehre(t.Context(),
			vorlageMitFassung(`write_image_file("bild.png", file: "x.png")`),
			eineKarte(), nil, nil)
		if err == nil || !strings.Contains(err.Error(), "nur Karten schreiben") {
			t.Fatalf("Fehler = %v", err)
		}
	})

	// **Ein Verweis, der seinen Ort schon trägt, bleibt unberührt.**
	t.Run("ein absoluter Pfad bleibt", func(t *testing.T) {
		t.Parallel()

		gestalt := &symbolgestaltprobe{}

		if _, _, err := mitSetsymbol(&beigabenmitschrift{}, &baumrastererprobe{}, gestalt).
			Fuehre(t.Context(),
				vorlageMitFassung(`write_image_file(file: "s.png",`+
					` symbol_variation(symbol: "/fremd.mse-style/s.mse-symbol",`+
					` variation: "rare"))`),
				eineKarte(), nil, nil); err != nil {
			t.Fatalf("ausführen: %v", err)
		}

		if len(gestalt.gefragt) != 1 || gestalt.gefragt[0] != "/fremd.mse-style/s.mse-symbol" {
			t.Errorf("gefragt wurde %q", gestalt.gefragt)
		}
	})

	// **Eine Breite, die keine Zahl ist, ist keine Angabe.**
	t.Run("eine unlesbare Breite", func(t *testing.T) {
		t.Parallel()

		rasterer := &baumrastererprobe{}

		if _, _, err := mitSetsymbol(&beigabenmitschrift{}, rasterer, &symbolgestaltprobe{}).
			Fuehre(t.Context(),
				vorlageMitFassung(`write_image_file("b.png", file: "x.png", width: "breit")`),
				eineKarte(), nil, nil); err != nil {
			t.Fatalf("ausführen: %v", err)
		}

		if len(rasterer.groesse) != 1 || rasterer.groesse[0] != [2]int{0, 0} {
			t.Errorf("verlangt wurde %v", rasterer.groesse)
		}
	})

	// **Und ein Schreibfehler kommt durch** (§8).
	t.Run("die Platte ist voll", func(t *testing.T) {
		t.Parallel()

		_, _, err := mitSetsymbol(
			&beigabenmitschrift{schreibfehte: errors.New("kein Platz")},
			&baumrastererprobe{}, &symbolgestaltprobe{},
		).Fuehre(t.Context(),
			vorlageMitFassung(`write_image_file("b.png", file: "x.png")`),
			eineKarte(), nil, nil)
		if err == nil || !strings.Contains(err.Error(), "kein Platz") {
			t.Fatalf("Fehler = %v", err)
		}
	})
}

// stummerBildschreiber schreibt Karten, aber keine Bildwerte.
type stummerBildschreiber struct{}

func (stummerBildschreiber) Schreibe(
	_ context.Context, _ int64, verzeichnis, datei string,
	_ service.Bildformat, _ int,
) (*service.Bildschreibung, error) {
	return &service.Bildschreibung{Pfad: verzeichnis + "/" + datei, Breite: 1, Hoehe: 1}, nil
}
