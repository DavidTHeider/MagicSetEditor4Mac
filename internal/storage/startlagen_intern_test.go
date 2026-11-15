package storage

// **Was beim Öffnen scheitert, sagt seine Lage** — F05
// (#2344),
// Befunde A-05 und A-21.
//
// Belegt, voll und nur lesend kannte `einordnen` schon. Eine beschädigte
// Datei, ein Fehler des Datenträgers und ein verweigerter Zugriff kamen roh
// heraus — und der Start machte aus allem Rohen „unbrauchbar, spiele eine
// Sicherung zurück". Für einen gesunden Bestand ist das der Rat, der Daten
// kostet.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	sqlite "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinordnenTrenntDieLagenDesStarts(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name string
		code error
		art  error
	}{
		{"eine beschädigte Datei", sqlite.CORRUPT, domain.ErrBeschaedigt},
		{"eine Datei, die kein Bestand ist", sqlite.NOTADB, domain.ErrBeschaedigt},
		{"ein Lesefehler des Datenträgers", sqlite.IOERR_READ, domain.ErrEA},
		{"ein Fehler beim Festschreiben", sqlite.IOERR_FSYNC, domain.ErrEA},
		{"eine Datei, die sich nicht öffnen lässt", sqlite.CANTOPEN, domain.ErrGesperrt},
		{"ein verweigerter Zugriff", sqlite.PERM, domain.ErrGesperrt},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			eingeordnet := einordnen(fmt.Errorf("öffnen: %w", fall.code))

			if !errors.Is(eingeordnet, fall.art) {
				t.Errorf("%v ist nicht %v", eingeordnet, fall.art)
			}

			// **Der Grund reist mit** (§8): wer im Protokoll nachsieht, findet
			// den Code von SQLite.
			if domain.GrundVon(eingeordnet) == nil {
				t.Errorf("%v trägt keinen Grund", eingeordnet)
			}
		})
	}
}

// angelegterBestand legt einen migrierten Bestand an und schließt ihn.
func angelegterBestand(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	return pfad
}

// oeffnenUndHeben öffnet und hebt — so wie der Start.
func oeffnenUndHeben(t *testing.T, pfad string) error {
	t.Helper()

	db, err := Oeffnen(probekontext(t), pfad)
	if err != nil {
		return err
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gelesen

	return Migriere(probekontext(t), db, nil)
}

func TestEinBestandMitZerschriebenerSeiteIstBeschaedigt(t *testing.T) {
	t.Parallel()

	pfad := angelegterBestand(t)

	// **Die Schemaseite zerschrieben, der Kopf bleibt**: SQLite erkennt die
	// Datei noch als Datenbank und stolpert erst beim Lesen.
	//
	// **Und nicht eine Datei aus fremdem Text**: deren Kopf ist nicht
	// `SQLite format 3`, und dann gilt sie als verschlüsselt — eine Chiffre
	// lässt sich nicht wiedererkennen (`bestandsschluessel.go`). Ohne
	// Schlüssel heißt das „der Schlüssel fehlt", und das ist Absicht.
	datei, err := os.OpenFile(pfad, os.O_WRONLY, 0) // #nosec G304 -- der Pfad kommt aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := datei.WriteAt(make([]byte, 3000), 100); err != nil {
		t.Fatal(err)
	}

	if err := datei.Close(); err != nil {
		t.Fatal(err)
	}

	if err := oeffnenUndHeben(t, pfad); !errors.Is(err, domain.ErrBeschaedigt) {
		t.Errorf("%v ist nicht beschädigt", err)
	}
}

func TestEinBestandOhneLeserechtIstGesperrt(t *testing.T) {
	t.Parallel()

	pfad := angelegterBestand(t)
	if err := os.Chmod(pfad, 0); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(pfad, 0o600) }) //nolint:errcheck // Aufräumen

	if err := oeffnenUndHeben(t, pfad); !errors.Is(err, domain.ErrGesperrt) {
		t.Errorf("%v ist nicht gesperrt", err)
	}
}

func TestEinBestandOhneSchreibrechtIstNurLesbar(t *testing.T) {
	t.Parallel()

	pfad := angelegterBestand(t)
	if err := os.Chmod(pfad, 0o400); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(pfad, 0o600) }) //nolint:errcheck // Aufräumen

	if err := oeffnenUndHeben(t, pfad); !errors.Is(err, domain.ErrNurLesend) {
		t.Errorf("%v ist nicht nur lesbar", err)
	}
}
