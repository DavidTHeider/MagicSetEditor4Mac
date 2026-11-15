package storage

// **Die Fehlerwege der Datenpflege** — Werkstück O8 von Phase 5.2
// (#1432).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Vier Wege gibt
// eine echte SQLite-Datenbank nicht her: eine Zeile, die sich nicht in ihr
// Feld legen lässt; ein Abbruch **mitten** im Durchgang; ein `DELETE`, das
// scheitert, obwohl das Lesen davor glückte; und ein `Commit`, das abbricht.
//
// **Warum sie geprüft gehören**: eine Analyse, die still eine leere Liste
// gibt, sagt „der Bestand ist sauber" — und eine Pflege, die still glückt,
// sagt „es ist weggeräumt". Beides sind Auskünfte, die niemals falsch sein
// dürfen.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Pflegeweg stolpern kann.
const (
	beimPflegescan stolperstelle = iota + 2000
	beimPflegeende
	beimPflegeloeschen
	beimUmfangscan
	beimUmfangende
	beimUmfangfragen
)

func init() {
	sql.Register("stolper-pflege-scan", stolpertreiber{stelle: beimPflegescan})
	sql.Register("stolper-pflege-ende", stolpertreiber{stelle: beimPflegeende})
	sql.Register("stolper-pflege-loeschen", stolpertreiber{stelle: beimPflegeloeschen})
	sql.Register("stolper-umfang-scan", stolpertreiber{stelle: beimUmfangscan})
	sql.Register("stolper-umfang-ende", stolpertreiber{stelle: beimUmfangende})
	sql.Register("stolper-umfang-fragen", stolpertreiber{stelle: beimUmfangfragen})
}

// pflegeabfrage beantwortet die Kennungsabfragen der Pflege — oder stolpert.
//
// **Ohne Fehlerergebnis**: die Wege, an denen die Abfrage **selbst** scheitert,
// deckt `beimFragen` ab; hier geht es um das, was danach kommt.
func pflegeabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool) {
	if !strings.Contains(abfrage, "FROM card_values") &&
		!strings.Contains(abfrage, "FROM set_values") {
		return nil, false
	}

	switch stelle {
	case beimPflegescan:
		// **Ein NULL in der Kennung** passt in keine Zeichenkette.
		return &werteZeile{spalten: []string{"kennung"}, werte: []driver.Value{nil}}, true
	case beimPflegeende:
		return &stolperzeile{spalten: []string{"kennung"}}, true

	// **`beimCommit` gehört dazu** — der Weg braucht eine Zeile, um überhaupt
	// bis zum Festschreiben zu kommen. Eine eigene Stelle dafür wäre eine
	// zweite Schreibweise für dieselbe Sache.
	case beimPflegeloeschen, beimCommit:
		return &werteZeile{
			spalten: []string{"kennung"}, werte: []driver.Value{"card_values:1:name"},
		}, true
	}

	return nil, false
}

// umfangabfrage beantwortet die zwei Abfragen einer Zählstelle — oder
// stolpert.
//
// **Die Zahl kommt immer durch**: ohne sie käme der Weg gar nicht bis zu den
// Namen, um die es hier geht (O4, #1428).
func umfangabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool) {
	geglaettet := strings.TrimSpace(abfrage)

	if stelle != beimUmfangscan && stelle != beimUmfangende &&
		stelle != beimUmfangfragen {
		return nil, false
	}

	switch {
	case strings.HasPrefix(geglaettet, "SELECT count(*) FROM sets"):
		return &werteZeile{spalten: []string{"anzahl"}, werte: []driver.Value{int64(1)}}, true
	case strings.HasPrefix(geglaettet, "SELECT name FROM sets"):
		// **`beimUmfangfragen` fällt durch** — dann scheitert die Abfrage
		// selbst, und das ist der Zweig **vor** dem Durchgang.
		if stelle == beimUmfangfragen {
			return nil, false
		}

		if stelle == beimUmfangende {
			return &stolperzeile{spalten: []string{"name"}}, true
		}

		// **Ein NULL im Namen** passt in keine Zeichenkette.
		return &werteZeile{spalten: []string{"name"}, werte: []driver.Value{nil}}, true
	}

	return nil, false
}

