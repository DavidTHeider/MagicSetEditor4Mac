package storage

// **Die Stolperstellen des Turnierspeichers** — Werkstück 3 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `stolperwerk_test.go` und
// `stolperabfragen_test.go` unter der Zeilengrenze aus §2 bleiben. Der
// Schnitt folgt der Frage: dort steht, wie **allgemein** gestolpert wird,
// hier, wo der **Turnierweg** es tut.
//
// # Warum die Anweisung entscheidet und kein Zähler
//
// Dieselbe Überlegung wie in `stolperwerk_test.go`: ein Zähler hinge an der
// Zahl der Vorgänge, und die ändert sich, sobald jemand einen dazwischen
// einfügt. Der Präfix einer Anweisung ändert sich nur, wenn sie sich ändert.

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
)

// Die Stellen, an denen der Turnierweg stolpern kann.
const (
	beimTeilnehmerabfrage stolperstelle = iota + 900
	beimTeilnehmerscan
	beimTeilnehmerende
	beimVerlaufabfrage
	beimVerlaufscan
	beimRundenbeginn
	beimRundencommit
	beimRundenkennung

	// **`beimPaarungszaehlen` stand hier**, und mit ihm ein Lauf über das
	// Zählen der berührten Paarung. Beides ist mit Befund 7 des Nachlaufs
	// über Phase 4.7 weggefallen: `ErgebnisSetzen` zählt keine Zeilen mehr,
	// sondern fragt vorher, zu welchem Turnier die Paarung gehört — und
	// **diese** Frage beantwortet zugleich, ob es sie gibt. Eine
	// Stolperstelle ohne Weg dahinter ist eine Zusage ohne Rumpf.
	beimTurnierlistescan
	beimTurnierlisteende
	beimAbzugabfrage
	beimAbzugscan
	beimAbzugkennung
	beimTeilnehmerkennung

	// **Der Anlegeweg mit Teilnehmenden** — DOM-04.AK-01 (#1699): ein Zug,
	// der nicht beginnt, und eine Kennung, die sich nicht lesen lässt.
	beimTurnieranlagebeginn
	beimTurnierkennung
)

func init() {
	sql.Register("stolper-teilnehmer-abfrage", stolpertreiber{stelle: beimTeilnehmerabfrage})
	sql.Register("stolper-teilnehmer-scan", stolpertreiber{stelle: beimTeilnehmerscan})
	sql.Register("stolper-teilnehmer-ende", stolpertreiber{stelle: beimTeilnehmerende})
	sql.Register("stolper-verlauf-abfrage", stolpertreiber{stelle: beimVerlaufabfrage})
	sql.Register("stolper-verlauf-scan", stolpertreiber{stelle: beimVerlaufscan})
	sql.Register("stolper-runden-beginn", stolpertreiber{stelle: beimRundenbeginn})
	sql.Register("stolper-runden-commit", stolpertreiber{stelle: beimRundencommit})
	sql.Register("stolper-runden-kennung", stolpertreiber{stelle: beimRundenkennung})
	sql.Register("stolper-turnierliste-scan", stolpertreiber{stelle: beimTurnierlistescan})
	sql.Register("stolper-turnierliste-ende", stolpertreiber{stelle: beimTurnierlisteende})
	sql.Register("stolper-abzug-abfrage", stolpertreiber{stelle: beimAbzugabfrage})
	sql.Register("stolper-abzug-scan", stolpertreiber{stelle: beimAbzugscan})
	sql.Register("stolper-abzug-kennung", stolpertreiber{stelle: beimAbzugkennung})
	sql.Register("stolper-teilnehmer-kennung",
		stolpertreiber{stelle: beimTeilnehmerkennung})
	sql.Register("stolper-turnieranlage-beginn",
		stolpertreiber{stelle: beimTurnieranlagebeginn})
	sql.Register("stolper-turnier-kennung", stolpertreiber{stelle: beimTurnierkennung})
}

