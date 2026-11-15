package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Sprachlücken, die das Initskript von `magic.mse-game` aufgehalten haben
// (#214).
//
// **Jede ist gemessen, nicht vermutet**, und jede steht mit ihrer Quelle in
// `docs/FORMATE.md`. Sie stehen in einer eigenen Datei, weil sie eine
// gemeinsame Geschichte haben: die Sperre wandert von Lücke zu Lücke, und wer
// sie hier beieinander liest, sieht den Weg.
//
// | Zeile im eingebundenen Initskript | Lücke |
// | --- | --- |
// | 196 | hinter `name:` darf der Wert umbrechen |
// | 8448 | `mod` ist ein Operator |
// | 8590 | eine Zahl ist ein Abbildungsschlüssel |
// | 9726 | ein `if` darf rechts von einem Operator stehen |
// | 12650 | ein Argumentname darf ein Schlüsselwort sein |

// **Nach `name:` darf der Wert auf der nächsten Zeile beginnen.**
//
// # Woher der Fall kommt
//
// Aus dem Initskript von `magic.mse-game`, Zeile 196 — gemessen, nicht
// vermutet (#214):
//
//	word_list_artifact:
//	  "Altmetall,"
//	+ "Attraction,"
//
// **Diese eine Lücke hielt alle 190 Stile auf.** Der Zählauf über das bezogene
// Pack meldete sie 190-mal — einmal je Stil, denn jeder lädt dasselbe
// Spielskript. Was danach kam, hat niemand gesehen: der Auswerter bricht beim
// ersten Zerteilfehler ab.
//
// # Die Regel
//
// Ein Umbruch **hinter dem Doppelpunkt** ist keiner. Dieselbe Regel gilt seit
// jeher hinter `:=` und hinter einem Operator (`stufenoperator`); hier fehlte
// sie, und zwar an beiden Stellen, an denen `name:` vorkommt — in einer
// Abbildung und in einem Aufruf.
func TestNachEinemDoppelpunktDarfDerWertUmbrechen(t *testing.T) {
	faelle := map[string]string{
		"in einer Abbildung": "[a:\n  \"x\", b: \"y\"]",
		"in einem Aufruf":    "f(a:\n  \"x\")",

		// **Und über mehrere Umbrüche**, wie es im Spielskript steht: dort
		// folgen der Zeile mit dem Doppelpunkt eine leere und dann der Wert.
		"über zwei Umbrüche": "[a:\n\n  \"x\"]",
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := msescript.Zerteile(quelle); err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}

// **Ein Doppelpunkt am Ende der Quelle bleibt ein Fehler.**
//
// Die Gegenprobe zur Regel darüber: „der Wert darf umbrechen" heißt nicht „der
// Wert darf fehlen". Ohne diesen Fall führe der Zerteiler über das Ende hinaus
// und meldete eine Stelle, die niemand wiederfindet.
func TestEinDoppelpunktOhneWertBleibtEinFehler(t *testing.T) {
	if _, err := msescript.Zerteile("[a:\n]"); err == nil {
		t.Error("kein Fehler")
	}

	if _, err := msescript.Zerteile("f(a:\n"); err == nil {
		t.Error("kein Fehler am Quellende")
	}
}

// **`mod` ist ein Operator und bindet wie `*`.**
//
// # Der Beleg, ausgezählt
//
// 272 Vorkommen in `M15-Magic-Pack@3f7f8ce/data` — `(input mod 2) == 0`,
// `colors[(to_int(offset) + 0) mod 4]`, `input mod 10`. Die Klammern in den
// echten Vorlagen sagen die Vorrangstufe: um die Addition steht eine, um `mod`
// gegen `==` auch — beides passt zur Stufe von `*` und `/` und zu keiner
// anderen (#214).
//
// # Warum nicht auch `div`
//
// **Weil es nicht belegt ist.** Die 38 Vorkommen von `div` in demselben Pack
// sind sämtlich HTML — `document.createElement("div")` in Exportvorlagen. Was
// nicht belegt ist, wird nicht gebaut, und der Grund gehört aufgeschrieben
// (§11, und die Lehre aus Phase 3 bis 5).
func TestModIstEinOperator(t *testing.T) {
	faelle := map[string]string{
		"einfach":            "10",
		"vor dem Vergleich":  "0",
		"nach dem Vorrang":   "2",
		"über einen Umbruch": "10",
	}

	quellen := map[string]string{
		"einfach":            "310 mod 100",
		"vor dem Vergleich":  "if 4 mod 2 == 0 then 0 else 1",
		"nach dem Vorrang":   "1 + 6 mod 3 + 1",
		"über einen Umbruch": "310 mod\n  100",
	}

	for name, erwartet := range faelle {
		t.Run(name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), quellen[name], msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("%v", err)
			}

			if wert.AlsText() != erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), erwartet)
			}
		})
	}
}

