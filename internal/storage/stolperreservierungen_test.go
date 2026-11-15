package storage

// **Die Fehlerwege der Reservierung, die kein echter Treiber hergibt** —
// Werkstück `T2` von Phase 5.6
// (#1467).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen — dieselbe Form
// wie `stolperleihen_test.go`, und eine eigene Datei aus demselben Grund (§2).
//
// # Warum es dafür einen Haken braucht
//
// **Weil die vorhandenen Treiber nur eine Spalte geben.** `Festhalten` liest
// drei (Menge, Fassung, Wunsch) und sechs (die Zusagen); unter
// `stolper-schreiben` scheiterte schon die **erste** Abfrage, und das
// Einfügen wurde nie erreicht. Mit dem Haken hier antworten beide Abfragen
// brauchbar — und die Fehler fallen dort an, wo sie gemeint sind.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen die Zusagen stolpern können.
const (
	beimZusagenscan stolperstelle = iota + 2300
	beimZusagenende
	beimBesitzscan
	beimZusagenzaehlen

	// **Der Zähler über alle Zusagen** (`ReserviertJeExemplar`) — er liest
	// eine andere Spaltenreihe als die zwei darüber und stolpert deshalb an
	// eigenen Stellen.
	beimZaehlerscan
	beimZaehlerfrist
	beimZaehlerende

	// **Was ein Vorgang von einer Variante schon hält** (#2492) — eine
	// eigene Abfrage mit eigenen Stellen.
	beimGebundenFragen
	beimGebundenScan
	beimGebundenEnde
)

func init() {
	sql.Register("stolper-zusagen-scan", stolpertreiber{stelle: beimZusagenscan})
	sql.Register("stolper-zusagen-ende", stolpertreiber{stelle: beimZusagenende})
	sql.Register("stolper-besitz-scan", stolpertreiber{stelle: beimBesitzscan})
	sql.Register("stolper-zusagen-zaehlen",
		stolpertreiber{stelle: beimZusagenzaehlen})
	sql.Register("stolper-zaehler-scan", stolpertreiber{stelle: beimZaehlerscan})
	sql.Register("stolper-zaehler-frist", stolpertreiber{stelle: beimZaehlerfrist})
	sql.Register("stolper-zaehler-ende", stolpertreiber{stelle: beimZaehlerende})
	sql.Register("stolper-gebunden-fragen", stolpertreiber{stelle: beimGebundenFragen})
	sql.Register("stolper-gebunden-scan", stolpertreiber{stelle: beimGebundenScan})
	sql.Register("stolper-gebunden-ende", stolpertreiber{stelle: beimGebundenEnde})
}

// stolperexemplar ist die Besitzzeile, die der Treiber nennt — und
// stolpervertrag ein Vertrag, den sie erfüllt (#2492): sonst endete jeder
// Lauf an der Prüfung des Vertrags statt an seiner Stelle.
var stolpervertrag = domain.Zusagevertrag{{
	Printing: "p-stolper", Menge: 9, Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
}}

