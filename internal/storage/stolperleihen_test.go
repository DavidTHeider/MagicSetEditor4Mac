package storage

// **Die Fehlerwege von Freigabe und Leihe, die kein echter Treiber hergibt** —
// Werkstücke G5 und G6 von Phase 5.5
// (#1462,
// #1463).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen — dieselbe Form
// wie bei `stolperpartien_test.go`, und eine eigene Datei aus demselben
// Grund (§2).
//
// **Was er prüft, gibt SQLite nicht her**: eine Zeile, die sich nicht lesen
// lässt, und ein Zeilensatz, der **nach** der Schleife abbricht. Ohne diesen
// Treiber stünden die Zweige daneben ungeprüft da und sähen aus wie geprüfte
// Zusagen (§10).

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen Freigabe und Leihe stolpern können.
const (
	beimOffenlegungsscan stolperstelle = iota + 2100
	beimOffenlegungsende
	beimLeihscan
	beimLeihenende
	beimOffenlegungRaeumen
	beimOffenlegungEintragen
	beimOffenlegungCommit
	beimVergessenZaehlen
	beimBindungRaeumen
	beimBindungSchreiben
	beimBindungCommit
	beimMeldungsscan
	beimMeldungsende
)

func init() {
	sql.Register("stolper-offenlegung-scan",
		stolpertreiber{stelle: beimOffenlegungsscan})
	sql.Register("stolper-offenlegungen-ende",
		stolpertreiber{stelle: beimOffenlegungsende})
	sql.Register("stolper-leih-scan", stolpertreiber{stelle: beimLeihscan})
	sql.Register("stolper-leihen-ende", stolpertreiber{stelle: beimLeihenende})
	sql.Register("stolper-offenlegung-raeumen",
		stolpertreiber{stelle: beimOffenlegungRaeumen})
	sql.Register("stolper-offenlegung-eintragen",
		stolpertreiber{stelle: beimOffenlegungEintragen})
	sql.Register("stolper-offenlegung-commit",
		stolpertreiber{stelle: beimOffenlegungCommit})
	sql.Register("stolper-vergessen-zaehlen",
		stolpertreiber{stelle: beimVergessenZaehlen})
	sql.Register("stolper-bindung-raeumen",
		stolpertreiber{stelle: beimBindungRaeumen})
	sql.Register("stolper-bindung-schreiben",
		stolpertreiber{stelle: beimBindungSchreiben})
	sql.Register("stolper-bindung-commit",
		stolpertreiber{stelle: beimBindungCommit})
	sql.Register("stolper-meldungs-scan",
		stolpertreiber{stelle: beimMeldungsscan})
	sql.Register("stolper-meldungen-ende",
		stolpertreiber{stelle: beimMeldungsende})
}

// leihschreiben beantwortet die Anweisungen beider Wege.
func leihschreiben(
	stelle stolperstelle, abfrage string,
) (driver.Result, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "DELETE FROM partiefreigaben"):
		if stelle == beimOffenlegungRaeumen {
			return nil, true, errGestolpert
		}

		return leereWirkung{}, true, nil

	case strings.HasPrefix(geglaettet, "INSERT INTO partiefreigaben"):
		if stelle == beimOffenlegungEintragen {
			return nil, true, errGestolpert
		}

		return leereWirkung{}, true, nil

	case strings.HasPrefix(geglaettet, "DELETE FROM matchbindungen"):
		if stelle == beimBindungRaeumen {
			return nil, true, errGestolpert
		}

		return leereWirkung{}, true, nil

	case strings.HasPrefix(geglaettet, "INSERT INTO matchbindungen"):
		if stelle == beimBindungSchreiben {
			return nil, true, errGestolpert
		}

		return leereWirkung{}, true, nil

	case strings.HasPrefix(geglaettet, "DELETE FROM partien WHERE id"):
		if stelle == beimVergessenZaehlen {
			// **Ein Ergebnis, dessen Zählung klagt.** SQLite gibt hier nie
			// einen Fehler; ohne diesen Treiber stünde der Zweig daneben
			// ungeprüft da (§10).
			return klagendeZaehlung{}, true, nil
		}

		return leereWirkung{}, true, nil
	}

	return nil, false, nil
}

