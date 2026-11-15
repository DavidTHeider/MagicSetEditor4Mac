package storage

// **Die Stolperstellen der Turnierfassung** — Befund 7 des Nachlaufs über
// Phase 4.7.
//
// **Eine eigene Datei**, damit `stolperturnier_test.go` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, wo der
// **Turnierweg** stolpert, hier, wo seine **Buchführung über die Fassung** es
// tut.
//
// **Warum sie überhaupt gebraucht werden.** Die Fassung entscheidet, ob zwei
// Leute am selben Tisch einander überschreiben. Ihre Fehlerwege sind gegen
// eine echte SQLite-Datei nicht zu betreten — `RowsAffected` scheitert dort
// nie —, und im Ernstfall unterscheiden sie „das Turnier gibt es nicht" von
// „jemand anderes war schneller" von „das Zählen ging schief".

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
)

// Die Stellen, an denen die Buchführung stolpern kann.
const (
	beimFassungschreiben stolperstelle = iota + 950
	beimFassungzaehlen
	beimFassungnachschlagen
	beimFassungfehlt
	beimTeilnehmerturnier
	beimPaarungsturnier
	beimEintragenbeginnen
	beimAbmeldenbeginnen

	// **Abschluss und Protokoll** — DOM-04.AK-04 (#1699).
	beimAbschlusszaehlen
	beimProtokollschreiben
	beimPaarungschreiben
	beimProtokolllesen
	beimProtokollscan
)

func init() {
	sql.Register("stolper-fassung-schreiben",
		stolpertreiber{stelle: beimFassungschreiben})
	sql.Register("stolper-fassung-zaehlen",
		stolpertreiber{stelle: beimFassungzaehlen})
	sql.Register("stolper-fassung-nachschlagen",
		stolpertreiber{stelle: beimFassungnachschlagen})
	sql.Register("stolper-fassung-fehlt", stolpertreiber{stelle: beimFassungfehlt})
	sql.Register("stolper-teilnehmer-turnier",
		stolpertreiber{stelle: beimTeilnehmerturnier})
	sql.Register("stolper-paarungs-turnier",
		stolpertreiber{stelle: beimPaarungsturnier})
	sql.Register("stolper-eintragen-beginnen",
		stolpertreiber{stelle: beimEintragenbeginnen})
	sql.Register("stolper-abmelden-beginnen",
		stolpertreiber{stelle: beimAbmeldenbeginnen})
	sql.Register("stolper-abschluss-zaehlen", stolpertreiber{stelle: beimAbschlusszaehlen})
	sql.Register("stolper-protokoll-schreiben", stolpertreiber{stelle: beimProtokollschreiben})
	sql.Register("stolper-paarung-schreiben", stolpertreiber{stelle: beimPaarungschreiben})
	sql.Register("stolper-protokoll-lesen", stolpertreiber{stelle: beimProtokolllesen})
	sql.Register("stolper-protokoll-scan", stolpertreiber{stelle: beimProtokollscan})
}

// fassungsabfrage beantwortet die Abfragen der Buchführung — oder stolpert.
func fassungsabfrage(stelle stolperstelle, abfrage string) (
	driver.Rows, bool, error,
) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	// **Das Turnier des Teilnehmers** — eine Spalte, beim Abmelden.
	case strings.HasPrefix(geglaettet, "SELECT turnier_id FROM turnierteilnehmer"):
		if stelle == beimTeilnehmerturnier {
			return nil, true, errGestolpert
		}

		return &einezahl{}, true, nil

	// **Das Turnier der Paarung** — eine Spalte, beim Eintragen.
	//
	// **`beimFassungfehlt` gibt hier eine gültige Zeile.** Der Weg dahin muss
	// offen sein, sonst käme das `ErrNotFound` der Paarung und nicht das der
	// Fassung — zwei verschiedene Auskünfte, und geprüft würde die falsche.
	case strings.HasPrefix(geglaettet, "SELECT r.turnier_id"):
		if stelle == beimPaarungsturnier {
			return nil, true, errGestolpert
		}

		return &einezahl{}, true, nil

	// **Die gegenwärtige Fassung** — gefragt wird sie nur, wenn das
	// Fortschreiben keine Zeile getroffen hat.
	//
	// **Drei Spalten seit dem Abschluss** (DOM-04.AK-04, #1699): Fassung,
	// Name und ob das Turnier abgeschlossen ist.
	// **Das Protokoll** (DOM-04.AK-04) — vier Spalten, oder ein Stolpern.
	case strings.HasPrefix(geglaettet, "SELECT zeit, vorgang"):
		switch stelle {
		case beimProtokolllesen:
			return nil, true, errGestolpert
		case beimProtokollscan:
			return &krummeViererzeile{}, true, nil
		}

		return &leerezeile{spalten: 4}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT fassung, name, abgeschlossen"):
		switch stelle {
		case beimFassungnachschlagen:
			return nil, true, errGestolpert
		case beimFassungfehlt:
			return &leerezeile{spalten: 3}, true, nil
		}

		return &fassungszeile{}, true, nil
	}

	return nil, false, nil
}