// pflegeExec beantwortet die Löschbefehle der Pflege.
func (s stolperstmt) pflegeExec() (driver.Result, bool, error) {
	if !strings.HasPrefix(strings.TrimSpace(s.abfrage), "DELETE FROM card_values") {
		return nil, false, nil
	}

	if s.stelle == beimPflegeloeschen {
		return nil, true, errGestolpert
	}

	return druckergebnis{}, true, nil
}

// **Jeder Weg der Pflege reicht seinen Fehler durch.**
func TestJederWegDerPflegeReichtSeinenFehlerDurch(t *testing.T) {
	analysieren := func(db *sql.DB) error {
		_, fehler := Pflegeanalyse(probekontext(t), db, time.Now())

		return fehler
	}

	for name, treiber := range map[string]string{
		"eine Zeile lässt sich nicht in ihr Feld legen": "stolper-pflege-scan",
		"der Durchgang bricht ab":                       "stolper-pflege-ende",
	} {
		t.Run(name, func(t *testing.T) {
			db, fehler := sql.Open(treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // Probe

			if lauffehler := analysieren(db); lauffehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Und beim Ausführen sind es drei Wege** — das Lesen in der Transaktion,
// das Löschen und das Festschreiben.
func TestJederSchrittDerAusfuehrungReichtSeinenFehlerDurch(t *testing.T) {
	for name, treiber := range map[string]string{
		"das Lesen in der Transaktion bricht ab": "stolper-pflege-ende",
		"das Löschen scheitert":                  "stolper-pflege-loeschen",
		"das Festschreiben scheitert":            "stolper-commit",
	} {
		t.Run(name, func(t *testing.T) {
			db, fehler := sql.Open(treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // Probe

			zeilen, lauffehler := PflegeAusfuehren(probekontext(t), db, time.Now(),
				map[string]string{"kartenwerteOhneKarte": revisionAus([]string{"card_values:1:name"})})
			if lauffehler != nil {
				t.Fatalf("ausführen: %v", lauffehler)
			}

			if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeGescheitert {
				t.Errorf("die Bilanz ist %+v", zeilen)
			}

			if zeilen[0].Grund == "" {
				t.Error("die gescheiterte Maßnahme nennt keinen Grund")
			}
		})
	}
}

// **Die zwei Wege, an denen das Benennen scheitert** (O4, #1428).
//
// Ohne sie stünde eine Gruppe mit ihrer Zahl und **ohne Namen** da — und „drei
// Sets" ohne ihre Namen ist die Auskunft, gegen die es diese Vorschau gibt.
func TestEinScheiterndesBenennenMeldetSich(t *testing.T) {
	for name, treiber := range map[string]string{
		"ein Name lässt sich nicht in sein Feld legen": "stolper-umfang-scan",
		"der Durchgang bricht ab":                      "stolper-umfang-ende",
		"die Abfrage der Namen scheitert":              "stolper-umfang-fragen",
	} {
		t.Run(name, func(t *testing.T) {
			db, fehler := sql.Open(treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // Probe

			if _, lauffehler := Profilumfang(probekontext(t), db, 1); lauffehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Und eine Abfrage, die selbst scheitert, ebenso.**
func TestEineGescheiterteZaehlungMeldetSich(t *testing.T) {
	db, fehler := sql.Open("stolper-fragen", "egal")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // Probe

	if _, lauffehler := Profilumfang(probekontext(t), db, 1); lauffehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}

// **Und das Zählen der zurückgesetzten Einstellungen** (O4, #1428).
//
// Ob `RowsAffected` scheitern kann, gibt kein echter Treiber her — und ohne
// diesen Weg stünde „null zurückgesetzt" da, wo etwas schiefging.
func TestDasZaehlenDerRuecksetzungReichtSeinenFehlerDurch(t *testing.T) {
	db, fehler := sql.Open("stolper-zaehlen", "egal")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // Probe

	if _, lauffehler := EinstellungenZuruecksetzen(
		probekontext(t), db, nil); lauffehler == nil {
		t.Error("das Zählen der Einstellungen meldete nichts")
	}

	if _, lauffehler := NutzereinstellungenZuruecksetzen(
		probekontext(t), db, 1); lauffehler == nil {
		t.Error("das Zählen der Nutzereinstellungen meldete nichts")
	}
}
