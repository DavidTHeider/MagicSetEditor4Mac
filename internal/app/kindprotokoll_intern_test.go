package app

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// TestEinDienstSchreibtInsProtokollDesFensters hält Nachlauf über Phase 6.4,
// N3: ein vom Fenster gestarteter Dienst schrieb seine Zeilen ins Nullgerät,
// und die Vorgangskennung, die Dokumentation und Antwort nennen, führte ins
// Leere. **Zeile für Zeile**, auch wenn eine in zwei Stücken ankommt.
func TestEinDienstSchreibtInsProtokollDesFensters(t *testing.T) {
	t.Parallel()

	var puffer bytes.Buffer

	kind := neuesKindprotokoll("mse-rest", slog.New(slog.NewTextHandler(&puffer, nil)))

	for _, stueck := range []string{"level=INFO vorgang=0a1b", "2c3d msg=erste\nzweite\n", "\n"} {
		if geschrieben, fehler := kind.Write([]byte(stueck)); fehler != nil || geschrieben != len(stueck) {
			t.Fatalf("schreiben: %d, %v", geschrieben, fehler)
		}
	}

	zeilen := strings.Split(strings.TrimSpace(puffer.String()), "\n")
	if len(zeilen) != 2 || !strings.Contains(zeilen[0], "vorgang=0a1b2c3d") ||
		!strings.Contains(zeilen[0], "dienst=mse-rest") || !strings.Contains(zeilen[1], "zweite") {
		t.Errorf("%q", zeilen)
	}
}

// **Eine Zeile ohne Ende füllt den Speicher nicht** — ab 64 KiB wird sie
// geschrieben, wie sie ist.
func TestEineEndloseZeileWirdTrotzdemGeschrieben(t *testing.T) {
	t.Parallel()

	var puffer bytes.Buffer

	kind := neuesKindprotokoll("mse-mcp", slog.New(slog.NewTextHandler(&puffer, nil)))

	if _, fehler := kind.Write(bytes.Repeat([]byte("x"), hoechsteZeile+1)); fehler != nil {
		t.Fatal(fehler)
	}

	if puffer.Len() < hoechsteZeile || len(kind.rest) != 0 {
		t.Errorf("geschrieben %d, übrig %d", puffer.Len(), len(kind.rest))
	}
}
