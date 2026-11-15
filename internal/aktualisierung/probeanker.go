//go:build probeanker

package aktualisierung

// **Ein Probeanker für die Läufe über eigene Programme** — SEC03
// (#2356).
//
// # Wozu
//
// `internal/flaechen` fährt die Aktualisierung über `mse` und `mse-rest`,
// und die prüfen gegen den Anker, der **in ihnen** steht. Bis SEC03
// signierten die Läufe deshalb mit dem echten Freigabeschlüssel aus dem
// Schlüsselbund dieses Rechners; das Issue verlangt, dass echte Schlüssel
// außerhalb der automatischen Probe bleiben.
//
// **Unter diesem Tag** liest ein Programm seinen Anker aus der Datei, die
// `MSE_PROBEANKER` nennt — einen öffentlichen Schlüssel, den der Lauf eben
// erzeugt hat. Ohne die Variable gilt der eingebettete.
//
// # Warum das keine Hintertür ist
//
// **Kein Freigabeweg setzt das Tag**: `scripts/bauen.sh` und
// `scripts/release.sh` bauen ohne, und `TestKeinFreigabewegBautMitProbeanker`
// (`internal/werkzeuge`) hält es. Ein ausgeliefertes Programm kennt die
// Variable gar nicht — dieser Code ist nicht in ihm.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// Probeankerumgebung nennt die Datei mit dem öffentlichen Probeschlüssel.
const Probeankerumgebung = "MSE_PROBEANKER"

func init() {
	pfad := os.Getenv(Probeankerumgebung)
	if pfad == "" {
		return
	}

	datei, err := os.Open(pfad) //nolint:gosec // nur in Testbauten, Pfad vom Testlauf
	if err != nil {
		panic(fmt.Sprintf("den Probeanker öffnen: %v", err))
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	// **Begrenzt**, wie jedes ganze Lesen (`lesebudget_test.go`): ein
	// öffentlicher Schlüssel ist ein paar Kilobyte, [HoechstensSignatur] das
	// Vielfache davon.
	roh, err := io.ReadAll(io.LimitReader(datei, HoechstensSignatur))
	if err != nil {
		panic(fmt.Sprintf("den Probeanker lesen: %v", err))
	}

	gelesen, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(roh))
	if err != nil || len(gelesen) != 1 {
		panic(fmt.Sprintf("der Probeanker ist kein einzelner Schlüssel: %v", err))
	}

	freigabeschluessel = string(roh)
	freigabeanker = []string{strings.ToUpper(fmt.Sprintf("%X", gelesen[0].PrimaryKey.Fingerprint))}
}