// reservierungsabfrage beantwortet die zwei Abfragen der Reservierung.
//
// **Mit Werten und nicht mit einer leeren Zeile**: `ReservierungPasst` rechnet
// gegen Menge und Fassung, und eine Zeile aus Nullen ließe jede Zusage an der
// Prüfung scheitern statt am Schreiben.
func reservierungsabfrage(
	stelle stolperstelle, abfrage string,
) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "SELECT menge, fassung, wunsch"):
		spalten := []string{
			"menge", "fassung", "wunsch", "printing", "druck_id", "zustand",
			"sprache", "finish",
		}

		// **Ein Abbruch und kein „keine Zeile"**: „gibt es nicht" ist der
		// gute Fall und hat seinen eigenen Satz; hier geht es um den
		// **anderen** — die Zeile ist da und lässt sich nicht lesen.
		if stelle == beimBesitzscan {
			return &abbrechendeZeile{spalten: spalten}, true, nil
		}

		return &werteZeile{
			spalten: spalten,
			werte: []driver.Value{
				int64(3), int64(0), int64(0), "p-stolper", int64(0), "nearMint", "de", "nonfoil",
			},
		}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT r.menge, s.printing"):
		spalten := []string{"menge", "printing", "druck_id", "zustand", "sprache", "finish"}

		switch stelle {
		case beimGebundenFragen:
			return nil, true, errGestolpert
		case beimGebundenScan:
			return &krummeZeile{spalten: spalten}, true, nil
		case beimGebundenEnde:
			return &abbrechendeZeile{spalten: spalten}, true, nil
		}

		return &leerezeile{spalten: len(spalten)}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT COUNT(*) FROM reservierungen"):
		// **Null übergebene Zusagen**, damit `Freigeben` bis zum Löschen
		// kommt. Ohne diese Zeile scheiterte es schon am Zählen, und der
		// Zweig darunter bliebe unbetreten.
		if stelle == beimZusagenzaehlen {
			return &abbrechendeZeile{spalten: []string{"anzahl"}}, true, nil
		}

		return &werteZeile{
			spalten: []string{"anzahl"},
			werte:   []driver.Value{int64(0)},
		}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT exemplar_id, menge, frist"):
		// **`beimFragen` geht durch, nicht hier hinein.** Die allgemeine
		// Stelle lässt die Abfrage selbst scheitern; wer sie hier abfinge,
		// gäbe eine leere Antwort zurück — und der Zweig, der einen
		// Abfragefehler meldet, bliebe unbetreten.
		if stelle == beimFragen {
			return nil, false, nil
		}

		spalten := []string{"exemplar_id", "menge", "frist", "uebergeben"}

		switch stelle {
		case beimZaehlerscan:
			return &krummeZeile{spalten: spalten}, true, nil

		case beimZaehlerende:
			return &abbrechendeZeile{spalten: spalten}, true, nil

		// **Eine Frist, die keine ist.** Sie kommt aus der Datei und nicht
		// aus diesem Programm; wer sie für einen Nullwert hielte, machte aus
		// einer kaputten Zeile eine ohne Frist — also eine, die **ewig**
		// festhält (§8).
		case beimZaehlerfrist:
			return &werteZeile{
				spalten: spalten,
				werte:   []driver.Value{int64(1), int64(1), "vorgestern", int64(0)},
			}, true, nil
		}

		return &leerezeile{spalten: len(spalten)}, true, nil

	case strings.HasPrefix(geglaettet, "SELECT vorgang, exemplar_id, revision"):
		spalten := []string{
			"vorgang", "exemplar_id", "revision", "menge", "fassung",
			"frist", "uebergeben",
		}

		switch stelle {
		case beimZusagenscan:
			return &krummeZeile{spalten: spalten}, true, nil

		case beimZusagenende:
			return &abbrechendeZeile{spalten: spalten}, true, nil
		}

		return &leerezeile{spalten: len(spalten)}, true, nil
	}

	return nil, false, nil
}

// **Jeder Weg der Reservierung reicht seinen Fehler durch.**
//
// **Vier Stellen, und jede trifft eine andere**: die zweite Abfrage bricht im
// Scan ab, sie bricht nach der Schleife ab, das Einfügen scheitert, das
// Commit scheitert. Ohne diesen Treiber stünden die Zweige daneben ungeprüft
// da und sähen aus wie geprüfte Zusagen (§10).
func TestJederReservierungswegReichtSeinenFehlerDurch(t *testing.T) {
	for _, treiber := range []string{
		"stolper-zusagen-scan", "stolper-zusagen-ende", "stolper-besitz-scan",
		"stolper-gebunden-fragen", "stolper-gebunden-scan", "stolper-gebunden-ende",
		"stolper-schreiben", "stolper-commit",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			ablage, ctx := NeueReservierungsablage(db), probekontext(t)

			if fehler := ablage.Festhalten(ctx, domain.Reservierung{
				Vorgang: "h1", Revision: 1, ExemplarID: 1, Menge: 1,
				Frist: time.Now().Add(time.Hour),
			}, stolpervertrag); fehler == nil {
				t.Error("das Festhalten kam ohne Fehler zurück")
			}
		})
	}
}

// **Und das Lösen einer Zusage ebenso.**
//
// **Ein eigener Lauf, weil `Freigeben` einen anderen Weg geht**: erst zählen,
// dann löschen. Wer beides in einen Lauf legte, bekäme bei einem roten
// Ergebnis zwei Verdächtige.
func TestDasLoesenEinerZusageReichtSeinenFehlerDurch(t *testing.T) {
	for _, treiber := range []string{
		"stolper-zusagen-zaehlen", "stolper-schreiben",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if fehler := NeueReservierungsablage(db).Freigeben(
				probekontext(t), "h1"); fehler == nil {
				t.Error("das Freigeben kam ohne Fehler zurück")
			}
		})
	}
}