// **Auch eine Zahl zählt als Schlüssel.**
//
// # Der Beleg
//
// `magic.mse-game`, Zeile 8590 des eingebundenen Initskripts — gemessen
// (#214):
//
//	russian_number_map :=
//	[
//	0	: "ноль"
//	1	: "один"
//
// # Warum das dieselbe Klasse ist wie die Zeichenkette
//
// #195 hat hier den **Zeichenkettenschlüssel** aufgenommen, weil
// `magic.mse-game/language` `["Deutsch": …]` schreibt. Der Grund war schon
// damals nicht „Zeichenketten sind erlaubt", sondern: **was fest in der Datei
// steht, ist ein Schlüssel** — es lässt sich lesen, ohne etwas auszurechnen.
// Eine Zahl steht ebenso fest. Ein **Ausdruck** ist weiterhin keiner.
func TestAuchEineZahlIstEinSchluessel(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		"[\n0\t: \"null\"\n1\t: \"eins\"\n].1", msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("%v", err)
	}

	if wert.AlsText() != "eins" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Ein `if` darf rechts von einem Operator stehen.**
//
// # Der Beleg
//
// `magic.mse-game`, Zeile 9726 des eingebundenen Initskripts — gemessen
// (#214):
//
//	"…</sym> or 2 life. " +
//	if multiple then "For each …"
//	else             "If life was paid, …"
//
// # Die Regel
//
// **`if` ist ein Ausdruck**, kein Satzanfang. Bis hierher stand es nur an der
// schwächsten Stufe (`bedingung`), und rechts von einem `+` suchte der
// Zerteiler einen Wert und fand ein Schlüsselwort.
//
// **Es reicht bis ans Ende**, wie in jeder Sprache mit Ausdruck-`if`:
// `"a" + if b then "c" else "d"` verkettet `"a"` mit dem ganzen `if`. Anders
// gälte der `else`-Zweig nur für sich, und die Vorlage schriebe die Hälfte.
func TestEinIfDarfRechtsVonEinemOperatorStehen(t *testing.T) {
	faelle := map[string]string{
		`"a" + if 1 == 1 then "b" else "c"`:         "ab",
		`"a" + if 1 == 2 then "b" else "c"`:         "ac",
		"\"a\" +\nif 1 == 1 then \"b\"\nelse \"c\"": "ab",
	}

	for quelle, erwartet := range faelle {
		t.Run(quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("%v", err)
			}

			if wert.AlsText() != erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), erwartet)
			}
		})
	}
}

// **Auch ein Schlüsselwort darf ein Argumentname sein.**
//
// # Der Beleg, ausgezählt
//
// `number_of_items(in: colors)` — 25 Vorkommen von `(in: ` in
// `M15-Magic-Pack@3f7f8ce/data`, davon eines in `magic.mse-game`, Zeile 12650
// des eingebundenen Initskripts
// (#214).
//
// # Warum das keine Zweideutigkeit ist
//
// `in` ist ein Schlüsselwort — `for x in y`. Als **Name** steht es nur dort, wo
// unmittelbar ein `:` folgt, und diese Form gibt es in keiner Schleife. Zwei
// Merkmale weit vorausgeschaut, wie schon für `f(order: "XY")`.
func TestAuchEinSchluesselwortIstEinArgumentname(t *testing.T) {
	// **Zerteilt, nicht ausgewertet.** `number_of_items` steht nicht in der
	// Bibliothek — was hier zählt, ist die **Form**: ein Schlüsselwort als
	// Argumentname. Wer dafür eine Funktion bräuchte, prüfte die Bibliothek.
	if _, err := msescript.Zerteile(`number_of_items(in: colors)`); err != nil {
		t.Fatalf("%v", err)
	}

	// **Und die Schleife bleibt eine Schleife.** Ohne diese Gegenprobe wäre
	// offen, ob das Vorausschauen die Form `for x in y` zerschneidet.
	schleife, fehler := msescript.AuswertenQuelle(t.Context(),
		`for each x in ["a", "b"] do x`, msescript.Bibliothek().Darueber())
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if schleife.AlsText() == "" {
		t.Error("die Schleife gab nichts")
	}
}

