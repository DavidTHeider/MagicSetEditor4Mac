package msescript

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// **Ein abgebrochener Lauf wird gemeldet, nicht verschluckt** (§8).
//
// # Der Befund, aus dem dieser Test entstanden ist
//
// Der Zweitblick über die Portierung fand, dass [Muster.Passt] aus **jedem**
// Fehler der Rückfalllinie schlicht `false` machte und [Muster.jederTreffer]
// die Schleife kommentarlos verließ. Ein Muster, das an einer ungünstigen
// Eingabe in die Zeitgrenze lief, sah danach aus wie „kein Treffer" — oder,
// bei [Muster.Ersetze], wie ein halb ersetzter Text.
//
// Das ist genau der stille Weg, den `AGENTS.md` §8 verbietet: ein Set, von dem
// etwas nicht lesbar ist, wird mit einem **benannten** Befund geöffnet.
//
// # Warum das ein interner Test ist
//
// Der Abbruch braucht ein Muster, das sich in der Rückverfolgung verrennt, und
// eine Frist, die früh genug zuschlägt. Mit der echten Sekunde dauerte dieser
// Test fünf Sekunden. [musterFrist] ist deshalb eine Veränderliche, und nur von
// hier aus wird sie geschrieben.
func TestEinAbgebrochenerLaufWirdGemeldet(t *testing.T) {
	vorher := musterFrist
	musterFrist = time.Millisecond

	t.Cleanup(func() { musterFrist = vorher })

	// **Vorausblick, damit RE2 ablehnt** — sonst käme die Rückfalllinie gar
	// nicht zum Zug. Dahinter die klassische Verrennung `(a+)+b`: für jedes
	// zusätzliche `a` verdoppeln sich die Wege.
	muster, err := UebersetzeMuster(`(?=a)(a+)+b`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	if muster.schnell != nil {
		t.Fatal("RE2 hat das Muster genommen — dann prüft dieser Test nichts")
	}

	text := strings.Repeat("a", 40) + "c"

	t.Run("Passt", func(t *testing.T) {
		if _, err := muster.Passt(text); !errors.Is(err, ErrLauf) {
			t.Errorf("%v", err)
		}
	})

	t.Run("AlleTreffer", func(t *testing.T) {
		if _, err := muster.AlleTreffer(text); !errors.Is(err, ErrLauf) {
			t.Errorf("%v", err)
		}
	})

	t.Run("Teile", func(t *testing.T) {
		if _, err := muster.Teile(text); !errors.Is(err, ErrLauf) {
			t.Errorf("%v", err)
		}
	})

	// **Und `Ersetze` gibt keinen halben Text.** Der Rückgabewert ist leer,
	// nicht „so weit wir gekommen sind" — ein halb ersetzter Kartentext sähe
	// aus wie ein Ergebnis.
	t.Run("Ersetze", func(t *testing.T) {
		gebaut, err := muster.Ersetze(text, "X")
		if !errors.Is(err, ErrLauf) {
			t.Errorf("%v", err)
		}

		if gebaut != "" {
			t.Errorf("halber Stand %q", gebaut)
		}
	})

	// **Und `ErsetzeMitFunktion` genauso.** Sie geht über denselben Lauf, und
	// eine Ersatzfunktion änderte daran nichts — der Abbruch kommt aus dem
	// Mustervergleich, nicht aus dem Rumpf.
	t.Run("ErsetzeMitFunktion", func(t *testing.T) {
		gebaut, err := muster.ErsetzeMitFunktion(t.Context(), text, Bibliothek(),
			func(_ context.Context, _ *Umgebung) (Wert, error) { return Text("X"), nil }, false)
		if !errors.Is(err, ErrLauf) {
			t.Errorf("%v", err)
		}

		if gebaut != "" {
			t.Errorf("halber Stand %q", gebaut)
		}
	})

	// **Die Meldung nennt das Muster.** Ohne das steht in einem Lauf über
	// dreißig Module „ein Vergleich wurde abgebrochen" und niemand weiß, welcher.
	if _, err := muster.Passt(text); !strings.Contains(err.Error(), "(a+)+b") {
		t.Errorf("die Meldung nennt das Muster nicht: %v", err)
	}
}

// **Und jede der fünf Funktionen reicht ihn durch.**
//
// Ohne diesen Test wäre belegt, dass [Muster] meldet — nicht, dass die
// Bibliothek die Meldung weitergibt. Genau dazwischen sitzt der stille Weg.
func TestJedeMusterfunktionReichtDenAbbruchDurch(t *testing.T) {
	vorher := musterFrist
	musterFrist = time.Millisecond

	t.Cleanup(func() { musterFrist = vorher })

	quellen := map[string]string{
		"match":       `match(eingabe, match: muster)`,
		"split_text":  `split_text(eingabe, match: muster)`,
		"break_text":  `break_text(eingabe, match: muster)`,
		"filter_text": `filter_text(eingabe, match: muster)`,
		"replace":     `replace(eingabe, match: muster, replace: "X")`,

		// **Auch die beiden Wege mit Rekursion und mit Ersatzfunktion.** Sie
		// laufen durch [Muster.ErsetzeMitFunktion] statt durch
		// [Muster.Ersetze], und ein Abbruch dort ginge sonst still durch.
		"replace rekursiv":  `replace(eingabe, match: muster, replace: "X", recursive: true)`,
		"replace mit Rumpf": `replace(eingabe, match: muster, replace: { "X" })`,
	}

	for name, quelle := range quellen {
		t.Run(name, func(t *testing.T) {
			umgebung := Bibliothek().Darueber()
			umgebung.Setze("muster", Text(`(?=a)(a+)+b`))
			umgebung.Setze("eingabe", Text(strings.Repeat("a", 40)+"c"))

			_, err := AuswertenQuelle(t.Context(), quelle, umgebung)
			if err == nil {
				t.Fatal("kein Fehler — der Abbruch ging still durch")
			}

			if !strings.Contains(err.Error(), strings.Fields(name)[0]) {
				t.Errorf("die Meldung nennt die Funktion nicht: %v", err)
			}
		})
	}
}
