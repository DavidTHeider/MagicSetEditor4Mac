package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Umwandlungsregeln von MSE, portiert statt erschlossen
// (#240).
//
// # Herkunft
//
// `src/script/value.cpp` aus twanvl/MagicSetEditor2, Stand 2026-08-16. Die
// Klasse `ScriptNil` (Zeile 441) legt fest, was aus `nil` wird:
//
//	toString() → ""      toDouble() → 0.0
//	toInt()    → 0       toBool()   → false
//
// Seit dem Lizenzwechsel auf GPL-2.0 darf das übernommen werden
// ([0019](../../docs/entscheidungen/0019-gpl-2-0-statt-mit.md)); die Herkunft
// steht laut `AGENTS.md` §11 an der Stelle.
//
// # Warum das die teuerste Zeile des Projekts war
//
// Gemessen am 2026-08-16 über die 190 Stile von `M15-Magic-Pack`: **1.606**
// Feldstile scheiterten an `font size: … * …`, weil eine Seite `nichts` war,
// und **327** an `>`, weil beide Seiten `nichts` waren. Zusammen ein Drittel
// aller Befunde — und nichts davon war ein Fehler in der Vorlage.
//
// Fassung f9d9356 vom 2021-07-06.
func TestNichtsRechnetAlsNull(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt float64
	}{
		// **Der gemessene Hauptfall.** So steht es in `magic-m15.mse-style`:
		// eine Schriftgröße mal einem Faktor, der aus einem leeren Stilwert
		// kommt.
		{"nichts mal Zahl", `nil * 12`, 0},
		{"Zahl mal nichts", `12 * nil`, 0},
		{"Zahl minus nichts", `12 - nil`, 12},
		{"nichts minus Zahl", `nil - 12`, -12},
		{"nichts geteilt durch Zahl", `nil / 12`, 0},

		// Ein Text, der eine Zahl ist, rechnet mit — `ScriptString::toDouble`
		// zerteilt ihn und wirft nur, wenn das misslingt (value.cpp:329).
		{"Text als Zahl", `"12" * 2`, 24},
		{"Kommazahl als Text", `"1.5" * 2`, 3},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Art != msescript.ArtZahlwert {
				t.Fatalf("Art %s — erwartet war eine Zahl", wert.Art)
			}

			if wert.Zahl != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Zahl, fall.erwartt)
			}
		})
	}
}

// **Ein Text, der keine Zahl ist, bleibt ein Fehler** — `ScriptString::toDouble`
// wirft `ScriptErrorConversion` (value.cpp:333).
//
// Die Gegenprobe zur Lockerung darüber: ohne sie hieße „nichts rechnet als
// null" in Wahrheit „alles rechnet als null", und ein vertippter Feldname
// ergäbe stillschweigend 0 statt einer Meldung.
func TestEinTextOhneZahlBleibtEinFehler(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(), `"links" * 2`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler — ein unlesbarer Text rechnete stillschweigend mit")
	}

	if !strings.Contains(err.Error(), "links") {
		t.Errorf("die Meldung nennt den Wert nicht: %v", err)
	}
}

// **Ein Wahrheitswert rechnet nicht mit.** In `value.cpp` steht der Grund als
// Kommentar an `ScriptBool` (Zeile 271): „bools don't autoconvert to int".
//
// Das ist die zweite Gegenprobe: die Lockerung gilt für `nil` und für Text,
// nicht für alles, was sich irgendwie in eine Zahl biegen ließe.
func TestEinWahrheitswertRechnetNichtMit(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(), `true * 2`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler — ein Wahrheitswert rechnete mit")
	}
}

// **Vergleiche rechnen genauso** — MSE führt `<`, `>`, `<=` und `>=` über
// dasselbe Makro `OPERATOR_DI` wie `-` und `*` (context.cpp:436).
func TestVergleichRechnetNichtsAlsNull(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt bool
	}{
		// **Der gemessene Fall**: 327-mal `top: … > …` mit zwei leeren Seiten.
		{"nichts größer nichts", `nil > nil`, false},
		{"nichts kleiner Zahl", `nil < 5`, true},
		{"Zahl größer nichts", `5 > nil`, true},
		{"nichts gleich groß", `nil >= nil`, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Art != msescript.ArtWahrheitswert || wert.Wahrheit != fall.erwartt {
				t.Errorf("%+v, erwartet war %v", wert, fall.erwartt)
			}
		})
	}
}

