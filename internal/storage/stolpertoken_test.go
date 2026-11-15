package storage

// **Die Fehlerwege der Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Was er prüft,
// gibt kein echter Treiber her: eine Abfrage, die scheitert; eine Zeile, die
// sich nicht lesen lässt; ein Abbruch **nach** der Schleife; eine Kennung, die
// sich nach dem Einfügen nicht lesen lässt.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Tokenweg stolpern kann.
const (
	beimTokenscan stolperstelle = iota + 1400
	beimTokenende
	beimTokenkennung
	beimTokenzaehlen
	beimTokennachschlagen
)

func init() {
	sql.Register("stolper-token-scan", stolpertreiber{stelle: beimTokenscan})
	sql.Register("stolper-token-ende", stolpertreiber{stelle: beimTokenende})
	sql.Register("stolper-token-kennung", stolpertreiber{stelle: beimTokenkennung})
	sql.Register("stolper-token-zaehlen", stolpertreiber{stelle: beimTokenzaehlen})
	sql.Register("stolper-token-nachschlagen",
		stolpertreiber{stelle: beimTokennachschlagen})
}

// tokenabfrage beantwortet die Abfragen des Tokenwegs — oder stolpert.
func tokenabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	// **Das einzelne Token samt gehüteter Formen — vierzehn Spalten.**
	//
	// **Es steht vor der Liste**, und das ist keine Laune: sein Anfang
	// **enthält** den der Liste. Stünde es dahinter, träfe die Liste zuerst,
	// und dieser Zweig wäre nicht erreichbar — der erste Anlauf war grün und
	// prüfte nichts.
	// **Die Spaltenliste ist die aus `tokenspalten`** — seit O2 (#1426) mit
	// Principal, letzter Nutzung und Vorgängerfrist, seit #1819 mit der
	// Sitzbindung, dazu die zwei gehüteten. Ein Präfix, das die alte Liste
	// nennt, trifft nichts mehr, und der Lauf wäre grün, weil der Fehler
	// anderswo entstünde (§10). **Genau das geschah mit #1819** — gefunden
	// hat es die Abdeckung, nicht dieser Lauf.
	case strings.HasPrefix(geglaettet,
		"SELECT id, profil_id, bezeichnung, scope, ablauf, angelegt, fassung, "+
			"principal, letzte_nutzung, vorgaenger_bis, sitz_partie, sitz, gehuetet"):
		if stelle == beimTokennachschlagen {
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 14}, true, nil

	// **Die Liste eines Profils — sieben Spalten.**
	case strings.HasPrefix(geglaettet, "SELECT id, profil_id, bezeichnung, scope"):
		switch stelle {
		case beimTokenscan:
			return &krummeTokenzeile{}, true, nil
		case beimTokenende:
			return &abbrechendeTokenzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: tokenspaltenzahl}, true, nil

	// **Die Frage, ob es das Token gibt** — eine Zahl.
	case strings.HasPrefix(geglaettet, "SELECT count(*) FROM zugriffstoken"):
		if stelle == beimTokennachschlagen {
			return nil, true, errGestolpert
		}

		return &zahlenzeile{zahl: 1}, true, nil
	}

	return nil, false, nil
}

// tokenspaltenzahl ist die Zahl der Spalten einer Tokenzeile — **aus
// [tokenspalten] gezählt**: hier stand eine Zehn, als längst zwölf gelesen
// wurden, und die krumme Zeile scheiterte an der Spaltenzahl statt am Scan
// (Nachlauf des Zweitblicks über Phase B).
var tokenspaltenzahl = len(strings.Split(tokenspalten, ","))

// krummeTokenzeile scheitert am Scan einer Tokenzeile.
type krummeTokenzeile struct{ gegeben bool }

func (r *krummeTokenzeile) Columns() []string { return nutzerspaltennamen(tokenspaltenzahl) }

func (r *krummeTokenzeile) Close() error { return nil }