// turnierabfrage beantwortet die Abfragen des Turnierwegs — oder stolpert.
//
// **Sie wird von `stolperstmt.Query` gerufen**, bevor die allgemeinen
// Präfixe drankommen; wer hier nichts erkennt, gibt `nil, false` zurück und
// lässt den bisherigen Weg unverändert.
func turnierabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	// **Die nächste freie Stelle und die nächste Rundennummer** — beide
	// geben genau eine Zahl. Ohne diese Zeile bekämen sie die Zeile eines
	// anderen Falls und scheiterten am Scan statt an der Stelle.
	// **Der Kartenname für den Deckabzug** — eine Spalte.
	case strings.HasPrefix(geglaettet, "SELECT coalesce((SELECT wert FROM card_values"):
		if stelle == beimFragen {
			return nil, true, errGestolpert
		}

		return &einezahl{}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT coalesce(max("):
		if stelle == beimFragen {
			return nil, true, errGestolpert
		}

		return &einezahl{}, true, nil

	// **Die Liste liest dieselben acht Spalten wie der Kopf** — erkannt
	// wird sie an ihrem `FROM turniere` ohne `WHERE id`. Sieben seit dem
	// Format, acht seit dem Abschluss (DOM-04, #1699).
	case strings.Contains(geglaettet, "FROM turniere\n\t\t  WHERE profil_id"):
		switch stelle {
		case beimTurnierlistescan:
			return &krummeZeile{spalten: nutzerspaltennamen(8)}, true, nil
		case beimTurnierlisteende:
			return &abbrechendeSechserzeile{}, true, nil
		case beimFragen:
			return nil, true, errGestolpert
		}

		return &leerezeile{spalten: 8}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT id, fassung, name, spiel, einzelspiel"):
		// **Die allgemeine Stelle behält ihren Vorrang.** Ohne diese Zeile
		// bekäme `stolper-fragen` hier eine gültige Zeile, und der Fehlerweg
		// des Kopfes bliebe unbetreten — die Vorschaltung dieser Datei hätte
		// eine bestehende Prüfung stillgelegt.
		if stelle == beimFragen {
			return nil, true, errGestolpert
		}

		return &turnierkopfzeile{}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT id, turnier_id, name"):
		switch stelle {
		case beimTeilnehmerabfrage:
			return nil, true, errGestolpert
		case beimTeilnehmerscan:
			return &krummeSechserzeile{}, true, nil
		case beimTeilnehmerende:
			return &abbrechendeSechserzeile{}, true, nil
		}

		return &leerezeile{spalten: 6}, true, nil

	// **Die eingereichten Listen** — so viele Spalten, wie die Anweisung
	// liest. Hier standen sieben, als acht gelesen wurden, und die krumme
	// Zeile scheiterte an der Spaltenzahl statt am Scan (Nachlauf des
	// Zweitblicks über Phase B).
	case strings.HasPrefix(geglaettet, "SELECT p.teilnehmer_id"):
		switch stelle {
		case beimAbzugabfrage:
			return nil, true, errGestolpert
		case beimAbzugscan:
			return &krummeZeile{spalten: nutzerspaltennamen(spaltenzahlAus(geglaettet))}, true, nil
		}

		return &leerezeile{spalten: spaltenzahlAus(geglaettet)}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT r.id, r.nummer"):
		switch stelle {
		case beimVerlaufabfrage:
			return nil, true, errGestolpert
		case beimVerlaufscan:
			return &krummeNeunerzeile{}, true, nil
		}

		return &leerezeile{spalten: 9}, true, nil
	}

	return nil, false, nil
}

// einezahl gibt eine Zeile mit einer Null.
type einezahl struct{ gegeben bool }

func (r *einezahl) Columns() []string { return []string{"zahl"} }
func (r *einezahl) Close() error      { return nil }

func (r *einezahl) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = int64(1)

	return nil
}

// turnierkopfzeile gibt einen brauchbaren Turnierkopf.
//
// **Damit die Wege dahinter überhaupt erreicht werden.** Ein Kopf, der
// scheitert, ließe `teilnehmer` und `runden` ungeprüft — und die sind der
// Grund für diese Datei.
type turnierkopfzeile struct{ gegeben bool }

func (r *turnierkopfzeile) Columns() []string {
	return []string{"id", "fassung", "name", "spiel", "einzelspiel", "notiz", "format_id", "abgeschlossen"}
}
func (r *turnierkopfzeile) Close() error { return nil }

func (r *turnierkopfzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0], werte[1] = int64(1), int64(1)
	werte[2], werte[3] = "Küchentisch", "magic"
	werte[4], werte[5], werte[6], werte[7] = int64(0), "", nil, ""

	return nil
}

// leerezeile gibt gar keine Zeile — nur die Spaltenzahl stimmt.
type leerezeile struct{ spalten int }

func (r *leerezeile) Columns() []string         { return make([]string, r.spalten) }
func (r *leerezeile) Close() error              { return nil }
func (r *leerezeile) Next([]driver.Value) error { return io.EOF }

