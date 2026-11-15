//go:build darwin

// **Über einem verschlüsselten Bestand richtet sich kein Plan ein** — OPS01
// (#2347),
// Zweitblick über Phase 4, Befund 4. **Darwin**, weil die Verschlüsselung den
// Schlüsselbund braucht ([gestarteteAnwendung]).
package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestUeberEinemVerschluesseltenBestandRichtetSichKeinPlanEin(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	gefragt := false
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		gefragt = true

		return t.TempDir(), nil
	})

	if an := anwendung.VerschluesselungUmschalten(t.Context(), true); an.Fehler != nil {
		t.Fatalf("einschalten: %+v", an.Fehler)
	}

	antwort := anwendung.SicherungsplanEinrichten(t.Context())
	if antwort.Fehler == nil || antwort.Fehler.Schluessel != domain.SchluesselPlanVerschluesselt {
		t.Fatalf("eingerichtet: %+v", antwort)
	}

	// **Gar nicht erst gefragt** — wer einen Ordner wählt, um danach zu hören,
	// dass es nicht geht, hat umsonst gewählt.
	if gefragt {
		t.Error("der Dialog kam trotzdem")
	}

	// **Und über die Einstellung ebenso** (Nachlauf über Phase 4, Befund 4):
	// `sicherung.planordner` setzen Fensterbindung, Kommandozeile, MCP und
	// REST über denselben Verwalter.
	if gesetzt := anwendung.EinstellungSetzen(t.Context(), domain.SicherungsplanOrdner, t.TempDir()); gesetzt.Fehler == nil ||
		gesetzt.Fehler.Schluessel != domain.SchluesselPlanVerschluesselt {
		t.Errorf("über die Einstellung: %+v", gesetzt.Fehler)
	}

	if stand := anwendung.Sicherungsstand(t.Context()); stand.Stand.Plan.Ordner != "" {
		t.Errorf("der Plan steht trotzdem: %+v", stand.Stand.Plan)
	}
}
