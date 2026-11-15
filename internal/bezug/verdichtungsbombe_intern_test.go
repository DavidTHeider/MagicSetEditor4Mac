// **Eine Kompressionsbombe wird abgewiesen** — Befund Z7-01, Werkstück 4
// (#1163).
//
// # Warum eine Datei und nicht ein gebautes Archiv
//
// Die drei Läufe in `mengengrenzen_intern_test.go` bauen ihre Archive selbst.
// Das prüft die Grenzen — aber gegen etwas, das der Test **kennt**. Hier liegt
// eine Datei, die niemand beim Lauf erzeugt: was sie tut, tut sie unabhängig
// davon, was der Test erwartet.
//
// # Wie sie entstanden ist
//
// Selbst geschrieben (§11), aus Nullbytes:
//
//	kopf := &tar.Header{Name: "…/data/gross.mse-style/style", Size: 64 << 20}
//	schreiber.WriteHeader(kopf)
//	schreiber.Write(make([]byte, 64<<20))
//
// 65.349 Byte verdichtet, 67.108.864 ausgepackt — **1026 : 1**.
//
// # Was dieser Lauf belegt, und was nicht
//
//   - **Belegt**: die Grenzen greifen an einer Datei, die sich beim Auspacken
//     vertausendfacht, und der unfertige Stand bleibt nicht liegen.
//   - **Nicht belegt**: dass zwei und vier Gigabyte die richtigen Zahlen sind.
//     Der Lauf dreht sie herunter; die Zahlen selbst sind eine Entscheidung
//     und keine Messung.
package bezug

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dieBombe ist der Pfad zur eingecheckten Bombe.
const dieBombe = "../../testdaten/bomben/verdichtungsbombe.tar.gz"

// **Die Bombe kommt nicht durch — und lässt nichts liegen.**
func TestEineVerdichtungsbombeWirdAbgewiesen(t *testing.T) {
	// **Ein Megabyte je Eintrag, zwei zusammen.** Die Bombe bringt 64 MiB in
	// einem Eintrag; sie reißt beide.
	mitMengengrenzen(t, 1<<20, 2<<20, 100)

	inhalt, err := os.ReadFile(dieBombe) // #nosec G304 -- eingecheckte Testdatei
	if err != nil {
		t.Fatalf("die Bombe lesen: %v", err)
	}

	// **Das Verhältnis wird mitgeprüft.** Ohne diese Zusicherung wäre der Lauf
	// auch dann grün, wenn jemand die Datei durch ein harmloses Archiv
	// ersetzte — und er hieße weiter „Bombe".
	if len(inhalt) > 1<<20 {
		t.Fatalf("die Bombe wiegt %d Byte — das ist keine mehr", len(inhalt))
	}

	ziel := t.TempDir()

	_, fehler := auspacken(bytes.NewReader(inhalt), "", ziel)
	if fehler == nil {
		t.Fatal("die Kompressionsbombe kam durch")
	}

	if !strings.Contains(fehler.Error(), "zu groß") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}

	// **Was liegen bleibt, bleibt unter der Grenze.** Der Aufrufer räumt das
	// Zielverzeichnis weg (`service.Paketbezug` packt in ein
	// `os.MkdirTemp`-Verzeichnis und benennt erst am Ende um); hier zählt,
	// dass das Auspacken nicht vorher die Platte füllt.
	if gewicht := verzeichnisgewicht(t, ziel); gewicht > 2<<20 {
		t.Errorf("es liegen %d Byte im Ziel — die Grenze hat zu spät gegriffen", gewicht)
	}
}

// verzeichnisgewicht summiert die Dateigrößen unter einem Verzeichnis.
func verzeichnisgewicht(t *testing.T, wurzel string) int64 {
	t.Helper()

	var gesamt int64

	if err := filepath.WalkDir(wurzel, func(_ string, eintrag os.DirEntry, unterwegs error) error {
		if unterwegs != nil || eintrag.IsDir() {
			return unterwegs
		}

		angaben, fehler := eintrag.Info()
		if fehler != nil {
			return fehler
		}

		gesamt += angaben.Size()

		return nil
	}); err != nil {
		t.Fatalf("zählen: %v", err)
	}

	return gesamt
}
