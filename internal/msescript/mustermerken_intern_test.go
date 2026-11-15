package msescript

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// **Ein Muster wird einmal übersetzt, nicht bei jedem Aufruf** —
// #2006.
//
// # Der Befund
//
// Gemessen am 2026-10-03 an der Ausfuhr von `Black.mse-set` (3.788 Karten): von
// den 56 Sekunden, die die Kartenskripte im Hauptfaden brauchten, gingen 45 in
// `regexp.Compile` (CPU-Profil, unter Last, also die Anteile und nicht die
// Sekunden). `replace@(match: …)` übersetzt sein Muster bei jedem
// Aufruf neu, und das Spiel ruft es über 142 Wertskripte je Karte tausendfach.
// Das Muster ist dasselbe, der Text ein anderer.
//
// MSE merkt sich ebenfalls nichts — `regex_from_script` übersetzt eine
// Zeichenkette bei jedem Aufruf und trägt die Randbemerkung `TODO: introduce
// some kind of caching?` (`src/script/functions/regex.cpp`, Zeilen 56–64,
// Fassung f9d9356 vom 2021-07-06). **adapted**: hier ist das Übersetzen eines
// Musters teurer als der Treffer auf einem Kartentext, und das Merken ändert
// kein Ergebnis.
//
// Die Tests stehen im Paket, weil sie in den Speicher hineinsehen.

func TestEinMusterWirdNurEinmalUebersetzt(t *testing.T) {
	musterspeicher.leeren()

	erstes, err := UebersetzeMuster(`a+b`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	zweites, err := UebersetzeMuster(`a+b`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	anderes, err := UebersetzeMuster(`a+c`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	if erstes != zweites {
		t.Error("dasselbe Muster wurde zweimal übersetzt")
	}

	if erstes == anderes {
		t.Error("zwei verschiedene Muster teilen sich eine Übersetzung")
	}
}

// **Die Frist gehört zur Übersetzung**: ein Muster der Rückfalllinie trägt sie
// als `MatchTimeout`, und wer sie ändert, bekommt kein Muster mit der alten.
func TestEinMusterMitAndererFristWirdNeuUebersetzt(t *testing.T) {
	musterspeicher.leeren()

	vorher := musterFrist

	t.Cleanup(func() { musterFrist = vorher })

	// Vorausblick, damit RE2 ablehnt und die Rückfalllinie zum Zug kommt.
	erstes, err := UebersetzeMuster(`(?=a)a+`)
	if err != nil || erstes.breit == nil {
		t.Fatalf("übersetzen: %v (breit: %v)", err, erstes != nil && erstes.breit != nil)
	}

	musterFrist = 5 * time.Millisecond

	zweites, err := UebersetzeMuster(`(?=a)a+`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	if zweites == erstes || zweites.breit.MatchTimeout != 5*time.Millisecond {
		t.Errorf("Frist %v, erwartet 5ms; dasselbe Muster: %v",
			zweites.breit.MatchTimeout, zweites == erstes)
	}
}

// **Ein Muster ohne Rückfalllinie hängt an keiner Frist.** Es bleibt gemerkt,
// auch wenn sie sich ändert.
func TestEinMusterOhneRueckverfolgungBleibtBeiAndererFristGemerkt(t *testing.T) {
	musterspeicher.leeren()

	vorher := musterFrist

	t.Cleanup(func() { musterFrist = vorher })

	erstes, err := UebersetzeMuster(`a+`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	musterFrist = 5 * time.Millisecond

	zweites, err := UebersetzeMuster(`a+`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	if zweites != erstes {
		t.Error("ein RE2-Muster wurde wegen der Frist neu übersetzt")
	}
}

// **Ein Muster, das keiner liest, wird nicht gemerkt** — und der Fehler kommt
// jedes Mal.
func TestEinUnbrauchbaresMusterWirdNichtGemerkt(t *testing.T) {
	musterspeicher.leeren()

	for range 2 {
		if _, err := UebersetzeMuster(`(`); err == nil {
			t.Fatal("kein Fehler")
		}
	}

	if musterspeicher.groesse() != 0 {
		t.Errorf("%d gemerkt, erwartet 0", musterspeicher.groesse())
	}
}

// **Der Speicher wächst nicht ohne Grenze**: ein Skript kann Muster aus Kartennamen
// bauen (`regex_escape(card.name)`), und ein Dienst läuft lange.
func TestDerMusterspeicherBleibtBegrenzt(t *testing.T) {
	musterspeicher.leeren()

	for i := range musterspeichergrenze + 10 {
		if _, err := UebersetzeMuster(fmt.Sprintf("karte%d", i)); err != nil {
			t.Fatalf("übersetzen: %v", err)
		}
	}

	if gemerkt := musterspeicher.groesse(); gemerkt > musterspeichergrenze {
		t.Errorf("%d gemerkt, die Grenze ist %d", gemerkt, musterspeichergrenze)
	}

	// Das Muster danach geht trotzdem — es wird neu übersetzt.
	if _, err := UebersetzeMuster("karte0"); err != nil {
		t.Errorf("nach dem Überlauf: %v", err)
	}
}

// **Parallele Aufrufe teilen den Speicher** — die Pipeline fährt diese Tests
// unter dem Detektor.
func TestDerMusterspeicherVertraegtParallelenZugriff(t *testing.T) {
	musterspeicher.leeren()

	var gruppe sync.WaitGroup

	for i := range 8 {
		gruppe.Go(func() {
			for j := range 50 {
				muster, err := UebersetzeMuster(fmt.Sprintf("x%d", (i+j)%10))
				if err != nil || muster == nil {
					t.Errorf("übersetzen: %v", err)
				}
			}
		})
	}

	gruppe.Wait()
}

// **Dieselbe Quelle gibt dieselbe RE2-Übersetzung** für die Auszeichnungsmuster,
// die `remove_tag` und seine Geschwister bauen.
func TestEinAusdruckWirdNurEinmalUebersetzt(t *testing.T) {
	musterspeicher.leeren()

	erster := gemerkterAusdruck(`</?i[^>]*>`)
	zweiter := gemerkterAusdruck(`</?i[^>]*>`)

	if erster != zweiter {
		t.Error("derselbe Ausdruck wurde zweimal übersetzt")
	}

	if gemerkterAusdruck(`</?b[^>]*>`) == erster {
		t.Error("zwei Ausdrücke teilen sich eine Übersetzung")
	}
}
