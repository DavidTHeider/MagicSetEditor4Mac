package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestEineGezeichneteZeichnungWirdZuEinerDatei prüft den Weg, den es bis Phase
// X nicht gab: aus einem Modell entsteht eine `.mse-symbol`.
//
// **Die Erwartung steht als ganze Datei da**, nicht als Feldliste. Ein Test,
// der Feld für Feld zusichert, sagt nichts über Reihenfolge und Einrückung —
// und genau die liest MSE2.
//
// **Zehn Nachkommastellen sind kein Versehen.** MSE schreibt jeden Vektor als
// `(%.10lf,%.10lf)` (`src/util/io/writer.cpp`, Zeile 131, Fassung f9d9356 vom
// 2021-07-06). Der erste Anlauf dieses Tests erwartete `(0.5,0)` — geraten,
// obwohl der Quelltext dalag.
func TestEineGezeichneteZeichnungWirdZuEinerDatei(t *testing.T) {
	t.Parallel()

	zeichnung := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{{
		Art:         msepaket.TeilForm,
		Name:        "Dreieck",
		Kombination: msepaket.KombinationVereinen,
		Punkte: []msepaket.Kontrollpunkt{
			{Stelle: msepaket.Vektor{X: 0.5, Y: 0}},
			{Stelle: msepaket.Vektor{X: 1, Y: 1}},
			{Stelle: msepaket.Vektor{X: 0, Y: 1}},
		},
	}}}

	will := "\ufeffmse_version: 0.3.5\n" +
		"part:\n" +
		"\ttype: shape\n" +
		"\tname: Dreieck\n" +
		"\tcombine: merge\n" +
		"\tpoint:\n" +
		"\t\tposition: (0.5000000000,0.0000000000)\n" +
		"\t\tlock: free\n" +
		"\t\tline_after: line\n" +
		"\tpoint:\n" +
		"\t\tposition: (1.0000000000,1.0000000000)\n" +
		"\t\tlock: free\n" +
		"\t\tline_after: line\n" +
		"\tpoint:\n" +
		"\t\tposition: (0.0000000000,1.0000000000)\n" +
		"\t\tlock: free\n" +
		"\t\tline_after: line\n"

	if hat := string(msepaket.SchreibeSymbolzeichnung(zeichnung)); hat != will {
		t.Errorf("die geschriebene Datei:\n%s\nwill:\n%s", hat, will)
	}
}

// TestEineGeschriebeneZeichnungLiestSichSelbstZurueck ist der Rundlauf, der
// zählt: Schreiber und Leser sind zwei Seiten derselben Zusage.
func TestEineGeschriebeneZeichnungLiestSichSelbstZurueck(t *testing.T) {
	t.Parallel()

	will := msepaket.Symbolzeichnung{Teile: []msepaket.Symbolteil{
		{
			Art:         msepaket.TeilForm,
			Name:        "Bogen",
			Kombination: msepaket.KombinationAbziehen,
			Punkte: []msepaket.Kontrollpunkt{
				{
					Stelle:      msepaket.Vektor{X: 0.25, Y: 0.5},
					GriffDanach: msepaket.Vektor{X: 0.1, Y: -0.2},
					GriffDavor:  msepaket.Vektor{X: -0.1, Y: 0.2},
					KurveDanach: true,
					KurveDavor:  true,
					Sperre:      "size",
				},
				{
					Stelle:      msepaket.Vektor{X: 0.75, Y: 0.5},
					GriffDavor:  msepaket.Vektor{X: -0.1, Y: -0.2},
					GriffDanach: msepaket.Vektor{X: 0.1, Y: 0.2},
					KurveDavor:  true,
					KurveDanach: true,
					Sperre:      "free",
				},
			},
		},
		{
			Art:        msepaket.TeilSymmetrie,
			Name:       "Spiegel",
			Spiegelart: msepaket.SpiegelungAchse,
			Kopien:     4,
			Mitte:      msepaket.Vektor{X: 0.5, Y: 0.5},
			Griff:      msepaket.Vektor{X: 0, Y: -0.5},
			Teile: []msepaket.Symbolteil{{
				Art:         msepaket.TeilForm,
				Name:        "Kind",
				Kombination: msepaket.KombinationUeberdecken,
				Punkte:      []msepaket.Kontrollpunkt{{Stelle: msepaket.Vektor{X: 0, Y: 0}, Sperre: "free"}},
			}},
		},
		{
			Art:  msepaket.TeilGruppe,
			Name: "Gruppe",
			Teile: []msepaket.Symbolteil{{
				Art:         msepaket.TeilForm,
				Name:        "Enkel",
				Kombination: msepaket.KombinationUmranden,
				Punkte:      []msepaket.Kontrollpunkt{{Stelle: msepaket.Vektor{X: 1, Y: 1}, Sperre: "free"}},
			}},
		},
	}}

	roh := msepaket.SchreibeSymbolzeichnung(will)

	baum, err := msepaket.Zerteile("symbol", roh)
	if err != nil {
		t.Fatalf("zerteilen:\n%s\n%v", roh, err)
	}

	hat, befunde := msepaket.Symbolzeichnungsmodell(baum)
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	vergleiche(t, "", hat.Teile, will.Teile)
}