func (r *krummeTokenzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return errGestolpert
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// abbrechendeTokenzeile bricht **nach** der Schleife ab.
type abbrechendeTokenzeile struct{}

func (r *abbrechendeTokenzeile) Columns() []string { return nutzerspaltennamen(tokenspaltenzahl) }

func (r *abbrechendeTokenzeile) Close() error { return nil }

func (r *abbrechendeTokenzeile) Next([]driver.Value) error { return errGestolpert }

// **Jeder Tokenweg reicht seinen Fehler durch.**
func TestJederTokenwegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		lauf    func(*Zugriffstokenablage) error
	}{
		"das Anlegen scheitert": {"stolper-schreiben", func(a *Zugriffstokenablage) error {
			_, err := a.TokenAnlegen(probekontext(t), stolpertoken(), "gehütet")

			return err
		}},
		"die Kennung lässt sich nicht lesen": {
			"stolper-token-kennung", func(a *Zugriffstokenablage) error {
				_, err := a.TokenAnlegen(probekontext(t), stolpertoken(), "gehütet")

				return err
			},
		},
		"die Tokenabfrage scheitert": {"stolper-fragen", func(a *Zugriffstokenablage) error {
			_, err := a.Token(probekontext(t), 1)

			return err
		}},
		"eine Tokenzeile lässt sich nicht lesen": {
			"stolper-token-scan", func(a *Zugriffstokenablage) error {
				_, err := a.Token(probekontext(t), 1)

				return err
			},
		},
		"die Tokenzeilen brechen ab": {
			"stolper-token-ende", func(a *Zugriffstokenablage) error {
				_, err := a.Token(probekontext(t), 1)

				return err
			},
		},
		"das einzelne Token lässt sich nicht lesen": {
			"stolper-token-nachschlagen", func(a *Zugriffstokenablage) error {
				_, _, err := a.TokenMitKennung(probekontext(t), 1)

				return err
			},
		},
		"das Ändern scheitert": {"stolper-schreiben", func(a *Zugriffstokenablage) error {
			_, err := a.TokenAendern(probekontext(t), gestolpertesToken())

			return err
		}},
		"das Zählen beim Ändern scheitert": {
			"stolper-token-zaehlen", func(a *Zugriffstokenablage) error {
				_, err := a.TokenAendern(probekontext(t), gestolpertesToken())

				return err
			},
		},
		"das Nachschlagen beim Ändern scheitert": {
			"stolper-token-nachschlagen", func(a *Zugriffstokenablage) error {
				_, err := a.TokenAendern(probekontext(t), gestolpertesToken())

				return err
			},
		},
		"das Löschen scheitert": {"stolper-schreiben", func(a *Zugriffstokenablage) error {
			return a.TokenLoeschen(probekontext(t), 1)
		}},
		"das Zählen beim Löschen scheitert": {
			"stolper-token-zaehlen", func(a *Zugriffstokenablage) error {
				return a.TokenLoeschen(probekontext(t), 1)
			},
		},

		// **Rotation, Vermerk und Sammelwiderruf** — Werkstück O2 (#1426).
		"das Rotieren scheitert": {"stolper-schreiben", func(a *Zugriffstokenablage) error {
			_, err := a.TokenRotieren(probekontext(t), 1, "neu", "2026-09-08T13:00:00Z")
			return err
		}},
		"das Zählen beim Rotieren scheitert": {
			"stolper-token-zaehlen", func(a *Zugriffstokenablage) error {
				_, err := a.TokenRotieren(probekontext(t), 1, "neu", "2026-09-08T13:00:00Z")
				return err
			},
		},
		"der Vermerk scheitert": {"stolper-schreiben", func(a *Zugriffstokenablage) error {
			return a.TokenNutzungVermerken(probekontext(t), 1, "2026-09-08T12:00:00Z")
		}},
		"das Zählen beim Vermerk scheitert": {
			"stolper-token-zaehlen", func(a *Zugriffstokenablage) error {
				return a.TokenNutzungVermerken(probekontext(t), 1, "2026-09-08T12:00:00Z")
			},
		},
		"der Sammelwiderruf scheitert": {"stolper-schreiben", func(a *Zugriffstokenablage) error {
			_, err := a.TokenAlleLoeschen(probekontext(t), 1)
			return err
		}},
		"das Zählen beim Sammelwiderruf scheitert": {
			"stolper-token-zaehlen", func(a *Zugriffstokenablage) error {
				_, err := a.TokenAlleLoeschen(probekontext(t), 1)
				return err
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "egal")
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() {
				if fehler := db.Close(); fehler != nil {
					t.Errorf("schließen: %v", fehler)
				}
			})

			pruefeStolperfehler(t, fall.lauf(NeueZugriffstokenablage(db)))
		})
	}
}

// stolpertoken gibt ein gültiges Token für die Läufe oben.
func stolpertoken() domain.Zugriffstoken {
	return domain.Zugriffstoken{
		ProfilID: 1, Bezeichnung: "Skript", Scope: domain.ScopeLesen,
		Principal: "Nachtrechner", Angelegt: "2026-09-05T12:00:00Z",
	}
}

// gestolpertesToken ist eines mit Kennung und Fassung — zum Ändern.
func gestolpertesToken() domain.Zugriffstoken {
	eines := stolpertoken()
	eines.ID = 1
	eines.Fassung = 1

	return eines
}

// tokenExec beantwortet ein `Exec` auf einem Tokenweg.
//
// **Es steht hier und nicht in `stolperwerk_test.go`** — der Schnitt folgt der
// Frage: dort steht der Treiber, hier, **was der Tokenweg antwortet**. Und die
// Datei dort ging mit diesem Block über die Zeilengrenze aus §2.
//
// **`meiner` und kein Fehler**: „dieser Weg ist nicht meiner" ist keine
// Störung, sondern die Auskunft, dass der Aufrufer weitersuchen soll.
func (s stolperstmt) tokenExec() (ergebnis driver.Result, meiner bool) {
	geglaettet := strings.TrimSpace(s.abfrage)

	// **Eine Kennung, die sich nach dem Einfügen nicht lesen lässt** — den
	// Zweig gibt kein echter Treiber her.
	if strings.HasPrefix(geglaettet, "INSERT INTO zugriffstoken") {
		if s.stelle == beimTokenkennung {
			return kennungsfehler{}, true
		}

		if s.stelle == beimSchreiben {
			return nil, false
		}

		return kennung(1), true
	}

	// **Und das Ändern und Löschen zählen ihre Zeilen** — aus null wird dort
	// „gibt es nicht" oder „hat sich geändert", und wer das liest, hört auf zu
	// suchen.
	if (strings.HasPrefix(geglaettet, "UPDATE zugriffstoken") ||
		strings.HasPrefix(geglaettet, "DELETE FROM zugriffstoken")) &&
		s.stelle != beimSchreiben {
		if s.stelle == beimTokenzaehlen {
			return zaehlfehler{}, true
		}

		return driver.RowsAffected(0), true
	}

	return nil, false
}
