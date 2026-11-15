package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine `.mse-symbol` ist ein Baum aus Formen und Symmetrien.**
//
// # Was hier gelesen wird
//
// Dieselbe Syntax wie jedes andere MSE-Paket; neu ist die **Bedeutung** der
// Blöcke (`src/data/symbol.cpp`, Zeilen 17–25, 126–166 und 225–285, Fassung
// f9d9356 vom 2021-07-06):
//
//   - `type: shape` ist ein **geschlossener** Zug aus `point:`-Blöcken, dazu
//     `combine:` — wie er mit dem verrechnet wird, was schon dasteht.
//   - `type: symmetry` vervielfältigt seine Kinder: `kind:` (`rotation` oder
//     `reflection`), `copies:`, `center:`, `handle:`.
//   - `type: group` fasst Kinder zusammen, ohne sie zu ändern.
//
// Ein Punkt trägt `position:` und, wo `line after: curve` steht, die
// Bézier-Griffe — **relativ zur Position**, nicht absolut. Alle Koordinaten
// sind auf [0,1] normiert
// (#323).
//
// # Das Prüfstück ist selbst geschrieben
//
// `testdaten/` führt keine fremden Vorlagen (§11), und das lokal installierte
// `M15-Magic-Pack` enthält **keine** `.mse-symbol`. Die 21 Dateien, an denen
// sich ein Rasterer messen ließe, liegen in `Full-Magic-Pack` — hier liegen sie
// nicht. Geprüft wird deshalb gegen eine minimale Fassung derselben Syntax.
func TestEineSymbolzeichnungWirdGelesen(t *testing.T) {
	pfad := probesymbol(t, `mse version: 0.3.5
part:
	type: symmetry
	name: Reflection (2)
	kind: reflection
	copies: 2
	center: (0.5000000000,0.5000000000)
	handle: (0.2000000000,0.0000000000)
	part:
		type: shape
		name: Rechteck
		combine: subtract
		point:
			position: (0.1880949567,0.3189438621)
			lock: free
			line after: curve
			handle before: (0.0100000000,0.0200000000)
			handle after: (-0.0564330311,-0.0268493516)
		point:
			position: (0.8000000000,0.3000000000)
			lock: size
			line after: line
`)

	zeichnung, befunde, err := msepaket.LiesSymbolzeichnung(pfad)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(zeichnung.Teile) != 1 {
		t.Fatalf("%d Teile, erwartet eines: %+v", len(zeichnung.Teile), zeichnung.Teile)
	}

	symmetrie := zeichnung.Teile[0]

	if symmetrie.Art != msepaket.TeilSymmetrie || symmetrie.Spiegelart != msepaket.SpiegelungAchse {
		t.Errorf("das äußere Teil ist %+v", symmetrie)
	}

	if symmetrie.Kopien != 2 || symmetrie.Mitte != (msepaket.Vektor{X: 0.5, Y: 0.5}) {
		t.Errorf("Kopien %d, Mitte %+v", symmetrie.Kopien, symmetrie.Mitte)
	}

	if len(symmetrie.Teile) != 1 {
		t.Fatalf("%d Kinder, erwartet eines", len(symmetrie.Teile))
	}

	form := symmetrie.Teile[0]

	if form.Art != msepaket.TeilForm || form.Kombination != msepaket.KombinationAbziehen {
		t.Errorf("das innere Teil ist %+v", form)
	}

	if len(form.Punkte) != 2 {
		t.Fatalf("%d Punkte, erwartet zwei", len(form.Punkte))
	}

	erst := form.Punkte[0]
	if !erst.KurveDanach || erst.GriffDanach != (msepaket.Vektor{X: -0.0564330311, Y: -0.0268493516}) {
		t.Errorf("der erste Punkt ist %+v", erst)
	}

	// **Die Kurve gilt für die Strecke, nicht für den Punkt.** MSE setzt
	// `p2->segment_before = p1->segment_after` (`enforceConstraints`, Zeilen
	// 205–212); der zweite Punkt hat also eine Kurve **davor** und eine Gerade
	// danach.
	zweit := form.Punkte[1]
	if !zweit.KurveDavor || zweit.KurveDanach {
		t.Errorf("der zweite Punkt ist %+v", zweit)
	}

	// **Und der letzte Punkt schließt zum ersten.** Ein Zug ist geschlossen;
	// die Strecke vom letzten zum ersten trägt dessen `line after`.
	if erst.KurveDavor {
		t.Errorf("der erste Punkt trägt eine Kurve davor — der Zug schließt gerade")
	}
}