// leereWirkung ist ein Ergebnis ohne Wirkung — und ohne Klage.
type leereWirkung struct{}

func (leereWirkung) LastInsertId() (int64, error) { return 0, nil }
func (leereWirkung) RowsAffected() (int64, error) { return 1, nil }

// klagendeZaehlung klagt beim Zählen der geänderten Zeilen.
type klagendeZaehlung struct{}

func (klagendeZaehlung) LastInsertId() (int64, error) { return 0, nil }
func (klagendeZaehlung) RowsAffected() (int64, error) { return 0, errGestolpert }

// leihabfrage beantwortet die Abfragen beider Wege — oder stolpert.
func leihabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "SELECT von, an FROM partiefreigaben"):
		switch stelle {
		case beimOffenlegungsscan:
			return &krummeZeile{spalten: []string{"von", "an"}}, true
		case beimOffenlegungsende:
			return &abbrechendeZeile{spalten: []string{"von", "an"}}, true
		}

		return &leerezeile{spalten: 2}, true

	case strings.HasPrefix(geglaettet, "SELECT von, fassung, partie_id"):
		spalten := []string{
			"von", "fassung", "partie_id", "spiele_a", "spiele_b",
			"unentschieden", "kette",
		}

		switch stelle {
		case beimMeldungsscan:
			return &krummeZeile{spalten: spalten}, true
		case beimMeldungsende:
			return &abbrechendeZeile{spalten: spalten}, true
		}

		return &leerezeile{spalten: len(spalten)}, true

	case strings.HasPrefix(geglaettet, "SELECT dokument, abzug, verliehen"):
		spalten := []string{
			"dokument", "abzug", "verliehen", "verbraucht", "widerrufen",
		}

		switch stelle {
		case beimLeihscan:
			return &krummeZeile{spalten: spalten}, true
		case beimLeihenende:
			return &abbrechendeZeile{spalten: spalten}, true
		}

		return &leerezeile{spalten: len(spalten)}, true
	}

	return nil, false
}

// krummeZeile gibt in jeder Spalte ein NULL — daran scheitert der Scan.
//
// **NULL und nicht „keine Zeichenkette"**: die Spalten sind TEXT, und
// `database/sql` macht aus jedem anderen Wert klaglos eine.
type krummeZeile struct {
	spalten []string
	gegeben bool
}

func (k *krummeZeile) Columns() []string { return k.spalten }
func (k *krummeZeile) Close() error      { return nil }

func (k *krummeZeile) Next(ziel []driver.Value) error {
	if k.gegeben {
		return io.EOF
	}

	k.gegeben = true

	for stelle := range ziel {
		ziel[stelle] = nil
	}

	return nil
}

// abbrechendeZeile bricht **vor** der ersten Zeile ab — der Fehler steht dann
// in `zeilen.Err()` und nicht im Scan.
type abbrechendeZeile struct{ spalten []string }

func (a *abbrechendeZeile) Columns() []string         { return a.spalten }
func (a *abbrechendeZeile) Close() error              { return nil }
func (a *abbrechendeZeile) Next([]driver.Value) error { return errGestolpert }

// **Jeder Weg reicht seinen Fehler durch.**
func TestJederFreigabeUndLeihwegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*sql.DB) error
	}{
		"eine Freigabezeile lässt sich nicht lesen": {
			"stolper-offenlegung-scan", func(db *sql.DB) error {
				_, fehler := NeuePartieablage(db).Freigaben(
					probekontext(t), "p1")

				return fehler
			},
		},
		"die Freigabezeilen brechen ab": {
			"stolper-offenlegungen-ende", func(db *sql.DB) error {
				_, fehler := NeuePartieablage(db).Freigaben(
					probekontext(t), "p1")

				return fehler
			},
		},
		"eine Leihzeile lässt sich nicht lesen": {
			"stolper-leih-scan", func(db *sql.DB) error {
				_, fehler := NeueLeihablage(db).Leihen(probekontext(t))

				return fehler
			},
		},
		"die Leihzeilen brechen ab": {
			"stolper-leihen-ende", func(db *sql.DB) error {
				_, fehler := NeueLeihablage(db).Leihen(probekontext(t))

				return fehler
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if fehler := fall.fahre(
				stolperbestand(t, fall.treiber)); fehler == nil {
				t.Fatal("der Fehler kam nicht an")
			}
		})
	}
}

