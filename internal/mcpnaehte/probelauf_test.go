// Der Probelauf des MCP-Servers und sein Blick auf das Ziel — UXR-F7
// (#2015).
//
// **Er legt nichts an**, sieht aber nach, ob die Wache den Ort bekäme: liegt
// dort schon etwas, antwortet er wie der schreibende Aufruf.

package mcpnaehte

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Ziel gibt es nichts anzusehen, und ein freier Ort bleibt frei.**
func TestOrtFreiLaesstEinenFreienOrtFrei(t *testing.T) {
	t.Parallel()

	if err := ortFrei(""); err != nil {
		t.Errorf("ohne Ziel: %v", err)
	}

	ziel := filepath.Join(t.TempDir(), "frei.pdf")
	if err := ortFrei(ziel); err != nil {
		t.Errorf("ein freier Ort: %v", err)
	}

	// **Und danach ist er immer noch frei.** Der Probelauf legt nichts an —
	// die Wache dagegen reserviert.
	if _, err := os.Lstat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("der Probelauf hat etwas angelegt: %v", err)
	}
}

// **Dasselbe Nein wie die Wache — im Wortlaut.**
func TestOrtFreiWeistWieDieWacheAb(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()

	datei := filepath.Join(ordner, "da.pdf")
	if err := os.WriteFile(datei, []byte("x"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	verzeichnis := filepath.Join(ordner, "verzeichnis")
	if err := os.Mkdir(verzeichnis, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// Ein Pfad durch eine Datei: weder frei noch belegt, sondern unanlegbar.
	durchDatei := filepath.Join(datei, "tiefer.pdf")

	faelle := map[string]struct {
		ziel string
		art  error
	}{
		"eine Datei":                {datei, domain.ErrConflict},
		"ein Verzeichnis":           {verzeichnis, domain.ErrConflict},
		"ein Pfad durch eine Datei": {durchDatei, domain.ErrValidation},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := ortFrei(fall.ziel)
			_, wache := reserviereOrt(fall.ziel)

			if !errors.Is(probe, fall.art) || wache == nil || probe.Error() != wache.Error() {
				t.Errorf("Probelauf %v, Wache %v", probe, wache)
			}

			// **Der Grund fürs Protokoll nennt bei einem belegten Ort denselben
			// Pfad.** Bei einem unanlegbaren sagt er, welcher Aufruf scheiterte
			// (`lstat` oder `open`) — und das darf sich unterscheiden.
			if errors.Is(fall.art, domain.ErrConflict) && grundText(probe) != grundText(wache) {
				t.Errorf("Grund: Probelauf %q, Wache %q", grundText(probe), grundText(wache))
			}
		})
	}

	if roh, err := os.ReadFile(datei); err != nil || string(roh) != "x" { // #nosec G304 -- t.TempDir()
		t.Errorf("der vorhandene Inhalt ist hin: %q, %v", roh, err)
	}
}

// **Der Bogen: dieselben Zahlen wie das Legen, und ein belegtes Ziel wird
// abgewiesen, bevor gerechnet wird.**
func TestDerProbelaufDesMCPServersRechnetDenBogen(t *testing.T) {
	db := frischerBestand(t)

	ergebnis, fehler := app.Einfuhrnaht(db).Set(probekontext(t),
		filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if fehler != nil {
		t.Fatalf("einfahren: %v", fehler)
	}

	probe, err := probelauf{db: db}.BogenProbelauf(probekontext(t), ergebnis.SetID, "")
	if err != nil {
		t.Fatalf("Probelauf: %v", err)
	}

	if probe.Karten == 0 || probe.Seiten == 0 || probe.JeSeite != 9 || probe.DPI != 300 {
		t.Errorf("%+v", probe)
	}

	belegt := filepath.Join(t.TempDir(), "da.pdf")
	if schreibfehler := os.WriteFile(belegt, []byte("x"), 0o600); schreibfehler != nil {
		t.Fatalf("schreiben: %v", schreibfehler)
	}

	if _, err := (probelauf{db: db}).BogenProbelauf(probekontext(t), ergebnis.SetID, belegt); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("ein belegtes Ziel: %v", err)
	}

	if _, err := (probelauf{db: db}).BogenProbelauf(probekontext(t), 4711, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein Set, das es nicht gibt: %v", err)
	}
}

// **Die Vorlage: ein unbekanntes Paket ist ein Nicht-Gefunden, ein belegtes
// Ziel ein Konflikt.**
func TestDerProbelaufDerVorlageImMCPServer(t *testing.T) {
	db := frischerBestand(t)

	ergebnis, fehler := app.Einfuhrnaht(db).Set(probekontext(t),
		filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if fehler != nil {
		t.Fatalf("einfahren: %v", fehler)
	}

	_, err := probelauf{db: db}.VorlageProbelauf(probekontext(t), ergebnis.SetID,
		"gibtsnicht.mse-export-template", "", nil)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("eine unbekannte Vorlage: %v", err)
	}

	belegt := filepath.Join(t.TempDir(), "da.txt")
	if schreibfehler := os.WriteFile(belegt, []byte("x"), 0o600); schreibfehler != nil {
		t.Fatalf("schreiben: %v", schreibfehler)
	}

	if _, err := (probelauf{db: db}).VorlageProbelauf(probekontext(t), ergebnis.SetID,
		"egal", belegt, nil); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("ein belegtes Ziel: %v", err)
	}

	// **Ein geschlossener Bestand ist ein Fehler** und kein Absturz.
	if schliessen := db.Close(); schliessen != nil {
		t.Fatalf("schließen: %v", schliessen)
	}

	zu := probelauf{db: db}

	if _, err := zu.VorlageProbelauf(probekontext(t), 1, "p", "", nil); err == nil {
		t.Error("ein geschlossener Bestand ging durch")
	}

	if _, err := zu.BogenProbelauf(probekontext(t), 1, ""); err == nil {
		t.Error("ein geschlossener Bestand ging durch")
	}
}
