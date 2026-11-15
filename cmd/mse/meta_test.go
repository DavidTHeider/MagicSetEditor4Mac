package main

// **`mse meta`** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestMetaNenntFassungUndFlaeche(t *testing.T) {
	code, aus, fehler := fahre(t, "meta")

	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	// **Bau und Bindungsvertrag** — ENG-06.AK-04 (#1731), dieselben Angaben
	// wie im Diagnosetext und am MCP-Werkzeug.
	for _, teil := range []string{
		"fassung\t" + app.Fassung, "flaeche\tkommandozeile",
		"stand\t" + app.DiesenBau().Stand, "bindungsvertrag\t" + app.Vertragsabdruck(),
	} {
		if !strings.Contains(aus, teil) {
			t.Errorf("die Ausgabe nennt %q nicht: %q", teil, aus)
		}
	}
}

// **Und als JSON dasselbe.**
func TestMetaAlsJSON(t *testing.T) {
	code, aus, fehler := fahre(t, "meta", "--json")

	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, `"flaeche": "kommandozeile"`) {
		t.Errorf("JSON %q", aus)
	}
}

// **Ein Argument ist ein Bedienfehler** — der Befehl nimmt keines.
func TestMetaNimmtKeineArgumente(t *testing.T) {
	if code, _, _ := fahre(t, "meta", "zuviel"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}
