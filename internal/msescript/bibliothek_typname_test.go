package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`type_name` nennt die Art eines Werts — mit MSEs Namen.**
//
//	SCRIPT_FUNCTION(type_name) {
//	  SCRIPT_PARAM_C(ScriptValueP, input);
//	  SCRIPT_RETURN(input->typeName());
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 231–234, Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum die Namen englisch sind und nicht deutsch
//
// Die Vorlagen **vergleichen** sie: `if type_name(x) != type_name(nil)`
// (`magic-m15-split-fusable.mse-style`, Zeile 72). Für den Vergleich reichte
// jeder Name — aber `magic.mse-game` schreibt an anderer Stelle
// `type_name(input) == "nil"`, und dann zählt die Zeichenkette selbst.
//
// **Eine Zeichenkette nennt ihren Inhalt mit:** `string ("abc")`, gekürzt nach
// dreißig Zeichen (`src/script/value.cpp`, Zeile 324). 113 Befunde über die 190
// Stile von `M15-Magic-Pack`
// (#270).
func TestTypeNameNenntDieArtEinesWerts(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"eine ganze Zahl", "type_name(5)", "integer"},
		{"eine gebrochene Zahl", "type_name(5.5)", "double"},
		{"ein Wahrheitswert", "type_name(true)", "boolean"},
		{"nichts", "type_name(nil)", "nil"},
		{"eine Funktion", "type_name({ 1 })", "function"},
		{"eine Abbildung", "type_name([1, 2])", "collection"},

		// **Der Inhalt gehört zum Namen** — und wird nach dreißig Zeichen
		// abgeschnitten.
		{"eine Zeichenkette", `type_name("abc")`, `string ("abc")`},
		{
			"eine lange Zeichenkette",
			`type_name("0123456789012345678901234567890123456789")`,
			`string ("012345678901234567890123456789...")`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **`reverse` dreht eine Zeichenkette um.**
//
//	SCRIPT_FUNCTION(reverse) {
//	  SCRIPT_PARAM_C(String, input);
//	  SCRIPT_RETURN(reverse_string(input));
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 333–336.
//
// **In Zeichen, nicht in Bytes** — ein Kartenname mit Umlaut gäbe sonst
// Bruchstücke. 159 Befunde über die 190 Stile
// (#270).
//
// Fassung f9d9356 vom 2021-07-06.
func TestReverseDrehtEineZeichenketteUm(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"ein Wort", `reverse("abc")`, "cba"},
		{"mit Umlaut", `reverse("Öl")`, "lÖ"},
		{"leer", `reverse("")`, ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Beide brauchen genau ein Argument.**
//
// MSE liest `input` als Pflichtparameter; ein Aufruf ohne Argument ist ein
// Fehler und keine leere Antwort (§8).
func TestTypeNameUndReverseBrauchenEinArgument(t *testing.T) {
	for _, quelle := range []string{"type_name()", "reverse()"} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek()); err == nil {
				t.Error("erwartet einen Fehler")
			}
		})
	}
}

// **Eine Farbe heißt `color`.**
//
// `ScriptColor::typeName`, `src/script/value.cpp`, Zeile 402.
//
// Fassung f9d9356 vom 2021-07-06.
func TestTypeNameEinerFarbe(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), "type_name(rgb(1,2,3))", msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "color" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "color")
	}
}

// **Zwei Stücke der Bibliothek stehen dem Aufrufer offen.**
//
// `internal/service` setzt eigene eingebaute Funktionen in die Umgebung —
// `exists_in_package` und `symbol_variation` (§2: die Bibliothek trägt die
// Sprache, was die Umgebung darüber hinaus weiß, setzt der Aufrufer hinein).
// Dafür braucht es zwei Dinge von hier: eine Meldung mit der Stelle des Aufrufs
// und eine Farbe als Wert
// (#323).
func TestDieBibliothekOeffnetZweiStueckeFuerDenAufrufer(t *testing.T) {
	fehler := msescript.Ruf{Zeile: 7, Spalte: 3}.FehlerAn("so nicht")
	if fehler.Zeile != 7 || fehler.Spalte != 3 || fehler.Meldung != "so nicht" {
		t.Errorf("der Fehler ist %+v", fehler)
	}

	farbe := msescript.Farbwert(msescript.Farbe{R: 1, G: 2, B: 3, A: 4})

	gelesen, istFarbe := farbe.AlsFarbe()
	if !istFarbe || gelesen != (msescript.Farbe{R: 1, G: 2, B: 3, A: 4}) {
		t.Errorf("die Farbe ist %+v (%v)", gelesen, istFarbe)
	}
}
