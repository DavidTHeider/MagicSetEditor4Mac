package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
)

// nurGrundlinie sind keine Wanderungen nach der Grundlinie — der Stand, den
// 0.7.0 und 0.8.0 ausgeliefert haben.
var nurGrundlinie = fstest.MapFS{"migrationen/LIESMICH.md": {Data: []byte("leer")}}

// danach gibt die Wanderungen dieser Fassung nach der Grundlinie.
func danach(t *testing.T) []string {
	t.Helper()

	namen, fehler := migrationsnamen(migrationen)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return namen
}

// kuenftigeWanderung gibt eine Wanderung mit der **nächsten freien Nummer**
// und ihren Namen.
//
// **Nicht fest `0098`**: so hießen die Proben, bis es die erste echte gab
// (M01, #2345), und sechs Läufe hielten danach den heutigen Bestand für einen
// neueren.
func kuenftigeWanderung(t *testing.T, bezeichnung, inhalt string) (quelle fstest.MapFS, name string) {
	t.Helper()

	nummer := len(grundlinienumfangOder(t)) + len(danach(t)) + 1
	name = fmt.Sprintf("%04d_%s.sql", nummer, bezeichnung)

	return fstest.MapFS{"migrationen/" + name: {Data: []byte(inhalt)}}, name
}

// grundlinienumfangOder gibt die Wanderungen der Grundlinie.
func grundlinienumfangOder(t *testing.T) []string {
	t.Helper()

	namen, _, fehler := grundlinienumfang(grundlinie)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return namen
}

// fstestOhneOrdner ist eine Quelle ohne `migrationen/`.
func fstestOhneOrdner() fstest.MapFS {
	return fstest.MapFS{"anderes/LIESMICH.md": {Data: []byte("leer")}}
}

// verweigerndesFS meldet für jede Datei einen Fehler, der nicht „fehlt" ist.
type verweigerndesFS struct{}

func (verweigerndesFS) Open(string) (fs.File, error) { return nil, fs.ErrPermission }

// **Nur was fehlt, fällt auf heute zurück** — Zweitblick über Phase 2,
// Befund 12: eine vorhandene künftige Datei, die sich nicht lesen ließ, wurde
// still durch die heutige ersetzt, und die Probe führte etwas anderes aus als
// vorgesehen.
func TestDieUeberlagerungVerschlucktKeinenLesefehler(t *testing.T) {
	t.Parallel()

	_, fehler := ueberlagert{heute: migrationen, weitere: verweigerndesFS{}}.Open("migrationen/0098_sicherungsvermerke.sql")
	if !errors.Is(fehler, fs.ErrPermission) {
		t.Errorf("gab %v, erwartet den Lesefehler der weiteren Quelle", fehler)
	}

	// **Und was fehlt, kommt von heute.**
	if datei, fehler := (ueberlagert{heute: migrationen, weitere: fstest.MapFS{}}).Open("migrationen/0098_sicherungsvermerke.sql"); fehler != nil {
		t.Errorf("eine fehlende weitere Datei: %v", fehler)
	} else {
		_ = datei.Close() //nolint:errcheck // nur geöffnet
	}
}
