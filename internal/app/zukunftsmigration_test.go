// **Eine Migration, die es hier nicht gibt** — die Vorrichtung für #940.
package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
	// Der Treiber ohne cgo — derselbe wie in `internal/storage`.
)

// zukunftEintragen verbucht in einer Datei eine Migration aus einer neueren
// Fassung.
//
// **Eingetragen und nicht nachgestellt.** Eine Attrappe, die „aus der Zukunft"
// behauptet, prüfte den Aufrufer; hier steht die Zeile, die eine echte neuere
// Fassung hinterließe.
func zukunftEintragen(t *testing.T, pfad string) {
	t.Helper()

	db, err := storage.FuerLaeufe("file:" + pfad)
	if err != nil {
		t.Fatalf("die Sicherung öffnen: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO migrationen (name, angewendet, pruefsumme)
		 VALUES ('9999_aus_der_zukunft.sql', '2026-08-23T00:00:00.000Z', 'x')`); err != nil {
		t.Fatalf("die künftige Migration eintragen: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}
}

// zukunftImArchiv trägt die künftige Migration in den Bestand **innerhalb**
// einer Sicherung ein.
//
// **Warum der Umweg**: eine Sicherung ist seit 0.3.0 ein Archiv
// (#1162).
// Sie als Datenbank zu öffnen gäbe „file is not a database" — und ein Test,
// der stattdessen eine nackte `.db` baute, prüfte den Weg von gestern.
func zukunftImArchiv(t *testing.T, pfad string) {
	t.Helper()

	imArchiv(t, pfad, zukunftEintragen)
}

// imArchiv ändert den Bestand **innerhalb** einer Sicherung.
func imArchiv(t *testing.T, pfad string, aendern func(*testing.T, string)) {
	t.Helper()

	offen := t.TempDir()

	beschreibung, fehler := sicherung.Entpacken(pfad, offen)
	if fehler != nil {
		t.Fatalf("die Sicherung auspacken: %v", fehler)
	}

	aendern(t, filepath.Join(offen, app.Bestandsname))

	teile := make([]sicherung.Teil, 0, len(beschreibung.Teile))
	for _, teil := range beschreibung.Teile {
		teile = append(teile, sicherung.Teil{
			Name:   teil.Name,
			Quelle: filepath.Join(offen, filepath.FromSlash(teil.Name)),
		})
	}

	// **Daneben und dann darüber.** `Packen` überschreibt nichts — das ist die
	// Zusage der Sicherung und keine Unbequemlichkeit, die man hier umgeht.
	neben := pfad + ".neu"

	if fehler := sicherung.Packen(neben, teile, beschreibung.Angaben); fehler != nil {
		t.Fatalf("die Sicherung wieder packen: %v", fehler)
	}

	if fehler := os.Rename(neben, pfad); fehler != nil {
		t.Fatalf("die neue Sicherung an ihren Platz: %v", fehler)
	}
}