// fassungsschreiben beantwortet das Fortschreiben — oder stolpert.
func fassungsschreiben(stelle stolperstelle, abfrage string) (
	driver.Result, bool, error,
) {
	geglaettet := strings.TrimSpace(abfrage)

	// **Das Abmelden trifft seine Zeile.** Ohne diese Zeile gäbe der Treiber
	// null betroffene Zeilen, `Aussteigen` machte daraus ein `ErrNotFound`,
	// und die Buchführung dahinter bliebe unbetreten — der Lauf wäre grün,
	// weil ein Fehler kam, und zwar der falsche.
	if strings.HasPrefix(geglaettet, "UPDATE turnierteilnehmer SET ausgestiegen") &&
		stelle != beimSchreiben && stelle != beimZaehlen {
		return driver.RowsAffected(1), true, nil
	}

	if ergebnis, erkannt, fehler := abschlussschreibweg(stelle, geglaettet); erkannt {
		return ergebnis, erkannt, fehler
	}

	if !strings.HasPrefix(geglaettet, "UPDATE turniere SET fassung") {
		return nil, false, nil
	}

	switch stelle {
	case beimFassungschreiben:
		return nil, true, errGestolpert
	case beimFassungzaehlen:
		return zaehlfehler{}, true, nil
	case beimFassungnachschlagen, beimFassungfehlt:
		// **Keine Zeile getroffen** — erst dann wird nachgeschlagen.
		return driver.RowsAffected(0), true, nil
	}

	return driver.RowsAffected(1), true, nil
}

// fassungszeile ist ein offenes Turnier bei Fassung 1 — das Nachschlagen nach
// einem Fortschreiben, das keine Zeile traf.
type fassungszeile struct{ gegeben bool }

func (r *fassungszeile) Columns() []string { return []string{"fassung", "name", "abgeschlossen"} }
func (r *fassungszeile) Close() error      { return nil }

func (r *fassungszeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0], werte[1], werte[2] = int64(1), "Küchentisch", int64(0)

	return nil
}

// abschlussschreibweg beantwortet die Schreibwege von Abschluss und Protokoll
// — DOM-04.AK-04 (#1699). **Die allgemeine Stelle `beimSchreiben` behält
// ihren Vorrang**, sonst legte diese Antwort den Fehlerweg des Abschlusses
// still.
func abschlussschreibweg(stelle stolperstelle, geglaettet string) (driver.Result, bool, error) {
	switch {
	case stelle == beimProtokollschreiben && strings.HasPrefix(geglaettet, "INSERT INTO turnierprotokoll"),
		stelle == beimPaarungschreiben && strings.HasPrefix(geglaettet, "UPDATE turnierpaarungen"):
		return nil, true, errGestolpert

	case strings.HasPrefix(geglaettet, "UPDATE turniere") && strings.Contains(geglaettet, "abgeschlossen ="):
		switch stelle {
		case beimSchreiben:
			return nil, true, errGestolpert
		case beimAbschlusszaehlen:
			return zaehlfehler{}, true, nil
		}

		return driver.RowsAffected(1), true, nil
	}

	return nil, false, nil
}
