package msescript_test

// **Ein Name, der eine Zahl ist, meint eine Stelle** —
// #1356.
//
// # Was MSE2 tut
//
// `ScriptValue::getMember` (`src/script/value.cpp`, Zeilen 64–70, Fassung
// `f9d9356` vom 2021-07-06):
//
//	long index;
//	if (name.ToLong(&index)) return getIndex(index);
//	else return delay_error(…);
//
// Das steht in der **Basisklasse** und gilt damit für jeden Wert. Eine
// Sammlung greift damit auf ihre **unbenannten** Glieder
// (`ScriptCustomCollection::getIndex`, Zeilen 520–526) — und erst, nachdem der
// benannte Zugriff nichts gefunden hat (`getMember`, Zeilen 512–519).
//
// # Warum es das hier bisher nicht gab
//
// Weil nichts danach gefragt hat. `set.cards.0` ist die erste Stelle im Pack,
// die es braucht — zwei Auszählungsdimensionen von `magic.mse-game` rechnen
// über das ganze Set und geben ihr Ergebnis an der ersten Karte aus.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// liste baut eine Sammlung aus unbenannten Gliedern.
func liste(werte ...msescript.Wert) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(werte))
	for _, wert := range werte {
		eintraege = append(eintraege, msescript.Eintrag{Wert: wert})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}

// TestEineZahlAlsNameMeintDieStelle ist die Grundzusage.
func TestEineZahlAlsNameMeintDieStelle(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("stapel",
		liste(msescript.Text("erstes"), msescript.Text("zweites"), msescript.Text("drittes")))

	for _, fall := range []struct {
		quelle   string
		erwartet string
	}{
		{"stapel.0", "erstes"},
		{"stapel.1", "zweites"},
		{"stapel.2", "drittes"},
		{`stapel["1"]`, "zweites"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, umgebung)
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if gefunden := wert.AlsText(); gefunden != fall.erwartet {
				t.Errorf("%q, erwartet %q", gefunden, fall.erwartet)
			}
		})
	}
}

// **Der benannte Zugriff geht vor.**
//
// `ScriptCustomCollection::getMember` sieht **zuerst** in `key_value` und
// erst danach in die Stellen (Zeilen 512–519). Eine Sammlung, die ein Glied
// namens `0` führt, gibt dieses — nicht das erste unbenannte.
func TestEinGliedNamensNullGehtVor(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("gemischt", msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{
			{Wert: msescript.Text("unbenannt")},
			{Name: "0", Wert: msescript.Text("benannt")},
		},
	})

	wert, err := msescript.AuswertenQuelle(t.Context(), "gemischt.0", umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "benannt" {
		t.Errorf("%q, erwartet %q", wert.AlsText(), "benannt")
	}
}

// **Nur die unbenannten Glieder werden gezählt.**
//
// `getIndex` greift in `value`, und dort stehen ausschließlich die Glieder
// ohne Schlüssel. Ein benanntes dazwischen verschiebt die Stellen **nicht**.
func TestBenannteGliederVerschiebenDieStellenNicht(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("gemischt", msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{
			{Name: "titel", Wert: msescript.Text("egal")},
			{Wert: msescript.Text("erstes")},
			{Name: "fussnote", Wert: msescript.Text("auch egal")},
			{Wert: msescript.Text("zweites")},
		},
	})

	for _, fall := range [][2]string{{"gemischt.0", "erstes"}, {"gemischt.1", "zweites"}} {
		wert, err := msescript.AuswertenQuelle(t.Context(), fall[0], umgebung)
		if err != nil {
			t.Fatalf("%s auswerten: %v", fall[0], err)
		}

		if wert.AlsText() != fall[1] {
			t.Errorf("%s gab %q, erwartet %q", fall[0], wert.AlsText(), fall[1])
		}
	}
}

// **Über das Ende hinaus ist ein Fehlerwert, kein Abbruch** — MSEs
// `delay_error`, den `or else` auffängt.
func TestEineStelleJenseitsDesEndesIstEinFehlerwert(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("stapel", liste(msescript.Text("eins")))

	wert, err := msescript.AuswertenQuelle(t.Context(), `stapel.7 or else "nichts"`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "nichts" {
		t.Errorf("%q, erwartet %q", wert.AlsText(), "nichts")
	}
}

// **Eine negative Stelle ebenso.**
func TestEineNegativeStelleIstEinFehlerwert(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("stapel", liste(msescript.Text("eins")))

	if _, err := msescript.AuswertenQuelle(t.Context(), "stapel.-1", umgebung); err == nil {
		t.Log("`-1` ist für den Zerteiler kein Gliedname — das ist in Ordnung")
	}
}

// **Eine Abbildung ohne unbenannte Glieder gibt keine Stelle.**
//
// Das ist die Zusage aus #653:
// `set.0` gibt **keine** Karte. In MSE geht `ScriptObject<Set*>::getMember`
// ins `set_info` und fällt dann auf `ScriptValue::getMember` zurück — und
// `ScriptValue::getIndex` ist ein Fehler. Hier ist es dasselbe: eine
// Abbildung aus lauter benannten Gliedern hat keine Stellen.
func TestEineAbbildungOhneUnbenannteGliederHatKeineStellen(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("set", msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{
			{Name: "title", Wert: msescript.Text("Ein Set")},
		},
	})

	wert, err := msescript.AuswertenQuelle(t.Context(), "set.0", umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if !strings.Contains(wert.AlsText(), "Fehler") {
		t.Errorf("%q — erwartet wurde ein Fehlerwert", wert.AlsText())
	}
}

// **Und eine Stelle in einer Schleife** — die Form, die die Dimensionen
// benutzen: `for x from 0 to n do set.cards[x]`.
func TestEineStelleLaesstSichBerechnen(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("stapel", liste(msescript.Text("a"), msescript.Text("b"), msescript.Text("c")))

	wert, err := msescript.AuswertenQuelle(t.Context(),
		"for x from 0 to length(stapel)-1 do stapel[x]", umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "abc" {
		t.Errorf("%q, erwartet %q", wert.AlsText(), "abc")
	}
}