// **Ein unbekannter Teiltyp ist ein Befund, kein Abbruch.**
//
// MSE wirft dort einen Fehler (`read_new<SymbolPart>`); hier gilt §8: eine
// Zeichnung mit drei brauchbaren Teilen und einem fremden ist dreimal
// nützlicher als gar keine.
func TestEinUnbekannterSymbolteilWirdBenannt(t *testing.T) {
	pfad := probesymbol(t, `mse version: 0.3.5
part:
	type: gibtsnicht
part:
	type: shape
	point:
		position: (0.1,0.2)
`)

	zeichnung, befunde, err := msepaket.LiesSymbolzeichnung(pfad)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	if len(zeichnung.Teile) != 1 || zeichnung.Teile[0].Art != msepaket.TeilForm {
		t.Errorf("die Teile sind %+v, erwartet die eine brauchbare Form", zeichnung.Teile)
	}
}

// **Ohne `type:` ist ein Teil eine Form.**
//
// `read_new<SymbolPart>` liest `type.empty()` als `shape` (Zeile 139).
func TestEinTeilOhneTypIstEineForm(t *testing.T) {
	pfad := probesymbol(t, `mse version: 0.3.5
part:
	point:
		position: (0.25,0.75)
`)

	zeichnung, _, err := msepaket.LiesSymbolzeichnung(pfad)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(zeichnung.Teile) != 1 || zeichnung.Teile[0].Art != msepaket.TeilForm {
		t.Fatalf("die Teile sind %+v", zeichnung.Teile)
	}

	// **Und `combine:` fehlt: das ist `overlap`**, der Nullwert des Aufzählers
	// (`SymbolShapeCombine`, Zeilen 153–160).
	if zeichnung.Teile[0].Kombination != msepaket.KombinationUeberdecken {
		t.Errorf("die Kombination ist %v, erwartet überdecken",
			zeichnung.Teile[0].Kombination)
	}

	if stelle := zeichnung.Teile[0].Punkte[0].Stelle; stelle != (msepaket.Vektor{X: 0.25, Y: 0.75}) {
		t.Errorf("die Stelle ist %+v", stelle)
	}
}

// probesymbol legt eine `.mse-symbol` mit dem gegebenen Inhalt an.
//
// **Eine Datei und kein Verzeichnis.** Eine `.mse-symbol` liegt in einem
// Stilpaket als einzelne Datei — anders als `.mse-style` oder `.mse-game`,
// die Verzeichnisse sind.
func probesymbol(t *testing.T, inhalt string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.mse-symbol")
	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// **Was kein `(x,y)` ist, ist der Nullpunkt — und keine halbe Zahl.**
//
// Eine Zeichnung von 2007 kann alles enthalten (§8). Ein halb gelesener Vektor
// wäre eine Zahl, die aussieht wie eine gemessene; der Nullpunkt ist wenigstens
// als solcher zu erkennen.
func TestEinUnlesbarerVektorIstDerNullpunkt(t *testing.T) {
	faelle := []struct {
		name    string
		angabe  string
		erwartt msepaket.Vektor
	}{
		{"gut", "(0.25,0.75)", msepaket.Vektor{X: 0.25, Y: 0.75}},
		{"mit Leerraum", "( 0.5 , 0.5 )", msepaket.Vektor{X: 0.5, Y: 0.5}},
		{"ohne Klammer auf", "0.25,0.75)", msepaket.Vektor{}},
		{"ohne Klammer zu", "(0.25,0.75", msepaket.Vektor{}},
		{"nur eine Zahl", "(0.25)", msepaket.Vektor{}},
		{"keine Zahlen", "(a,b)", msepaket.Vektor{}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			pfad := probesymbol(t, "mse version: 0.3.5\npart:\n\tpoint:\n\t\tposition: "+
				fall.angabe+"\n")

			zeichnung, _, err := msepaket.LiesSymbolzeichnung(pfad)
			if err != nil {
				t.Fatalf("lesen: %v", err)
			}

			if stelle := zeichnung.Teile[0].Punkte[0].Stelle; stelle != fall.erwartt {
				t.Errorf("die Stelle ist %+v, erwartet %+v", stelle, fall.erwartt)
			}
		})
	}
}

