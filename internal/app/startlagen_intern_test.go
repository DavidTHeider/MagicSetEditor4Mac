package app

// **Jede Lage des Starts hat ihre Klasse und ihren Rat** — F05
// (#2344),
// Befunde A-05 und A-21.
//
// Bis hierher galt alles außer Schlüssel und Fassung als „unbrauchbar —
// spiele eine Sicherung zurück". Ein belegter, voller, gesperrter oder nur
// lesbarer Bestand ist aber **in Ordnung**, und wer ihn auf diesen Rat hin
// ersetzt, verliert, was seit der Sicherung dazukam.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeStartlageHatIhreKlasse(t *testing.T) {
	t.Parallel()

	grund := errors.New("roh")
	pfadfehler := func(art syscall.Errno) error {
		return fmt.Errorf("Datenverzeichnis anlegen: %w", &fs.PathError{Op: "mkdir", Path: "/probe", Err: art})
	}

	for _, fall := range []struct {
		name   string
		fehler error
		klasse domain.Schluessel
	}{
		{"belegt", domain.MitGrund(domain.ErrBelegt, grund, "gleich noch einmal"), domain.SchluesselBelegt},
		{"voll", domain.MitGrund(domain.ErrVoll, grund, "Platz schaffen"), domain.SchluesselVoll},
		{"voll beim Anlegen des Ordners", pfadfehler(syscall.ENOSPC), domain.SchluesselVoll},
		{"nur lesbar", domain.MitGrund(domain.ErrNurLesend, grund, "Schreibrecht"), domain.SchluesselBestandNurLesbar},
		{"gesperrt", domain.MitGrund(domain.ErrGesperrt, grund, "freigeben"), domain.SchluesselGesperrt},
		{"gesperrt beim Anlegen des Ordners", pfadfehler(syscall.EACCES), domain.SchluesselGesperrt},
		{"Ein- und Ausgabe", domain.MitGrund(domain.ErrEA, grund, "Datenträger prüfen"), domain.SchluesselBestandEA},
		{"Ein- und Ausgabe bei knappem Platz", fmt.Errorf("%w: %w", domain.ErrKaumPlatz, domain.MitGrund(domain.ErrEA, grund, "Datenträger prüfen")), domain.SchluesselKaumPlatz},
		{"beschädigt", domain.MitGrund(domain.ErrBeschaedigt, grund, "zurückspielen"), domain.SchluesselBestandUnbrauchbar},
		{"ungeklärt", errors.New("etwas ganz anderes"), domain.SchluesselStartUngeklaert},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if klasse := startklasse(fmt.Errorf("Datenbank öffnen: %w", fall.fehler)); klasse != fall.klasse {
				t.Errorf("Klasse %q, erwartet %q", klasse, fall.klasse)
			}
		})
	}
}

// **Die rohe Eingabe durch die Startnaht** (§3) — so, wie der Speicher sie
// gibt, und bis in den Befund, den das Fenster liest.
func TestEinBelegterBestandRaetNichtZumZurueckspielen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.oeffnen = func(context.Context, string) (*sql.DB, error) {
		return nil, fmt.Errorf("Datenbank öffnen: %w", domain.MitGrund(domain.ErrBelegt,
			errors.New("database is locked (5) (SQLITE_BUSY)"), "gleich noch einmal"))
	}

	if err := anwendung.Starten(t.Context()); err == nil {
		t.Fatal("der belegte Start kam durch")
	}

	if befund := anwendung.Startbefund(); befund.Schluessel != string(domain.SchluesselBelegt) {
		t.Errorf("der Befund rät %q, erwartet %q", befund.Schluessel, domain.SchluesselBelegt)
	}
}
