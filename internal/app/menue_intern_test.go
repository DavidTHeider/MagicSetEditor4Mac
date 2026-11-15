// Der Wächter über die Beschriftungen der Menüleiste.
//
// **Kein anderer sieht sie.** `keintextimjsx.test.ts` liest JSX, die Menüleiste
// ist nativ; der Fehlerschlüsselvertrag deckt Meldungen ab, keine
// Beschriftungen. Genau in dieser Lücke standen „Ablage" und „Einstellungen …"
// eine ganze Phase lang, und im Kopf von `menue.go` stand, dass es eine Lücke
// ist — eine benannte Lücke ist keine geschlossene.
//
// # Was er nicht prüft
//
// Ob die Beschriftung **passt**. `texte.Ablage` an einem Bearbeiten-Menü wäre
// hier gedeckt und trotzdem falsch. Er sieht die Form, nicht den Sinn.
package app

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// beschriftungImQuelltext findet einen Menüeintrag oder einen Dialog mit
// fester Zeichenkette — in der Gestalt von Wails 3 (`Add`, `AddSubmenu`,
// `SetTitle`, `SetMessage`, `AddFilter`, `SetLabel`) und in der von Wails 2.
var beschriftungImQuelltext = regexp.MustCompile(
	`(?:AddText|AddSubmenu|AddCheckbox|AddRadio|\.Add|SetLabel|SetTitle|SetMessage|AddFilter)\("|` +
		`(?:Title|DisplayName):\s*"`)

func TestKeineBeschriftungImQuelltext(t *testing.T) {
	// **Jede native Stelle**, nicht nur die, an der es einmal aufgefallen ist:
	// der Dateidialog hat denselben Fehler ein zweites Mal gemacht. Seit
	// ENG-02 baut die Anwendung das Modell, und die Hülle setzt es um — beide
	// stehen hier.
	for _, datei := range []string{"menue.go", "../fenster/menue.go", "../fenster/wails.go"} {
		t.Run(datei, func(t *testing.T) {
			quelle, err := os.ReadFile(datei) //nolint:gosec // Quelltext im Repository
			if err != nil {
				t.Fatalf("%s nicht lesbar: %v", datei, err)
			}

			treffer := beschriftungImQuelltext.FindAllString(string(quelle), -1)
			if len(treffer) > 0 {
				t.Errorf("%d Beschriftung(en) im Quelltext: %s — sie gehören nach menuetexte.go",
					len(treffer), strings.Join(treffer, ", "))
			}
		})
	}
}

// **Jede Sprache ist vollständig.** Ein leeres Feld wäre ein Menüeintrag ohne
// Beschriftung — sichtbar erst im gebauten Fenster.
func TestJedeSpracheIstVollstaendig(t *testing.T) {
	for _, sprache := range Sprachen {
		t.Run(sprache, func(t *testing.T) {
			texte := MenuetexteFuer(sprache)

			spiegel := reflect.ValueOf(texte)
			for i := range spiegel.NumField() {
				if strings.TrimSpace(spiegel.Field(i).String()) == "" {
					t.Errorf("%s ist in %q leer", spiegel.Type().Field(i).Name, sprache)
				}
			}
		})
	}
}

// **Auch die Dialoge sind zweisprachig und vollständig.**
//
// Der erste Anlauf des Dateidialogs trug seine Beschriftungen als deutsche
// Zeichenketten im Quelltext — dieselbe Lücke, die `menue.go` eine ganze Phase
// lang hatte. Ein nativer Text sieht kein Wächter, wenn ihn keiner sucht.
func TestJedeSpracheHatVollstaendigeDialogtexte(t *testing.T) {
	for _, sprache := range Sprachen {
		t.Run(sprache, func(t *testing.T) {
			spiegel := reflect.ValueOf(DialogtexteFuer(sprache))
			for i := range spiegel.NumField() {
				if strings.TrimSpace(spiegel.Field(i).String()) == "" {
					t.Errorf("%s ist in %q leer", spiegel.Type().Field(i).Name, sprache)
				}
			}
		})
	}

	if DialogtexteFuer("de") == DialogtexteFuer("en") {
		t.Error("der deutsche und der englische Dialogkatalog sind gleich")
	}

	if DialogtexteFuer("kl") != DialogtexteFuer("de") {
		t.Error("eine unbekannte Sprache fällt nicht auf Deutsch zurück")
	}
}

// **Und die Wörter des Turnierberichts** — DOM-04.AK-03 (#1699): er entsteht
// in Go, und ein leeres Wort stünde als Lücke in einer Tabelle.
func TestJedeSpracheHatVollstaendigeBerichttexte(t *testing.T) {
	for _, sprache := range Sprachen {
		spiegel := reflect.ValueOf(BerichttexteFuer(sprache))
		for i := range spiegel.NumField() {
			if strings.TrimSpace(spiegel.Field(i).String()) == "" {
				t.Errorf("%s ist in %q leer", spiegel.Type().Field(i).Name, sprache)
			}
		}
	}

	if BerichttexteFuer("de") == BerichttexteFuer("en") || BerichttexteFuer("kl") != BerichttexteFuer("de") {
		t.Error("die Kataloge sind gleich, oder eine unbekannte Sprache fällt nicht auf Deutsch zurück")
	}
}

// **Und keine Sprache ist die andere.** Zwei Kataloge, die dieselben Wörter
// führen, sind einer — und dann ist die Übersetzung vergessen worden.
func TestDieSprachenUnterscheidenSich(t *testing.T) {
	if MenuetexteFuer("de") == MenuetexteFuer("en") {
		t.Error("der deutsche und der englische Katalog sind gleich")
	}
}

// Eine unbekannte Sprache gibt den deutschen Katalog, keinen leeren.
func TestEineUnbekannteSpracheFaelltAufDeutschZurueck(t *testing.T) {
	if MenuetexteFuer("kl") != MenuetexteFuer("de") {
		t.Errorf("%+v", MenuetexteFuer("kl"))
	}

	if IstSprache("kl") || !IstSprache("en") {
		t.Error("IstSprache sagt das Falsche")
	}
}

// Der Katalog führt genau die angemeldeten Sprachen.
func TestSprachenUndKatalogeGehenZusammen(t *testing.T) {
	if len(Sprachen) != len(menuekataloge) {
		t.Errorf("%d Sprachen, %d Kataloge", len(Sprachen), len(menuekataloge))
	}

	for _, sprache := range Sprachen {
		if !IstSprache(sprache) {
			t.Errorf("%q ist angemeldet, hat aber keinen Katalog", sprache)
		}
	}
}
