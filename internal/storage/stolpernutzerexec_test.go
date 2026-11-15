package storage

// **Was der Nutzerweg auf ein `Exec` antwortet** — Werkstück 1 von Phase 4.12
// (#1351),
// herausgelöst am 2026-09-05.
//
// **Eine eigene Datei**, weil `stolpernutzer_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **was der Nutzerweg auf
// eine Abfrage antwortet**, hier, **was er auf ein Schreiben antwortet**.

import (
	"database/sql/driver"
	"strings"
)

// nutzerExec beantwortet ein `Exec` auf einem Nutzer- oder Profilweg.
//
// **Es steht hier und nicht in `stolperwerk_test.go`** — derselbe Schnitt wie
// beim Token- und beim Vorgabenweg: dort steht der Treiber, hier, **was der
// Nutzerweg antwortet**. Und die Datei dort ging über die Zeilengrenze aus §2,
// als der Vorgabenweg dazukam (Werkstück 15, #1351).
func (s stolperstmt) nutzerExec() (ergebnis driver.Result, meiner bool, fehler error) {
	geglaettet := strings.TrimSpace(s.abfrage)

	// **Eine Kennung, die sich nach dem Einfügen nicht lesen lässt.** Ein
	// Nutzer entsteht samt seinem ersten Profil; ohne diese Zeilen scheiterte
	// das Einfügen schon an der Kennung, und das zweite bliebe unbetreten.
	if s.stelle == beimNutzerkennung &&
		strings.HasPrefix(geglaettet, "INSERT INTO nutzer") {
		return kennungsfehler{}, true, nil
	}

	if s.stelle == beimProfilkennung &&
		strings.HasPrefix(geglaettet, "INSERT INTO profile") {
		return kennungsfehler{}, true, nil
	}

	if !strings.HasPrefix(geglaettet, "INSERT INTO nutzer") &&
		!strings.HasPrefix(geglaettet, "INSERT INTO profile") {
		return nil, false, nil
	}

	switch s.stelle {
	case beimSchreiben:
		return nil, true, errGestolpert

	// **Der Zähler wird auch hier geführt.** Ein Nutzer entsteht mit zwei
	// Anweisungen; „das zweite Einfügen" meint die zweite davon, und ohne
	// diese Zeilen liefe die Stelle an diesem Weg vorbei.
	case beimZweitenSchreiben:
		zweiterSchreibzaehler++
		if zweiterSchreibzaehler >= 2 {
			return nil, true, errGestolpert
		}
	}

	return kennung(1), true, nil
}
