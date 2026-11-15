package flaechen_test

// **Der Code eines Fehlers an jeder Fläche** — F03
// (#2342).
//
// `abweisungGleich` fährt jede Abweisung einmal so, wie ein Mensch sie
// liest, und einmal maschinenlesbar: mit `--json` an der Kommandozeile und
// mit `structuredContent` an MCP. **Die Marke gilt je Test**, nicht je
// Lauf: die Aufrufe eines Falls sind Funktionen, die ihren Test kennen, und
// über die Marke wissen die Helfer, wie sie diesmal rufen.

import (
	"sync"
	"testing"
)

// codeVerlangt hält die Tests, deren Aufrufe gerade maschinenlesbar laufen.
var codeVerlangt sync.Map

// mitCode ruft einmal maschinenlesbar und gibt, was die Fläche dabei sagt.
func mitCode(t *testing.T, ruf func() (string, int)) string {
	t.Helper()

	codeVerlangt.Store(t, true)
	defer codeVerlangt.Delete(t)

	aus, _ := ruf()

	return aus
}

// maschinenlesbar sagt, ob die Aufrufe dieses Tests gerade so laufen sollen.
func maschinenlesbar(t *testing.T) bool {
	_, verlangt := codeVerlangt.Load(t)

	return verlangt
}
