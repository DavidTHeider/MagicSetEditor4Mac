// Der Rundlauf über **jedes** Paket in `testdaten/`.
//
// # Warum das eigens steht
//
// `schreiber_test.go` prüft den Rundlauf an getippten Zeichenketten und an
// einer Datei. Das deckt, woran der Schreibende gedacht hat. Dieser Test hier
// deckt, was **dasteht** — jede Datei in jedem Paket, ohne Auswahl.
//
// Er ist damit auch der Wächter über künftige Testdaten: wer ein Paket
// hinzufügt, prüft den Rundlauf mit, ohne etwas dafür zu tun.
package msepaket_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDerRundlaufUeberAllePakete(t *testing.T) {
	wurzel := filepath.Join("..", "..", "testdaten")

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("testdaten/ nicht lesbar: %v", err)
	}

	pakete := 0

	for _, eintrag := range eintraege {
		if !strings.Contains(eintrag.Name(), ".mse-") {
			continue
		}

		pakete++

		t.Run(eintrag.Name(), func(t *testing.T) {
			paket, err := msepaket.Oeffne(filepath.Join(wurzel, eintrag.Name()))
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() {
				if fehler := paket.Schliessen(); fehler != nil {
					t.Errorf("schließen: %v", fehler)
				}
			})

			namen, err := paket.Dateien()
			if err != nil {
				t.Fatalf("aufzählen: %v", err)
			}

			for _, datei := range namen {
				t.Run(datei, func(t *testing.T) {
					roh, err := paket.Datei(datei)
					if err != nil {
						t.Fatalf("lesen: %v", err)
					}

					if !istText(roh) {
						// **Ein Bild ist keine Datendatei.** Es reist im Paket
						// mit und wird nie zerteilt; es hier zu überspringen
						// ist die Regel, nicht die Ausnahme.
						t.Skip("keine Textdatei")
					}

					baum, err := msepaket.Zerteile(datei, roh)
					if err != nil {
						t.Fatalf("zerteilen: %v", err)
					}

					gebaut := msepaket.Schreibe(baum)
					if bytes.Equal(gebaut, roh) {
						return
					}

					t.Errorf("der Rundlauf änderte %s/%s:\n%s",
						eintrag.Name(), datei, ersteAbweichung(roh, gebaut))
				})
			}
		})
	}

	// **Ein Test über eine leere Liste ist grün.** Dass er Pakete gesehen hat,
	// gehört geprüft — sonst wäre eine verschobene `testdaten/` ein grüner Lauf.
	if pakete < 5 {
		t.Errorf("nur %d Pakete geprüft — greift der Test ins Leere?", pakete)
	}
}

// ersteAbweichung nennt Zeile und Inhalt der ersten Abweichung.
//
// **Nicht die ganzen Dateien.** Ein Vergleich, der zweimal 700 Zeilen ausgibt,
// wird nicht gelesen; die eine Zeile, an der es auseinandergeht, wird gelesen.
func ersteAbweichung(erwartet, gebaut []byte) string {
	// **Getrennt am Zeilenvorschub, ohne ihn wegzuwerfen**: ein `\r` am Ende
	// bleibt damit im Text stehen und wird von `quote` sichtbar gemacht.
	alt := strings.SplitAfter(string(erwartet), "\n")
	neu := strings.SplitAfter(string(gebaut), "\n")

	for i := range max(len(alt), len(neu)) {
		links, rechts := "", ""
		if i < len(alt) {
			links = alt[i]
		}

		if i < len(neu) {
			rechts = neu[i]
		}

		if links != rechts {
			return "Zeile " + strconv.Itoa(i+1) + ":\n  gelesen:     " + quote(links) +
				"\n  geschrieben: " + quote(rechts)
		}
	}

	return "die Zeilen sind gleich — der Unterschied liegt am Dateiende"
}

// quote macht Unsichtbares sichtbar.
//
// **Sonst steht zweimal dieselbe Zeile da.** Die erste Gegenprobe meldete
// „gelesen: …" und „geschrieben: …" mit identischem Text — der Unterschied war
// ein `\r`, und eine Meldung, die ihren eigenen Befund verschluckt, ist keine.
func quote(zeile string) string {
	sichtbar := strings.NewReplacer(
		"\t", "\\t",
		"\r", "\\r",
		"\n", "\\n",
		"\ufeff", "<BOM>",
	).Replace(zeile)

	return "\"" + sichtbar + "\""
}

// istText sagt, ob eine Datei als MSE-Datenbaum gemeint sein kann.
//
// **Kein Nullbyte — mehr nicht.** Der erste Anlauf verlangte zusätzlich gültiges
// UTF-8, und der Nachlauf des Zweitblicks hat gezeigt, warum das falsch ist:
// „Zeichenkodierung im Detail — die BOM sagt UTF-8, aber alte Pakete?" steht in
// `docs/FORMATE.md` unter *Noch offen*. Eine Datei in fremder Kodierung wäre als
// Bild durchgegangen — und zwar ausgerechnet im Fremdpaketlauf, der genau solche
// Dateien finden soll.
//
// Für den Rundlauf ist die Kodierung ohnehin gleichgültig: der Schreiber gibt
// die Rohzeile Byte für Byte zurück, gleich was in ihr steht.
func istText(roh []byte) bool {
	return !bytes.ContainsRune(roh, 0)
}
