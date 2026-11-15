package msescript_test

// **Zwei Werte mit derselben Kennung sind derselbe Gegenstand** —
// Befund 1 des Zweitblicks über Phase 4.11.
//
// # Was MSE2 tut
//
// `equal` (`src/script/value.cpp` Zeilen 96–132, Fassung `f9d9356` vom
// 2021-07-06) fragt am Ende `compareAs`. Ein `ScriptObject<T>` antwortet
// `COMPARE_AS_POINTER` (`src/script/to_value.hpp` Zeilen 184–187): **zwei
// Karten sind gleich, wenn es dieselbe Karte ist** — nicht, wenn sie dieselben
// Felder tragen.
//
// # Warum das hier gefehlt hat
//
// Dieses Programm vergleicht Abbildungen über Namen und Werte. Für zwei
// **strukturgleiche** Karten gab das „gleich", und der Ausdruck
//
//	if card != set.cards.0 then "" else (…)
//
// lief damit an jeder Karte, die zufällig dieselben Felder trägt — eine
// Auszählung zählte still doppelt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// gleichgestaltet baut zwei Abbildungen mit denselben Feldern.
func gleichgestaltet() (links, rechts msescript.Wert) {
	bauen := func() msescript.Wert {
		return msescript.Wert{
			Art: msescript.ArtAbbildung,
			Eintraege: []msescript.Eintrag{
				{Name: "name", Wert: msescript.Text("Wald")},
				{Name: "type", Wert: msescript.Text("Land")},
			},
		}
	}

	return bauen(), bauen()
}

// TestZweiStrukturgleicheOhneKennungSindGleich hält den Stand fest, der ohne
// Kennung gilt — und der für zwei Abbildungen richtig ist.
func TestZweiStrukturgleicheOhneKennungSindGleich(t *testing.T) {
	t.Parallel()

	links, rechts := gleichgestaltet()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("a", links)
	umgebung.Setze("b", rechts)

	wert, err := msescript.AuswertenQuelle(t.Context(), `a == b`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "true" {
		t.Errorf("%q — zwei gleiche Abbildungen ohne Kennung sind gleich", wert.AlsText())
	}
}

// **Mit verschiedenen Kennungen sind sie verschieden** — der Befund.
func TestZweiStrukturgleicheMitVerschiedenerKennungSindVerschieden(t *testing.T) {
	t.Parallel()

	links, rechts := gleichgestaltet()
	links.Kennung = "karte:1"
	rechts.Kennung = "karte:2"

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("a", links)
	umgebung.Setze("b", rechts)

	for _, fall := range [][2]string{{"a == b", "false"}, {"a != b", "true"}} {
		wert, err := msescript.AuswertenQuelle(t.Context(), fall[0], umgebung)
		if err != nil {
			t.Fatalf("%s auswerten: %v", fall[0], err)
		}

		if wert.AlsText() != fall[1] {
			t.Errorf("%s gab %q, erwartet %q", fall[0], wert.AlsText(), fall[1])
		}
	}
}

// **Und mit derselben Kennung sind sie gleich, auch wenn die Felder
// auseinandergehen.**
//
// Das ist die andere Hälfte von `COMPARE_AS_POINTER`: derselbe Gegenstand
// bleibt derselbe, auch nachdem sich etwas an ihm geändert hat.
func TestDieselbeKennungIstDerselbeGegenstand(t *testing.T) {
	t.Parallel()

	links, rechts := gleichgestaltet()
	links.Kennung = "karte:1"
	rechts.Kennung = "karte:1"
	rechts.Eintraege[0].Wert = msescript.Text("Gebirge")

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("a", links)
	umgebung.Setze("b", rechts)

	wert, err := msescript.AuswertenQuelle(t.Context(), `a == b`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "true" {
		t.Errorf("%q — dieselbe Kennung ist derselbe Gegenstand", wert.AlsText())
	}
}

// **Hat nur eine Seite eine Kennung, sind sie verschieden.**
//
// # Hier stand das Gegenteil, und es war ein Lesefehler
//
// „MSE fällt in `equal` auf den Zeichenkettenvergleich zurück, wenn nur eine
// Seite ein Zeiger ist (`value.cpp` Zeilen 127–129)." Das steht dort nicht.
// Nachgelesen am 2026-09-01, `src/script/value.cpp`, Zeilen 122–130, Fassung
// `f9d9356` vom 2021-07-06:
//
//	if (aw == COMPARE_AS_STRING || bw == COMPARE_AS_STRING) {
//	  return as == bs;
//	} else if (aw == COMPARE_AS_POINTER || bw == COMPARE_AS_POINTER) {
//	  return ap == bp;
//	}
//
// Der Zeichenkettenzweig greift, wenn eine Seite **`COMPARE_AS_STRING`**
// meldet — das ist der Vorgabewert für Text und Zahlen (`value.cpp`, Zeilen
// 60–63), nicht für Abbildungen. Eine Abbildung meldet `COMPARE_AS_POINTER`
// (`to_value.hpp`, Zeilen 182–187), und dann greift der **zweite** Zweig mit
// seinem `||`: eine Seite genügt, und die Zeiger sind verschieden.
//
// Berichtigt mit
// #1366.
func TestEineKennungAlleinTrennt(t *testing.T) {
	t.Parallel()

	links, rechts := gleichgestaltet()
	links.Kennung = "karte:1"

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("a", links)
	umgebung.Setze("b", rechts)

	wert, err := msescript.AuswertenQuelle(t.Context(), `a == b`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "false" {
		t.Errorf("%q — ein Nachbau ohne Kennung ist nicht der Gegenstand",
			wert.AlsText())
	}
}
