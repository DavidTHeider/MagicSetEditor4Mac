package service

// **Ein Ziel, dessen Regeln anderswo stehen** — OPS03
// (#2349).
//
// Die Verknüpfungen von `mse` und `mse-mcp` in `~/.local/bin` richtet
// `internal/app` seit OPS-03 (#1720) ein, und **was eine eigene Verknüpfung
// ist**, steht dort (`befehlsweg.go`). Bis OPS03 standen sie neben den
// Zielen: „Alles widerrufen" ließ sie liegen, und wer das Programm danach in
// den Papierkorb legte, behielt zwei Verweise ins Leere (Befund A-10).
//
// Hier werden sie ein Ziel wie jedes andere — dieselben Stände, dieselben
// Regeln: **fremd bleibt fremd**, eingerichtet einzurichten ändert nichts,
// offen zu widerrufen auch nicht. Was dabei geschieht, reicht das Ziel herein.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ArtVerknuepfung ist ein Ziel aus Verknüpfungen in einem Ordner.
const ArtVerknuepfung Schnittstellenart = "verknuepfung"

// Verknuepfungsziel trägt Stand und Schritte eines Ziels der
// [ArtVerknuepfung].
type Verknuepfungsziel struct {
	Stand      func() Schnittstellenstand
	Einrichten func() error
	Widerrufen func() error

	// Zeilen sind die Verknüpfungen, die einrichten anlegt — für die
	// Vorschau (AK-01 von OPS-09: jede Zeile vorher).
	Zeilen []string
}

// verknuepfungEinrichten richtet ein Verknüpfungsziel ein.
func verknuepfungEinrichten(ziel Schnittstellenziel) (Schnittstellenausgang, error) {
	switch stand := ziel.Verknuepfung.Stand(); stand {
	case SchnittstelleEingerichtet:
		return Schnittstellenausgang{Stand: stand}, nil
	case SchnittstelleFremd:
		return Schnittstellenausgang{Stand: stand}, fmt.Errorf(
			"%w: dort steht schon ein fremder Eintrag — er bleibt, wie er ist", domain.ErrConflict)
	}

	fehler := ziel.Verknuepfung.Einrichten()

	return Schnittstellenausgang{Stand: ziel.Verknuepfung.Stand()}, fehler
}

// verknuepfungWiderrufen nimmt die eigenen Verknüpfungen.
func verknuepfungWiderrufen(ziel Schnittstellenziel) (Schnittstellenausgang, error) {
	switch stand := ziel.Verknuepfung.Stand(); stand {
	case SchnittstelleFremd:
		return Schnittstellenausgang{Stand: stand}, fmt.Errorf(
			"%w: dort steht kein eigener Eintrag, nur ein fremder — er bleibt", domain.ErrConflict)
	case SchnittstelleEingerichtet, SchnittstelleVeraltet:
	default:
		return Schnittstellenausgang{Stand: stand, Bytegleich: true}, nil
	}

	fehler := ziel.Verknuepfung.Widerrufen()

	return Schnittstellenausgang{Stand: ziel.Verknuepfung.Stand(), Bytegleich: fehler == nil}, fehler
}
