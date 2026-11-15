package storage

// **Die Fehlerwege des Wiederherstellungsbuchs** — Werkstück O7 von Phase 5.2
// (#1431).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Acht Wege gibt
// eine echte SQLite-Datenbank nicht her: ein `INSERT`, das scheitert, obwohl
// die Transaktion begann; eine Kennung, die sich nach dem Einfügen nicht lesen
// lässt; ein `Commit`, das nach lauter geglückten Schritten abbricht; ein
// Zählen der Quittung, das scheitert; und je ein Abbruch **mitten** im Lesen
// und ein Scan, der nicht passt — für die Epochen wie für ihre Mengen.
//
// **Warum sie geprüft gehören**: das Buch ist der Nachweis über eine Rückkehr.
// Ein Weg, der still eine leere Liste gibt, sagt „es hat keine Rückkehr
// gegeben" — und das ist die eine Auskunft, die niemals falsch sein darf.

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Rückkehrweg stolpern kann.
const (
	beimRueckkehrschreiben stolperstelle = iota + 1900
	beimRueckkehrkennung
	beimRueckkehrende
	beimRueckkehrscan
	beimMengenende
	beimMengenscan
	beimMengenfragen
	beimSammlungsscan
	beimSammlungsende
)

func init() {
	sql.Register("stolper-rueckkehr-schreiben",
		stolpertreiber{stelle: beimRueckkehrschreiben})
	sql.Register("stolper-rueckkehr-kennung",
		stolpertreiber{stelle: beimRueckkehrkennung})
	sql.Register("stolper-rueckkehr-ende", stolpertreiber{stelle: beimRueckkehrende})
	sql.Register("stolper-rueckkehr-scan", stolpertreiber{stelle: beimRueckkehrscan})
	sql.Register("stolper-mengen-ende", stolpertreiber{stelle: beimMengenende})
	sql.Register("stolper-mengen-scan", stolpertreiber{stelle: beimMengenscan})
	sql.Register("stolper-mengen-fragen", stolpertreiber{stelle: beimMengenfragen})
	sql.Register("stolper-sammlung-scan", stolpertreiber{stelle: beimSammlungsscan})
	sql.Register("stolper-sammlung-ende", stolpertreiber{stelle: beimSammlungsende})
}

// rueckkehrspalten sind die sechs Spalten einer Epochenzeile.
var rueckkehrspalten = []string{
	"id", "zurueckgespielt", "datenstand", "quelle", "klaerung", "geklaert_am",
}

// mengenspalten sind die acht Spalten einer Mengenzeile.
var mengenspalten = []string{
	"schluessel", "zustand", "sprache", "foil", "menge", "name", "setcode", "nummer",
}

// sammlungsstandspalten sind die zehn Spalten einer gruppierten Sammlungszeile.
var sammlungsstandspalten = []string{
	"profil_id", "druck_id", "printing", "zustand", "sprache", "foil",
	"menge", "name", "setcode", "nummer",
}

// rueckkehrabfrage beantwortet die zwei Abfragen des Buchs — oder stolpert.
func rueckkehrabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "SELECT id, zurueckgespielt"):
		if stelle == beimFragen {
			return nil, true, errGestolpert
		}

		return epochenzeilen(stelle), true, nil
	case strings.HasPrefix(geglaettet, "SELECT profil_id, coalesce(druck_id"):
		return sammlungszeilen(stelle), true, nil
	case strings.HasPrefix(geglaettet, "SELECT schluessel, zustand"):
		// **Zwei Stellen für zwei Abfragen.** `beimFragen` lässt die erste
		// scheitern; wer die zweite prüfen will, kommt dort nie an.
		if stelle == beimMengenfragen {
			return nil, true, errGestolpert
		}

		return mengenzeilen(stelle), true, nil
	}

	return nil, false, nil
}

// epochenzeilen gibt eine Epoche — oder scheitert unterwegs.
func epochenzeilen(stelle stolperstelle) driver.Rows {
	stempel := Zeitstempel(time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC))

	switch stelle {
	case beimRueckkehrende:
		return &stolperzeile{spalten: rueckkehrspalten}
	case beimRueckkehrscan:
		// **Eine Kennung, die NULL ist**, passt in kein `int64`.
		return &werteZeile{spalten: rueckkehrspalten, werte: []driver.Value{
			nil, stempel, stempel, "probe", "offen", "",
		}}
	}

	return &werteZeile{spalten: rueckkehrspalten, werte: []driver.Value{
		int64(1), stempel, stempel, "probe", "offen", "",
	}}
}

