// Das Lesen von go.mod-Dateien für die Wächter in diesem Paket.
//
// Von `pinnung_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, *was*
// geprüft wird, hier, *wie* eine go.mod gelesen wird.
package werkzeuge

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

// wurzel ist das Projektverzeichnis, von diesem Paket aus gesehen.
const wurzel = "../.."

// modul hält die Angaben einer go.mod-Datei, soweit dieser Wächter sie braucht.
type modul struct {
	// goVersion ist der Wert der `go`-Direktive.
	goVersion string
	// anforderungen bildet Modulpfad auf Version ab (require-Block).
	anforderungen map[string]string
	// werkzeuge sind die Paketpfade der tool-Direktiven.
	werkzeuge []string
}

// version sucht zu einem Werkzeugpfad die Version des Moduls, aus dem er
// stammt. Die tool-Direktive nennt das Paket (`…/cmd/gosec`), der require-Block
// das Modul (`…/gosec/v2`); verbunden sind beide über das längste passende
// Präfix. Ohne Fund liefert sie die leere Zeichenkette.
func (m modul) version(werkzeugpfad string) string {
	var bestesPraefix, gefunden string

	for pfad, wert := range m.anforderungen {
		if werkzeugpfad != pfad && !strings.HasPrefix(werkzeugpfad, pfad+"/") {
			continue
		}
		if len(pfad) > len(bestesPraefix) {
			bestesPraefix, gefunden = pfad, wert
		}
	}

	return gefunden
}

// werkzeugmodule liest die go.mod-Dateien mit tool-Direktiven.
func werkzeugmodule(t *testing.T) map[string]modul {
	t.Helper()

	module := make(map[string]modul, len(werkzeugdateien))
	for _, datei := range werkzeugdateien {
		module[datei] = gomodLesen(t, datei)
	}

	return module
}

// gomodLesen zerteilt eine go.mod-Datei. Von Hand statt über
// golang.org/x/mod/modfile: der Wächter soll das Wurzelmodul nicht um eine
// Abhängigkeit erweitern, die er selbst mitprüfen müsste.
func gomodLesen(t *testing.T, pfad string) modul {
	t.Helper()

	roh := projektdatei(t, pfad)

	gelesen := modul{anforderungen: make(map[string]string)}

	var block string

	for _, roheZeile := range strings.Split(string(roh), "\n") {
		felder := strings.Fields(ohneKommentar(roheZeile))
		if len(felder) == 0 {
			continue
		}

		if block != "" {
			if felder[0] == ")" {
				block = ""

				continue
			}
			gelesen.eintragen(block, felder)

			continue
		}

		switch {
		case felder[0] == "go" && len(felder) == 2:
			gelesen.goVersion = felder[1]
		case felder[0] != "require" && felder[0] != "tool":
		case len(felder) == 2 && felder[1] == "(":
			block = felder[0]
		default:
			gelesen.eintragen(felder[0], felder[1:])
		}
	}

	return gelesen
}

// eintragen übernimmt eine Direktivenzeile in das gelesene Modul.
func (m *modul) eintragen(block string, felder []string) {
	switch {
	case block == "require" && len(felder) >= 2:
		m.anforderungen[felder[0]] = felder[1]
	case block == "tool" && len(felder) >= 1:
		m.werkzeuge = append(m.werkzeuge, felder[0])
	}
}

// projektdatei liest eine Datei unterhalb der Projektwurzel.
//
// Über `os.DirFS` statt über einen zusammengesetzten Pfad an `os.ReadFile`:
// damit ist der Lesezugriff auf das Projektverzeichnis beschränkt, und zwar
// sichtbar — ein zusammengesetzter Pfad wäre dieselbe Absicht ohne Beleg.
func projektdatei(t *testing.T, pfad string) []byte {
	t.Helper()

	roh, err := fs.ReadFile(os.DirFS(wurzel), pfad)
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", pfad, err)
	}

	return roh
}

// ohneKommentar schneidet einen Zeilenkommentar ab.
func ohneKommentar(zeile string) string {
	if stelle := strings.Index(zeile, "//"); stelle >= 0 {
		return zeile[:stelle]
	}

	return zeile
}
