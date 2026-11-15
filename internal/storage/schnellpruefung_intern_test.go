package storage

// **Ein Fehler der Prüfung ist kein Bestanden** — F05
// (#2344).

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineGescheiterteSchnellpruefungIstKeinBestanden(t *testing.T) {
	t.Parallel()

	db, err := oeffner(ImArbeitsspeicher, "")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if err := schnellPruefen(probekontext(t), db); err == nil {
		t.Error("eine Prüfung, die nicht lief, gilt als bestanden")
	}
}

func TestEineSicherungDieSichNichtOeffnenLaesstBestehtNicht(t *testing.T) {
	t.Parallel()

	if err := geschriebeneSicherungPruefen(probekontext(t), filepath.Join(t.TempDir(), "fehlt.db")); err == nil {
		t.Error("eine fehlende Sicherung bestand ihre Prüfung")
	}
}

func TestEineSicherungOhneTreiberBestehtNicht(t *testing.T) {
	vorher := oeffner
	oeffner = func(string, string) (*sql.DB, error) { return nil, errors.New("kein Treiber") }
	t.Cleanup(func() { oeffner = vorher })

	if err := geschriebeneSicherungPruefen(probekontext(t), filepath.Join(t.TempDir(), "x.db")); err == nil {
		t.Error("eine Sicherung, die sich nicht öffnen ließ, bestand ihre Prüfung")
	}
}

// **Nur lesbar heißt: lesen geht.** Lässt sich die Datei gar nicht öffnen,
// bleibt der Zugriff verweigert.
func TestNurLesbarVerlangtEineLesbareDatei(t *testing.T) {
	t.Parallel()

	gesperrt := domain.MitGrund(domain.ErrGesperrt, errors.New("roh"), "freigeben")

	if err := nurLesbar(filepath.Join(t.TempDir(), "fehlt.db"), gesperrt); errors.Is(err, domain.ErrNurLesend) {
		t.Errorf("eine fehlende Datei gilt als nur lesbar: %v", err)
	}
}