// mengenzeilen gibt eine unklare Menge — oder scheitert unterwegs.
func mengenzeilen(stelle stolperstelle) driver.Rows {
	switch stelle {
	case beimMengenende:
		return &stolperzeile{spalten: mengenspalten}
	case beimMengenscan:
		return &werteZeile{spalten: mengenspalten, werte: []driver.Value{
			nil, "nm", "en", false, int64(1), "", "", "",
		}}
	}

	return &werteZeile{spalten: mengenspalten, werte: []driver.Value{
		"scryfall:abc", "nm", "en", false, int64(1), "", "", "",
	}}
}

// sammlungszeilen gibt eine gruppierte Sammlungszeile — oder scheitert.
//
// **Nur für die zwei Stellen.** Wer hier immer antwortete, nähme jedem anderen
// Stolperlauf die Möglichkeit, an dieser Abfrage zu scheitern.
func sammlungszeilen(stelle stolperstelle) driver.Rows {
	if stelle == beimSammlungsende {
		return &stolperzeile{spalten: sammlungsstandspalten}
	}

	// **Ein NULL im Profil** passt in kein `int64`.
	return &werteZeile{spalten: sammlungsstandspalten, werte: []driver.Value{
		nil, int64(0), "scryfall:abc", "nm", "en", false, int64(1), "", "", "",
	}}
}

// zaehlendeAnweisung sagt, ob eine Anweisung aus null Zeilen einen Fehler
// macht — und damit vom Zählen abhängt.
//
// **Hier und nicht in `stolperwerk_test.go`**: die Aufzählung wuchs mit jedem
// Werkstück, und jene Datei ginge mit ihr über die Zeilengrenze aus §2. Was
// die Einträge gemeinsam haben: `UPDATE … WHERE <Bedingung>` oder `DELETE …`,
// bei denen null betroffene Zeilen ein `ErrNotFound` oder einen Konflikt
// bedeuten — und ob `RowsAffected` selbst scheitern kann, gibt kein echter
// Treiber her.
func zaehlendeAnweisung(abfrage string) bool {
	geglaettet := strings.TrimSpace(abfrage)

	for _, anfang := range []string{
		"UPDATE turnierteilnehmer",
		"DELETE FROM turniere",

		// **Die Kontakte aus dem Adressbuch** — ID-04 (#1703): Ersetzen und
		// Löschen gehen gegen die Fassung.
		"UPDATE kontakte",
		"DELETE FROM kontakte",

		// **Die zwei Wege am Ausgang** — Werkstück T5 (#1470): ein Versuch
		// wird gezählt, eine Bestätigung räumt weg, und beide machen aus
		// null Zeilen ein `ErrNotFound`.
		"UPDATE handelsausgang",
		"DELETE FROM handelsausgang",

		// **Die beiden Faktorwege** — Werkstück 7 (#1351): das Herausnehmen
		// eines Faktors und das Abstempeln eines Wiederherstellungscodes.
		"DELETE FROM nutzerfaktoren",
		"UPDATE wiederherstellungscodes",

		// **Das Einlösen eines Nachweises** — Werkstück O1 von Phase 5.2
		// (#1425): `UPDATE … WHERE eingeloest = 0`.
		"UPDATE freigabenachweise",

		// **Die Quittung einer Rückkehr** — Werkstück O7 von Phase 5.2
		// (#1431): `UPDATE … WHERE klaerung = 'offen'`.
		"UPDATE wiederherstellungen SET",

		// **Die zurückgesetzten Einstellungen** — Werkstück O4 von Phase 5.2
		// (#1428): die Zahl geht als Bilanz nach oben, und „null
		// zurückgesetzt" wäre dort eine Aussage.
		"DELETE FROM settings",
		"DELETE FROM nutzereinstellungen",

		// **Der Widerruf einer Sammlungsfreigabe** — Werkstück T3 von
		// Phase 5.6 (#1468): „widerrufen" ohne Wirkung sähe aus wie ein
		// Erfolg, und deshalb zählt dieser Weg seine Zeilen.
		"DELETE FROM sammlungsfreigaben",
	} {
		if strings.HasPrefix(geglaettet, anfang) {
			return true
		}
	}

	return false
}

// rueckkehrExec beantwortet die Schreibbefehle des Buchs.
func (s stolperstmt) rueckkehrExec() (driver.Result, bool, error) {
	geglaettet := strings.TrimSpace(s.abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "INSERT INTO wiederherstellungen"):
		switch s.stelle {
		case beimRueckkehrschreiben:
			return nil, true, errGestolpert
		case beimRueckkehrkennung:
			return kennungsfehler{}, true, nil
		}

		// **`druckergebnis` und nicht `driver.RowsAffected`**: jenes gibt auf
		// `LastInsertId` einen Fehler, und dann käme **jeder** Fall bis hier
		// und nicht weiter — der Lauf über das Festschreiben wäre grün, ohne
		// je ein `Commit` gesehen zu haben (§10).
		return druckergebnis{}, true, nil
	case strings.HasPrefix(geglaettet, "INSERT INTO wiederherstellungsmengen") &&
		s.stelle != beimSchreiben:
		// **`beimSchreiben` fällt durch** — dort scheitert das allgemeine
		// `Exec` dahinter, und das ist der Weg für „eine unklare Menge lässt
		// sich nicht eintragen".
		return druckergebnis{}, true, nil
	}

	return nil, false, nil
}

