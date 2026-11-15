// Package biometrie fragt den Rechner, ob sein Eigentümer davorsitzt.
//
// **Ein Blatt** (§2): es kennt `internal/domain` für seine Fehlerklassen und
// sonst nichts aus diesem Projekt. Werkstück 7 von Phase 4.12
// (#1351),
// entschieden in
// [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md):
// „Beides vorsehen, auch Biometrie. Vor 1.0 ist kein Hinderungsgrund."
//
// # Was hier eingekauft wird, und was es kostet
//
// **`LocalAuthentication` hat keine Kommandozeile.** Anders als beim
// Schlüsselbund, wo `/usr/bin/security` den Weg ohne cgo öffnet, gibt es hier
// keinen von Apple dokumentierten. Der Weg geht über
// [`tmc/apple`](https://github.com/tmc/apple) und `purego` — cgo-frei, MIT.
//
// **Der Preis ist ausgesprochen und nachgemessen** (2026-09-05): `tmc/apple`
// steht bei **v0.6.18** und hängt selbst an `purego v0.11.0-**alpha.6**`.
// purego ist seit dem 2026-10-11 auf der Endfassung v0.11.1 (#2305), aber
// weiter vor 1.0, und tmc/apple bleibt es auch.
// Zwei unfertige Schichten unter einem Sicherheitsmerkmal — deshalb steht
// alles hier hinter [Gefangen], und deshalb ist **TOTP** der Weg, der immer
// geht.
//
// # Was gemessen ist und was nicht
//
// **Gemessen am 2026-09-05 auf diesem Rechner**: `CanEvaluatePolicyError`
// gibt für beide Richtlinien `true`, `BiometryType` meldet `TouchID`.
//
// **Nicht gemessen**: dass eine Abfrage einen Finger erkennt. Dazu müsste
// jemand den Sensor berühren, und ein Testlauf, der darauf wartet, bleibt
// stehen. Der Weg dorthin steht in `.covignore` mit diesem Grund.
package biometrie

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Auskunft sagt, ob Biometrie hier geht — und warum nicht.
//
// **Beides zusammen und nicht zwei Rufe**: „geht nicht" ohne Grund ist eine
// Meldung, mit der niemand etwas anfangen kann (§8). Die Oberfläche schreibt
// den Grund neben den ausgegrauten Schalter.
type Auskunft struct {
	// Vorhanden sagt, ob eine Abfrage überhaupt versucht werden kann.
	Vorhanden bool

	// Grund steht da, wenn sie es nicht kann — leer, wenn sie es kann.
	Grund string

	// Art ist das Merkmal, das dieser Rechner hat — „Touch ID", „Face ID"
	// oder leer.
	Art string
}

// Stand fragt den Rechner, ob eine Abfrage möglich wäre.
//
// **Ohne etwas anzustoßen**: `canEvaluatePolicy` zeigt keinen Dialog und
// verlangt keinen Finger. Wer diese Auskunft holt, hat noch nichts getan —
// dieselbe Zusage wie bei [app.App.Anbieter] für die Anmeldung.
func Stand(ctx context.Context) Auskunft { return stand(ctx) }

// Bestaetige fragt den Eigentümer — mit Dialog und Sensor.
//
// **Der Grund steht im Dialog**, und macOS verlangt ihn: `localizedReason`
// ist der Satz, den die Person liest, bevor sie den Finger auflegt.
func Bestaetige(ctx context.Context, grund string) error {
	return Gefangen(func() error { return bestaetige(ctx, grund) })
}

// Gefangen ruft etwas und macht aus einem Absturz eine Antwort.
//
// # Warum das nötig ist
//
// **Weil die Bindung vor 1.0 ist und ihre eigene Abhängigkeit im
// Alpha-Stand.** Ein `panic` aus einer Fremdbibliothek risse ohne diesen
// Riegel den ganzen Prozess mit — und mit ihm das Fenster, an dem jemand
// gerade arbeitet.
//
// **Ein gewöhnlicher Fehler geht unverändert durch.** Der Riegel fängt
// Abstürze, nicht Antworten: „der Finger passte nicht" ist eine Antwort.
func Gefangen(arbeit func() error) (fehler error) {
	defer func() {
		if gefangen := recover(); gefangen != nil {
			fehler = fmt.Errorf(
				"%w: die Biometrie dieses Rechners hat ausgesetzt (%v) — "+
					"der Code aus der Authenticator-App geht weiter",
				domain.ErrValidation, gefangen)
		}
	}()

	return arbeit()
}
