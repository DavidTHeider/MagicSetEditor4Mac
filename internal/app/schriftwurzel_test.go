// **Die Naht zum Ort der Paketschriften** —
// #1170.
//
// # Warum es sie gibt
//
// Die **Stilwurzeln** sind nicht die **Schriftwurzel**. `msepaket.Paketwurzeln`
// gibt die Verzeichnisse, die Pakete *enthalten* — bei `M15-Magic-Pack` also
// `…/data`; die Schriften liegen daneben unter `Magic - Fonts/` und fallen
// heraus. Im Betrieb fällt das nicht auf, weil beide unter derselben Wurzel im
// Zuhause liegen — in einem Messlauf mit umgelenktem `HOME` schon, und dort
// hat es zwei Freigaben lang jede Karte ohne Text gezeichnet.
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Sie lässt sich setzen und kommt zurück.**
//
// **Beides in einem Lauf, und das ist die Zusage**: eine Naht, die man setzen,
// aber nicht zurücknehmen kann, färbt jeden folgenden Test — und ein
// Messlauf, der sie setzt, liefe danach mit einer Wurzel weiter, die niemand
// mehr meint.
func TestDieSchriftwurzelLaesstSichSetzenUndZuruecknehmen(t *testing.T) {
	vorher := app.Schriftwurzel()

	zurueck := app.MitSchriftwurzel(func() string { return "/anderswo" })

	if gesetzt := app.Schriftwurzel(); gesetzt != "/anderswo" {
		t.Errorf("die gesetzte Wurzel ist %q", gesetzt)
	}

	zurueck()

	if wieder := app.Schriftwurzel(); wieder != vorher {
		t.Errorf("nach dem Zurücknehmen steht %q, vorher stand %q", wieder, vorher)
	}
}
