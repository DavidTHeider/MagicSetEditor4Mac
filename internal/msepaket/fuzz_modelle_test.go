package msepaket_test

// **Die Modelle aus dem Baum** — QA-06b
// (#1967),
// Fläche 6 aus `docs/FUZZZIELE.md`: [msepaket.Paket.Spielmodell] und
// [msepaket.Paket.Stilmodell].
//
// # Warum über dem Zerteiler und nicht neben ihm
//
// **`FuzzZerteile` hält den Baum, nicht seine Deutung.** Was aus einem
// `card field:` ein Feld macht, was aus einem `card style:` einen Feldstil,
// liest ein eigener Leser — und dessen Ergebnis geht ungeprüft in die
// Vorschau, die Kartenliste und den Bestand.
//
// **Das Paket wird aus dem Baum gebaut und nicht von der Platte geöffnet**:
// beide Modelle lesen nur Kopf, Pfad und kurzen Namen, und ein Ordner je
// Ausführung kostete das Hundertfache, ohne einen Weg mehr zu erreichen.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic
//   - zweimal derselbe Text, zwei Modelle oder zwei Befundlisten
//   - ein Befund ohne Meldung
//   - **ein Block, der weder Feld noch Befund wird** — jedes `card field:`,
//     `set field:`, `extra card field:` und `styling field:` kommt genau
//     einmal heraus, als Eintrag oder als benannter Vorbehalt (§8)
//   - ein Feld ohne Namen; ein `boolean`, dessen Wahlen nicht mit `yes`,
//     `no` beginnen (`src/data/field/boolean.cpp`, Zeilen 14–18, Fassung
//     f9d9356 vom 2021-07-06); ein Farbfeld mit anderen Längen bei Wahlen
//     und Farben, ein anderes Feld mit Farben
//   - Feldstile, die nicht nach `z index` geordnet sind
//
// # Was nicht zugesichert wird
//
// **Dass ein Feldname als Schlüssel kanonisch gelesen wird** — `card_name:`
// unter `card style:` trifft das Feld `card name` nicht. Die Schlüssel des
// Formats selbst (`z_index:`, `italic_name:`) liest der Feldstil seit #1972
// wie MSE2; ein Name, der als Schlüssel steht, ist eine andere Frage und
// offen.
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzSpielmodell -fuzztime 2m
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzStilmodell -fuzztime 2m

import (
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// spielsaat sind Gestalten aus `docs/FORMATE.md` und den Tests daneben.
var spielsaat = []string{
	"mse version: 2.0.0\nshort name: Probe\ninit script:\n\tf := 1\n" +
		"card field:\n\ttype: text\n\tname: name\n\tmulti line: true\n" +
		"card field:\n\ttype: choice\n\tname: farbe\n\tchoice: weiß\n\tchoice:\n\t\tname: blau\n" +
		"card field:\n\ttype: color\n\tname: rand\n\tchoice: white\n\tchoice:\n\t\tname: schwarz\n\t\tcolor: rgb(0,0,0)\n" +
		"card field:\n\ttype: boolean\n\tname: legendär\n\tinitial: no\n" +
		"set field:\n\ttype: symbol\n\tname: symbol\n\tlocalized caption:\n\t\tde: Symbol\n" +
		"card field:\n\ttype: text\n\tcard list column: 0400\n",
	"card field:\n\tname:\ncard field: frei\nset field:\n\ttype: unbekannt\n\tname: x\n\tscript: f(value)\n\t\t# Kommentar\nset field:\n\ttype: text\n",
}

// **Jedes Feld kommt heraus — als Feld oder als Befund.**
func FuzzSpielmodell(f *testing.F) {
	for _, saat := range append(slices.Clone(spielsaat), "", "card field:\n") {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, text string) {
		erstes, zweites := paketAusText(text, "probe.mse-game"), paketAusText(text, "probe.mse-game")
		if erstes == nil {
			return
		}

		spiel, befunde := erstes.Spielmodell()
		nochmal, nochmalBefunde := zweites.Spielmodell()

		if !reflect.DeepEqual(spiel, nochmal) || !reflect.DeepEqual(befunde, nochmalBefunde) {
			t.Fatal("derselbe Text, zwei Spielmodelle")
		}

		bloecke := len(erstes.Kopf.Alle("card field")) + len(erstes.Kopf.Alle("set field"))
		if heraus := len(spiel.Fields) + len(spiel.SetFields) + len(befunde); heraus != bloecke {
			t.Errorf("%d Feldblöcke, %d Felder und Befunde", bloecke, heraus)
		}

		befundePruefen(t, befunde)

		for _, feld := range append(slices.Clone(spiel.Fields), spiel.SetFields...) {
			feldPruefen(t, feld.Name, feld.Kind, feld.Choices, feld.Wahlfarben)
		}
	})
}

