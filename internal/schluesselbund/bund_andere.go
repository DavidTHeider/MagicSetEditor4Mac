//go:build !darwin

package schluesselbund

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Auf anderen Plattformen gibt es hier keinen Schlüsselbund** — beschrieben,
// nicht gebaut (§12).
//
// Windows hätte den Credential Manager, Linux den Secret Service über D-Bus.
// Beides steht in `docs/PORTIERUNG.md` als offener Punkt. Was hier steht, ist
// **kein stiller Ersatz**: ein Geheimnis in einer Datei neben dem Bestand wäre
// schlechter als keines, weil es aussähe wie eines.
//
// **Ein Fehler und keine leere Zeichenkette.** Wer hier „nichts" bekäme,
// bekäme dieselbe Antwort wie bei einem Dienst ohne Schlüssel — und das ist
// ein anderer Zustand (Vermerk 0063, Entscheidung 3).
var errKeinBund = fmt.Errorf(
	"%w: auf dieser Plattform gibt es keinen Schlüsselbund — siehe docs/PORTIERUNG.md",
	domain.ErrValidation)

func (b *Bund) schreibe(_ context.Context, _, _, _ string) error { return errKeinBund }

func (b *Bund) lies(_ context.Context, _, _ string) (string, error) {
	return "", errKeinBund
}

func (b *Bund) loesche(_ context.Context, _, _ string) error { return errKeinBund }

// OrtFehlt gibt es hier auch — mit derselben Antwort wie alles andere.
func (b *Bund) OrtFehlt() error { return errKeinBund }
