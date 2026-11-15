package flaechen_test

// **Eine Sicherung mit Kennwort über die Flächen ohne Fenster** — OPS-06.AK-02
// (#1723).
//
// Die Kommandozeile liest das Kennwort aus `MSE_SICHERUNGSKENNWORT` und nie
// aus einem Argument; der Werkzeugserver nimmt keines — ein Kennwort, das
// durch ein Modell geht, ist keines mehr. Dass der Dienst es im Rumpf und im
// Kopf annimmt, hält `internal/rest`.

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestEineSicherungMitKennwortUeberDieFlaechen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	mitKennwort := append(slices.Clone(umgebung),
		"MSE_SICHERUNGSKENNWORT=korrekt pferd batterie heftklammer")
	ziel := filepath.Join(t.TempDir(), "geheim.msesicherung")

	mussKommandozeile(t, mitKennwort, "bestand-sichern", ziel)

	// **Ohne Kennwort zeigt sie nichts** — und sagt, woran es liegt.
	aus, fehler, code := kommandozeile(t, umgebung, "bestand-schauen", ziel)
	if code == 0 || !strings.Contains(aus+fehler, "Kennwort") {
		t.Errorf("ohne Kennwort: Code %d\n%s%s", code, aus, fehler)
	}

	if roh := mussKommandozeile(t, mitKennwort, "--json", "bestand-schauen", ziel); !strings.Contains(roh, `"verschluesselt": true`) ||
		!strings.Contains(roh, `"karten": 3`) {
		t.Errorf("mit Kennwort:\n%s", roh)
	}

	// **Der Werkzeugserver kennt kein Kennwort** und sagt es, statt eine
	// beschädigte Datei zu melden.
	if text, istFehler := mcpRuf(t, umgebung, "lesen.sicherung",
		map[string]any{"pfad": ziel}); !istFehler || !strings.Contains(text, "Kennwort") {
		t.Errorf("lesen.sicherung: %v %s", istFehler, text)
	}
}
