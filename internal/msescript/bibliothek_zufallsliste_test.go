package msescript_test

// **Die drei Zufallsfunktionen über Listen** — Werkstück A von Phase 4.9
// (#1324).
//
// # Warum die erwarteten Werte hier Eigenschaften sind
//
// **Weil das Ergebnis gewürfelt ist.** Ein Test, der eine bestimmte Reihenfolge
// erwartet, prüfte den Würfel und nicht die Funktion. Geprüft wird deshalb,
// was **immer** gilt: die Länge, die Zugehörigkeit, die Vollständigkeit.
//
// Am Orakel gemessen (Fassung f9d9356 vom 2021-07-06, am 2026-09-02) — die
// Werte wechseln je Lauf, die Gestalt nicht:
//
//	random_shuffle(["a","b","c"])                → ["c","a","b"]
//	random_select(["a","b","c"])                 → "b"
//	random_select_many(count: 2, ["a","b","c"])  → ["b","c"]

import (
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// glieder gibt die Glieder einer Abbildung als Text, sortiert.
func glieder(t *testing.T, quelle string) []string {
	t.Helper()

	wert, err := msescript.AuswertenQuelle(t.Context(), quelle,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	gefunden := make([]string, 0, len(wert.Eintraege))
	for _, eintrag := range wert.Eintraege {
		gefunden = append(gefunden, eintrag.Wert.AlsText())
	}

	sort.Strings(gefunden)

	return gefunden
}

// **`random_shuffle` gibt dieselben Glieder in anderer Reihenfolge.**
//
//	ScriptValueP it = input->makeIterator();
//	while (ScriptValueP v = it->next()) { ret->value.push_back(v); }
//	shuffle(ret->value.begin(), ret->value.end());
//
// `src/script/functions/basic.cpp`, Zeilen 604–615, Fassung f9d9356 vom
// 2021-07-06.
func TestZufallsmischungBehaeltAlleGlieder(t *testing.T) {
	t.Parallel()

	gemischt := glieder(t, `random_shuffle(["a", "b", "c", "d"])`)

	if len(gemischt) != 4 {
		t.Fatalf("%d Glieder, erwartet 4", len(gemischt))
	}

	if strings.Join(gemischt, "") != "abcd" {
		t.Errorf("gemischt kamen %v, erwartet dieselben vier", gemischt)
	}
}

// **`random_select` gibt genau eines — und zwar eines der Glieder.**
func TestZufallswahlGibtEinGlied(t *testing.T) {
	t.Parallel()

	// **Hundertmal**, damit nicht ein glücklicher Wurf die Zusage trägt.
	for range 100 {
		wert, err := msescript.AuswertenQuelle(t.Context(),
			`random_select(["a", "b", "c"])`, msescript.Bibliothek())
		if err != nil {
			t.Fatalf("auswerten: %v", err)
		}

		// **Eines der drei und nichts sonst.** `strings.Contains` herum
		// gedreht wäre hier die Frage „steht abc in der Antwort", und das
		// ist eine andere.
		if gewaehlt := wert.AlsText(); gewaehlt != "a" && gewaehlt != "b" &&
			gewaehlt != "c" {
			t.Fatalf("gewählt wurde %q, erwartet a, b oder c", gewaehlt)
		}
	}
}

// **`random_select_many` gibt so viele, wie verlangt.**
func TestZufallsauswahlGibtDieVerlangteZahl(t *testing.T) {
	t.Parallel()

	gewaehlt := glieder(t, `random_select_many(["a", "b", "c", "d"], count: 2)`)

	if len(gewaehlt) != 2 {
		t.Fatalf("%d Glieder, erwartet 2", len(gewaehlt))
	}

	// **Ohne `replace:` kommt keines doppelt** — im Original wird gemischt
	// und dann abgeschnitten.
	if gewaehlt[0] == gewaehlt[1] {
		t.Errorf("zweimal %q — ohne replace: darf sich keines wiederholen",
			gewaehlt[0])
	}
}

// **Mit `replace: true` darf sich eines wiederholen** — und dann geht auch
// mehr, als die Liste hat.
//
//	bool with_replace = replace && replace->type() != SCRIPT_FUNCTION
//	                    && replace->toBool();
//
// Zeilen 630–631.
func TestZufallsauswahlMitZuruecklegen(t *testing.T) {
	t.Parallel()

	gewaehlt := glieder(t,
		`random_select_many(["a"], count: 3, replace: true)`)

	if len(gewaehlt) != 3 {
		t.Fatalf("%d Glieder, erwartet 3", len(gewaehlt))
	}

	if strings.Join(gewaehlt, "") != "aaa" {
		t.Errorf("gewählt kamen %v, erwartet dreimal a", gewaehlt)
	}
}

// **`replace:` mit einer Funktion heißt „ohne Zurücklegen".**
//
//	bool with_replace = replace && replace->type() != SCRIPT_FUNCTION && …
//
// Zeilen 630–631. Die Prüfung steht dort, weil derselbe Name bei `replace`
// eine Funktion trägt — wer sie hier mitgibt, meint etwas anderes.
func TestEineFunktionAlsReplaceHeisstOhneZuruecklegen(t *testing.T) {
	t.Parallel()

	_, err := msescript.AuswertenQuelle(t.Context(),
		`random_select_many(["a"], count: 3, replace: { input })`,
		msescript.Bibliothek())

	if err == nil {
		t.Fatal("drei aus einem Glied glückten — mit einer Funktion als " +
			"replace: gilt »ohne Zurücklegen«, und dann geht das nicht")
	}
}

// **Und die Fehlerwege stehen im Original** (§3).
//
//	throw ScriptError(_("Can not select a random item from an empty collection"));
//	throw ScriptError(String::Format(
//	    _("Can not select %d items from a collection conaining only %d items"), …));
//
// Zeilen 622 und 643. Der Tippfehler „conaining" steht dort; hier steht die
// Meldung auf Deutsch (§9), und sie nennt dieselben Zahlen.
func TestZufallUeberListenMeldetWasNichtGeht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, nennt string }{
		{"wählen aus dem Leeren", `random_select([])`, "leer"},
		{
			"viele wählen aus dem Leeren",
			`random_select_many([], count: 1, replace: true)`, "leer",
		},
		{
			"mehr wollen, als dasteht",
			`random_select_many(["a"], count: 2)`, "2",
		},
		{"mischen ohne Eingabe", `random_shuffle()`, "random_shuffle"},
		{"wählen ohne Eingabe", `random_select()`, "random_select"},
		{"viele wählen ohne count", `random_select_many(["a"])`, "count"},
		{
			"viele wählen mit unbrauchbarem count",
			`random_select_many(["a"], count: [1])`, "count",
		},
		{
			"viele wählen ohne Eingabe",
			`random_select_many(count: 1)`, "random_select_many",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil {
				t.Fatal("der Aufruf glückte, erwartet ein Fehler")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", err, fall.nennt)
			}
		})
	}
}

