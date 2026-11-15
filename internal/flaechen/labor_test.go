package flaechen_test

// **Das Ein-Mac-Labor: getrennt ist, was getrennt wirkt** — QA-10.AK-01 und
// NAT-01 (#1741).
//
// # Was hier bewiesen wird, und wie
//
// **Nicht, dass Umgebungsvariablen gesetzt sind** — das Issue sagt es
// ausdrücklich. Zwei Instanzen tun etwas, das eine Spur hinterlässt, und
// danach wird nachgesehen, **wo** sie steht: ein Set, eine Einstellung, eine
// Identität, eine Datei und ein horchender Anschluss je Instanz — und nichts
// davon bei der anderen. Jede Instanz hat ein eigenes Zuhause — und damit
// eigene Pakete, die dem Rechner gehören —, einen eigenen Bestand samt
// Bildern und einen eigenen Schlüsselbund ([stelleAuf]).
//
// # Was es nicht beweist
//
// **Zwei Rechner.** Beide Instanzen laufen auf dieser Maschine, über die
// Schleifenadresse; was WLAN, ein Router oder eine Firewall dazwischenlegen,
// bleibt eine Abnahme von Hand (QA-10.AK-04, `docs/LABOR.md`).
//
// **Das Protokoll in der Datei.** Es schreibt nur das Fenster; wo es liegt,
// hält `TestWasDemBestandGehoertFolgtIhm`. Was ein horchender
// Prozess protokolliert, geht auf seine eigene Fehlerausgabe.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// anwendungsortVon gibt den Ort, an dem eine Instanz ihren Bestand hält.
func anwendungsortVon(t *testing.T, rechner einRechner) string {
	t.Helper()

	return filepath.Dir(bestandsortAus(t, rechner.umgebung))
}

// zuhauseVon gibt das Zuhause einer Instanz — den letzten Eintrag, wie
// `os/exec` ihn liest.
func zuhauseVon(t *testing.T, rechner einRechner) string {
	t.Helper()

	zuhause := ""

	for _, eintrag := range rechner.umgebung {
		if wert, gefunden := strings.CutPrefix(eintrag, "HOME="); gefunden {
			zuhause = wert
		}
	}

	if zuhause == "" {
		t.Fatal("die Umgebung nennt kein Zuhause")
	}

	return zuhause
}

// dateienMit nennt jede Datei unter einer Wurzel, deren Inhalt den Text trägt.
func dateienMit(t *testing.T, wurzel, text string) []string {
	t.Helper()

	var treffer []string

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil || eintrag.IsDir() || eintrag.Type()&fs.ModeSymlink != 0 {
			return fehler
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // unter einem t.TempDir()
		if lesefehler == nil && strings.Contains(string(roh), text) {
			treffer = append(treffer, pfad)
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", wurzel, fehler)
	}

	return treffer
}

func TestDasLaborTrenntZweiInstanzen(t *testing.T) {
	anna := stelleAuf(t)
	bert := stelleAuf(t)

	// **Die Orte sind verschieden** — die Voraussetzung, nicht der Beweis.
	if anwendungsortVon(t, anna) == anwendungsortVon(t, bert) ||
		zuhauseVon(t, anna) == zuhauseVon(t, bert) {
		t.Fatal("beide Instanzen teilen sich einen Ort")
	}

	// **Jede hinterlässt eine Spur**, die es nur bei ihr geben darf.
	const spur = "Nur bei Anna"

	mussKommandozeile(t, anna.umgebung, "set-neu", spur, "probe", "probe-schlicht")
	mussKommandozeile(t, anna.umgebung, "einstellungen", "setzen", "sprache", "en")
	mussKommandozeile(t, bert.umgebung, "einstellungen", "setzen", "sprache", "de")

	t.Run("der Bestand", func(t *testing.T) {
		if sets := mussKommandozeile(t, anna.umgebung, "sets"); !strings.Contains(sets, spur) {
			t.Errorf("Anna sieht ihr eigenes Set nicht: %q", sets)
		}

		if sets := mussKommandozeile(t, bert.umgebung, "sets"); strings.Contains(sets, spur) {
			t.Errorf("Bert sieht Annas Set: %q", sets)
		}
	})

	t.Run("die Einstellungen", func(t *testing.T) {
		for _, fall := range []struct {
			rechner  einRechner
			erwartet string
		}{{anna, "sprache\ten"}, {bert, "sprache\tde"}} {
			if stand := mussKommandozeile(t, fall.rechner.umgebung, "einstellungen"); !strings.Contains(stand, fall.erwartet) {
				t.Errorf("erwartet %q in %q", fall.erwartet, stand)
			}
		}
	})

	t.Run("die Identität", func(t *testing.T) {
		if anna.kennung == bert.kennung {
			t.Fatal("beide Instanzen tragen dieselbe Peerkennung")
		}

		for _, rechner := range []einRechner{anna, bert} {
			if eigene := zeilenwert(t, mussKommandozeile(t, rechner.umgebung, "peer"), "kennung"); eigene != rechner.kennung {
				t.Errorf("die Instanz nennt %q, angelegt war %q", eigene, rechner.kennung)
			}
		}
	})

	t.Run("die Dateien", func(t *testing.T) {
		if bei := dateienMit(t, zuhauseVon(t, anna), spur); len(bei) == 0 {
			t.Error("die Spur steht in keiner Datei unter Annas Zuhause — die Suche greift ins Leere")
		}

		if bei := dateienMit(t, zuhauseVon(t, bert), spur); len(bei) != 0 {
			t.Errorf("Annas Spur steht unter Berts Zuhause: %v", bei)
		}
	})

	t.Run("die Anschlüsse", func(t *testing.T) {
		annasAdresse, _ := horcheImHintergrund(t, anna)
		bertsAdresse, _ := horcheImHintergrund(t, bert)

		if annasAdresse == bertsAdresse {
			t.Errorf("beide horchen an %s", annasAdresse)
		}
	})
}
