package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein fehlgeschlagener Zugriff ist ein Wert, kein Abbruch — und `or else`
// fängt ihn.**
//
//	case I_OR_ELSE:
//	  if (at == SCRIPT_ERROR) a = b;
//	  break;
//
// `src/script/context.cpp`, Zeilen 535–537, Fassung f9d9356 vom 2021-07-06.
//
// **`or else` fängt genau den Fehlerwert**, nicht `nil`. Und dieser Wert
// entsteht bei jedem Zugriff, der ins Leere geht:
//
//	ScriptValueP ScriptValue::getMember(const String& name) const {
//	  return delay_error(ScriptErrorNoMember(typeName(), name));
//	}
//
// `src/script/value.cpp`, Zeilen 67–74. Der Fehler wirft erst, wenn jemand den
// Wert **benutzt** (Zeilen 140–167).
//
// # Was das gekostet hat
//
// `flavor_bar_top` in `magic.mse-game` schreibt
// `(style.layout.blocks[1].bottom or else 0) <= 0` — ein Block, den es bei
// einer Karte ohne Flavortext nicht gibt. **406 Befunde** über die 190 Stile
// von `M15-Magic-Pack` lauteten auf diesen einen Zugriff
// (#345).
func TestEinFehlgeschlagenerZugriffIstEinWert(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "ein Glied an nichts",
			quelle:   `to_string([1] or else "aufgefangen")`,
			erwartet: "1",
		},
		{
			name:     "ein Glied an einer Zahl",
			quelle:   `to_string(5.feld or else "aufgefangen")`,
			erwartet: "aufgefangen",
		},
		{
			name:     "ein Glied, das es nicht gibt",
			quelle:   `to_string([a: 1].b.c or else "aufgefangen")`,
			erwartet: "aufgefangen",
		},
		{
			// **`nil` ist kein Fehler** — `or else` lässt es stehen.
			name:     "nichts bleibt nichts",
			quelle:   `to_string(nil or else "aufgefangen")`,
			erwartet: "",
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

// **Wer den Fehlerwert benutzt, bekommt den Fehler.**
//
// Alle Umwandlungen von `ScriptDelayedError` werfen (`src/script/value.cpp`,
// Zeilen 146–156). Ein Zugriff ins Leere ist damit keine stille Null: er wird
// zum Befund, sobald jemand mit ihm rechnet.
//
// Fassung f9d9356 vom 2021-07-06.
func TestWerDenFehlerwertBenutztBekommtDenFehler(t *testing.T) {
	for _, quelle := range []string{`5.feld + 1`, `to_string(5.feld)`, `if 5.feld then 1 else 2`} {
		t.Run(quelle, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
			if err == nil {
				t.Error("erwartet einen Fehler")
			}

			if err != nil && !strings.Contains(err.Error(), "feld") {
				t.Errorf("die Meldung nennt das Glied nicht: %v", err)
			}
		})
	}
}

// **`to_int` auf einen unzahligen Text gibt ebenfalls einen Fehlerwert.**
//
//	} else {
//	  return delay_error(ScriptErrorConversion(str, input->typeName(), _TYPE_("integer")));
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 116–124. **Ein leerer Text ist
// dagegen null** (Zeile 121).
//
// 607 Befunde über die 190 Stile lauteten auf `to_int` — die meisten davon
// stehen hinter einem `or else`
// (#345).
//
// Fassung f9d9356 vom 2021-07-06.
func TestToIntAufUnzahligemTextGibtEinenFehlerwert(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`to_string(to_int("kein Zahlwort") or else 7)`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "7" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "7")
	}

	leer, err := msescript.AuswertenQuelle(t.Context(),
		`to_string(to_int(""))`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if leer.AlsText() != "0" {
		t.Errorf("ein leerer Text ergibt %q, erwartet %q", leer.AlsText(), "0")
	}
}

// **Auch eine Umwandlung in eine Zahl wirft.**
//
// Alle `to*`-Methoden von `ScriptDelayedError` werfen (`src/script/value.cpp`,
// Zeilen 146–156). Geprüft wird hier die, die kein anderer Test betritt: eine
// Maßangabe rechnet über [msescript.Wert.AlsZahl], und dort muss die Meldung
// ankommen, nicht eine Null.
//
// Fassung f9d9356 vom 2021-07-06.
func TestDerFehlerwertWirftAuchAlsZahl(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(), `5.feld`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if !wert.IstFehler() {
		t.Fatalf("%v ist kein Fehlerwert", wert)
	}

	if _, err := wert.AlsZahl(); err == nil {
		t.Error("AlsZahl gibt keine Meldung")
	}

	if _, err := wert.AlsWahrheitswert(); err == nil {
		t.Error("AlsWahrheitswert gibt keine Meldung")
	}
}

// **Ein Wahrheitswert wird zu eins oder null.**
//
//	if (t == SCRIPT_BOOL) {
//	  result = input->toBool() ? 1 : 0;
//	}
//
// `to_int`, `src/script/functions/basic.cpp`, Zeilen 111–112, Fassung f9d9356
// vom 2021-07-06. **Nur in `to_int`**, nicht überall: `ScriptBool` überschreibt
// `toInt` nicht, und der Kommentar dort lautet „bools don't autoconvert to
// int". Die Umwandlung steht in der Funktion, nicht am Typ.
//
// **226 Befunde** über die 190 Stile von `M15-Magic-Pack` — `to_int` auf
// `"true"` und `"false"`
// (#345).
func TestToIntMachtAusEinemWahrheitswertEinsOderNull(t *testing.T) {
	faelle := map[string]string{
		"to_string(to_int(true))":  "1",
		"to_string(to_int(false))": "0",

		// **Die Wörter nicht.** `to_int("true")` geht bei MSE in den
		// Zeichenkettenzweig, und `str.ToLong` scheitert dort — es ergibt
		// einen verzögerten Fehler, keine Eins (Zeilen 116–124). Der
		// Unterschied ist der Typ des Werts, nicht sein Aussehen.
		`to_string(to_int("true") or else "kein Zahlwort")`: "kein Zahlwort",
	}

	for quelle, erwartet := range faelle {
		t.Run(quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), erwartet)
			}
		})
	}
}

// **Ein Fehlerwert, der gerufen wird, bleibt ein Fehlerwert.**
//
//	ScriptValueP eval(Context&, bool openScope) const override {
//	  return delay_error(error);
//	}
//
// `ScriptDelayedError`, `src/script/value.cpp`, Zeile 163. Die Vorlagen
// schreiben `f(x) or else "…"`, wo `f` selbst aus einem Zugriff kommt — bräche
// der Aufruf ab, käme `or else` nie zum Zug. **198 Befunde** über die 190
// Stile (#345).
//
// Fassung f9d9356 vom 2021-07-06.
func TestEinGerufenerFehlerwertBleibtEinFehlerwert(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`to_string([a: 1].b(1) or else "aufgefangen")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "aufgefangen" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "aufgefangen")
	}
}
