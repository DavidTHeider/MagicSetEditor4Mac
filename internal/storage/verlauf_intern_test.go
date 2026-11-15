// Die Wege des Verlaufs, die eine funktionierende Datenbank nicht hergibt.
//
// **Paketintern, weil sie am Stolpertreiber hängen** — derselbe Grund wie bei
// der Feldänderung: ein `Commit`, der scheitert, obwohl alles davor geglückt
// ist, kommt mit SQLite auf einer echten Platte nicht zustande. Im Betrieb
// schon: eine volle Platte, eine abgebrochene Verbindung.
package storage

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

// **Jeder Weg des Bewegens meldet sich.**
//
// Zurücknehmen und Wiederherstellen führen dieselbe Transaktion: Eintrag
// suchen, Weg gehen, Zeiger setzen, bestätigen. Wer einen davon ungeprüft
// ließe, hätte eine Zusage über eine Rücknahme, die im Ernstfall schweigt.
func TestDasBewegenMeldetJedenWeg(t *testing.T) {
	for name, treiber := range map[string]string{
		"beim Fragen":       "stolper-fragen",
		"beim Lesen":        "stolper-lesen",
		"beim Umwandeln":    "stolper-umwandeln",
		"beim Schreiben":    "stolper-verlauf-schreiben",
		"beim Bestätigen":   "stolper-verlauf-commit",
		"beim Zurückrollen": "stolper-zurueckrollen",
	} {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			leser := NeuerVerlaufsleser(db)

			// **Beide Richtungen**, denn sie nehmen verschiedene Abfragen.
			_, _, zurueck := leser.Zuruecknehmen(probekontext(t), 1)
			_, _, vor := leser.Wiederherstellen(probekontext(t), 1)

			if zurueck == nil && vor == nil {
				t.Errorf("%s: kein Fehler in beiden Richtungen", treiber)
			}
		})
	}
}

// **Der zweite Schreibweg einer Verlaufsbewegung.**
//
// Zurücknehmen schreibt zweimal: den Weg in den Bestand und den Zeiger in den
// Verlauf. Wer nur den ersten prüft, hat den zweiten nie betreten.
func TestDasBewegenMeldetAuchDenZweitenSchreibweg(t *testing.T) {
	// **An der Anweisung erkannt, nicht gezählt** (#1036). Hier stand
	// `stolper-zweites-schreiben` — „der zweite Schreibvorgang" war der
	// Zeiger. Seit die Rücknahme auch die Herkunft führt, sind es drei.
	db, err := sql.Open("stolper-verlauf-zeiger", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, _, fehler := NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), 1); fehler == nil {
		t.Error("kein Fehler")
	}
}

// Und das bloße Lesen des Verlaufs ebenso.
func TestDasVerlaufslesenMeldetJedenWeg(t *testing.T) {
	for name, treiber := range map[string]string{
		"beim Fragen":    "stolper-fragen",
		"beim Lesen":     "stolper-lesen",
		"beim Umwandeln": "stolper-umwandeln",
	} {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeuerVerlaufsleser(db).Verlauf(probekontext(t), 1); fehler == nil {
				t.Errorf("%s: kein Fehler", treiber)
			}
		})
	}
}

// **Der Löschweg einer Rücknahme meldet sich ebenso.**
//
// Gab es das Feld vorher nicht, wird es entfernt statt geleert (0010) — und
// auch dieser Weg kann scheitern.
func TestDasBewegenMeldetDenLoeschweg(t *testing.T) {
	db, err := sql.Open("stolper-verlauf-schreiben", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	verlaufOhneFeld = true
	t.Cleanup(func() { verlaufOhneFeld = false })

	if _, _, fehler := NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), 1); fehler == nil {
		t.Error("kein Fehler")
	}
}

// stolperfall beschreibt einen Stolperweg einer Feldänderung.
//
// **Die Meldung gehört zum Fall.** Ein Test, der nur `!= nil` verlangt, ist
// grün, sobald *irgendwo* etwas scheitert — und das war er hier: drei von fünf
// Fällen kamen nie über die Vorprüfung hinaus und prüften den Weg nicht, den ihr
// Name nennt. Wer die Stelle nennt, an der es scheitern soll, merkt das.
type stolperfall struct {
	name     string
	treiber  string
	meldung  string
	stolpert bool
}

