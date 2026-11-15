package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Der Rückweg von `combined_editor`** —
// #696.
func TestVerbundAus(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		skript  string
		da      bool
		felder  []string
		trenner []string
	}{
		{
			name: "magic.mse-game",
			skript: `combined_editor(field1: card.rule_text, ` +
				`separator: "<line>\n</line>", field2: card.flavor_text)`,
			da:      true,
			felder:  []string{"rule_text", "flavor_text"},
			trenner: []string{"<line>\n</line>"},
		},
		{
			name: "forward_editor ist derselbe Weg",
			skript: `forward_editor(field1: card.rule_text, ` +
				`separator: " ", field2: card.flavor_text)`,
			da:      true,
			felder:  []string{"rule_text", "flavor_text"},
			trenner: []string{" "},
		},
		{
			// **So schreibt das Pack die Reihe** — `separator1` neben
			// `field1`, nicht `separator` neben `field2`.
			name: "drei Felder, zwei Trenner",
			skript: `combined_editor(field1: card.a, separator1: "-", ` +
				`field2: card.b, separator2: "+", field3: card.c)`,
			da:      true,
			felder:  []string{"a", "b", "c"},
			trenner: []string{"-", "+"},
		},
		{
			// **Und die Reihe endet an der ersten Lücke.** `separator`
			// ohne Zahl ist der erste; `separator2` daneben wird nie
			// gelesen, weil `separator1` fehlt — genau MSEs `break`.
			name: "eine Lücke in der Reihe beendet sie",
			skript: `combined_editor(field1: card.a, separator: "-", ` +
				`field2: card.b, separator2: "+", field3: card.c)`,
		},
		{
			// **So steht es wirklich im Pack** — der Aufruf liegt im
			// `else`-Zweig zweier Bedingungen. Die frühere Fassung verlangte
			// den Aufruf als **ganzen** Ausdruck und fand ihn deshalb an
			// keinem einzigen Feld von `magic.mse-game`.
			name: "magic.mse-game, wie es dasteht",
			skript: `if mainframe_walker() then
	mainframe_walker_text_script()
else if alt_text() then
	alt_text_script()
else
	combined_editor(field1: card.rule_text, separator: "<line>
</line>", field2: card.flavor_text)`,
			da:      true,
			felder:  []string{"rule_text", "flavor_text"},
			trenner: []string{"<line>\n</line>"},
		},
		{
			// **Zwei Aufrufe sind keine Antwort.** Welcher gilt, entscheidet
			// sich beim Rechnen; wer einen davon nähme, schriebe die Hälfte
			// der Karten in die falschen Felder.
			name: "zwei Verbünde in einem Skript",
			skript: `if a() then combined_editor(field1: card.a, separator: "-", field2: card.b) ` +
				`else combined_editor(field1: card.c, separator: "-", field2: card.d)`,
		},
		{
			// **Ein gerechneter Trenner bleibt draußen.** `pt` und `type`
			// schreiben `separator: language().pt_separator` — was dabei
			// herauskommt, weiß erst der Auswerter, und ohne den Trenner
			// lässt sich nicht spalten.
			name: "ein gerechneter Trenner im Pack",
			skript: `combined_editor(field1: card.power, ` +
				`separator: language().pt_separator, field2: card.toughness)`,
		},
		{
			name:   "ein gewöhnliches Skript ist kein Rückweg",
			skript: `card.name + " " + card.rule_text`,
		},
		{
			name:   "ein gerechnetes Argument ist kein Rückweg",
			skript: `combined_editor(field1: to_upper(card.a), separator: "-", field2: card.b)`,
		},
		{
			name:   "ein Trenner, der gerechnet wird, ist kein Rückweg",
			skript: `combined_editor(field1: card.a, separator: card.x, field2: card.b)`,
		},
		{
			name:   "ohne Trenner geht es nicht zurück",
			skript: `combined_editor(field1: card.a, field2: card.b)`,
		},
		{
			name:   "ein Zugriff auf etwas anderes als die Karte zählt nicht",
			skript: `combined_editor(field1: set.a, separator: "-", field2: card.b)`,
		},
		{
			// **Nur ein Name wird gerufen.** `card.combined_editor(…)`
			// ruft ein Glied der Karte; was es tut, weiß niemand.
			name:   "ein gerufenes Glied ist kein Rückweg",
			skript: `card.combined_editor(field1: card.a, separator: "-", field2: card.b)`,
		},
		{
			name:   "ein anderer Aufruf ist kein Rückweg",
			skript: `to_upper(card.rule_text)`,
		},
		{
			name:   "unzerteilbar",
			skript: `combined_editor(field1:`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			verbund, da := msescript.VerbundAus(fall.skript)
			if da != fall.da {
				t.Fatalf("gefunden = %v, erwartet %v", da, fall.da)
			}

			if !da {
				return
			}

			if strings.Join(verbund.Felder, "|") != strings.Join(fall.felder, "|") {
				t.Errorf("Felder = %q, erwartet %q", verbund.Felder, fall.felder)
			}

			if strings.Join(verbund.Trenner, "|") != strings.Join(fall.trenner, "|") {
				t.Errorf("Trenner = %q, erwartet %q", verbund.Trenner, fall.trenner)
			}
		})
	}
}