// werteZeile gibt genau eine Zeile mit festen Werten.
type werteZeile struct {
	spalten []string
	werte   []driver.Value
	gegeben bool
}

func (z *werteZeile) Columns() []string { return z.spalten }

func (z *werteZeile) Close() error { return nil }

func (z *werteZeile) Next(ziel []driver.Value) error {
	if z.gegeben {
		return io.EOF
	}

	z.gegeben = true
	copy(ziel, z.werte)

	return nil
}

// stolperzeile bricht **nach** der Schleife ab — ohne je eine Zeile zu geben.
type stolperzeile struct{ spalten []string }

func (z *stolperzeile) Columns() []string { return z.spalten }

func (z *stolperzeile) Close() error { return nil }

func (z *stolperzeile) Next([]driver.Value) error { return errGestolpert }

// **Jeder Weg des Buchs reicht seinen Fehler durch.**
func TestJederWegDesWiederherstellungsbuchsReichtSeinenFehlerDurch(t *testing.T) {
	vermerken := func(b *Wiederherstellungsbuch) error {
		_, fehler := b.Vermerken(probekontext(t), domain.Wiederherstellungsepoche{
			Quelle: "probe.msesicherung",
			Mengen: []domain.Sammlungsmenge{{Schluessel: "scryfall:abc", Menge: 1}},
		})

		return fehler
	}

	lesen := func(b *Wiederherstellungsbuch) error {
		_, fehler := b.Offene(probekontext(t))

		return fehler
	}

	for name, fall := range map[string]struct {
		treiber string
		lauf    func(*Wiederherstellungsbuch) error
	}{
		"die Transaktion beginnt nicht":                       {"stolper-vorgaben-beginn", vermerken},
		"das Eintragen der Rückkehr scheitert":                {"stolper-rueckkehr-schreiben", vermerken},
		"die Kennung lässt sich nicht lesen":                  {"stolper-rueckkehr-kennung", vermerken},
		"eine unklare Menge lässt sich nicht eintragen":       {"stolper-schreiben", vermerken},
		"das Festschreiben scheitert":                         {"stolper-commit", vermerken},
		"die Abfrage der Rückkehren scheitert":                {"stolper-fragen", lesen},
		"der Durchgang über die Rückkehren bricht ab":         {"stolper-rueckkehr-ende", lesen},
		"eine Rückkehr lässt sich nicht in ihre Felder legen": {"stolper-rueckkehr-scan", lesen},
		"der Durchgang über die Mengen bricht ab":             {"stolper-mengen-ende", lesen},
		"eine Menge lässt sich nicht in ihre Felder legen":    {"stolper-mengen-scan", lesen},
	} {
		t.Run(name, func(t *testing.T) {
			db, fehler := sql.Open(fall.treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // Probe

			if lauffehler := fall.lauf(NeuesWiederherstellungsbuch(db)); lauffehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Die zwei Wege, an denen das Lesen der Sammlung scheitert.**
//
// Ohne sie stünde `Bestandsstand` mit einer **leeren** Sammlung da und
// meldete `SammlungGelesen: true` — also „nichts unklar", wo nichts gelesen
// wurde. Das ist die eine Auskunft, die niemals falsch sein darf.
func TestEineUnlesbareSammlungGiltNichtAlsGelesen(t *testing.T) {
	for name, treiber := range map[string]string{
		"eine Zeile lässt sich nicht in ihre Felder legen": "stolper-sammlung-scan",
		"der Durchgang bricht ab":                          "stolper-sammlung-ende",
	} {
		t.Run(name, func(t *testing.T) {
			db, fehler := sql.Open(treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // Probe

			if stand := Bestandsstand(probekontext(t), db); stand.SammlungGelesen {
				t.Error("die unlesbare Sammlung gilt als gelesen")
			}
		})
	}
}

// **Und das Zählen der Quittung.** Ob es scheitern kann, gibt kein echter
// Treiber her — der Zweig dahinter macht aus null Zeilen ein `ErrNotFound`,
// und ohne ihn wäre eine Quittung auf eine Epoche, die es nicht gibt, ein
// stiller Erfolg.
func TestDasZaehlenDerQuittungReichtSeinenFehlerDurch(t *testing.T) {
	db, fehler := sql.Open("stolper-zaehlen", "egal")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // Probe

	if quittierfehler := NeuesWiederherstellungsbuch(db).Quittieren(
		probekontext(t), 1, time.Now()); quittierfehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}
