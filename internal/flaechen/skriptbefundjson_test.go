// **`--json` ändert die Ausgabe und nicht den Ausgang** — #2400, Vermerk 0033.
//
// `mse paketdatei-pruefen` endete bei einem Befund mit 1, mit `--json` mit
// 0: ein Ablauf, der maschinenlesbar prüft, sah einen kaputten Stil als
// Erfolg. Gefunden im Nachlauf über Phase 1 von 0.9.0 (#2322).
package flaechen_test

import (
	"encoding/json"
	"testing"
)

func TestMitJSONEndetDiePaketpruefungWieOhne(t *testing.T) {
	umgebung := eigenesZuhause(t)
	paket := probepaket(t, umgebung)

	for name, fall := range map[string]struct {
		inhalt string
		code   int
		zeile  int
	}{
		"ein Befund":       {"mse version: 0.3.8\nscript: f(1, 2\n", 1, 2},
		"eine heile Datei": {"mse version: 0.3.8\nscript: f(1, 2)\n", 0, 0},
	} {
		schreibePaketdatei(t, umgebung, paket, fall.inhalt)

		_, _, ohne := kommandozeile(t, umgebung, "paketdatei-pruefen", paket, "style")
		aus, fehler, mit := kommandozeile(t, umgebung, "--json", "paketdatei-pruefen", paket, "style")

		var bericht map[string]any
		if err := json.Unmarshal([]byte(aus), &bericht); err != nil {
			t.Fatalf("%s: kein JSON: %v — %s%s", name, err, aus, fehler)
		}

		if ohne != fall.code || mit != fall.code {
			t.Errorf("%s: ohne --json %d, mit %d, erwartet %d — %s", name, ohne, mit, fall.code, aus)
		}

		zeile := 0.0
		if wert, steht := bericht["zeile"].(float64); steht {
			zeile = wert
		}

		if int(zeile) != fall.zeile {
			t.Errorf("%s: der Bericht nennt Zeile %v: %s", name, bericht["zeile"], aus)
		}
	}
}
