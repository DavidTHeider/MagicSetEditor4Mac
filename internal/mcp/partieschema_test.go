package mcp_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Das Schema nennt jedes Feld, das das Werkzeug liest** — angestrichen beim
// Werkstück zu #2041: `Sitzwunsch` las `commander`, das Schema nannte es
// nicht, und ein Modell erfuhr nie, dass es einen Commander setzen kann.
// Geprüft wird über die JSON-Namen der Struktur, damit ein neues Feld nicht
// still am Schema vorbeiläuft.
func TestDasSchemaVonPartieAnlegenNenntJedesFeldEinesSitzes(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Partien: &partienprobe{},
	}) {
		if werkzeug.Name != "schreiben.partie.anlegen" {
			continue
		}

		sitze := karteAus(t, karteAus(t, werkzeug.Schema["properties"])["sitze"])
		eigenschaften := karteAus(t, karteAus(t, sitze["items"])["properties"])

		typ := reflect.TypeFor[mcp.Sitzwunsch]()
		for stelle := range typ.NumField() {
			name, _, _ := strings.Cut(typ.Field(stelle).Tag.Get("json"), ",")
			if _, da := eigenschaften[name]; !da {
				t.Errorf("das Schema nennt %q nicht, das Werkzeug liest es aber", name)
			}
		}

		return
	}

	t.Fatal("das Werkzeug schreiben.partie.anlegen gibt es nicht")
}
