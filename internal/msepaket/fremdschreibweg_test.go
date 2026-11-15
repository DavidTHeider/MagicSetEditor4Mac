//go:build fremdpakete

// **Der Schreibweg über ein bezogenes Paket** — Werkstück 2 von Phase 3.3
// (#1164).
//
// # Was er beantwortet, das `fremdpakete_test.go` offenlässt
//
// Der Lauf daneben prüft den Rundlauf **einer Datei durch Zerteiler und
// Schreiber**. Dieser hier prüft den Rundlauf **eines ganzen Pakets durch den
// Schreibweg**: eine Datei wird unverändert zurückgeschrieben, und danach muss
// jede Datei des Pakets Byte für Byte dieselbe sein.
//
// Das ist die Zusage aus dem Issue — „ein MSE2-Rundlauf weist unveränderte
// Bereiche bytegenau nach" — an echten Daten statt an nachgebauten.
//
// # Was er nicht zeigt
//
// **Dass MSE2 das Ergebnis liest.** Das steht in `docs/RELEASE.md` als
// Handgriff; ein Programm, das hier nicht läuft, kann es nicht behaupten
// (§10).
//
//	MSE_FREMDPAKETE=~/full-magic-pack \
//	  go test -tags fremdpakete ./internal/msepaket/ -run FremdenSchreibweg -v
package msepaket_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDerSchreibwegUeberFremdenSchreibweg(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt — kein Verzeichnis mit bezogenen Paketen")
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", wurzel, err)
	}

	geprueft := 0

	for _, eintrag := range eintraege {
		if msepaket.ArtAusName(eintrag.Name()) == msepaket.ArtUnbekannt {
			continue
		}

		geprueft++

		t.Run(eintrag.Name(), func(t *testing.T) {
			// **Gearbeitet wird auf einer Kopie.** Ein Lauf, der ein bezogenes
			// Paket anfasst, wäre kein Test, sondern ein Eingriff — und beim
			// zweiten Durchgang prüfte er den Stand des ersten.
			ort := kopiertesPaket(t, filepath.Join(wurzel, eintrag.Name()))

			vorher := alleDateien(t, ort)

			name, inhalt := irgendeine(t, vorher)
			if name == "" {
				t.Skip("das Paket hat keine Datei zum Zurückschreiben")
			}

			if fehler := msepaket.DateiSchreiben(t.Context(), ort, name, inhalt, summe(inhalt)); fehler != nil {
				t.Fatalf("%s zurückschreiben: %v", name, fehler)
			}

			nachher := alleDateien(t, ort)

			if len(nachher) != len(vorher) {
				t.Fatalf("aus %d Dateien wurden %d", len(vorher), len(nachher))
			}

			for datei, alt := range vorher {
				if !bytes.Equal(nachher[datei], alt) {
					t.Errorf("%s hat sich geändert — %d statt %d Byte",
						datei, len(nachher[datei]), len(alt))
				}
			}
		})
	}

	if geprueft == 0 {
		t.Fatalf("in %s liegt kein einziges Paket", wurzel)
	}

	t.Logf("%d Pakete durch den Schreibweg", geprueft)
}

// kopiertesPaket legt ein Paket unter einem eigenen Ort noch einmal ab.
func kopiertesPaket(t *testing.T, quelle string) string {
	t.Helper()

	ziel := filepath.Join(t.TempDir(), filepath.Base(quelle))

	if err := os.CopyFS(ziel, os.DirFS(quelle)); err == nil {
		return ziel
	}

	// **Keine Verzeichnisse, also eine Datei** — ein Paket ist beides (§2),
	// und `os.CopyFS` nimmt nur das eine.
	roh, err := os.ReadFile(quelle) // #nosec G304,G703 -- aus MSE_FREMDPAKETE, von Hand gesetzt
	if err != nil {
		t.Fatalf("%s lesen: %v", quelle, err)
	}

	if fehler := os.WriteFile(ziel, roh, 0o600); fehler != nil { // #nosec G703 -- aus t.TempDir()
		t.Fatalf("%s schreiben: %v", ziel, fehler)
	}

	return ziel
}

// alleDateien liest jede Datei eines Pakets.
func alleDateien(t *testing.T, ort string) map[string][]byte {
	t.Helper()

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	namen, fehler := paket.Dateien()
	if fehler != nil {
		t.Fatalf("aufzählen: %v", fehler)
	}

	alle := make(map[string][]byte, len(namen))

	for _, name := range namen {
		roh, lesefehler := paket.Datei(name)
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", name, lesefehler)
		}

		alle[name] = roh
	}

	return alle
}

// irgendeine wählt eine Datei — die erste in sortierter Reihenfolge.
//
// **Sortiert und nicht die erste beste**: eine Map hat keine Ordnung, und ein
// Lauf, der jedes Mal eine andere Datei nähme, wäre bei jedem Durchgang ein
// anderer Test.
func irgendeine(t *testing.T, dateien map[string][]byte) (name string, inhalt []byte) {
	t.Helper()

	gewaehlt := ""

	for name := range dateien {
		if gewaehlt == "" || name < gewaehlt {
			gewaehlt = name
		}
	}

	return gewaehlt, dateien[gewaehlt]
}
