package storage

// **Die Profilklemme im Stolpertreiber** —
// #1815.
//
// Seit #1815 fragt jede Ablage, die eine Kennung annimmt, zuerst, ob sie dem
// gezeigten Profil gehört. Der Stolpertreiber kannte diese Frage nicht und gab
// ihr die Zeile, die für die **eigentliche** Abfrage gedacht war: die Läufe
// scheiterten am Scan der Klemme statt an der Stelle, um die es ihnen geht.
//
// **Deshalb steht sie ganz vorn** und antwortet jedem Lauf mit „gehört dir" —
// außer an ihrer eigenen Stelle. So nimmt sie keinem Lauf seine Antwort ab und
// bekommt doch ihren eigenen Fehlerweg.

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// beimProfilklemme stolpert an der Frage, ob eine Kennung dem Profil gehört.
const beimProfilklemme stolperstelle = iota + 2800

func init() {
	sql.Register("stolper-profilklemme", stolpertreiber{stelle: beimProfilklemme})
}

// besitzabfrage beantwortet die Frage der Profilklemme — oder stolpert.
//
// **Erkannt an Anfang und Ende**: jede Frage aus `besitz.go` beginnt mit
// `SELECT 1 FROM` und endet mit der Bedingung auf das Profil. Eine andere
// Abfrage dieser Gestalt gibt es in der Ablage nicht.
func besitzabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)
	if !strings.HasPrefix(geglaettet, "SELECT 1 FROM") ||
		!strings.HasSuffix(geglaettet, "profil_id = ?") {
		return nil, false, nil
	}

	if stelle == beimProfilklemme {
		return nil, true, errGestolpert
	}

	return &kennungszeilen{}, true, nil
}

// **Scheitert die Frage selbst, sagt der Fehler, woran** — und er ist kein
// „nicht gefunden": ein Lesefehler, der sich als fehlende Kennung ausgäbe,
// verbärge einen kaputten Bestand hinter einer harmlosen Auskunft.
func TestDieProfilklemmeMeldetIhrenWeg(t *testing.T) {
	t.Parallel()

	db, fehler := sql.Open("stolper-profilklemme", "")
	if fehler != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	ctx := MitProfil(t.Context(), 1)

	for name, ruf := range map[string]func() error{
		"über eine Zahl":   func() error { return setImProfil(ctx, db, 1) },
		"über eine Partie": func() error { return partieImProfil(ctx, db, "partie") },
	} {
		fehler := ruf()
		if !errors.Is(fehler, errGestolpert) || errors.Is(fehler, domain.ErrNotFound) ||
			!strings.Contains(fehler.Error(), "nachschlagen") {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}

// **Ohne Profil fragt die Klemme gar nicht erst** — ein Aufruf, der kein
// Profil trägt, ist ein Fehler des Aufrufers und kein „nicht gefunden".
func TestDieProfilklemmeVerlangtEinProfil(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)

	for name, ruf := range map[string]func() error{
		"über eine Zahl":   func() error { return setImProfil(t.Context(), db, 1) },
		"über eine Partie": func() error { return partieImProfil(t.Context(), db, "partie") },
	} {
		if fehler := ruf(); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}
