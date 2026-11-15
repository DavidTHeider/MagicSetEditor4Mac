package storage

// **Die Fehlerwege des Faktorspeichers** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Was er prüft,
// gibt kein echter Treiber her: eine Abfrage, die scheitert; eine Zeile, die
// sich nicht lesen lässt; ein Abbruch **nach** der Schleife.

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Faktorweg stolpern kann.
const (
	beimFaktorscan stolperstelle = iota + 1300
	beimFaktorende
	beimCodescan
	beimCodeende
	beimBestandskennung
	beimEinstellungsscan
	beimEinstellungsende
)

func init() {
	sql.Register("stolper-faktor-scan", stolpertreiber{stelle: beimFaktorscan})
	sql.Register("stolper-faktor-ende", stolpertreiber{stelle: beimFaktorende})
	sql.Register("stolper-code-scan", stolpertreiber{stelle: beimCodescan})
	sql.Register("stolper-code-ende", stolpertreiber{stelle: beimCodeende})
	sql.Register("stolper-bestandskennung", stolpertreiber{stelle: beimBestandskennung})
	sql.Register("stolper-einstellung-scan", stolpertreiber{stelle: beimEinstellungsscan})
	sql.Register("stolper-einstellung-ende", stolpertreiber{stelle: beimEinstellungsende})
}

// faktorabfrage beantwortet die Abfragen des Faktorwegs — oder stolpert.
func faktorabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "SELECT id, nutzer_id, art, bezeichnung"):
		switch stelle {
		case beimFaktorscan:
			return &krummeViererzeile{}, true, nil
		case beimFaktorende:
			return &abbrechendeViererzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 4}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT id, gehuetet FROM wiederherstellungscodes"):
		switch stelle {
		case beimCodescan:
			return &krummeZweierzeile{}, true, nil
		case beimCodeende:
			return &abbrechendeZweierzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 2}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT schluessel, wert FROM nutzereinstellungen"):
		switch stelle {
		case beimEinstellungsscan:
			// **NULL und nicht „keine Zahl"**: beide Spalten sind Text, und
			// `database/sql` macht aus jedem anderen Wert klaglos eine
			// Zeichenkette. Nur ein `NULL` scheitert am Scan.
			return &nullzweierzeile{}, true, nil
		case beimEinstellungsende:
			return &abbrechendeZweierzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 2}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT wert FROM bestandskennung"):
		if stelle == beimBestandskennung {
			return nil, true, errGestolpert
		}

		return &namenszeile{name: "abc"}, true, nil
	}

	return nil, false, nil
}

