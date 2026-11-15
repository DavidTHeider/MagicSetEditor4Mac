// **Die Zahlen im README sind gezählt** —
// #1170.
//
// # Der Befund
//
// Die Übersichtstabelle unter *Schnittstellen* nannte am 2026-08-28 noch „37
// von 37" — eine Zahl aus Phase T. Tatsächlich waren es 87 Fähigkeiten, 86 am
// Fenster, 59 über MCP, 61 über die Kommandozeile und 59 über den Dienst.
//
// **Das ist genau die Gestalt, gegen die dieses Projekt schon zweimal einen
// Wächter gebaut hat**: der Einleitungssatz von `FAEHIGKEITEN.md` und seine
// Übersichtstabelle. Beide stehen unter Aufsicht — das README nicht, und es
// ist die Datei, die ein Fremder zuerst liest.
//
// # Was er hält
//
// Jede Zeile der Schnittstellentabelle nennt dieselbe Zahl wie die
// Übersichtstabelle in `docs/FAEHIGKEITEN.md` — und die ist ihrerseits aus den
// Zeilen gezählt (`faehigkeiten_test.go`). Damit hängt das README an einer
// Messung und nicht an einer Erinnerung.
//
// # Was er nicht hält
//
// **Den Fließtext daneben.** „Dreizehn Fähigkeiten hängen an einem nativen
// Dateidialog" steht dort als Wort und nicht als Ziffer; ein Wächter, der
// Zahlwörter läse, wäre eine Übung in Grammatik. Was ihn hält, ist dieselbe
// Zahl in `FAEHIGKEITEN.md`, und die ist gezählt.
package werkzeuge_test

import (
	"regexp"
	"strconv"
	"testing"
)

// zeileImReadme findet „| **Fenster** | **86 von 87** …".
var zeileImReadme = regexp.MustCompile(`\|\s*\*\*(Fenster|MCP|CLI|REST)[^|]*\|\s*\*\*(\d+) von (\d+)\*\*`)

// zeileImKatalog findet „| davon am **Fenster** | **86** …".
var zeileImKatalog = regexp.MustCompile(`\|\s*davon (?:am|über) \*\*(\w+)\*\*\s*\|\s*\*\*(\d+)\*\*`)

// insgesamtImKatalog findet „| Fähigkeiten insgesamt | **87** |".
var insgesamtImKatalog = regexp.MustCompile(`\|\s*Fähigkeiten insgesamt\s*\|\s*\*\*(\d+)\*\*`)

func TestDieZahlenImReadmeStimmenMitDemKatalog(t *testing.T) {
	katalog := lies(t, "docs/FAEHIGKEITEN.md")

	insgesamt := einzelfund(t, insgesamtImKatalog, katalog,
		"docs/FAEHIGKEITEN.md nennt keine Gesamtzahl")

	imKatalog := map[string]string{}
	for _, treffer := range zeileImKatalog.FindAllStringSubmatch(katalog, -1) {
		imKatalog[treffer[1]] = treffer[2]
	}

	if len(imKatalog) != 4 {
		t.Fatalf("die Übersicht nennt %d Flächen — der Leser greift daneben", len(imKatalog))
	}

	treffer := zeileImReadme.FindAllStringSubmatch(lies(t, "README.md"), -1)
	if len(treffer) != 4 {
		t.Fatalf("das README nennt %d Flächen mit Zahl — erwartet waren vier", len(treffer))
	}

	for _, zeile := range treffer {
		flaeche, hat, von := zeile[1], zeile[2], zeile[3]

		if von != insgesamt {
			t.Errorf("das README nennt für %s »von %s« — der Katalog zählt %s",
				flaeche, von, insgesamt)
		}

		if erwartet := imKatalog[flaeche]; hat != erwartet {
			t.Errorf("das README nennt für %s die Zahl %s — der Katalog zählt %s",
				flaeche, hat, erwartet)
		}
	}

	// **Und die Gesamtzahl ist eine Zahl.** Ohne diese Zeile bliebe ein
	// `insgesamt` von „0" unbemerkt, und jeder Vergleich darüber wäre grün.
	if wert, err := strconv.Atoi(insgesamt); err != nil || wert < 1 {
		t.Errorf("die Gesamtzahl %q ist keine brauchbare Zahl", insgesamt)
	}
}
