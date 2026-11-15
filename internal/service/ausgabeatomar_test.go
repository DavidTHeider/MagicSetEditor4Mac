// **Eine gescheiterte Ausgabe zerstört keine vorhandene Datei** — der zweite
// Zweitblick über die Freigabe 0.3.0, Befund 5
// (#1170).
//
// # Der Befund
//
// `schreibeAtomar` steht seit
// #853 da und
// beschreibt in seinem Kopf genau diese Fehlerklasse. **Zwei Ausgabewege
// benutzten ihn nicht**: MTG JSON und der Druckbogen schrieben ihr endgültiges
// Ziel mit `os.WriteFile`.
//
// Über dem einen stand dabei wörtlich „Erst bauen, dann schreiben. Ein PDF,
// das zur Hälfte dasteht, sieht im Dateibrowser aus wie ein ganzes — dieselbe
// Regel wie bei den Bildern." **Es war nicht dieselbe Regel.** Der Satz fängt
// einen Fehler des Erzeugers; `O_TRUNC` fängt er nicht.
//
// # Wie der rote Lauf herbeigeführt wird
//
// An dem Preis, den `schreibeAtomar` in seinem eigenen Kopf nennt: in einem
// **nur lesbaren Verzeichnis** mit vorhandener Zieldatei scheitert es, wo
// `os.WriteFile` noch überschrieben hätte — Schreibrecht braucht das eine am
// Verzeichnis, das andere nur an der Datei.
//
// Damit ist der Unterschied **beobachtbar**: vorher war die alte Datei weg,
// nachher steht sie unverändert da. Das ist die Zusage selbst und nicht ihr
// Stellvertreter.
package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// altbestand legt eine Zieldatei mit Inhalt in ein danach nur lesbares
// Verzeichnis und gibt ihren Pfad.
func altbestand(t *testing.T, name string) string {
	t.Helper()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, name)

	if err := os.WriteFile(ziel, []byte("der vorige Stand"), 0o600); err != nil {
		t.Fatalf("die alte Datei anlegen: %v", err)
	}

	if err := os.Chmod(ordner, 0o500); err != nil { // #nosec G302 -- ein Verzeichnis, und der Entzug ist der Zweck
		t.Fatalf("das Verzeichnis schließen: %v", err)
	}

	// **Zurückgedreht, damit `t.TempDir` aufräumen kann.** Ein Verzeichnis
	// ohne Schreibrecht lässt sich nicht leeren, und der Aufräumfehler machte
	// jeden Lauf rot, der diese Vorrichtung benutzt.
	t.Cleanup(func() { _ = os.Chmod(ordner, 0o700) }) //nolint:errcheck,gosec // nur fürs Aufräumen

	return ziel
}

// pruefeUnveraendert sichert zu, dass die alte Datei noch dasteht.
func pruefeUnveraendert(t *testing.T, ziel string) {
	t.Helper()

	inhalt, err := os.ReadFile(ziel) // #nosec G304 -- der Pfad stammt aus t.TempDir()
	if err != nil {
		t.Fatalf("die alte Datei lesen: %v", err)
	}

	if string(inhalt) != "der vorige Stand" {
		t.Errorf("die alte Datei ist %q — eine gescheiterte Ausgabe hat sie "+
			"überschrieben", string(inhalt))
	}
}

func TestEineGescheiterteMTGJSONAusfuhrLaesstDieAlteDateiStehen(t *testing.T) {
	t.Parallel()

	ziel := altbestand(t, "PRB.json")

	if _, err := service.NeuerExporteur(probenbestand()).
		MTGJSON(t.Context(), 7, "magic", ziel); err == nil {
		t.Fatal("kein Fehler — in einem nur lesbaren Verzeichnis muss die " +
			"Ausfuhr scheitern, statt die vorhandene Datei zu ersetzen")
	}

	pruefeUnveraendert(t, ziel)
}

func TestEinGescheiterterDruckbogenLaesstDieAlteDateiStehen(t *testing.T) {
	t.Parallel()

	ziel := altbestand(t, "Bogen.pdf")

	if _, err := bogenleger(t, 1).Lege(t.Context(), 7, ziel,
		service.StandardBogen(), domain.Kartenauswahl{}); err == nil {
		t.Fatal("kein Fehler — in einem nur lesbaren Verzeichnis muss der " +
			"Bogen scheitern, statt die vorhandene Datei zu ersetzen")
	}

	pruefeUnveraendert(t, ziel)
}
