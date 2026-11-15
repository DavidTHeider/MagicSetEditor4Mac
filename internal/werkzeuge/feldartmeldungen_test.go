// **Die Feldartbefunde der Werkstatt stehen in beiden Katalogen** — FMT-02
// (#1691).
//
// # Der Befund
//
// Die Prüfung einer Paketdatei schickt Schlüssel und Werte neben dem
// deutschen Satz, und das Fenster zeigte trotzdem den deutschen Satz — auch
// im englischen Katalog. Seit FMT-02 übersetzt es; die zwei Schlüssel der
// Feldart stehen dafür unter `seiten.vorlagen` und nicht im Block `skript`,
// den [TestJedeSkriptmeldungStehtInBeidenKatalogen] gegen den Zerteiler hält.
//
// # Was er nicht hält
//
// **Ob das Fenster den Schlüssel nachschlägt.** Das hält
// `frontend/src/seiten/vorlagenbefund.test.tsx`, und die Werte in den Sätzen
// zählt [TestDieFeldartsaetzeHabenIhrePlatzhalter].
package werkzeuge_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// feldartsaetze gibt die zwei Sätze eines Katalogs, nach Schlüssel.
func feldartsaetze(t *testing.T, datei string) map[string]string {
	t.Helper()

	inhalt := lies(t, datei)
	saetze := map[string]string{}

	for _, schluessel := range []string{
		string(service.SchluesselFeldartFehlt), string(service.SchluesselFeldartUnbekannt),
	} {
		treffer := regexp.MustCompile(`(?m)^  ` + schluessel + `:\s*\n?\s*'([^']*)'`).FindStringSubmatch(inhalt)
		if treffer == nil {
			t.Errorf("%s kennt den Befund %q nicht", datei, schluessel)

			continue
		}

		saetze[schluessel] = treffer[1]
	}

	return saetze
}

func TestDieFeldartsaetzeHabenIhrePlatzhalter(t *testing.T) {
	t.Parallel()

	for _, datei := range []string{
		"frontend/src/sprache/de/seitenvorlagen.ts",
		"frontend/src/sprache/en/seitenvorlagen.ts",
	} {
		saetze := feldartsaetze(t, datei)

		// **Die Art und die Liste der bekannten** — so viele Werte setzt
		// `feldartfehler` in `internal/service/paketpruefung.go`.
		if satz := saetze[string(service.SchluesselFeldartUnbekannt)]; !strings.Contains(satz, "{0}") ||
			!strings.Contains(satz, "{1}") {
			t.Errorf("%s: %q nennt die Art oder die bekannten nicht", datei, satz)
		}

		if satz := saetze[string(service.SchluesselFeldartFehlt)]; strings.Contains(satz, "{") {
			t.Errorf("%s: %q hat einen Platzhalter, und Go setzt keinen Wert", datei, satz)
		}
	}
}
