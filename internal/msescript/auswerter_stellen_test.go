// Die Prüfungen über die **Stelle** eines Laufzeitfehlers.
//
// Von `auswerter_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, welchen
// Wert ein Ausdruck ergibt, hier, wohin der Finger zeigt, wenn keiner entsteht.
package msescript_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Jeder Laufzeitfehler nennt die Stelle im Skript.
//
// **Das ist die dritte Zusage aus #23**, und sie ist der Grund, warum jeder
// Knoten seine Stelle trägt: ein Stil hat 1100 Zeilen, und „ein Name ist nicht
// bekannt" ohne Zeilennummer schickt jemanden auf die Suche.
func TestJederLaufzeitfehlerNenntSeineStelle(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
		spalte int
	}{
		{"ein unbekannter Name", `gibtsnicht`, 1},
		// **Eine Zahl ist seit #216 eine gültige Bedingung** — der Fall steht
		// jetzt in `bibliothek_stelle_test.go`. Hier bleibt der, für den keine
		// Fundstelle da ist: eine Zeichenkette.
		{"eine Bedingung ohne Wahrheitswert", `if "ja" then 2 else 3`, 4},
		{"nicht auf einer Zahl", `not 1`, 1},
		{"ein Vorzeichen auf einer Zeichenkette", `-"a"`, 1},
		{"und mit einer Zahl links", `1 and true`, 1},
		{"und mit einer Zahl rechts", `true and 1`, 10},
		// **Ein Zugriff ins Leere steht hier nicht mehr.** Er wirft nicht
		// sofort, sondern gibt einen Fehlerwert, den `or else` auffängt —
		// `fehlerwert_test.go` führt ihn (siehe [msescript.Fehlerwert]).
		{"eine Zahl rufen", `1()`, 1},
		{"an eine Zahl binden", `1@(a: 1)`, 1},
		{"eine Funktion verketten", `"a" + {1}`, 1},
		{"rechnen mit Zeichenketten", `"a" - 1`, 1},
		// **`1 / 0` steht hier nicht mehr.** Es ist seit
		// #555
		// kein Fehler, sondern `inf` — am Orakel bestätigt; siehe
		// `nullteilung_test.go`.
		{"Arten der Größe nach vergleichen", `1 < "a"`, 1},
		{"ein Fehler in der linken Seite", `gibtsnicht + 1`, 1},
		{"ein Fehler in der rechten Seite", `1 + gibtsnicht`, 5},
		{"ein Fehler im then-Zweig", `if true then gibtsnicht else 1`, 14},
		{"ein Fehler unter einem Minus", `-gibtsnicht`, 2},
		{"ein Fehler in einer Abbildung", `[a: gibtsnicht]`, 5},
		{"ein Fehler im Index", `[a: 1][gibtsnicht]`, 8},
		{"ein Fehler im indizierten Wert", `gibtsnicht["a"]`, 1},
		{"ein Fehler vor einem Punkt", `gibtsnicht.a`, 1},
		{"ein Fehler in einer Zuweisung", `a := gibtsnicht`, 6},
		{"ein Fehler in einer Folge", "1\ngibtsnicht", 1},
		{"ein Fehler im Aufrufziel", `gibtsnicht()`, 1},
		{"ein Fehler in der Bedingung", `if gibtsnicht then 1 else 2`, 4},
		{"ein Fehler links von und", `gibtsnicht and true`, 1},
		{"ein Fehler rechts von und", `true and gibtsnicht`, 10},
		{"ein Fehler in einem Argument", `{1}(a: gibtsnicht)`, 8},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, nil)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Spalte != fall.spalte {
				t.Errorf("Spalte %d, erwartet war %d — %s",
					befund.Spalte, fall.spalte, befund.Meldung)
			}
		})
	}
}
