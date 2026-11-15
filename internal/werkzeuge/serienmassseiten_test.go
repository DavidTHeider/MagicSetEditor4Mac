// **Das Serienmaß misst jede Anwendungsseite** — Werkstück A von Phase 4.10
// (#1325).
//
// # Warum es diesen Wächter gibt
//
// `scripts/serienmass.mjs` führte seine Seiten als **Aufzählung von Hand**.
// Sie ist beim ersten Lauf am 2026-08-30 geschrieben worden, und die Phasen
// 4.6 und 4.7 haben sie nicht mitgezogen: **Decks und Turniere fehlten**,
// seit es sie gibt. Die Tabelle in `docs/SERIENARBEIT.md` sah vollständig aus
// und war es nie.
//
// Das ist dieselbe Falle, vor der `AGENTS.md` §3 beim Barrierefreiheitstest
// ausdrücklich warnt: *„eine zu kurze Aufzählung ist grün, und niemand sieht
// es."* Dort kommt die Liste deshalb aus `navigation.ts`; hier geht das nicht
// — das Messskript läuft ohne Bündler und kann kein TypeScript einlesen. Also
// hält ein Wächter die beiden zusammen.
//
// **Seit dem 2026-09-03 hält er beide Listen.** Bis dahin sah er die
// **Laschen eines Sets** nicht — dreizehn von zweiundzwanzig Seiten, und
// darunter die beiden größten der Anwendung. Hier stand als Begründung, sie
// hätten „keinen eigenen Weg in der Leiste", und das stimmt; es ist ein Grund
// für einen anderen **Weg**, nicht für eine fehlende **Messung**. Das
// Messgerät fährt sie jetzt, nachdem es ein Set geöffnet hat.
//
// # Was er nicht hält
//
// **Ob die Messung stimmt.** Er sieht, dass jede Seite in der Liste steht —
// nicht, ob der Weg dorthin führt oder die Zahlen etwas taugen.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestDasSerienmassMisstJedeSeite(t *testing.T) {
	messliste := gelesen(t, filepath.Join("..", "..", "scripts", "serienmasslisten.mjs"))
	navigation := gelesen(t, filepath.Join("..", "..", "frontend", "src", "navigation.ts"))

	gemessen := namenAusMessliste(messliste)

	// **Beide Listen, und beide über ihren eigenen Namen.** Eine gemeinsame
	// Schleife über `seitenschluessel` ginge auch — sie verlöre nur die
	// Auskunft, *welche* Art fehlt, und genau die ist beim Beheben die halbe
	// Arbeit.
	for name, quelle := range map[string]string{
		"anwendungsschluessel": "export const anwendungsschluessel = [",
		"setschluessel":        "export const setschluessel = [",
	} {
		gefuehrt := schluesselAus(navigation, quelle)

		// **Eine leere Liste macht diesen Wächter hohl** — sein erster Anlauf
		// war genau das, und nur die Gegenprobe hat es gezeigt. Die Zahl steht
		// hier bewusst niedrig: sie soll einen **leeren** Leser fangen, nicht
		// eine Seite zählen.
		if len(gefuehrt) < 5 {
			t.Fatalf("%s führt %d Seiten — der Leser greift ins Leere",
				name, len(gefuehrt))
		}

		for _, seite := range gefuehrt {
			if !gemessen[seite] {
				t.Errorf("die Seite %q steht in navigation.ts unter %s, wird aber "+
					"nicht gemessen — scripts/serienmasslisten.mjs führt sie nicht "+
					"(AGENTS.md §3)", seite, name)
			}
		}
	}
}

// namenAusMessliste liest die `schluessel:`-Angaben aus dem Messskript.
//
// **Über den Schlüssel und nicht über den Namen** — und das ist der zweite
// Anlauf. Der erste las `name:` auf beiden Seiten und war **hohl**:
// `navigation.ts` führt gar keine Namen, die stehen im Sprachkatalog (§9).
// Die Schleife lief über eine leere Liste, der Wächter war grün, und erst die
// Gegenprobe hat es gezeigt — eine entfernte Seite ließ ihn kalt.
var messeintrag = regexp.MustCompile(`\{\s*schluessel:\s*"([^"]+)"`)

func namenAusMessliste(quelle string) map[string]bool {
	namen := map[string]bool{}

	for _, treffer := range messeintrag.FindAllStringSubmatch(quelle, -1) {
		namen[treffer[1]] = true
	}

	return namen
}

// schluesselAus liest die Schlüssel einer der beiden Listen.
//
// **Aus `anwendungsschluessel` und `setschluessel`, nicht aus `seiten`.** Die
// beiden sind genau die Antwort auf „hängt diese Seite an einem Set"; `seiten`
// führt sie zusammen, und `istSetseite` hier nachzubauen hieße, die
// Unterscheidung ein zweites Mal zu treffen.
//
// **Ohne Fundstelle gibt sie nichts zurück** — und der Aufrufer sieht eine
// leere Schleife statt eines Fehlers. Genau das war der erste Anlauf dieses
// Wächters, deshalb steht die Prüfung darauf unten im Test.
func schluesselAus(quelle, beginn string) []string {
	_, ab, gefunden := strings.Cut(quelle, beginn)
	if !gefunden {
		return nil
	}

	bis, _, gefunden := strings.Cut(ab, "] as const")
	if !gefunden {
		return nil
	}

	eintrag := regexp.MustCompile(`'([a-z]+)'`)

	var schluessel []string

	for _, treffer := range eintrag.FindAllStringSubmatch(bis, -1) {
		schluessel = append(schluessel, treffer[1])
	}

	return schluessel
}

// gelesen liest eine Datei oder lässt den Lauf scheitern.
func gelesen(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) // #nosec G304 -- fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", pfad, err)
	}

	return string(inhalt)
}