// pruefeStolperfall führt beide Feldänderungen und hält den Fehler an seiner
// Meldung fest.
//
// **Beide Richtungen**, denn `FeldEntfernen` führt dieselbe Transaktion mit
// einer anderen Änderung — derselbe Verlaufscode, ein anderer Aufruf. Der Zähler
// wird vor **jedem** Aufruf zurückgesetzt: sonst stolpert der zweite an einer
// früheren Stelle als der erste und prüft etwas anderes.
func pruefeStolperfall(t *testing.T, fall stolperfall) {
	t.Helper()

	db, err := sql.Open(fall.treiber, "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	schreiber := NeuerKartenschreiber(db)

	zweiterSchreibzaehler = 0
	_, _, gesetzt := schreiber.FeldSetzenMitVorwert(probekontext(t), 1, "name", "x")

	zweiterSchreibzaehler = 0
	_, _, _, entfernt := schreiber.FeldEntfernen(probekontext(t), 1, "name")

	for weg, fehler := range map[string]error{"setzen": gesetzt, "entfernen": entfernt} {
		if fehler == nil {
			t.Errorf("%s: kein Fehler beim %s", fall.treiber, weg)

			continue
		}

		if !strings.Contains(fehler.Error(), fall.meldung) {
			t.Errorf("%s beim %s: %v — erwartet war %q", fall.treiber, weg, fehler, fall.meldung)
		}

		if fall.stolpert && !errors.Is(fehler, errGestolpert) {
			t.Errorf("%s beim %s: %v trägt den Treiberfehler nicht", fall.treiber, weg, fehler)
		}
	}
}

// **Die Wege beim Schreiben des Verlaufs.**
//
// `schreibeVerlauf` schreibt zweimal: den Eintrag und die Kürzung auf
// `GrenzeVerlauf`. Dazu liest die Feldänderung davor den neuen Wert. Drei Wege,
// die eine funktionierende Datenbank nicht hergibt.
func TestDasVerlaufsschreibenMeldetJedenWeg(t *testing.T) {
	for _, fall := range []stolperfall{
		{
			"beim Lesen des neuen Werts", "stolper-verlauf-lesefehler",
			`den neuen Wert von "name" lesen`, true,
		},
		{
			"beim Schreiben des Eintrags", "stolper-verlauf-eintrag",
			`den Verlauf zu Feld "name" schreiben`, true,
		},
		{"beim Kürzen", "stolper-verlauf-kuerzen", "den Verlauf kürzen", true},
	} {
		t.Run(fall.name, func(t *testing.T) { pruefeStolperfall(t, fall) })
	}
}

// **Die Vorprüfung der Feldänderung meldet ebenso jeden Weg.**
//
// Sie fragt in **einer** Abfrage, ob es die Karte gibt und was im Feld stand.
// Diese Fälle standen bis heute unter dem Verlaufsschreiben und trugen dessen
// Namen — sie erreichen ihn nie, weil die Vorprüfung davor liegt. Hier stehen
// sie richtig und nennen die Stelle, an der sie wirklich scheitern.
func TestDieVorpruefungDerFeldaenderungMeldetJedenWeg(t *testing.T) {
	for _, fall := range []stolperfall{
		{"beim Fragen", "stolper-fragen", `den vorigen Wert von "name" lesen`, true},
		{"beim Lesen", "stolper-lesen", `den vorigen Wert von "name" lesen`, true},
		// Der Treiber gibt acht Spalten auf eine Abfrage mit zweien; der Scan
		// scheitert an der Anzahl, nicht am Treiberfehler.
		{"beim Umwandeln", "stolper-umwandeln", `den vorigen Wert von "name" lesen`, false},
	} {
		t.Run(fall.name, func(t *testing.T) { pruefeStolperfall(t, fall) })
	}
}
