package biometrie_test

// **Was ohne Finger auf dem Sensor prüfbar ist** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Was diese Läufe **nicht** belegen
//
// **Dass ein Fingerabdruck erkannt wird.** Ein Lauf, der
// [biometrie.Bestaetige] wirklich riefe, bliebe stehen, bis jemand den Sensor
// berührt — und geprüft wäre, dass macOS Touch ID kann. Das steht in
// `.covignore` mit derselben Begründung wie der Dateidialog.
//
// **Geprüft wird, was daneben liegt** und was im Betrieb wirklich schiefgeht:
// dass die Auskunft über die Verfügbarkeit ankommt, dass eine Plattform ohne
// Biometrie eine **Antwort** gibt statt eines Absturzes, und dass ein
// Aussetzer der Bindung die Anwendung nicht mitnimmt.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/biometrie"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Auskunft über die Verfügbarkeit kommt an — mit einem Grund.**
//
// **Ohne Zusage darüber, wie sie ausfällt**: auf einem Mac mit Touch ID ist
// sie wahr, auf einem MacBook im Clamshell-Betrieb nicht, und unter Linux
// nie. Ein Test, der `true` erwartete, prüfte den Rechner des Prüfenden.
//
// **Nicht parallel**: er ruft die **echte** Bindung, und die Läufe in
// `wege_darwin_intern_test.go` tauschen sie für ihre Fälle aus. Ein
// paralleler Lauf läse dann ein Doppel und hieße trotzdem „Verfügbarkeit".
func TestDieAuskunftUeberDieVerfuegbarkeitKommtAn(t *testing.T) {
	stand := biometrie.Stand(t.Context())

	// **Vorhanden und Grund gehören zusammen**: „geht nicht" ohne Grund wäre
	// eine Meldung, mit der niemand etwas anfangen kann (§8).
	if !stand.Vorhanden && stand.Grund == "" {
		t.Error("nicht vorhanden und kein Grund genannt")
	}

	if stand.Vorhanden && stand.Grund != "" {
		t.Errorf("vorhanden und trotzdem ein Grund: %q", stand.Grund)
	}
}

// **Ein Aussetzer der Bindung nimmt die Anwendung nicht mit.**
//
// # Warum das eigens geprüft wird
//
// **Weil die Bindung vor 1.0 ist und ihre eigene Abhängigkeit im
// Alpha-Stand** (`purego v0.11.0-alpha.6`, gemessen am 2026-09-05; seit dem
// 2026-10-11 die Endfassung v0.11.1, weiter vor 1.0 — #2305). Ein
// `panic` aus einer Fremdbibliothek risse ohne diesen Riegel den ganzen
// Prozess mit — und mit ihm das Fenster, an dem jemand gerade arbeitet.
//
// **§16, Regel 5, sinngemäß**: fällt der zweite Weg aus, fällt der zweite Weg
// aus und nicht die Anwendung.
func TestEinAussetzerDerBindungNimmtNichtsMit(t *testing.T) {
	t.Parallel()

	fehler := biometrie.Gefangen(func() error { panic("die Bindung ist weg") })

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("durchgelassen wurde: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "Biometrie") {
		t.Errorf("die Meldung sagt nicht, was ausgefallen ist: %v", fehler)
	}
}

// **Und ein gewöhnlicher Fehler geht unverändert durch.**
//
// **Der Riegel fängt Abstürze, nicht Antworten**: „der Finger passte nicht"
// ist eine Antwort und soll als solche ankommen.
func TestEinGewoehnlicherFehlerGehtDurch(t *testing.T) {
	t.Parallel()

	eigener := errors.New("der Finger passte nicht")

	if fehler := biometrie.Gefangen(func() error { return eigener }); !errors.Is(fehler, eigener) {
		t.Errorf("angekommen ist: %v", fehler)
	}
}

// **Und ein glücklicher Lauf gibt nichts zurück.**
func TestEinGlueckderLaufGibtNichts(t *testing.T) {
	t.Parallel()

	if fehler := biometrie.Gefangen(func() error { return nil }); fehler != nil {
		t.Errorf("angekommen ist: %v", fehler)
	}
}
