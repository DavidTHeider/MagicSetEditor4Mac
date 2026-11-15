//go:build darwin

package biometrie

// **Die drei Rufe ins Framework, und sonst nichts.**
//
// # Warum sie allein in dieser Datei stehen
//
// **Weil sie die einzigen sind, die ohne Sensor nicht laufen** — und
// `.covignore` nimmt Dateien aus, keine Zeilen. Was daneben liegt, entscheidet
// etwas und ist geprüft (`biometrie_darwin.go`); was hier steht, reicht durch.
//
// # Die Richtlinie und warum diese
//
// **`LAPolicyDeviceOwnerAuthenticationWithBiometrics` und nicht
// `LAPolicyDeviceOwnerAuthentication`.** Die zweite schließt den Rückfall auf
// das Anmeldekennwort des Rechners ein — und das ist hier kein zweiter
// Faktor, sondern derselbe erste: wer am Rechner angemeldet ist, kennt es.
//
// **Gelesen in `LocalAuthentication.framework/Headers/LAContext.h`** des
// macOS-SDK, festgehalten in
// [0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md):
// es gibt sie seit macOS 10.12.2, und die Untergrenze dieser Anwendung ist 13
// (§12).

import (
	"context"

	"github.com/tmc/apple/localauthentication"
)

// richtlinie ist die Prüfung mit Biometrie und ohne Kennwortrückfall.
const richtlinie = localauthentication.LAPolicyDeviceOwnerAuthenticationWithBiometrics

// neueBindung öffnet einen `LAContext` und reicht seine drei Rufe durch.
func neueBindung() bindung {
	kontext := localauthentication.NewLAContext()

	return bindung{
		kann: func() (bool, error) {
			return kontext.CanEvaluatePolicyError(richtlinie)
		},
		merkmal: func() string { return artNach(kontext.BiometryType()) },
		pruefe: func(ctx context.Context, grund string) (bool, error) {
			// **Der Kontext trägt die Frist.** Ein Dialog, den niemand
			// beantwortet, hinge sonst, bis die Anwendung endet — dieselbe
			// Regel wie bei einem Unterprozess ohne Frist (§16, Regel 3).
			return kontext.EvaluatePolicyLocalizedReasonReplySync(ctx, richtlinie, grund)
		},
		schliesse: kontext.Invalidate,
	}
}