// TestEinLeererNameBekommtKeinAbschlussleerzeichen hält die eine benannte
// Abweichung von MSE fest.
func TestEinLeererNameBekommtKeinAbschlussleerzeichen(t *testing.T) {
	t.Parallel()

	roh := string(msepaket.SchreibeSymbolzeichnung(msepaket.Symbolzeichnung{
		Teile: []msepaket.Symbolteil{{Art: msepaket.TeilForm}},
	}))

	if !strings.Contains(roh, "\tname:\n") {
		t.Errorf("kein `name:` ohne Wert in:\n%s", roh)
	}

	// **Und der Standardwert steht ausgeschrieben da.** Eine Datei, die sich
	// auf MSEs Lesart des leeren Feldes verlässt, ist eine Wette.
	for _, will := range []string{"\ttype: shape\n", "\tcombine: overlap\n"} {
		if !strings.Contains(roh, will) {
			t.Errorf("%q fehlt in:\n%s", will, roh)
		}
	}
}

// vergleiche prüft zwei Teillisten rekursiv und nennt den Weg zur Abweichung.
func vergleiche(t *testing.T, weg string, hat, will []msepaket.Symbolteil) {
	t.Helper()

	if len(hat) != len(will) {
		t.Fatalf("%s: %d Teile, will %d", weg, len(hat), len(will))
	}

	for stelle := range will {
		hier := weg + "/" + will[stelle].Name

		hatKinder, willKinder := hat[stelle].Teile, will[stelle].Teile
		hat[stelle].Teile, will[stelle].Teile = nil, nil

		// **Der Leser setzt `overlap` ein, wo nichts steht** — auch bei einer
		// Symmetrie, die gar kein `combine:` trägt. Beide Seiten werden
		// deshalb gleich behandelt; wer nur eine normiert, prüft seine eigene
		// Normierung.
		for _, seite := range []*msepaket.Symbolteil{&hat[stelle], &will[stelle]} {
			if seite.Kombination == "" {
				seite.Kombination = msepaket.KombinationUeberdecken
			}
		}

		if !gleich(hat[stelle], will[stelle]) {
			t.Errorf("%s:\nhat  %+v\nwill %+v", hier, hat[stelle], will[stelle])
		}

		vergleiche(t, hier, hatKinder, willKinder)
	}
}

func gleich(a, b msepaket.Symbolteil) bool {
	if a.Art != b.Art || a.Name != b.Name || a.Kombination != b.Kombination ||
		a.Spiegelart != b.Spiegelart || a.Kopien != b.Kopien ||
		a.Mitte != b.Mitte || a.Griff != b.Griff || len(a.Punkte) != len(b.Punkte) {
		return false
	}

	for stelle := range a.Punkte {
		if a.Punkte[stelle] != b.Punkte[stelle] {
			return false
		}
	}

	return true
}

// TestDieStandardwerteEinerSymmetrieUndEinerFremdenArt betritt die beiden
// Zweige, die kein gewöhnliches Schreiben nimmt.
//
// **Eine fremde Teilart ist keine Erfindung.** [msepaket.Symbolteilart] ist ein
// Zeichenkettentyp; ein Aufrufer kann jeden Wert hineinschreiben, und der
// Schreiber darf darüber keine Datei ohne `type:` erzeugen.
func TestDieStandardwerteEinerSymmetrieUndEinerFremdenArt(t *testing.T) {
	t.Parallel()

	t.Run("eine Symmetrie ohne kind ist eine Drehung", func(t *testing.T) {
		t.Parallel()

		roh := string(msepaket.SchreibeSymbolzeichnung(msepaket.Symbolzeichnung{
			Teile: []msepaket.Symbolteil{{Art: msepaket.TeilSymmetrie, Kopien: 2}},
		}))

		if !strings.Contains(roh, "\tkind: rotation\n") {
			t.Errorf("kein `kind: rotation` in:\n%s", roh)
		}
	})

	t.Run("eine fremde Art wird als Form geschrieben", func(t *testing.T) {
		t.Parallel()

		roh := string(msepaket.SchreibeSymbolzeichnung(msepaket.Symbolzeichnung{
			Teile: []msepaket.Symbolteil{{Art: msepaket.Symbolteilart("Kegelschnitt")}},
		}))

		if !strings.Contains(roh, "\ttype: shape\n") {
			t.Errorf("kein `type: shape` in:\n%s", roh)
		}
	})
}
