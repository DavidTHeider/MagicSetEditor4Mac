package scryfall_test

// **Dieselbe Antwort gibt dieselbe Meldung** — der Fund des Nachlaufs über
// Phase 4.2 zu
// #1289.
//
// # Der Fund
//
// `pflichtfelder` ging über ein `map`-Literal, und Go würfelt dessen
// Reihenfolge bei jedem Durchlauf. Eine Antwort, der **zwei** Pflichtfelder
// fehlen, bekam damit mal die eine und mal die andere Meldung — bei derselben
// Eingabe, im selben Prozess, ohne dass sich etwas geändert hätte.
//
// # Warum das mehr ist als ein Schönheitsfehler
//
// Wer eine Karte zweimal abruft und zwei verschiedene Gründe zu lesen bekommt,
// sucht den Fehler an der falschen Stelle. Und ein Fehlerbericht darüber ist
// nicht nachzustellen — die Meldung wechselt beim Nachfahren.
//
// Gefunden hat es die Zusage, die der Nachlauf verlangt hat: **zum Ergebnis
// gehört auch das Nein.** Der Fuzzlauf verglich bis dahin nur erfolgreiche
// Antworten; die erste Ablehnung kehrte zurück, ohne ein zweites Mal zu lesen.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

func TestDieselbeAntwortGibtDieselbeMeldung(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string]string{
		"gar nichts":              `{}`,
		"nur das Objekt":          `{"object":"card"}`,
		"Objekt und Kennung":      `{"object":"card","id":"a"}`,
		"alles außer dem Objekt":  `{"id":"a","name":"N"}`,
		"eine leere Zeichenkette": `{"object":"","id":"","name":""}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			erste := meldungVon(t, roh)

			// **Hundertmal**, nicht zweimal: bei drei Feldern trifft ein
			// gewürfelter Durchlauf mit einem Drittel Wahrscheinlichkeit
			// zufällig dasselbe. Ein Lauf, der das nicht bedenkt, ist grün und
			// sagt nichts.
			for i := 0; i < 100; i++ {
				if weitere := meldungVon(t, roh); weitere != erste {
					t.Fatalf("dieselbe Antwort gab zwei Meldungen:\n%q\n%q",
						erste, weitere)
				}
			}
		})
	}
}

// **Und die Meldung nennt das Feld, das am weitesten vorn fehlt.**
//
// `object` vor `id` vor `name` — die Reihenfolge, in der Scryfall sie in seiner
// Antwort führt. Sie ist willkürlich und deshalb festgeschrieben: was zählt,
// ist dass sie sich nicht ändert.
func TestDieMeldungNenntDasErsteFehlendePflichtfeld(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, roh, feld string
	}{
		{"alle drei fehlen", `{}`, "object"},
		{"das Objekt ist da", `{"object":"card"}`, "id"},
		{"nur der Name fehlt", `{"object":"card","id":"a"}`, "name"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if meldung := meldungVon(t, fall.roh); !strings.Contains(meldung, `"`+fall.feld+`"`) {
				t.Errorf("die Meldung nennt %q nicht: %s", fall.feld, meldung)
			}
		})
	}
}

// meldungVon liest eine Karte und gibt die Meldung der Ablehnung.
func meldungVon(t *testing.T, roh string) string {
	t.Helper()

	karte, _, fehler := scryfall.LiesKarte([]byte(roh))
	if fehler == nil {
		t.Fatalf("%s wurde angenommen: %+v", roh, karte)
	}

	return fehler.Error()
}