// **Ein negatives `count:` bringt den Prozess nicht um.**
//
// Gefunden vom Zweitblick über die Phasen 4.0 bis 4.9, Befund 2. Vorher
// erreichte der Wert entweder `glieder[:anzahl]` — ein Slice-Grenzfehler — oder
// bei `replace:` ein `make` mit negativer Kapazität. Beides ist eine Panik,
// und eine Panik beendet den Prozess.
//
// **MSE2 stirbt daran ebenfalls** — am Orakel gemessen am 2026-09-03:
// `1+1` gibt Rückgabewert 0, `random_select_many(count: -1, …)` gibt 1 und
// keine Antwort. Im Quelltext ist zu sehen, warum: `ret->value.resize(count)`
// mit negativem `count` (`src/script/functions/basic.cpp`, Zeile 653, Fassung
// f9d9356 vom 2021-07-06).
//
// **Das entschuldigt die Panik nicht.** §8 sagt, fremde Daten sind der
// Normalfall: ein Set, von dem 300 Karten lesbar sind und zwei nicht, wird
// mit 300 Karten geöffnet. Ein Skript, das den Prozess beendet, macht aus
// zwei unlesbaren Karten ein geschlossenes Fenster.
func TestEinNegativesCountBeendetNichtDenProzess(t *testing.T) {
	for name, quelle := range map[string]string{
		"ohne Zurücklegen": `random_select_many(count: -1, ["a", "b"])`,
		"mit Zurücklegen":  `random_select_many(count: -1, replace: true, ["a", "b"])`,
		"weit im Minus":    `random_select_many(count: -9999, ["a", "b"])`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek())
			if err == nil {
				t.Fatal("kein Fehler — erwartet war eine Abweisung")
			}

			// **Die Meldung nennt die Zahl.** Wer sie liest, sucht sonst in
			// einem Skript nach einer Stelle, die niemand benannt hat (§8).
			if !strings.Contains(err.Error(), "-1") &&
				!strings.Contains(err.Error(), "-9999") {
				t.Errorf("die Meldung nennt die Zahl nicht: %v", err)
			}
		})
	}
}

// **Eine sehr große Zahl bringt den Prozess nicht um** — Befund 1 des
// Nachlaufs über die Phasen 4.0 bis 4.9.
//
// # Was der erste Anlauf getroffen hat, und was nicht
//
// Abgewiesen wurden **negative** Zahlen. Mit `replace: true` geht der Wert
// aber ungeprüft als Kapazität an `make` — ein `count` von 10^12 gibt
// `makeslice: cap out of range` oder erschöpft den Speicher. **Dieselbe
// Klasse**: eine Größenangabe aus einem fremden Skript darf den Prozess nicht
// umbringen.
//
// Ohne `replace` gibt es die Schranke seit je (`anzahl > len(glieder)`); mit
// Zurücklegen fehlte sie, weil dort jede Zahl fachlich denkbar ist.
//
// # Warum eine Zahl und nicht „so viel der Speicher hergibt"
//
// **MSE2 begrenzt nicht** (`src/script/functions/basic.cpp`, Zeile 627,
// Fassung f9d9356 vom 2021-07-06) — es stirbt daran ebenso. Dieses Programm
// ist an dieser Stelle bewusst strenger, und dann braucht es eine Zahl.
//
// **Gemessen am 2026-09-03**: im vollen Pack kommt `random_select_many`
// **null**-mal vor. Es gibt also keine Vorlage, an der sich eine Grenze
// messen ließe; gewählt ist eine, die jede denkbare Ziehung durchlässt und
// jeden Vertipper fängt.
func TestEineSehrGrosseZahlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		zahl  string
		nennt string
	}{
		{"an der Grenze", "100000", ""},
		{"eins darüber", "100001", "höchstens"},
		{"unfassbar viel", "999999999999", "höchstens"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(),
				`random_select_many(count: `+fall.zahl+`, replace: true, ["a", "b"])`,
				msescript.Bibliothek())

			if fall.nennt == "" {
				if err != nil {
					t.Fatalf("abgewiesen: %v", err)
				}

				if len(wert.Eintraege) != 100000 {
					t.Errorf("%d Glieder, erwartet 100000", len(wert.Eintraege))
				}

				return
			}

			if err == nil {
				t.Fatal("kommt durch")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("der Fehler lautet %q und nennt nicht %q",
					err.Error(), fall.nennt)
			}
		})
	}
}
