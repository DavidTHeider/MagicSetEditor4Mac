// **Ein Bestand von vor der Grundlinie** — J02
// (#2266,
// Vermerk 0189): er wird abgewiesen, und das Fenster sagt, was zu tun ist.
package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// vergangenheitEintragen nimmt einem Bestand den letzten Schritt der
// Grundlinie — so steht ein Bestand aus 0.6.0 da, dem `0097` fehlt.
func vergangenheitEintragen(t *testing.T, pfad string) {
	t.Helper()

	db, err := storage.FuerLaeufe("file:" + pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t),
		`DELETE FROM migrationen WHERE name = '0097_faktorhaushalt.sql'`); err != nil {
		t.Fatalf("den Schritt austragen: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}
}

func TestEinBestandVorDerGrundlinieTraegtSeinenSchluessel(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")

	vorbereitet := app.NeueApp(ort)
	if fehler := vorbereitet.Starten(t.Context()); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	if fehler := vorbereitet.Beenden(); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	vergangenheitEintragen(t, ort)

	anwendung := app.NeueApp(ort)
	if fehler := anwendung.Starten(t.Context()); fehler == nil {
		t.Fatal("ein Bestand von vor der Grundlinie wurde geöffnet")
	}

	if befund := anwendung.Startbefund(); befund.Schluessel != string(domain.SchluesselBestandAeltereFassung) {
		t.Errorf("Schlüssel %q, erwartet war %q", befund.Schluessel, domain.SchluesselBestandAeltereFassung)
	}
}

// **Eine Sicherung von vor der Grundlinie wird benannt abgewiesen** — bevor
// sie den Bestand ersetzt.
func TestSicherungZurueckspielenWeistEineAeltereFassungAb(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	alt := filepath.Join(ort, "alt.db")

	anwendung := laufendeAnwendung(t, bestand)

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return alt, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	imArchiv(t, alt, vergangenheitEintragen)

	antwort := anwendung.SicherungZurueckspielen(t.Context(), alt, "")
	if antwort.Fehler == nil {
		t.Fatal("die Sicherung von vor der Grundlinie wurde angenommen")
	}

	if antwort.Fehler.Schluessel != domain.SchluesselSicherungAeltereFassung ||
		antwort.Fehler.Werte["grundlinie"] != "0097_faktorhaushalt.sql" {
		t.Errorf("Schlüssel %q mit Werten %v", antwort.Fehler.Schluessel, antwort.Fehler.Werte)
	}

	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("der laufende Bestand hat sich verändert: %q", got)
	}

	if _, err := os.Stat(filepath.Join(ort, app.Rueckspielsicherung)); !os.IsNotExist(err) {
		t.Error("es wurde etwas beiseitegelegt, obwohl gar nicht getauscht wurde")
	}
}
