package msescript_test

// **`trace`, `warning` und `error`** — Werkstück A von Phase 4.9
// (#1324).
//
// Am Orakel gemessen (Fassung f9d9356 vom 2021-07-06, am 2026-09-02):
//
//	to_string(trace("y"))                      → "y"   (und ein Hinweis daneben)
//	to_string(warning("x", condition: false))  → ""    (nil, keine Meldung)
//
// Die Rückgabe von `warning` und `error` ist `nil`; `trace` gibt seine Eingabe
// zurück. `src/script/functions/basic.cpp`, Zeilen 23–65, Fassung f9d9356
// vom 2021-07-06.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Was gemeldet wurde, ist danach lesbar.**
//
// # Warum an der Umgebung und nicht als zweiter Rückgabewert
//
// MSE2 legt seine Meldungen in eine Warteschlange, die die Oberfläche leert
// (`queue_message`, `src/script/functions/basic.cpp`, Zeilen 28, 37, 62,
// Fassung f9d9356 vom 2021-07-06).
// Hier reicht die Umgebung ohnehin durch jede Auswertung — ein zweiter
// Rückgabewert änderte zehn Aufrufstellen, ohne mehr zu können.
func TestMeldungenStehenAnDerUmgebung(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`trace("hallo") + warning("Vorsicht") + error("kaputt")`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	// **`trace` gibt seine Eingabe zurück**, die beiden anderen nichts.
	if wert.AlsText() != "hallo" {
		t.Errorf("herauskam %q, erwartet %q", wert.AlsText(), "hallo")
	}

	gemeldet := umgebung.Meldungen()
	if len(gemeldet) != 3 {
		t.Fatalf("%d Meldungen, erwartet 3", len(gemeldet))
	}

	for i, erwartet := range []struct {
		art  msescript.Meldungsart
		text string
	}{
		{msescript.MeldungHinweis, "hallo"},
		{msescript.MeldungWarnung, "Vorsicht"},
		{msescript.MeldungFehler, "kaputt"},
	} {
		if gemeldet[i].Art != erwartet.art {
			t.Errorf("Meldung %d ist %q, erwartet %q",
				i, gemeldet[i].Art, erwartet.art)
		}

		if gemeldet[i].Text != erwartet.text {
			t.Errorf("Meldung %d sagt %q, erwartet %q",
				i, gemeldet[i].Text, erwartet.text)
		}
	}
}

// **`condition:` entscheidet, ob gemeldet wird.**
//
//	SCRIPT_PARAM_DEFAULT_C(bool, condition, true);
//	if (condition) { queue_message(…); }
//
// Zeilen 34–38. **Die Vorgabe ist wahr** — wer nichts sagt, meldet.
func TestConditionEntscheidetObGemeldetWird(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name   string
		quelle string
		zahl   int
	}{
		{"ohne condition wird gemeldet", `warning("x")`, 1},
		{"mit condition: true wird gemeldet", `warning("x", condition: true)`, 1},
		{"mit condition: false nicht", `warning("x", condition: false)`, 0},
		{"und dasselbe für error", `error("x", condition: false)`, 0},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			umgebung := msescript.Bibliothek().Darueber()

			if _, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, umgebung); err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if gemeldet := umgebung.Meldungen(); len(gemeldet) != fall.zahl {
				t.Errorf("%d Meldungen, erwartet %d", len(gemeldet), fall.zahl)
			}
		})
	}
}

// **Eine Meldung trägt ihre Stelle.**
//
// `msescript` gibt für jeden Fehler Zeile und Spalte; eine Meldung, die das
// nicht täte, wäre die schlechtere Auskunft im selben Werkzeug.
func TestEineMeldungTraegtIhreStelle(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()

	if _, err := msescript.AuswertenQuelle(t.Context(),
		"1 + 1\nwarning(\"spät\")", umgebung); err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	gemeldet := umgebung.Meldungen()
	if len(gemeldet) != 1 {
		t.Fatalf("%d Meldungen, erwartet 1", len(gemeldet))
	}

	if gemeldet[0].Zeile != 2 {
		t.Errorf("die Meldung steht in Zeile %d, erwartet 2", gemeldet[0].Zeile)
	}
}

// **Ein unbrauchbares `condition:` ist ein Fehler.**
func TestEinUnbrauchbaresConditionIstEinFehler(t *testing.T) {
	t.Parallel()

	_, err := msescript.AuswertenQuelle(t.Context(),
		`warning("x", condition: [1, 2])`, msescript.Bibliothek().Darueber())

	if err == nil {
		t.Fatal("der Aufruf glückte, erwartet ein Fehler")
	}

	if !strings.Contains(err.Error(), "condition") {
		t.Errorf("die Meldung %q nennt condition: nicht", err)
	}
}

// **Ohne Sammler geht eine Meldung verloren — und hält nichts an.**
//
// Das ist derselbe Fall wie im Kopf von `Ruf`: wer eine eingebaute Funktion
// von Hand ruft, gibt keine Umgebung mit. Die Auswertung darf daran nicht
// scheitern.
func TestOhneSammlerGehtEineMeldungVerloren(t *testing.T) {
	t.Parallel()

	// **Die Bibliothek selbst sammelt nicht** — erst `Darueber` legt einen an.
	wert, err := msescript.AuswertenQuelle(t.Context(), `trace("hallo")`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "hallo" {
		t.Errorf("herauskam %q, erwartet %q", wert.AlsText(), "hallo")
	}

	if gemeldet := msescript.Bibliothek().Meldungen(); gemeldet != nil {
		t.Errorf("die Bibliothek sammelte %d Meldungen, erwartet keine",
			len(gemeldet))
	}
}

// **Und ein zweiter `Darueber` legt keinen zweiten Sammler an.**
//
// Sonst käme eine Meldung aus einem Funktionsrumpf nicht oben an, sondern
// bliebe in der Ebene liegen, in der sie entstand.
func TestEinZweiterAufbauSammeltInDenselbenTopf(t *testing.T) {
	t.Parallel()

	unten := msescript.Bibliothek().Darueber()
	oben := unten.Darueber()

	if _, err := msescript.AuswertenQuelle(
		t.Context(), `warning("x")`, oben); err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if gemeldet := unten.Meldungen(); len(gemeldet) != 1 {
		t.Errorf("unten stehen %d Meldungen, erwartet 1", len(gemeldet))
	}
}

// **Und die Fehlerwege** (§3).
func TestMeldenOhneEingabe(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"trace", "warning", "error"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), name+"()",
				msescript.Bibliothek().Darueber())
			if err == nil {
				t.Fatal("der Aufruf glückte, erwartet ein Fehler")
			}

			if !strings.Contains(err.Error(), name) {
				t.Errorf("die Meldung %q nennt %q nicht", err, name)
			}
		})
	}
}
