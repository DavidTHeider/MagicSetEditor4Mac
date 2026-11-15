package flaechen_test

// **Auch `gpg` und `security` bekommen eine eigene Umgebung** — C06-Rest
// aus dem Nachlauf über Phase C
// (#2284).
//
// Beide liefen mit der Umgebung des Testlaufs und erbten das echte HOME.
// `security` braucht davon nichts. **`gpg` ruft seit SEC03 (#2356) kein Lauf
// mehr** — er signierte mit dem Herausgeberschlüssel aus dem Schlüsselbund
// dieses Rechners; der Wächter unten hielte einen neuen Aufruf trotzdem.

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSecurityLaeuftOhneDasEchteZuhause(t *testing.T) {
	echt := echtesZuhause(t)
	umgebung := securityUmgebung(t)

	if zuhause := zuletzt(umgebung, "HOME"); zuhause == "" || zuhause == echt {
		t.Errorf("HOME ist %q — erwartet ein eigenes", zuhause)
	}

	// **Die Umgebung wird nicht ausgegeben** — sie trüge Marken des Laufs
	// in jedes Protokoll.
	if treffer := slices.IndexFunc(umgebung, func(e string) bool { return strings.Contains(e, echt+"/") }); treffer >= 0 {
		t.Errorf("die Umgebung nennt das echte Zuhause (Eintrag %d von %d)", treffer, len(umgebung))
	}
}

// securityUmgebung ist die Umgebung für `/usr/bin/security`: ein eigenes
// Zuhause und die Systempfade, sonst nichts.
func securityUmgebung(t *testing.T) []string {
	t.Helper()

	return []string{"HOME=" + t.TempDir(), "TMPDIR=" + t.TempDir(), "PATH=/usr/bin:/bin"}
}

// **Und jeder Aufruf nimmt sie** — die Helfer allein belegen nichts, wenn ein
// Aufruf sie nicht setzt. Gelesen wird der Quelltext dieses Pakets: auf
// jedes `gpg` und `/usr/bin/security` folgt binnen acht Zeilen ein `.Env =`.
// **Was er nicht hält**: ein Programm unter einem anderen Namen.
func TestJederAufrufVonGpgUndSecurityHatEineEigeneUmgebung(t *testing.T) {
	dateien, fehler := filepath.Glob("*_test.go")
	if fehler != nil || len(dateien) == 0 {
		t.Fatalf("keine Testdateien: %v", fehler)
	}

	gesehen := 0

	for _, datei := range dateien {
		roh, lesefehler := os.ReadFile(datei) // #nosec G304 -- eine Datei dieses Pakets
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		zeilen := strings.Split(string(roh), "\n")
		for stelle, zeile := range zeilen {
			if !strings.Contains(zeile, "exec.CommandContext(") ||
				(!strings.Contains(zeile, `"gpg"`) && !strings.Contains(zeile, `"/usr/bin/security"`)) {
				continue
			}

			gesehen++
			ende := min(stelle+8, len(zeilen))
			if !slices.ContainsFunc(zeilen[stelle:ende], func(z string) bool { return strings.Contains(z, ".Env = ") }) {
				t.Errorf("%s:%d startet ein Programm ohne eigene Umgebung", datei, stelle+1)
			}
		}
	}

	// **Beleg, dass der Wächter etwas sieht** (§3). **Drei seit SEC03**
	// (#2356): `gpg` ruft kein Lauf mehr — er signierte mit dem Schlüsselbund
	// dieses Rechners, und das hält jetzt `internal/werkzeuge`
	// (`echterschluessel_test.go`).
	if gesehen < 3 {
		t.Errorf("nur %d Aufrufe gefunden — erwartet drei von security", gesehen)
	}
}
