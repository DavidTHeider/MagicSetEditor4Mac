// Der Probelauf des Dienstes und sein Blick auf das Ziel — UXR-F7
// (#2015).
//
// **Er legt nichts an**, sieht aber nach, ob die Wache den Ort bekäme: liegt
// dort schon etwas, antwortet er wie der schreibende Weg.
package main

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

	faelle := map[string]struct {
		ziel string
		art  error
	}{
		"eine Datei":                {datei, domain.ErrConflict},
		"ein Verzeichnis":           {verzeichnis, domain.ErrConflict},
		"ein Pfad durch eine Datei": {filepath.Join(datei, "tiefer.pdf"), domain.ErrValidation},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := ortFrei(fall.ziel)
			_, wache := reserviereOrt(fall.ziel)

			if !errors.Is(probe, fall.art) || wache == nil || !errors.Is(wache, fall.art) {
				t.Fatalf("Probelauf %v, Wache %v", probe, wache)
			}

			// **Im Wortlaut gleich, soweit er nicht den Aufruf nennt**: bei einem
			// unanlegbaren Ort steht im Satz das Betriebssystem (`lstat` oder
			// `open`), und das darf sich unterscheiden.
			if errors.Is(fall.art, domain.ErrConflict) && probe.Error() != wache.Error() {
				t.Errorf("Probelauf %q, Wache %q", probe, wache)
			}
		})
	}
}

// **Der Bogen: dieselben Zahlen wie das Legen, ein belegtes Ziel wird vorher
// abgewiesen, und ein Set, das es nicht gibt, ist ein Nicht-Gefunden.**
func TestDerProbelaufDesDienstesRechnetDenBogen(t *testing.T) {
	db := frischerTokenbestand(t)

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
// Ziel ein Konflikt, ein geschlossener Bestand ein Fehler.**
func TestDerProbelaufDerVorlageImDienst(t *testing.T) {
	db := frischerTokenbestand(t)

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
