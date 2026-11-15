package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Bildskript gibt ein Bild zurück, keinen Dateinamen.**
//
// # Der Befund dahinter
//
// `src/gfx/generated_image.hpp` von MSE2 führt 21 Klassen, die alle von
// `GeneratedImage` erben: ein Blatt (`PackagedImage`, eine Datei im Paket) und
// zwanzig Verknüpfungen, die Bilder aus Bildern machen. Unser Zeichenplan trug
// bis hierher einen **Pfad** — und `recolor_image` hätte dort nichts
// zurückzugeben gehabt
// (#270).
//
// **965 der 2.729 Befunde** über die 190 Stile von `M15-Magic-Pack` lauten
// `der Name "recolor_image" ist hier nicht bekannt`. Es ist der größte
// Einzelposten überhaupt.
//
// # Warum `msescript` trotzdem nicht weiß, was ein Bild ist
//
// Es trägt einen **Baum** und deutet ihn nicht: eine Knotenart als Text, die
// Argumente als Einträge. Dieselbe Bauform wie [msescript.Farbe] — was `rgb`
// zurückgibt, weiß der Auswerter auch nicht.
//
// §2 verbietet `msescript` die Kenntnis dieses Projekts, nicht das Tragen einer
// Struktur. Wer hier eine Rasterung fände, hätte den Verstoß gefunden.
//
// Fassung f9d9356 vom 2021-07-06.
func TestRecolorImageGibtEinenBildbaum(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`recolor_image("rahmen.png", red: rgb(255,0,0))`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	baum, istBild := wert.AlsBild()
	if !istBild {
		t.Fatalf("die Art ist %q, erwartet ein Bild", wert.Art)
	}

	if baum.Knotenart != "recolor_image" {
		t.Errorf("die Knotenart ist %q, erwartet recolor_image", baum.Knotenart)
	}

	if len(baum.Kinder) != 1 {
		t.Fatalf("%d Kinder, erwartet eines", len(baum.Kinder))
	}

	// **Die Zeichenkette ist zum Blatt geworden.** Ein Pfad *ist* ein Bild —
	// das ist die Umwandlung, die MSE `to_image` nennt, und ohne sie müsste
	// jede Vorlage sie von Hand schreiben.
	if blatt := baum.Kinder[0].Baum; blatt.Knotenart != "datei" || blatt.Pfad != "rahmen.png" {
		t.Errorf("das Blatt ist %+v, erwartet die Datei rahmen.png", blatt)
	}
}

// **Ein Bild ist auch ein Bild, wenn es schon eines war.**
//
// `recolor_image(recolor_image(…))` kommt im Pack vor; die innere Verknüpfung
// darf nicht noch einmal als Dateiname gelesen werden.
func TestEinBildbaumWaechstInDieTiefe(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`recolor_image(recolor_image("a.png", red: rgb(1,2,3)), red: rgb(4,5,6))`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	baum, istBild := wert.AlsBild()
	if !istBild {
		t.Fatalf("die Art ist %q, erwartet ein Bild", wert.Art)
	}

	if len(baum.Kinder) != 1 || baum.Kinder[0].Baum.Knotenart != "recolor_image" {
		t.Fatalf("das Kind ist %+v, erwartet eine zweite Einfärbung", baum.Kinder)
	}

	if tiefer := baum.Kinder[0].Baum.Kinder; len(tiefer) != 1 || tiefer[0].Baum.Pfad != "a.png" {
		t.Errorf("das Blatt ist %+v, erwartet die Datei a.png", tiefer)
	}
}

// **Ohne Bild kein Baum, und ein benannter Fehler statt eines leeren.**
func TestRecolorImageOhneBildIstEinFehler(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`recolor_image(red: rgb(1,2,3))`, msescript.Bibliothek()); err == nil {
		t.Fatal("ohne Bild muss recolor_image scheitern")
	}
}

// **Was kein Bild ist, wird auch keines.**
//
// Eine Zahl, ein Wahrheitswert, eine Abbildung — keines davon ist ein Bild.
//
// **Die leere Zeichenkette und `nil` bleiben hier keines** — und das ist
// eine Aussage über **diesen** Weg: der `image:`-Block eines Feldstils fällt
// dann auf `default:` oder die Wahlbilder zurück.
//
// **Als Argument einer Bildfunktion sind sie sehr wohl ein Bild**, ein
// durchsichtiges: MSE wandelt dort über `toImage()` um
// (`src/script/value.cpp`, Zeilen 369–375 und 450, Fassung f9d9356 vom
// 2021-07-06). Der Wächter dafür steht in `leeresbild_test.go`
// (#509).
func TestWasKeinBildIstWirdKeines(t *testing.T) {
	faelle := map[string]msescript.Wert{
		"nichts":            msescript.Nichts,
		"eine Zahl":         msescript.Zahl(3),
		"leerer Text":       msescript.Text(""),
		"eine Abbildung":    {Art: msescript.ArtAbbildung},
		"ein Funktionswert": msescript.NeueFunktion(nil),
	}

	for name, wert := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, istBild := wert.AlsBild(); istBild {
				t.Errorf("%s gilt als Bild", name)
			}
		})
	}
}

