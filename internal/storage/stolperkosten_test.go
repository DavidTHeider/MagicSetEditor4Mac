package storage

// **Die Abbrüche mitten im Lesen des Kostenjournals** — Werkstück O12 von
// Phase 5.2 (#1439).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Einen Weg gibt
// eine echte SQLite-Datenbank nicht her: einen Durchgang, der **mitten** in der
// Zeilenschleife abbricht. `kostenfehler_test.go` deckt alles, was eine echte
// Datei kann — eine Tafel, die es nicht mehr gibt, eine Spalte, in die eine
// Zahl nicht passt, ein geschlossener Bestand.
//
// **Warum das geprüft gehört**: eine Summe, die still bei dem stehen bleibt,
// was sie bis zum Abbruch gezählt hat, sagt „so viel ist gebunden" — und das
// ist zu wenig. Aus einer halben Summe wird eine Freigabe, die es nicht gibt.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen das Kostenjournal stolpert.
const (
	// beimKostenende bricht **mitten** im Durchgang ab.
	beimKostenende stolperstelle = iota + 2100

	// beimKostenzaehlen lässt die **Summe** scheitern, während das Lesen
	// eines Postens glückt.
	//
	// **Eine echte Datei gibt das nicht her**: beide Abfragen gehen auf
	// dieselbe Tafel, und die eine liest eine Obermenge der Spalten der
	// anderen — wer eine Spalte wegnimmt, bricht zuerst das Lesen.
	beimKostenzaehlen
)

func init() {
	sql.Register("stolper-kosten-ende", stolpertreiber{stelle: beimKostenende})
	sql.Register("stolper-kosten-zaehlen", stolpertreiber{stelle: beimKostenzaehlen})
}

// kostenabfrage beantwortet die Abfragen des Journals — oder bricht ab.
//
// **Die Budgetzeile muss kommen**, sonst hielte schon die Grenze den Weg an,
// und die Schleife dahinter bliebe unbetreten (§10: eine Gegenprobe, die
// nichts ändert, ist keine).
func kostenabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	if stelle != beimKostenende && stelle != beimKostenzaehlen {
		return nil, false, nil
	}

	if strings.Contains(abfrage, "FROM kostenbudgets") {
		return &werteZeile{
			spalten: []string{"betrag", "waehrung"},
			werte:   []driver.Value{int64(100_000), "EUR"},
		}, true, nil
	}

	if !strings.Contains(abfrage, "FROM kostenposten") {
		return nil, false, nil
	}

	// **Die Suche nach der Identität gibt nichts** — der Auftrag ist neu.
	//
	// **Erkannt am Ende und nicht am `WHERE`**: seit Wanderung 0082 sucht sie
	// im Profil (`WHERE profil_id = ? AND kennung = ?`), und die Vorrichtung
	// hielt die Suche sonst für die Summe — beide Wege dahinter blieben
	// unbetreten, gefunden von der Abdeckung (Nachlauf über Phase 6.1).
	if strings.Contains(abfrage, "kennung = ?") {
		return &werteZeile{spalten: []string{"kennung"}, gegeben: true}, true, nil
	}

	if stelle == beimKostenzaehlen {
		return nil, true, errGestolpert
	}

	return &stolperzeile{spalten: []string{"hoechstbetrag"}}, true, nil
}

// **Jeder Durchgang durch die Posten meldet seinen Abbruch.**
func TestJederDurchgangDurchDieKostenMeldetSeinenAbbruch(t *testing.T) {
	db, fehler := sql.Open("stolper-kosten-ende", "")
	if fehler != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	pruefeAbbrueche(t, db)
}

// **Und die Summe scheitert für sich, während das Lesen glückt.**
func TestDieKostensummeScheitertFuerSich(t *testing.T) {
	db, fehler := sql.Open("stolper-kosten-zaehlen", "")
	if fehler != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	journal := NeuesKostenjournal(db)

	if _, fehler := journal.Reservieren(probekontext(t), domain.Kostenposten{
		Kennung: "eins", ProfilID: 1, Konto: "anbieter", Aufgabe: "karte.text",
		Hoechstbetrag: domain.Betrag{Wert: 100, Waehrung: "EUR"},
	}, domain.ZeitraumTag, time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)); fehler == nil {
		t.Error("die gescheiterte Summe kam durch")
	}
}

// pruefeAbbrueche fährt die drei Wege gegen einen Treiber, der mitten im
// Durchgang abbricht.
func pruefeAbbrueche(t *testing.T, db *sql.DB) {
	t.Helper()

	journal := NeuesKostenjournal(db)
	jetzt := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

	for name, wege := range map[string]func() error{
		"die Reservierung": func() error {
			_, fehler := journal.Reservieren(probekontext(t), domain.Kostenposten{
				Kennung: "eins", ProfilID: 1, Konto: "anbieter", Aufgabe: "karte.text",
				Hoechstbetrag: domain.Betrag{Wert: 100, Waehrung: "EUR"},
			}, domain.ZeitraumTag, jetzt)

			return fehler
		},
		"der Stand": func() error {
			_, fehler := journal.Stand(probekontext(t), 1, "anbieter",
				domain.ZeitraumTag, jetzt)

			return fehler
		},
		"die Posten": func() error {
			_, fehler := journal.Posten(probekontext(t), 1, "anbieter",
				domain.ZeitraumTag, jetzt, 5)

			return fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := wege(); fehler == nil {
				t.Error("der Abbruch ging verloren")
			}
		})
	}
}