// **Eine Kopienzahl, die keine ist, ist null.**
//
// Und `copies:` ganz ohne Angabe ebenso: was fehlt, wird nicht geraten.
func TestEineUnlesbareKopienzahlIstNull(t *testing.T) {
	pfad := probesymbol(t, `mse version: 0.3.5
part:
	type: symmetry
	copies: viele
part:
	type: symmetry
`)

	zeichnung, _, err := msepaket.LiesSymbolzeichnung(pfad)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for stelle, teil := range zeichnung.Teile {
		if teil.Kopien != 0 {
			t.Errorf("Teil %d nennt %d Kopien, erwartet null", stelle, teil.Kopien)
		}
	}
}

// **Eine Datei, die es nicht gibt, ist ein Fehler und kein leeres Symbol.**
//
// Ein leeres Symbol zeichnete nichts und sähe aus wie ein Symbol ohne Formen.
func TestEineFehlendeSymbolzeichnungIstEinFehler(t *testing.T) {
	_, _, err := msepaket.LiesSymbolzeichnung(filepath.Join(t.TempDir(), "gibtsnicht.mse-symbol"))
	if err == nil {
		t.Error("kein Fehler für eine fehlende Datei")
	}
}

// **Eine Gruppe trägt ihre Kinder, ohne sie zu ändern.**
func TestEineGruppeTraegtIhreKinder(t *testing.T) {
	pfad := probesymbol(t, `mse version: 0.3.5
part:
	type: group
	name: Gruppe
	part:
		type: shape
		point:
			position: (0.1,0.1)
	part:
		type: gibtsnicht
`)

	zeichnung, _, err := msepaket.LiesSymbolzeichnung(pfad)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(zeichnung.Teile) != 1 || zeichnung.Teile[0].Art != msepaket.TeilGruppe {
		t.Fatalf("die Teile sind %+v", zeichnung.Teile)
	}

	// **Das fremde Kind fällt weg, das brauchbare bleibt** — der Grund steht
	// bei [msepaket.kinderAus].
	if len(zeichnung.Teile[0].Teile) != 1 {
		t.Errorf("%d Kinder, erwartet eines", len(zeichnung.Teile[0].Teile))
	}
}

// **Eine Datei, die sich nicht zerteilen lässt, ist ein Fehler.**
//
// Ein Symbol ohne lesbaren Baum hat keine Formen, und eine leere Zeichnung
// zurückzugeben hieße, einen Lesefehler als „Symbol ohne Inhalt" auszugeben.
func TestEineUnzerteilbareSymbolzeichnungIstEinFehler(t *testing.T) {
	// Eine Zeile ohne Schlüssel — dieselbe Form, die auch ein Paket abweist.
	pfad := probesymbol(t, ": ohne schlüssel\n")

	if _, _, err := msepaket.LiesSymbolzeichnung(pfad); err == nil {
		t.Error("kein Fehler für eine unzerteilbare Datei")
	}
}
