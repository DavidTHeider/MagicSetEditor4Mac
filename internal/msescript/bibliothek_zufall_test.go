package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die Zufallsfunktionen**
// (#655).
//
//	SCRIPT_FUNCTION(random_int) {
//	  SCRIPT_PARAM_DEFAULT_C(int, begin, 0);
//	  SCRIPT_PARAM_C(        int, end);
//	  SCRIPT_RETURN( rand() % (end - begin) + begin );
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 248–263, Fassung f9d9356 vom
// 2021-07-06.
//
// **Geprüft wird der Bereich, nicht der Wert** — eine gewürfelte Zahl hat
// keinen erwarteten. Der Lauf geht deshalb oft genug, dass ein Ergebnis
// außerhalb der Grenzen auffiele.
func TestZufallsfunktionen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name        string
		quelle      string
		mindestens  float64
		unterhalb   float64
		verschieden bool
	}{
		{
			name:        "random_int liegt im halboffenen Bereich",
			quelle:      `random_int(begin: 10, end: 20)`,
			mindestens:  10,
			unterhalb:   20,
			verschieden: true,
		},
		// **`begin` ist vorbelegt, `end` nicht** — `random_int(end: 3)` gibt
		// 0, 1 oder 2.
		{
			name:        "ohne begin fängt es bei null an",
			quelle:      `random_int(end: 3)`,
			mindestens:  0,
			unterhalb:   3,
			verschieden: true,
		},
		{
			name:        "random_real liegt zwischen null und eins",
			quelle:      `random_real()`,
			mindestens:  0,
			unterhalb:   1,
			verschieden: true,
		},
		{
			name:        "random_real nimmt Grenzen",
			quelle:      `random_real(begin: 5, end: 6)`,
			mindestens:  5,
			unterhalb:   6,
			verschieden: true,
		},
		// **Ein Wahrheitswert ist null oder eins** — und beides kommt vor.
		{
			name:        "random_boolean gibt beides",
			quelle:      `if random_boolean() then 1 else 0`,
			mindestens:  0,
			unterhalb:   2,
			verschieden: true,
		},
		// **Eine Wahrscheinlichkeit von null gibt nie wahr.**
		{
			name:       "random_boolean(0) ist immer falsch",
			quelle:     `if random_boolean(0) then 1 else 0`,
			mindestens: 0,
			unterhalb:  1,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gesehen := map[float64]bool{}

			for range 200 {
				wert, err := msescript.AuswertenQuelle(
					t.Context(), fall.quelle, msescript.Bibliothek().Darueber())
				if err != nil {
					t.Fatalf("unerwarteter Fehler: %v", err)
				}

				if wert.Art != msescript.ArtZahlwert {
					t.Fatalf("Art = %s, erwartet eine Zahl", wert.Art)
				}

				if wert.Zahl < fall.mindestens || wert.Zahl >= fall.unterhalb {
					t.Fatalf("Ergebnis %v liegt nicht in [%v, %v)",
						wert.Zahl, fall.mindestens, fall.unterhalb)
				}

				gesehen[wert.Zahl] = true
			}

			// **Ein Würfel, der immer dasselbe zeigt, ist keiner.** Ohne diese
			// Zusicherung bliebe der Test grün, gäbe die Funktion stets die
			// untere Grenze zurück (§10).
			if fall.verschieden && len(gesehen) < 2 {
				t.Errorf("in 200 Läufen kam nur %v heraus", gesehen)
			}
		})
	}
}

// **Ein leerer Bereich meldet sich, statt abzubrechen.**
//
// `rand() % (end - begin)` ist in C++ bei `end == begin` undefiniert; in Go
// bricht die Division ab. Eine Vorlage darf das Programm nicht anhalten (§8).
func TestZufallszahlOhneBereich(t *testing.T) {
	t.Parallel()

	for _, quelle := range []string{
		`random_int(begin: 5, end: 5)`,
		`random_int(begin: 9, end: 3)`,
	} {
		t.Run(quelle, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber())
			if err == nil {
				t.Fatal("erwartet wurde ein Fehler, es kam keiner")
			}

			if !strings.Contains(err.Error(), "random_int") {
				t.Errorf("die Meldung nennt die Funktion nicht: %v", err)
			}
		})
	}
}

// **Was keine Zahl ist, wird benannt.**
//
// Die Meldung nennt Funktion und Argument. „random_int: `end` …" sagt, wo
// nachzusehen ist; „das ist keine Zahl" sagte es nicht.
func TestZufallszahlMitUnzahl(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		quelle string
		nennt  string
	}{
		{quelle: `random_int(begin: "x", end: 3)`, nennt: "random_int: `begin`"},
		{quelle: `random_int(begin: 0, end: "x")`, nennt: "random_int: `end`"},
		{quelle: `random_real(begin: "x")`, nennt: "random_real: `begin`"},
		{quelle: `random_real(end: "x")`, nennt: "random_real: `end`"},
		{quelle: `random_boolean("x")`, nennt: "random_boolean:"},
	}

	for _, fall := range faelle {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek().Darueber())
			if err == nil {
				t.Fatal("erwartet wurde ein Fehler, es kam keiner")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.nennt, err)
			}
		})
	}
}
