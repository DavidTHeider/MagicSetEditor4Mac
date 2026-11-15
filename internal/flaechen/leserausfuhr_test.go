// **Ein Leser exportiert an der Kommandozeile nicht** — Werkstück O1 von
// Phase 5.2 (#1425),
// nach [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Warum nur die Kommandozeile
//
// **Weil MCP und Dienst gar nicht als Leser hineinkommen** — `nurlesen_test.go`
// hält das als Zusage. Was dort dieselbe Frage stellt, ist ein Token, das nur
// lesen darf, und das prüft `token_test.go` am Dienst. Das Fenster fährt
// denselben Weg in `internal/app/leserausfuhr_test.go`.
package flaechen_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestEinLeserExportiertAnDerZeileNicht(t *testing.T) {
	t.Parallel()

	// **Und Scryfall zeigt ins Leere**: greift der Riegel nicht, endet der
	// Ruf an einer Adresse, die niemand bedient — nicht im Netz (§3).
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t), "MSE_SCRYFALL_ADRESSE=http://127.0.0.1:9")

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	fremder, fremdes := profilEinesFremden(t, angelegt)

	// **Der Eigner öffnet ein Set in seinem Profil und vergibt die Freigabe.**
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "leser")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)

	// **Und jetzt als Leser hinein.**
	mussKommandozeile(t, umgebung, "profil", fremdes)

	ziel := t.TempDir()

	faelle := map[string]struct {
		argumente []string
		datei     string
	}{
		"sichern": {[]string{"sichern", setID, filepath.Join(ziel, "kopie.mse-set")}, filepath.Join(ziel, "kopie.mse-set")},
		"bogen":   {[]string{"bogen", setID, filepath.Join(ziel, "bogen.pdf")}, filepath.Join(ziel, "bogen.pdf")},
		"mtgjson-schreiben": {
			[]string{"mtgjson-schreiben", setID, "magic", filepath.Join(ziel, "set.json")},
			filepath.Join(ziel, "set.json"),
		},

		// **Und nach draußen geht auch nichts** — `externSenden` verlangt
		// `mitarbeiter`. Preise holen fragt keine Zuordnung und kein Spiel;
		// der erste Schritt wäre die Anfrage.
		"preise-holen": {[]string{"preise-holen", "0000-probe"}, ""},
	}

	for name, fall := range faelle {
		aus, fehler, code := kommandozeile(t, umgebung, fall.argumente...)
		if code == 0 {
			t.Errorf("%s: der Leser schrieb — %s", name, aus)

			continue
		}

		gesagt := aus + fehler
		if !strings.Contains(gesagt, "nur gelesen") {
			t.Errorf("%s: die Meldung sagt nicht, was los ist: %s", name, gesagt)
		}

		if fall.datei != "" {
			if _, err := os.Stat(fall.datei); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("%s: %s liegt trotzdem da (%v)", name, fall.datei, err)
			}
		}

		if strings.Contains(gesagt, "127.0.0.1") {
			t.Errorf("%s: der Ruf ging bis ans Netz — %s", name, gesagt)
		}
	}

	// **Und der Weg hinaus bleibt offen.**
	mussKommandozeile(t, umgebung, "profil", meines)
}