// **Jeder Weg im Geschäft der Freigabe reicht seinen Fehler durch.**
//
// **Was SQLite nicht hergibt**: eine Transaktion, die beginnt, deren
// Anweisung aber scheitert — und ein Commit, das klagt.
func TestJederWegImFreigabegeschaeftReichtSeinenFehlerDurch(t *testing.T) {
	gabe := domain.Offenlegung{Von: "anna", An: []string{"ben"}}

	faelle := map[string]struct {
		treiber string
		fahre   func(*sql.DB) error
	}{
		"das Wegräumen scheitert": {
			"stolper-offenlegung-raeumen", func(db *sql.DB) error {
				return NeuePartieablage(db).Freigeben(
					probekontext(t), "p1", gabe)
			},
		},
		"das Eintragen scheitert": {
			"stolper-offenlegung-eintragen", func(db *sql.DB) error {
				return NeuePartieablage(db).Freigeben(
					probekontext(t), "p1", gabe)
			},
		},
		"das Commit scheitert": {
			"stolper-offenlegung-commit", func(db *sql.DB) error {
				return NeuePartieablage(db).Freigabeloeschen(
					probekontext(t), "p1", "anna")
			},
		},
		"die geänderten Zeilen lassen sich nicht zählen": {
			"stolper-vergessen-zaehlen", func(db *sql.DB) error {
				return NeuePartieablage(db).Vergessen(probekontext(t), "p1")
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if fehler := fall.fahre(
				stolperbestand(t, fall.treiber)); fehler == nil {
				t.Fatal("der Fehler kam nicht an")
			}
		})
	}
}

// **Jeder Weg der Matchablage reicht seinen Fehler durch.**
func TestJederMatchwegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*sql.DB) error
	}{
		"das Wegräumen der Bindung scheitert": {
			"stolper-bindung-raeumen", func(db *sql.DB) error {
				return NeueMatchablage(db).BindungSchreiben(
					probekontext(t), eineStolperbindung())
			},
		},
		"das Schreiben der Bindung scheitert": {
			"stolper-bindung-schreiben", func(db *sql.DB) error {
				return NeueMatchablage(db).BindungSchreiben(
					probekontext(t), eineStolperbindung())
			},
		},
		"das Commit scheitert": {
			"stolper-bindung-commit", func(db *sql.DB) error {
				return NeueMatchablage(db).BindungSchreiben(
					probekontext(t), eineStolperbindung())
			},
		},
		"eine Meldungszeile lässt sich nicht lesen": {
			"stolper-meldungs-scan", func(db *sql.DB) error {
				_, fehler := NeueMatchablage(db).Meldungen(probekontext(t), 7)

				return fehler
			},
		},
		"die Meldungszeilen brechen ab": {
			"stolper-meldungen-ende", func(db *sql.DB) error {
				_, fehler := NeueMatchablage(db).Meldungen(probekontext(t), 7)

				return fehler
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if fehler := fall.fahre(
				stolperbestand(t, fall.treiber)); fehler == nil {
				t.Fatal("der Fehler kam nicht an")
			}
		})
	}
}

// eineStolperbindung ist die kleinste, die geschrieben werden kann.
func eineStolperbindung() domain.Matchbindung {
	return domain.Matchbindung{
		TurnierID: 1, PaarungID: 7, Fassung: 3, PartieID: "p1",
		Sitze: []domain.Matchsitz{
			{TeilnehmerID: 10, Sitz: "anna", Deckabdruck: "a"},
			{TeilnehmerID: 11, Sitz: "bert", Deckabdruck: "b"},
		},
	}
}
