package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine Zeichenkette lässt sich indizieren — sie gibt das Zeichen an der
// Stelle** (#240).
//
// # Herkunft
//
// `ScriptString::getMember` in `src/script/value.cpp:376` von
// twanvl/MagicSetEditor2, Stand 2026-08-16:
//
//	long index;
//	if (name.ToLong(&index) && index >= 0 && (size_t)index < value.size()) {
//	  return to_script(String(1,value[index]));
//	} else {
//	  return delay_error(…);
//	}
//
// # Was daran hing
//
// `color_combination` in `magic-blends.mse-include/new-blends` schreibt
// `base := base_hybrid[shape][color_count]` — eine zweifache Indizierung, deren
// erste Stufe eine Zeichenkette liefert. Hier war das ein Fehler, und damit
// fiel `card color → image` aus.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEineZeichenketteLaesstSichIndizieren(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt string
	}{
		{"das erste Zeichen", `"wubrg"[0]`, "w"},
		{"ein Zeichen in der Mitte", `"wubrg"[2]`, "b"},
		{"das letzte Zeichen", `"wubrg"[4]`, "g"},

		// **Über das Ende hinaus ist ein Fehlerwert**, den `or else` auffängt
		// (`ScriptString::getMember`, `src/script/value.cpp`, Zeilen 377–384,
		// Fassung f9d9356 vom 2021-07-06). Hier stand bis zum 2026-08-17
		// `nil`, und damit fing `or else` den falschen Fall.
		{"über das Ende hinaus", `to_string("wu"[9] or else "")`, ""},
		{"unter null", `to_string("wu"[-1] or else "")`, ""},
		{"kein Zahlenschlüssel", `to_string("wu"["x"] or else "")`, ""},
		{"die leere Zeichenkette", `to_string(""[0] or else "")`, ""},

		// **In Zeichen, nicht in Bytes** — wie überall in dieser Sprache.
		{"mehrbytige Zeichen", `"äöü"[1]`, "ö"},
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

// **Eine Zahl lässt sich nicht indizieren.**
//
// Die Gegenprobe: „Abbildung oder Zeichenkette" ist die Regel, nicht „alles
// geht". Wer sie auf jeden Wert ausdehnte, machte aus einem Tippfehler ein
// stilles `nil`.
func TestEineZahlLaesstSichNichtIndizieren(t *testing.T) {
	// **Benutzt wirft es**, aufgefangen nicht — siehe [msescript.Fehlerwert].
	if _, err := msescript.AuswertenQuelle(
		t.Context(), `12[0] + 1`, msescript.Bibliothek()); err == nil {
		t.Fatal("kein Fehler")
	}

	wert, err := msescript.AuswertenQuelle(
		t.Context(), `to_string(12[0] or else "aufgefangen")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "aufgefangen" {
		t.Errorf("ergibt %q", wert.AlsText())
	}
}