// **Jede Stiloption kommt heraus, und die Feldstile stehen in
// Zeichenreihenfolge.**
func FuzzStilmodell(f *testing.F) {
	for _, saat := range []string{
		"mse version: 2.0.0\ngame: probe\ncard width: 375\ncard height: 523\ncard dpi: 150\n" +
			"styling field:\n\ttype: boolean\n\tname: rand\n" +
			"styling field:\n\ttype: color\n\tname: farbe\n\tchoice: white\n" +
			"styling field:\n\tname:\n" +
			"extra card field:\n\ttype: text\n\tname: pt box\n" +
			"extra card style:\n\tpt box:\n\t\tleft: 1\n\t\tz index: 3\n" +
			"card style:\n\tname:\n\t\tleft: 20\n\t\tright: 300\n\t\tz index: 2\n" +
			"\tbild:\n\t\tz index: kaputt\n\t\twidth:\n\t\t\tscript: 3 + 4\n\ttext:\n\t\tz index: -1\n" +
			"styling style:\n\trand:\n\t\trender style: checkbox\n",
		"card dpi: 0\ncard width: 9999999999999999999\ncard style:\n\ta:\n\t\tz index: 1\n\tb:\n\t\tz index: 1\n",
		"",
	} {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, text string) {
		erstes, zweites := paketAusText(text, "probe.mse-style"), paketAusText(text, "probe.mse-style")
		if erstes == nil {
			return
		}

		stil, befunde := erstes.Stilmodell()
		nochmal, nochmalBefunde := zweites.Stilmodell()

		if !reflect.DeepEqual(stil, nochmal) || !reflect.DeepEqual(befunde, nochmalBefunde) {
			t.Fatal("derselbe Text, zwei Stilmodelle")
		}

		befundePruefen(t, befunde)

		for _, option := range stil.Stiloptionen {
			feldPruefen(t, option.Name, option.Kind, option.Choices, option.Wahlfarben)
		}

		for _, feld := range stil.Zusatzfelder {
			feldPruefen(t, feld.Name, feld.Kind, feld.Choices, feld.Wahlfarben)
		}

		// **Höchstens so viele wie Blöcke, und was fehlt, ist benannt** — die
		// Befunde der Feldstile stehen in derselben Liste, deshalb „mindestens".
		bloecke := len(erstes.Kopf.Alle("styling field")) + len(erstes.Kopf.Alle("extra card field"))
		if gelesen := len(stil.Stiloptionen) + len(stil.Zusatzfelder); gelesen > bloecke || gelesen+len(befunde) < bloecke {
			t.Errorf("%d Options- und Zusatzblöcke, %d gelesen, %d Befunde", bloecke, gelesen, len(befunde))
		}

		for _, reihe := range [][]msepaket.Feldstil{stil.Felder, stil.Zusatzstile} {
			if !sort.SliceIsSorted(reihe, func(a, b int) bool { return reihe[a].ZIndex < reihe[b].ZIndex }) {
				t.Error("die Feldstile stehen nicht nach `z index`")
			}
		}
	})
}

// paketAusText baut ein Paket aus einer Hauptdatei — `nil`, wenn der
// Zerteiler sie abweist; das ist `FuzzZerteile`s Gegenstand und nicht dieser.
func paketAusText(text, name string) *msepaket.Paket {
	baum, fehler := msepaket.Zerteile(name+"/"+msepaket.ArtAusName(name).Hauptdatei(), []byte(text))
	if fehler != nil {
		return nil
	}

	kurz := ""
	if knoten := baum.Ersten("short name"); knoten != nil {
		kurz = knoten.Wert
	}

	return &msepaket.Paket{Pfad: name, Art: msepaket.ArtAusName(name), KurzerName: kurz, Kopf: baum}
}

// befundePruefen hält, dass jeder Befund sagt, was los ist und wo.
func befundePruefen(t *testing.T, befunde []msepaket.Befund) {
	t.Helper()

	for _, befund := range befunde {
		if befund.Meldung == "" || befund.Stelle.Datei == "" {
			t.Errorf("ein Befund ohne Meldung oder Stelle: %+v", befund)
		}
	}
}

// feldPruefen hält die Zusagen über ein gelesenes Feld oder eine Option.
func feldPruefen(t *testing.T, name string, art domain.FieldKind, wahlen, farben []string) {
	t.Helper()

	if name == "" {
		t.Error("ein Feld ohne Namen ist angenommen")
	}

	if art == domain.KindBoolean && (len(wahlen) < 2 || wahlen[0] != "yes" || wahlen[1] != "no") {
		t.Errorf("%q ist ein boolean mit den Wahlen %q", name, wahlen)
	}

	if art == domain.KindColor && len(farben) != len(wahlen) {
		t.Errorf("%q: %d Wahlen, %d Farben", name, len(wahlen), len(farben))
	}

	if art != domain.KindColor && len(farben) != 0 {
		t.Errorf("%q ist kein Farbfeld und trägt %d Farben", name, len(farben))
	}
}
