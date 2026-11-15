package app_test

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Drei Konflikte, drei Sätze — und ein voller Datenträger** — ENG-05.AK-03
// (#1730),
// nativer Befund F08, und
// #1812.
//
// Bis zum 2026-09-23 bekam **jeder** Konflikt den Schlüssel `konflikt`, und
// das Fenster sagte „Der Eintrag wurde zwischenzeitlich geändert. Bitte neu
// laden." — auch über einem zweiten Deck gleichen Namens und über dem letzten
// Nutzer, der bleiben muss. Ein voller Datenträger kam als „unerwarteter
// Fehler" an.
func TestDieKonfliktartenBekommenIhrenSatz(t *testing.T) {
	treiber := errors.New("sqlite3: constraint failed: UNIQUE constraint failed: decks.profil_id, decks.spiel, decks.name")

	for _, fall := range []struct {
		name       string
		fehler     error
		code       app.Fehlercode
		schluessel domain.Schluessel
		feld       string
		grund      string
	}{
		{
			name: "ein überholter Stand heißt lade neu",
			fehler: fmt.Errorf("das Deck ändern: %w",
				fmt.Errorf("%w — das Deck 3 hat sich seit dem Lesen geändert", domain.ErrUeberholt)),
			code:       app.CodeKonflikt,
			schluessel: domain.SchluesselKonflikt,
		},
		{
			name: "ein doppelter Name steht am Namen",
			fehler: fmt.Errorf("das Deck anlegen: %w", domain.NewFieldError("name",
				domain.SchluesselDoppelterName, "Diesen Namen gibt es hier schon.").MitGrund(treiber)),
			code:       app.CodeKonflikt,
			schluessel: domain.SchluesselDoppelterName,
			feld:       "name",
		},
		{
			name: "ein doppeltes Exemplar sagt, welche Angaben gleich sind",
			fehler: domain.NewFieldError("", domain.SchluesselDoppeltesExemplar,
				"Diesen Druck führt die Sammlung schon.").MitGrund(treiber),
			code:       app.CodeKonflikt,
			schluessel: domain.SchluesselDoppeltesExemplar,
		},
		{
			name: "ein anderer Konflikt nennt seinen Grund",
			fehler: fmt.Errorf("den Nutzer löschen: %w", domain.MitGrund(domain.ErrConflict, nil,
				"der letzte Nutzer bleibt — ohne ihn hätte die Anwendung kein Profil")),
			code:       app.CodeKonflikt,
			schluessel: domain.SchluesselKonfliktgrund,
			grund:      "den Nutzer löschen: der letzte Nutzer bleibt — ohne ihn hätte die Anwendung kein Profil",
		},
		{
			name: "ein voller Bestand",
			fehler: fmt.Errorf("das Set anlegen: %w", domain.MitGrund(domain.ErrVoll,
				errors.New("sqlite3: database or disk is full"), "schaffe Platz")),
			code:       app.CodeVoll,
			schluessel: domain.SchluesselVoll,
		},
		{
			// **Ein gesperrter Ort nennt sich** — FIX-F02, S01.NAT-03: beim
			// Sichern in einen Ordner ohne Schreibrecht stand „ein
			// unerwarteter Fehler" da (Hüllenprobe, 2026-09-26).
			name: "ein gesperrter Ordner beim Schreiben einer Datei",
			fehler: fmt.Errorf("das Set sichern: %w",
				&fs.PathError{Op: "mkdir", Path: "/Users/jemand/gesperrt/set.mse-set", Err: syscall.EACCES}),
			code:       app.CodeValidierung,
			schluessel: domain.SchluesselGesperrt,
		},
		{
			// **Ein Fehler des Datenträgers bei knappem Platz nennt beides**
			// (#2407): ohne `ENOSPC` ist er kein Beleg für eine volle Platte.
			name: "ein Fehler des Datenträgers bei knappem Platz",
			fehler: fmt.Errorf("die Sicherung schreiben: %w", fmt.Errorf("%w: %w", domain.ErrKaumPlatz,
				domain.MitGrund(domain.ErrEA, errors.New("sqlite3: disk I/O error"), "prüfe den Datenträger"))),
			code:       app.CodeVoll,
			schluessel: domain.SchluesselKaumPlatz,
		},
		{
			name: "eine volle Platte beim Schreiben einer Datei",
			fehler: fmt.Errorf("das Set sichern: %w",
				&fs.PathError{Op: "write", Path: "/Users/jemand/geheim.mse-set", Err: syscall.ENOSPC}),
			code:       app.CodeVoll,
			schluessel: domain.SchluesselVoll,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			uebersetzt := app.Uebersetze(t.Context(), fall.fehler)

			if uebersetzt.Code != fall.code || uebersetzt.Schluessel != fall.schluessel {
				t.Fatalf("Code %q, Schlüssel %q — erwartet %q, %q (%s)",
					uebersetzt.Code, uebersetzt.Schluessel, fall.code, fall.schluessel, uebersetzt.Meldung)
			}

			if uebersetzt.Feld != fall.feld {
				t.Errorf("am Feld %q, erwartet %q", uebersetzt.Feld, fall.feld)
			}

			if fall.grund != "" && uebersetzt.Werte["grund"] != fall.grund {
				t.Errorf("Grund %q, erwartet %q", uebersetzt.Werte["grund"], fall.grund)
			}

			// **Kein Treibertext und kein Pfad erreicht eine Fläche** (§8).
			for _, verboten := range []string{"sqlite3", "constraint failed", "/Users/jemand"} {
				if strings.Contains(uebersetzt.Meldung, verboten) {
					t.Errorf("die Meldung trägt %q: %q", verboten, uebersetzt.Meldung)
				}
			}
		})
	}
}
