//go:build !darwin

package biometrie

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne macOS gibt es keine Biometrie** — und das ist eine Antwort, kein
// Fehlen. `docs/PORTIERUNG.md` beschreibt Windows und Linux; gebaut wird für
// beide, damit plattformabhängiger Code nicht unbemerkt bricht (§12).
//
// **Windows hätte Hello, Linux hätte `polkit`** — beides ist hier weder
// gebaut noch geprüft, und deshalb steht hier kein halber Weg, sondern ein
// klares Nein. **TOTP geht auf beiden.**

// stand sagt, dass es hier nichts gibt.
func stand(_ context.Context) Auskunft {
	return Auskunft{Grund: "Biometrie gibt es in dieser Anwendung nur auf macOS."}
}

// bestaetige lehnt ab, statt so zu tun als ob.
func bestaetige(_ context.Context, _ string) error {
	return fmt.Errorf(
		"%w: Biometrie gibt es in dieser Anwendung nur auf macOS",
		domain.ErrValidation)
}