// **Zerlegen ist das Gegenstück zum Zusammensetzen** — `<sep>…</sep>` fällt
// samt Inhalt weg, und was dazwischen steht, ist ein Teil.
func TestVerbundZerlege(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		felder int
		wert   string
		teile  []string
	}{
		{
			name: "Regeltext und Flavortext", felder: 2,
			wert:  "Füge drei Schaden zu.<sep><line>\n</line></sep><i-flavor>Es brennt.</i-flavor>",
			teile: []string{"Füge drei Schaden zu.", "<i-flavor>Es brennt.</i-flavor>"},
		},
		{
			name: "ein weicher Trenner zählt genauso", felder: 2,
			wert:  "Kreatur<sep-soft> — </sep-soft>Goblin",
			teile: []string{"Kreatur", "Goblin"},
		},
		{
			name: "ohne Trenner bleibt ein Teil", felder: 2,
			wert:  "nur Regeltext",
			teile: []string{"nur Regeltext", ""},
		},
		{
			name: "mehr Teile als Felder werden abgeschnitten", felder: 2,
			wert:  "a<sep>-</sep>b<sep>-</sep>c",
			teile: []string{"a", "b"},
		},
		{
			name: "eine offene Marke frisst den Rest", felder: 2,
			wert:  "a<sep>-",
			teile: []string{"a", ""},
		},
		{
			name: "geschachtelt", felder: 2,
			wert:  "a<sep>x<sep>y</sep>z</sep>b",
			teile: []string{"a", "b"},
		},
		{
			name: "eine Marke ohne Ende ist keine", felder: 2,
			wert:  "a<sep",
			teile: []string{"a", ""},
		},
		{
			name: "ein Schluss ohne Ende frisst den Rest", felder: 2,
			wert:  "a<sep>b</sep",
			teile: []string{"a", ""},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			verbund := msescript.Verbund{Felder: make([]string, fall.felder)}

			teile := verbund.Zerlege(fall.wert)
			if strings.Join(teile, "|") != strings.Join(fall.teile, "|") {
				t.Errorf("Teile = %q, erwartet %q", teile, fall.teile)
			}
		})
	}
}

// **`rule_text` findet `rule text`** — dieselbe Regel wie beim Lesen eines
// Glieds, und der Rückweg braucht sie außerhalb dieses Pakets.
func TestGliedTrifft(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		gesucht, glied string
		trifft         bool
	}{
		{"rule_text", "rule text", true},
		{"rule_text", "rule texts", true},
		{"rule_texts", "rule text", false},
		{"flavor_text", "rule text", false},
	}

	for _, fall := range faelle {
		t.Run(fall.gesucht+"/"+fall.glied, func(t *testing.T) {
			t.Parallel()

			if msescript.GliedTrifft(fall.gesucht, fall.glied) != fall.trifft {
				t.Errorf("GliedTrifft(%q, %q) != %v", fall.gesucht, fall.glied, fall.trifft)
			}
		})
	}
}

// **Woraus ein Feld entsteht, auch wenn es sich nicht spalten lässt** —
// #687.
//
// `type` und `pt` schreiben `separator: language().type_separator`. Der
// Rückweg bleibt zu; die Frage der Oberfläche — *welche Felder sind das?* —
// beantwortet der Aufruf trotzdem.
func TestVerbundfelderAus(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		skript string
		felder []string
	}{
		{
			name: "type aus magic.mse-game",
			skript: `combined_editor(field1: card.super_type, ` +
				`separator: language().type_separator, field2: card.sub_type, ` +
				`soft_before_empty: true, hide_when_empty: true)`,
			felder: []string{"super_type", "sub_type"},
		},
		{
			name: "und der Rückweg, der offen ist, nennt sie auch",
			skript: `combined_editor(field1: card.rule_text, ` +
				`separator: "<line>\n</line>", field2: card.flavor_text)`,
			felder: []string{"rule_text", "flavor_text"},
		},
		{
			name:   "ein gewöhnliches Skript nennt keine",
			skript: `card.name + "!"`,
		},
		{
			name:   "und ein gerechnetes Feldargument auch nicht",
			skript: `combined_editor(field1: to_upper(card.a), separator: "-", field2: card.b)`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			felder := msescript.VerbundfelderAus(fall.skript)
			if strings.Join(felder, "|") != strings.Join(fall.felder, "|") {
				t.Errorf("Felder = %q, erwartet %q", felder, fall.felder)
			}
		})
	}
}
