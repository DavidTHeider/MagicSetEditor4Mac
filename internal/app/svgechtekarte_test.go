//go:build fremdpakete

// **Eine echte Karte als SVG** —
// #852.
//
// # Warum hinter dem Build-Tag
//
// Er braucht ein bezogenes Vorlagenpaket, und die kommen nach §11 nicht ins
// Repositorium. Er läuft nicht in der Pipeline und zählt nicht zur Abdeckung —
// was er belegt, belegt er dafür an einer Karte, die es wirklich gibt.
//
// # Was er prüft
//
//  1. Das Ergebnis ist **wohlgeformtes XML** — mit einem echten Zerteiler,
//     nicht mit `strings.Contains`. Der Zweitblick über Phase P hat
//     angestrichen, dass die Tests bis dahin nur Präfixe und Zeichenketten
//     verglichen: „gültiges XML" war eine Behauptung.
//  2. Es trägt **Geometrie und Text** und nicht nur ein eingebettetes Bild.
//     Eine Karte, die als ein einziges `<image>` herauskäme, wäre der Name
//     ohne die Sache.
//
// # Was er nicht prüft
//
// **Ob es aussieht wie die Karte.** Dafür gibt es keinen Maßstab: ein
// SVG-Goldbild bräuchte einen Betrachter, und der wäre der Rechner, nicht die
// Anwendung. Was der Schreiber vom Plan trägt, steht in `docs/FORMATE.md` und
// wird in `internal/svg` Eigenschaft für Eigenschaft geprüft.
package app_test

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

func TestEineEchteKarteAlsSVG(t *testing.T) {
	referenzen := os.Getenv("MSE_REFERENZBILDER")
	if referenzen == "" {
		t.Skip("MSE_REFERENZBILDER ist nicht gesetzt")
	}

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := os.ReadDir(referenzen)
	if err != nil {
		t.Fatalf("Referenzen lesen: %v", err)
	}

	geschrieben := 0

	for _, eintrag := range eintraege {
		name, istKarte := strings.CutSuffix(eintrag.Name(), ".karte")
		if !istKarte {
			continue
		}

		geschrieben++

		t.Run(name, func(t *testing.T) {
			schreibeAlsSVG(t, referenzen, wurzel, name)
		})
	}

	// **Ohne diese Zeile wäre ein leeres Verzeichnis ein Erfolg** (§10).
	if geschrieben == 0 {
		t.Fatalf("in %s liegt keine einzige .karte", referenzen)
	}
}

// schreibeAlsSVG baut den Plan einer echten Karte und prüft sein SVG.
func schreibeAlsSVG(t *testing.T, referenzen, wurzel, name string) {
	t.Helper()

	karte, stilname, _ := referenzkarte(t, filepath.Join(referenzen, name+".karte"))

	suche := schriften.Neue(schriften.Orte()).MitPaketen(filepath.Dir(wurzel))
	bilder := msepaket.NeueBildquelle(wurzel)

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(suche).MitBildern(bilder).Stilplan(
		probekontext(t), karte, fremdvorlage(t, fremdstil(t, wurzel, stilname), wurzel))
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// **Dieselben Teile wie im Betrieb** (`internal/app/app.go`): derselbe
	// Rasterer für die Bildbäume und die Symbolzüge, dieselbe Schriftsuche für
	// den Umbruch, dieselbe Bildquelle für die Symbolmaße.
	rasterer := render.NeuerRasterer(bilder).
		MitSchriften(suche).MitErsatzschrift(suche.Ersatz)

	geschrieben, befunde, err := svg.AusPlan(probekontext(t), plan, svg.Quellen{
		Baeume:    rasterer,
		Raster:    rasterer,
		Schriften: suche,
		Symbole:   bilder,
	})
	if err != nil {
		t.Fatalf("SVG schreiben: %v", err)
	}

	for _, befund := range befunde {
		t.Logf("Befund: %s", befund)
	}

	// **Ein echter Zerteiler und kein `strings.Contains`.** Ein unbalancierter
	// Baum, ein rohes `&` in einem Kartennamen, ein Steuerzeichen aus einer
	// Datei von 2007 — nichts davon fiele einem Zeichenkettenvergleich auf.
	marken := zaehleMarken(t, geschrieben)

	if marken["svg"] != 1 {
		t.Errorf("%d <svg>-Wurzeln", marken["svg"])
	}

	// **Geometrie und Text, nicht nur ein Bild.** Eine Karte, die als ein
	// einziges `<image>` herauskäme, wäre ein gewickeltes PNG.
	if marken["rect"] == 0 {
		t.Error("kein einziges <rect> — die Flächen sind keine Geometrie")
	}

	if marken["text"] == 0 {
		t.Error("kein einziges <text> — der Text ist kein Text")
	}

	t.Logf("%s: %d Marken, davon %d <rect>, %d <text>, %d <image>, %d <mask>",
		name, marken["*"], marken["rect"], marken["text"],
		marken["image"], marken["mask"])
}

// zaehleMarken zerteilt das SVG und zählt seine Elemente.
//
// **`xml.Decoder` liest bis zum Ende und meldet jeden Bruch** — ein
// unbalancierter Baum, ein ungültiges Zeichen, ein kaputtes Attribut.
func zaehleMarken(t *testing.T, dokument string) map[string]int {
	t.Helper()

	marken := map[string]int{}
	// **Die Eingabe ist unser eigenes Erzeugnis**, kein fremdes Dokument —
	// gelesen wird, was `svg.AusPlan` gerade geschrieben hat.
	zerteiler := xml.NewDecoder(strings.NewReader(dokument)) //#nosec G709 -- die Eingabe stammt aus diesem Lauf

	for {
		marke, err := zerteiler.Token()
		if errors.Is(err, io.EOF) {
			return marken
		}

		if err != nil {
			t.Fatalf("das SVG ist kein gültiges XML: %v", err)
		}

		if anfang, ist := marke.(xml.StartElement); ist {
			marken[anfang.Name.Local]++
			marken["*"]++
		}
	}
}
