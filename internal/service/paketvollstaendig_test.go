package service_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Paket erscheint unter seinem Namen erst, wenn es vollständig ist.**
//
// # Woran das hängt
//
// `internal/schriften` merkt sich seinen Namensindex **an der Änderungszeit der
// Paketwurzel**: solange sie gleich bleibt, gilt der Index als gültig. Diese
// Zusage trägt nur, wenn die Wurzel sich **zuletzt** ändert — also dann, wenn
// das Paket ganz dasteht.
//
// Der Nachlauf über Phase G hat gefunden, dass die Übernahme genau das brach:
// `os.CopyFS` legte das Zielverzeichnis **zuerst** an — damit änderte sich die
// Wurzel — und schrieb die Dateien danach hinein. Eine Schriftsuche im Fenster
// dazwischen merkte sich einen **halben** Bestand und behielt ihn, bis sich die
// Wurzel wieder änderte. Seit #263
// ist „beziehen und wieder hinsehen" der vorgesehene Weg.
//
// # Warum die Änderungszeit und kein Nebenläufigkeitstest
//
// **Weil sie die Zusage ist**, an der der Index hängt — nicht ein Anzeichen für
// sie. Ein Test, der den Kopiervorgang zu treffen versucht, prüfte dieselbe
// Aussage über einen Zufall.
func TestEinPaketErscheintErstVollstaendig(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "fremd.mse-style",
		"mse version: 2.0.0\ngame: probe\nshort name: Fremd\nversion: 2026-02-02\n")

	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	if _, err := verwalter.Installieren(t.Context(), filepath.Join(quelle, "fremd.mse-style")); err != nil {
		t.Fatalf("installieren: %v", err)
	}

	wurzel, err := os.Stat(ziel)
	if err != nil {
		t.Fatalf("Wurzel lesen: %v", err)
	}

	juengste := juengsteAenderung(t, filepath.Join(ziel, "fremd.mse-style"))

	if wurzel.ModTime().Before(juengste) {
		t.Errorf("die Wurzel steht auf %v, im Paket steht %v — "+
			"ein Index, der sich an die Wurzel bindet, hielte den halben Stand fest",
			wurzel.ModTime(), juengste)
	}
}

// juengsteAenderung gibt die späteste Änderungszeit unterhalb eines Pfades.
func juengsteAenderung(t *testing.T, wurzel string) time.Time {
	t.Helper()

	var spaeteste time.Time

	if err := filepath.WalkDir(wurzel, func(_ string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		angaben, err := eintrag.Info()
		if err != nil {
			return err
		}

		if angaben.ModTime().After(spaeteste) {
			spaeteste = angaben.ModTime()
		}

		return nil
	}); err != nil {
		t.Fatalf("Paket durchgehen: %v", err)
	}

	return spaeteste
}