// **Jeder Weg des Faktorspeichers reicht seinen Fehler durch.**
func TestJederFaktorwegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*Nutzerablage) error
	}{
		"die Faktorabfrage scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.Faktoren(probekontext(t), 1)

			return err
		}},
		"eine Faktorzeile lässt sich nicht lesen": {
			"stolper-faktor-scan", func(a *Nutzerablage) error {
				_, err := a.Faktoren(probekontext(t), 1)

				return err
			},
		},
		"die Faktorzeilen brechen ab": {"stolper-faktor-ende", func(a *Nutzerablage) error {
			_, err := a.Faktoren(probekontext(t), 1)

			return err
		}},
		"das Anlegen scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			_, err := a.FaktorAnlegen(probekontext(t),
				domain.Faktor{NutzerID: 1, Art: domain.FaktorTOTP})

			return err
		}},
		"die Kennung des Faktors": {"stolper-nutzer-kennung", func(a *Nutzerablage) error {
			_, err := a.FaktorAnlegen(probekontext(t),
				domain.Faktor{NutzerID: 1, Art: domain.FaktorTOTP})

			return err
		}},
		"das Löschen scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			return a.FaktorLoeschen(probekontext(t), 1, domain.FaktorTOTP)
		}},
		"das Zählen beim Löschen": {"stolper-zaehlen", func(a *Nutzerablage) error {
			return a.FaktorLoeschen(probekontext(t), 1, domain.FaktorTOTP)
		}},
		"der Beginn beim Setzen der Codes": {
			"stolper-eintragen-beginnen", func(a *Nutzerablage) error {
				return a.CodesSetzen(probekontext(t), 1, []string{"eins"})
			},
		},
		"das Entfernen der alten Codes": {"stolper-schreiben", func(a *Nutzerablage) error {
			return a.CodesSetzen(probekontext(t), 1, []string{"eins"})
		}},
		"das Eintragen eines Codes": {
			"stolper-zweites-schreiben", func(a *Nutzerablage) error {
				return a.CodesSetzen(probekontext(t), 1, []string{"eins"})
			},
		},
		"das Festschreiben der Codes": {"stolper-commit", func(a *Nutzerablage) error {
			return a.CodesSetzen(probekontext(t), 1, []string{"eins"})
		}},
		"die Codeabfrage scheitert": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.FreieCodes(probekontext(t), 1)

			return err
		}},
		"eine Codezeile lässt sich nicht lesen": {
			"stolper-code-scan", func(a *Nutzerablage) error {
				_, err := a.FreieCodes(probekontext(t), 1)

				return err
			},
		},
		"die Codezeilen brechen ab": {"stolper-code-ende", func(a *Nutzerablage) error {
			_, err := a.FreieCodes(probekontext(t), 1)

			return err
		}},
		"das Abstempeln scheitert": {"stolper-schreiben", func(a *Nutzerablage) error {
			return a.CodeVerbrauchen(probekontext(t), 1)
		}},
		"das Zählen beim Abstempeln": {"stolper-zaehlen", func(a *Nutzerablage) error {
			return a.CodeVerbrauchen(probekontext(t), 1)
		}},

		// **Die Einstellungen des Nutzers** — Werkstück 8 (#1351).
		"die Einstellungsabfrage scheitert": {
			"stolper-fragen", func(a *Nutzerablage) error {
				_, err := a.Nutzereinstellungen(probekontext(t), 1)

				return err
			},
		},
		"eine Einstellungszeile lässt sich nicht lesen": {
			"stolper-einstellung-scan", func(a *Nutzerablage) error {
				_, err := a.Nutzereinstellungen(probekontext(t), 1)

				return err
			},
		},
		"die Einstellungszeilen brechen ab": {
			"stolper-einstellung-ende", func(a *Nutzerablage) error {
				_, err := a.Nutzereinstellungen(probekontext(t), 1)

				return err
			},
		},
		// **Und derselbe Weg zum Menschen des gezeigten Profils** — Werkstück
		// 9 (#1351), Punkt 9. **Drei Fälle und nicht einer**: die Abfrage
		// scheitert, eine Zeile lässt sich nicht lesen, und die Zeilen
		// brechen nach der Schleife ab.
		"die Abfrage zum Profil scheitert": {
			"stolper-fragen", func(a *Nutzerablage) error {
				_, err := a.NutzerwerteZumProfil(probekontext(t))

				return err
			},
		},
		"eine Zeile zum Profil lässt sich nicht lesen": {
			"stolper-einstellung-scan", func(a *Nutzerablage) error {
				_, err := a.NutzerwerteZumProfil(probekontext(t))

				return err
			},
		},
		"die Zeilen zum Profil brechen ab": {
			"stolper-einstellung-ende", func(a *Nutzerablage) error {
				_, err := a.NutzerwerteZumProfil(probekontext(t))

				return err
			},
		},
		"das Setzen einer Einstellung scheitert": {
			"stolper-schreiben", func(a *Nutzerablage) error {
				return a.NutzereinstellungSetzen(probekontext(t), 1,
					domain.NutzerBewegung, domain.BewegungReduziert)
			},
		},

		// **Und die Nachfrage nach einem Schreiben, das null Zeilen traf.**
		//
		// **Sie kostet eine zweite Abfrage** — und wenn die scheitert, ist
		// die Frage unbeantwortet. Das ist etwas anderes als „gibt es nicht"
		// (§8).
		"die Nachfrage beim Nutzer": {"stolper-fragen", func(a *Nutzerablage) error {
			_, err := a.NutzerAendern(probekontext(t),
				domain.Nutzer{ID: 1, Name: "Wer", Fassung: 1})

			return err
		}},
		"die Nachfrage beim Profil": {
			"stolper-profil-nachschlagen", func(a *Nutzerablage) error {
				_, err := a.ProfilAendern(probekontext(t),
					domain.Profil{ID: 1, NutzerID: 1, Name: "Wer", Fassung: 1})

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

			t.Cleanup(func() { _ = db.Close() })

			pruefeStolperfehler(t, fall.fahre(NeueNutzerablage(db)))
		})
	}
}

// **Und die Bestandskennung ebenso.**
func TestDieBestandskennungReichtIhrenFehlerDurch(t *testing.T) {
	db, err := sql.Open("stolper-bestandskennung", "egal")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	_, fehler := Bestandskennung(probekontext(t), db)
	pruefeStolperfehler(t, fehler)
}

// krummeZweierzeile gibt eine Codezeile, die sich nicht lesen lässt.
type krummeZweierzeile struct{ gegeben bool }

func (krummeZweierzeile) Columns() []string { return []string{"id", "gehuetet"} }

func (krummeZweierzeile) Close() error { return nil }

func (k *krummeZweierzeile) Next(werte []driver.Value) error {
	if k.gegeben {
		return io.EOF
	}

	k.gegeben = true
	werte[0], werte[1] = "keine Zahl", 1

	return nil
}

// abbrechendeZweierzeile bricht **nach** der Schleife ab.
type abbrechendeZweierzeile struct{}

func (abbrechendeZweierzeile) Columns() []string { return []string{"id", "gehuetet"} }

func (abbrechendeZweierzeile) Close() error { return nil }

func (abbrechendeZweierzeile) Next([]driver.Value) error { return errGestolpert }

// nullzweierzeile gibt zwei NULL — der einzige Weg, einen Scan auf zwei
// Textspalten scheitern zu lassen.
type nullzweierzeile struct{ gegeben bool }

func (nullzweierzeile) Columns() []string { return []string{"schluessel", "wert"} }

func (nullzweierzeile) Close() error { return nil }

func (n *nullzweierzeile) Next(werte []driver.Value) error {
	if n.gegeben {
		return io.EOF
	}

	n.gegeben = true
	werte[0], werte[1] = nil, nil

	return nil
}
