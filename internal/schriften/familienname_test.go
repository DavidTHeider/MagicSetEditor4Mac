package schriften_test

// **Eine Schrift steht unter ihrem vollen Namen *und* unter ihrer Familie.**
//
// MSE reicht Familienname und Gewicht an wxWidgets weiter:
//
//	REFLECT(name);
//	REFLECT(weight);
//	REFLECT(italic_name);
//
// `IMPLEMENT_REFLECTION_NO_SCRIPT(Font)`, `src/data/font.cpp`, Zeilen
// 119–134, Fassung f9d9356 vom 2021-07-06. **Es gibt kein `bold name:`** —
// anders als bei Kursiv ist Fett ein *Attribut*, und wx sucht den passenden
// Schnitt der Familie.
//
// # Woran es aufgefallen ist
//
// Der Lauf über alle 23.168 kanonischen Karten meldete **48-mal** „die Schrift
// Beleren Small Caps ist nicht zu finden". Die Datei liegt sehr wohl im Pack:
//
//	belerensmallcaps-bold.ttf
//	    Familie        "Beleren Small Caps"
//	    Unterfamilie   "Bold"
//	    voller Name    "Beleren Small Caps Bold"
//
// Sie stand nur unter ihrem **vollen** Namen im Index; wer „Beleren Small
// Caps" verlangte — den Familiennamen, so wie MSE es tut —, fand nichts.

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

func TestEineSchriftStehtUnterFamilieUndVollemNamen(t *testing.T) {
	pfad, voll := ersteSystemschrift(t)

	roh, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	namen := schriften.NamenAus(roh)
	if len(namen) == 0 {
		t.Fatal("keine Namen gelesen")
	}

	// **Der volle Name steht vorn**, denn er ist der genauere: wer ihn nennt,
	// meint diesen Schnitt.
	if namen[0] != voll {
		t.Errorf("der erste Name ist %q, erwartet der volle Name %q", namen[0], voll)
	}

	for i := range namen {
		for j := i + 1; j < len(namen); j++ {
			if namen[i] == namen[j] {
				t.Errorf("%q steht zweimal — Doppelte gehören heraus", namen[i])
			}
		}
	}
}

// **Und die Suche findet sie unter beiden.**
//
// Der Fall aus dem Pack: eine Datei, deren Familie anders heißt als ihr voller
// Name. Ein Stil, der die Familie nennt, muss sie finden.
func TestDieSucheFindetEineSchriftUeberIhreFamilie(t *testing.T) {
	ort := t.TempDir()
	legeSchrift(t, ort, "irgendwie-anders.ttf")

	pfad, _ := ersteSystemschrift(t)

	roh, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	namen := schriften.NamenAus(roh)
	if len(namen) < 2 {
		t.Skip("diese Systemschrift führt Familie und vollen Namen gleich")
	}

	suche := schriften.Neue([]string{ort})

	for _, name := range namen {
		if _, err := suche.Fundort(name, false); err != nil {
			t.Errorf("unter %q nicht gefunden: %v", name, err)
		}
	}
}
