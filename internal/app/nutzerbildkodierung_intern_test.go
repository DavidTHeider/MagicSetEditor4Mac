package app

// **Ein Profilbild, das sich nicht kodieren lässt, fehlt in der Tabelle** —
// Werkstück 8 von Phase 4.12
// (#1351).
//
// **Ein interner Lauf, weil die Kodiernaht intern ist** — derselbe Grund wie
// bei `kartenbilder_intern_test.go`: sie von außen erreichbar zu machen
// hieße, eine Naht für einen Test zu öffnen, die im Betrieb niemand braucht,
// und `png.Encode` scheitert an einem `bytes.Buffer` nicht von selbst.
//
// **Und die Tabelle bleibt heil**: wer ein Bild hat, das sich kodieren lässt,
// bekommt es. Eine Antwort, die wegen eines Kodierfehlers gar nichts zeigte,
// wäre der schlechtere Tausch (§8).

import (
	"context"
	"image"
	"image/color"
	"io"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinNichtKodierbaresProfilbildFehltInDerTabelle(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	gesicht := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	gesicht.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 255})

	name, fehler := service.NeueBildablage(wurzel).NutzerbildAblegen(1, gesicht)
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	anwendung := NeueApp("")
	anwendung.packen = func(io.Writer, image.Image) error { return errKodierung }
	anwendung.bilderort = func() (string, error) { return wurzel, nil }
	anwendung.nutzer = &einNutzerMitBild{name: name}

	antwort := anwendung.Nutzerbilder(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Bilder) != 0 {
		t.Errorf("es kam ein Bild: %v", antwort.Bilder)
	}
}

// einNutzerMitBild führt genau einen Nutzer, und der hat ein Bild.
type einNutzerMitBild struct {
	Nutzerfuehrer

	name string
}

func (f *einNutzerMitBild) Uebersicht(context.Context) ([]service.Nutzerzeile, error) {
	return []service.Nutzerzeile{{Nutzer: domain.Nutzer{ID: 1, Name: "Ich", Bild: f.name}}}, nil
}