// **`+` ist überladen, und `nil` ist sein neutrales Element** (context.cpp:465).
//
//	if (at == SCRIPT_NIL) { a = b; } else if (bt == SCRIPT_NIL) { /* a = a */ }
//
// Der Unterschied zur Textverkettung ist sichtbar: `nil + 5` ist die **Zahl**
// fünf, nicht die Zeichenkette `"5"`. Wer das übersieht, bekommt aus
// `left: nil + 5` einen Text, und die nächste Rechnung darüber scheitert.
func TestNichtsIstDasNeutraleElementVonPlus(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		art    msescript.WertArt
		text   string
	}{
		{"nichts plus Zahl bleibt Zahl", `nil + 5`, msescript.ArtZahlwert, "5"},
		{"Zahl plus nichts bleibt Zahl", `5 + nil`, msescript.ArtZahlwert, "5"},
		{"nichts plus Text bleibt Text", `nil + "a"`, msescript.ArtTextwert, "a"},
		{"Text plus nichts bleibt Text", `"a" + nil`, msescript.ArtTextwert, "a"},
		{"nichts plus nichts bleibt nichts", `nil + nil`, msescript.ArtNichts, ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Art != fall.art {
				t.Errorf("Art %s, erwartet war %s", wert.Art, fall.art)
			}

			if wert.AlsText() != fall.text {
				t.Errorf("%q, erwartet war %q", wert.AlsText(), fall.text)
			}
		})
	}
}

// **`nil` als Bedingung eines Textes** — `ScriptString::toBool` kennt genau
// fünf Wörter und wirft für alles andere (value.cpp:345):
//
//	"yes", "true"           → wahr
//	"no", "false", ""       → falsch
//	sonst                   → Fehler
//
// Gemessen: 229 Feldstile scheiterten an einer Zeichenkette als Bedingung. Ein
// Teil davon ist der **leere** Text — ein Feld, das niemand ausgefüllt hat.
func TestEinTextAlsBedingung(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt string
	}{
		{"der leere Text ist falsch", `if "" then "ja" else "nein"`, "nein"},
		{"yes", `if "yes" then "ja" else "nein"`, "ja"},
		{"true", `if "true" then "ja" else "nein"`, "ja"},
		{"no", `if "no" then "ja" else "nein"`, "nein"},
		{"false", `if "false" then "ja" else "nein"`, "nein"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartt {
				t.Errorf("%q, erwartet war %q", wert.AlsText(), fall.erwartt)
			}
		})
	}
}

// **Und ein anderer Text bleibt ein Fehler.** Die Gegenprobe: `toBool` wirft
// dort, statt „nicht leer, also wahr" anzunehmen.
func TestEinBeliebigerTextIstKeineBedingung(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`if "vielleicht" then 1 else 2`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler — ein beliebiger Text galt als Bedingung")
	}
}

// **`nil` gerufen gibt sein `input` zurück** (value.cpp:456):
//
//	ScriptValueP eval(Context& ctx, bool) const override {
//	  // nil(input) == input
//	  return ctx.getVariable(SCRIPT_VAR_input);
//	}
//
// Gemessen: 168 Feldstile scheiterten mit „hier wird nichts gerufen". In den
// Vorlagen steht `{ module_card() }` — ein Modulname, den ein anderes Paket
// setzt. Fehlt das Paket, ist der Name `nil`, und MSE zeichnet dann schlicht
// nichts, statt die Karte abzubrechen.
func TestNichtsGerufenGibtDieEingabeZurueck(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`nil()`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Art != msescript.ArtNichts {
		t.Errorf("Art %s — ohne `input` ist das Ergebnis nichts", wert.Art)
	}
}