// **Hinter einer öffnenden Klammer darf der Ausdruck umbrechen.**
//
// # Der Beleg
//
// `magic.mse-game`, Zeile 12908 des eingebundenen Initskripts — der Index
// steht auf der nächsten Zeile:
//
//	else	colors[
//	font_color_positions[
//	if		chosen(choice:"horizontal")	then	"horizontal"
//
// # Die Regel, und sie ist dieselbe wie überall
//
// Ein Umbruch **mitten in einem Ausdruck** ist nichts. Hinter `:=`, hinter
// einem Operator, hinter `name:` und in einer Abbildung gilt das längst; beim
// **Index** und beim **Aufruf** fehlte es. Beide öffnen eine Klammer, hinter
// der der Ausdruck weitergeht.
func TestHinterEinerKlammerDarfDerAusdruckUmbrechen(t *testing.T) {
	faelle := map[string]string{
		"Index":           "[\"a\", \"b\"][\n  1]",
		"Index über zwei": "[\"a\", \"b\"][\n\n  1]",
		"Aufruf":          "to_text(\n  \"x\")",
		"Aufruf mit zwei": "replace(\n  \"abc\", match:\n  \"b\", replace: \"x\")",
	}

	erwartet := map[string]string{
		"Index":           "b",
		"Index über zwei": "b",
		"Aufruf":          "x",
		"Aufruf mit zwei": "axc",
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("%v", err)
			}

			if wert.AlsText() != erwartet[name] {
				t.Errorf("%q, erwartet %q", wert.AlsText(), erwartet[name])
			}
		})
	}
}

// **Die Wörter des Schleifenkopfs dürfen Namen sein.**
//
// # Der Beleg
//
// `magic.mse-game`, Zeile 13295 des eingebundenen Initskripts:
//
//	ar_position := {
//	pos := -1
//	for x from 0 to length(in)-1 do (
//	…
//	includes := {ar_position(of:input in:array) != -1}
//
// `in` ist hier ein **Argumentname** des Blocks — gerufen wird
// `ar_position(of:… in:…)` —, und gelesen wird er als `length(in)`.
//
// # Warum das notwendig folgt
//
// Ein Schlüsselwort darf seit dieser Reihe ein **Argumentname** sein
// (`number_of_items(in: …)`). Wer es nicht auch als **Wert** liest, hat eine
// Sprache, in der sich ein Argument übergeben, aber nicht benutzen lässt.
//
// **Nur `in`, und nicht jedes Schlüsselwort.** `if`, `for`, `not` und die
// übrigen beginnen eine Form; wer sie als Namen zuließe, machte
// `if x then y` zweideutig. `in` beginnt keine — es steht ausschließlich
// **hinter** einem Schleifennamen, und dort sucht `schleife` es ausdrücklich.
func TestInDarfEinNameSein(t *testing.T) {
	faelle := map[string]string{
		"in":   `{ length(in) }@(in: ["a", "b", "c"])()`,
		"from": `{ length(from) }@(from: ["a", "b", "c"])()`,
		"to":   `{ length(to) }@(to: ["a", "b", "c"])()`,
	}

	for name, quelle := range faelle {
		wert, err := msescript.AuswertenQuelle(
			t.Context(), quelle, msescript.Bibliothek().Darueber())
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if wert.AlsText() != "3" {
			t.Errorf("%s: %q", name, wert.AlsText())
		}
	}

	// **Und die Schleife bleibt eine Schleife** — die Gegenprobe zur Regel.
	if _, fehler := msescript.Zerteile(`for each x in ["a"] do x`); fehler != nil {
		t.Errorf("die Schleife: %v", fehler)
	}

	// **Ein Schlüsselwort, das eine Form beginnt, bleibt eines.** Ohne diesen
	// Fall stünde die Regel „nur die drei" nur im Kommentar.
	for _, wort := range []string{"then", "else", "each", "do", "not"} {
		if _, fehler := msescript.Zerteile(`length(` + wort + `)`); fehler == nil {
			t.Errorf("`%s` ging als Name durch", wort)
		}
	}
}
