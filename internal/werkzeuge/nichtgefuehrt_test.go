package werkzeuge

// **Was git in diesem Baum übergeht, zählt der Umfang nicht** — gefunden am
// 2026-09-24 mit der WebKit-Probe
// (#1828).
//
// # Warum der Wächter git fragt, statt eine Liste zu führen
//
// **Weil eine Liste den nächsten Ausstoß erst kennt, wenn er schon
// zugeschlagen hat.** Zweimal hat etwas, das nie eingecheckt wird, die Zahl
// im README verschoben: am 2026-09-07 die Stücklisten unter
// `build/freigabe/` (#1412), am 2026-09-24 die Berichte der WebKit-Abnahmen
// unter `.webkitabnahme/` — ihre Kopie trug die Dokumentation über den
// Spielraum, ohne dass eine Zeile im Baum dazugekommen wäre. `.gitignore`
// kennt einen solchen Ausstoß, sobald es ihn gibt.
//
// # Was er nicht übergeht
//
// **Was eingecheckt ist und trotzdem nicht zählen soll** — `graft/`, der
// Entwurf unter `docs/prototyp/`. Dafür bleibt die Liste in [uebergangen].

import (
	"os/exec"
	"strings"
	"testing"
)

// nichtGefuehrt sind die Pfade, die git übergeht — ein ganzes Verzeichnis
// mit `/` am Ende, sonst eine Datei, relativ zur Wurzel des Baums.
type nichtGefuehrt []string

// nichtGefuehrtIn fragt git nach allem, was `.gitignore` in diesem Baum
// ausnimmt.
//
// **`--directory`** fasst ein übergangenes Verzeichnis zu einer Zeile
// zusammen: `node_modules/` ist eine Zeile und nicht hunderttausend.
func nichtGefuehrtIn(t *testing.T) nichtGefuehrt {
	t.Helper()

	ausgabe, err := exec.CommandContext(t.Context(), "git", "-C", "../..",
		"ls-files", "--others", "--ignored", "--exclude-standard",
		"--directory").Output()
	if err != nil {
		t.Fatalf("git kennt diesen Baum nicht: %v", err)
	}

	var pfade nichtGefuehrt

	for zeile := range strings.SplitSeq(string(ausgabe), "\n") {
		if zeile != "" {
			pfade = append(pfade, zeile)
		}
	}

	// **Ohne einen einzigen Eintrag sähe der Lauf nichts** — `.go/` liegt
	// in jedem Baum, in dem dieser Wächter läuft (`scripts/activate.sh`).
	if len(pfade) == 0 {
		t.Fatal("git nennt nichts, was es übergeht — der Ruf trifft den Baum nicht")
	}

	return pfade
}

// uebergeht sagt, ob ein Pfad des Laufs zu dem gehört, was git übergeht.
func (n nichtGefuehrt) uebergeht(pfad string) bool {
	relativ := strings.TrimPrefix(pfad, "../../")

	for _, eintrag := range n {
		if relativ == strings.TrimSuffix(eintrag, "/") {
			return true
		}

		if strings.HasSuffix(eintrag, "/") && strings.HasPrefix(relativ, eintrag) {
			return true
		}
	}

	return false
}