// **Eine Verknüpfung, die kein Bild bekommt, sagt was sie bekam.**
func TestRecolorImageMitEinemNichtbildNenntDieArt(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`recolor_image(3, red: rgb(1,2,3))`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("eine Zahl ist kein Bild")
	}

	if !strings.Contains(err.Error(), "Zahl") {
		t.Errorf("die Meldung lautet %q und nennt die Art nicht", err)
	}
}

// **`input:` als Name, nicht als Stellung.**
//
// Hier stand bis zum 2026-08-18 `image:` und dazu der Satz „beide Formen
// stehen im Pack". **Er stimmte nicht**: über `M15-Magic-Pack` schreiben 643
// Aufrufe `input:` und **kein einziger** `image:` — über das volle Pack 787
// gegen null. Der Name kommt aus MSE2 (`SCRIPT_PARAM_C(GeneratedImageP,
// input)`, `src/script/functions/image.cpp`, Fassung f9d9356 vom
// 2021-07-06), und die Behauptung daneben war eine ungedeckte Aussage über
// fremdes Verhalten
// (#483).
func TestDasBildDarfBenanntKommen(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`recolor_image(input: "a.png", red: rgb(1,2,3))`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	baum, istBild := wert.AlsBild()
	if !istBild || len(baum.Kinder) != 1 || baum.Kinder[0].Baum.Pfad != "a.png" {
		t.Errorf("der Baum ist %+v, erwartet ein Blatt auf a.png", baum)
	}

	// **Und `input:` erscheint nicht doppelt**: einmal als Kind reicht.
	for _, wertchen := range baum.Werte {
		if wertchen.Name == "input" {
			t.Error("`input:` steht sowohl als Kind als auch als Wert")
		}
	}
}

// **`masked_blend` trägt drei Bilder, nicht eines.**
//
// # Der Aufbau, nachgelesen im Pack
//
// `magic-blends.mse-include/new-blends` ruft es mit `light:`, `dark:` und
// `mask:` — und schachtelt sich selbst als `light:`, um mehr als zwei Farben zu
// mischen. Gezählt über `M15-Magic-Pack`: 1.201 `dark:`, 898 `light:`, 866
// `mask:`.
//
// Mit **2.734 Vorkommen** ist es die häufigste Verknüpfung überhaupt
// (#270) — und
// der Hybridrahmen, der heute als nackte Goldfläche herauskommt, hängt daran
// (#306).
//
// # Warum die Kinder Namen brauchen
//
// Bei `recolor_image` gibt es ein Bild, und es ist das erste. Hier gibt es
// drei, und welches welches ist, steht **im Namen** — eine Reihenfolge allein
// trüge die Auskunft nicht.
func TestMaskedBlendTraegtDreiBilder(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`masked_blend(light: "hell.png", dark: "dunkel.png", mask: "maske.png")`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	baum, istBild := wert.AlsBild()
	if !istBild {
		t.Fatalf("die Art ist %q, erwartet ein Bild", wert.Art)
	}

	erwartet := map[string]string{
		"light": "hell.png", "dark": "dunkel.png", "mask": "maske.png",
	}

	if len(baum.Kinder) != len(erwartet) {
		t.Fatalf("%d Kinder, erwartet %d: %+v", len(baum.Kinder), len(erwartet), baum.Kinder)
	}

	for _, kind := range baum.Kinder {
		if kind.Baum.Pfad != erwartet[kind.Name] {
			t.Errorf("%s zeigt auf %q, erwartet %q", kind.Name, kind.Baum.Pfad, erwartet[kind.Name])
		}
	}
}

// **Ein Wahrheitswert ist kein Bild** — der letzte der sechs Arten.
func TestEinWahrheitswertIstKeinBild(t *testing.T) {
	if _, istBild := msescript.Wahrheit(true).AlsBild(); istBild {
		t.Error("ein Wahrheitswert gilt als Bild")
	}
}
