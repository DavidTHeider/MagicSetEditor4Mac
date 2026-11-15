package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Fallunterscheidung `case … of` und der Operator `xor`
// (#214).
//
// **Eine eigene Datei**, weil `sprachluecken_test.go` an der Zeilengrenze steht
// (§2) — und weil beides eine **Form** ist und nicht bloß ein Umbruch, der
// durchgehen muss.

// **`case … of` ist eine Form der Sprache.**
//
// # Der Beleg, ausgezählt
//
// Acht Vorkommen in vier Dateien von `M15-Magic-Pack@3f7f8ce/data`;
// `magic.mse-game`, Zeile 13429 des eingebundenen Initskripts:
//
//	token_shape := {
//	case input of
//	"token": true,
//	"emblem": true,
//	else: false;
//	}
//
// Und mit Ausdrücken statt Wahrheitswerten in `card_number_style`:
// `"0001/0099": card_number_mom() + cn_tag + (…)`.
//
// # Die Form
//
// `case <ausdruck> of` — danach Paare `<wert>: <ausdruck>`, durch Komma
// getrennt, und ein `else: <ausdruck>` als Vorgabe. Verglichen wird auf
// **Gleichheit**, wie `==` es tut.
//
// **`of` bleibt daneben ein Name.** Es steht 48-mal als Argumentname und
// 11-mal als Wert (`ar_position(of:of[x], …)`) — ein Schlüsselwort, das eine
// Form nur in ihrer Mitte trägt, darf beides sein (dieselbe Regel wie für `in`,
// `from` und `to`).
func TestCaseIstEineForm(t *testing.T) {
	faelle := map[string]string{
		`case "b" of "a": "eins", "b": "zwei", else: "sonst";`: "zwei",
		`case "x" of "a": "eins", "b": "zwei", else: "sonst";`: "sonst",
		`case 2 of 1: "eins", 2: "zwei", else: "sonst";`:       "zwei",
		"case \"a\" of\n\"a\": \"eins\",\nelse: \"sonst\";":    "eins",
		`"vor " + case "a" of "a": "eins", else: "sonst";`:     "vor eins",
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

// **Ein `case` ohne `of` und ohne `else` wird benannt.**
//
// Die Gegenprobe: „`case` ist eine Form" heißt nicht „irgendetwas hinter `case`
// geht durch". Ohne `else` steht kein Vorgabewert da — und ein `case`, das
// nichts trifft, gäbe sonst stillschweigend nichts.
func TestEinKaputtesCaseWirdBenannt(t *testing.T) {
	faelle := map[string]string{
		"ohne of":   `case "a" "a": 1;`,
		"ohne Wert": `case "a" of "a": ;`,
		"am Ende":   `case "a" of`,
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := msescript.Zerteile(quelle); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ohne Treffer und ohne `else` ist das Ergebnis `nil`.**
//
// MSE-Script kennt keinen Abbruch für diesen Fall, und einen zu erfinden hieße,
// eine Regel zu bauen, die keine Vorlage kennt. `nil` ist, was ein fehlender
// Wert in dieser Sprache sonst auch ist.
func TestEinCaseOhneTrefferGibtNichts(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`case "x" of "a": "eins";`, msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("%v", err)
	}

	if wert.AlsText() != "" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **`xor` ist ein Operator und bindet wie `or`.**
//
// # Der Beleg
//
// `magic.mse-game`, Zeile 16236 des eingebundenen Initskripts:
//
//	if (set.CARDNAME_replacement == "card's type" xor reverse)
//	then (…)
//
// Die Klammer um den Vergleich sagt die Vorrangstufe: `xor` bindet **loser**
// als `==` — dieselbe Stufe wie `or`.
//
// # Und es rechnet beide Seiten
//
// Anders als `and` und `or`: ein „entweder oder" lässt sich nicht abkürzen. Wer
// es wie `or` mit Abkürzung baute, wertete die rechte Seite nie aus, wenn die
// linke zutrifft — und das Ergebnis wäre für die Hälfte aller Fälle falsch.
func TestXorIstEinOperator(t *testing.T) {
	faelle := map[string]string{
		"true xor true":   "false",
		"true xor false":  "true",
		"false xor true":  "true",
		"false xor false": "false",

		// **Loser als `==`**: ohne Klammern liest die Sprache
		// `(1 == 1) xor false`.
		"1 == 1 xor false": "true",
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

// **`xor` braucht Wahrheitswerte — auf beiden Seiten.**
func TestXorBrauchtWahrheitswerte(t *testing.T) {
	for _, quelle := range []string{`"a" xor true`, `true xor "a"`} {
		if _, err := msescript.AuswertenQuelle(
			t.Context(), quelle, msescript.Bibliothek().Darueber()); err == nil {
			t.Errorf("%s: kein Fehler", quelle)
		}
	}
}

// **Jeder Fehlerweg der Fallunterscheidung, einzeln.**
//
// Ein Fehlerpfad, den kein Test betritt, ist eine Behauptung darüber, was im
// Ernstfall geschieht (§3) — und „im Ernstfall" heißt hier: eine Vorlage, die
// jemand halb geschrieben hat.
func TestJederFehlerwegDerFallunterscheidung(t *testing.T) {
	faelle := map[string]string{
		// **`case of …` ist kein Fehler mehr**: `of` ist seit dieser Reihe auch
		// ein Name, und der geprüfte Ausdruck ist dann eben `of`. Was scheitert,
		// ist ein Ausdruck, der keiner ist.
		"der geprüfte Ausdruck ist keiner": "case ] of \"a\": 1;",
		"nach dem Fall fehlt `:`":          `case "x" of "a" 1;`,
		"der Fallwert ist keiner":          `case "x" of then: 1;`,
		"nach `else` fehlt `:`":            `case "x" of "a": 1, else 2;`,
		"der Vorgabewert fehlt":            `case "x" of "a": 1, else: ;`,
		"der Wert eines Falls fehlt":       `case "x" of "a": , "b": 2;`,
		"nach `of` kommt nichts":           `case "x" of }`,
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := msescript.Zerteile(quelle); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein Fall, dessen Schlüssel sich nicht ausrechnen lässt, meldet das.**
//
// Der Schlüssel ist ein Wert und kein Ausdruck — aber ein Name ist ein Wert,
// und ein unbekannter Name scheitert beim Auswerten.
func TestEinUnbekannterFallschluesselWirdGemeldet(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`case "x" of gibtsnicht: 1, else: 2;`, msescript.Bibliothek().Darueber())
	if err == nil {
		t.Error("kein Fehler")
	}
}

// **Und ein geprüfter Ausdruck, der scheitert, kommt durch.**
func TestEinScheiterndesCaseSubjektKommtDurch(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`case gibtsnicht of "a": 1, else: 2;`, msescript.Bibliothek().Darueber())
	if err == nil {
		t.Error("kein Fehler")
	}
}