// krummeSechserzeile scheitert am Scan der Teilnehmer.
//
// **NULL und nicht „keine Zahl"** — derselbe Grund wie bei [krummezeile]:
// `database/sql` macht aus jedem anderen Wert klaglos eine Zeichenkette.
type krummeSechserzeile struct{ gegeben bool }

func (r *krummeSechserzeile) Columns() []string {
	return []string{"a", "b", "c", "d", "e", "f"}
}
func (r *krummeSechserzeile) Close() error { return nil }

func (r *krummeSechserzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// abbrechendeSechserzeile bricht **nach** der Schleife ab.
//
// **Der Zweig hinter `zeilen.Err()`.** Ein `io.EOF` beendet die Schleife
// still; nur ein anderer Fehler landet dort — und den gibt kein echter
// Treiber her.
type abbrechendeSechserzeile struct{}

func (r *abbrechendeSechserzeile) Columns() []string {
	return []string{"a", "b", "c", "d", "e", "f"}
}
func (r *abbrechendeSechserzeile) Close() error              { return nil }
func (r *abbrechendeSechserzeile) Next([]driver.Value) error { return errGestolpert }

// krummeNeunerzeile scheitert am Scan des Verlaufs.
type krummeNeunerzeile struct{ gegeben bool }

func (r *krummeNeunerzeile) Columns() []string {
	return []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
}
func (r *krummeNeunerzeile) Close() error { return nil }

func (r *krummeNeunerzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// kennungsfehler lässt `LastInsertId` scheitern und das Zählen glücken.
//
// **Das Gegenstück zu [zaehlfehler]**: dort scheitert `RowsAffected`, hier
// die Kennung. Beide sind beim echten Treiber unerreichbar und entscheiden im
// Ernstfall, ob die Anwendung sagt, was schiefging.
type kennungsfehler struct{}

func (kennungsfehler) LastInsertId() (int64, error) { return 0, errGestolpert }
func (kennungsfehler) RowsAffected() (int64, error) { return 1, nil }

// kennung ist ein Ergebnis mit einer brauchbaren Einfügekennung.
//
// **Das Gegenstück zu [driver.RowsAffected]**, dessen `LastInsertId` einen
// Fehler gibt: die Turnierwege lesen die Kennung nach jedem Einfügen, und ohne
// eine kämen sie nie bis zum Festschreiben.
type kennung int64

func (k kennung) LastInsertId() (int64, error) { return int64(k), nil }
func (k kennung) RowsAffected() (int64, error) { return 1, nil }

// spaltenzahlAus zählt die Spalten zwischen `SELECT` und `FROM` einer
// Anweisung — die Vorrichtung bietet, was die Anweisung liest, und nicht
// eine abgeschriebene Zahl.
func spaltenzahlAus(anweisung string) int {
	_, hinter, _ := strings.Cut(anweisung, "SELECT")
	spalten, _, _ := strings.Cut(hinter, "FROM")

	return len(strings.Split(spalten, ","))
}

// turnierExec beantwortet ein `Exec` auf einem Turnierweg.
//
// **Es steht hier und nicht in `stolperwerk_test.go`** — der Schnitt folgt der
// Frage: dort steht der Treiber, hier, **was der Turnierweg antwortet**. Und
// die Datei dort ging über die Zeilengrenze aus §2, als der Tokenweg dazukam
// (#1351, Werkstück 10).
//
// **`meiner` und kein Fehler**: „dieser Weg ist nicht meiner" ist keine
// Störung, sondern die Auskunft, dass der Aufrufer weitersuchen soll.
func (s stolperstmt) turnierExec() (ergebnis driver.Result, meiner bool) {
	geglaettet := strings.TrimSpace(s.abfrage)

	// **Die allgemeinen Stellen behalten ihren Vorrang** — sonst legten diese
	// Zeilen `stolper-schreiben` still, und der Fehlerweg jedes Einfügens
	// bliebe unbetreten.
	if s.stelle == beimTeilnehmerkennung &&
		strings.HasPrefix(geglaettet, "INSERT INTO turnierteilnehmer") {
		return kennungsfehler{}, true
	}

	if s.stelle == beimAbzugkennung &&
		strings.HasPrefix(geglaettet, "INSERT INTO turnierdeckplaetze") {
		return kennungsfehler{}, true
	}

	if s.stelle == beimTurnierkennung &&
		strings.HasPrefix(geglaettet, "INSERT INTO turniere ") {
		return kennungsfehler{}, true
	}

	return nil, false
}
