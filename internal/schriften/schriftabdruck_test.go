package schriften_test

// **Der Abdruck einer Schrift** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass eine ausgetauschte Schriftdatei auffällt.** Ein Zeichenplan trägt
// Namen; ohne diesen Abdruck hätte ein Schlüssel über ihm keine Ahnung, dass
// die Datei dahinter eine andere ist — der Text sähe anders aus, und der
// Spieltisch zeigte das Bild von vorhin.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// **Dieselbe Datei gibt denselben Abdruck.**
func TestDieselbeSchriftGibtDenselbenAbdruck(t *testing.T) {
	suche := schriften.Neue(nil).MitPaketen(paketMitSchrift(t, "mplantin.ttf"))

	erster := suche.Abdruck("MPlantin", false)
	zweiter := suche.Abdruck("MPlantin", false)

	if erster != zweiter {
		t.Errorf("zwei Abdrücke: %q und %q", erster, zweiter)
	}

	if len(erster) != 64 || strings.HasPrefix(erster, "fehlt:") {
		t.Errorf("der Abdruck %q sieht nicht aus wie einer", erster)
	}
}

// **Eine Schrift, die es nicht gibt, ist ein Wert und kein Fehler.**
//
// **Sonst verschwände die Karte ganz.** Der Rasterer zeichnet mit dem Ersatz
// und sagt es; dieses Bild soll wiedererkannt werden.
func TestEineFehlendeSchriftIstEinWert(t *testing.T) {
	suche := schriften.Neue(nil)

	abdruck := suche.Abdruck("GibtEsNichtSicher", false)

	if !strings.HasPrefix(abdruck, "fehlt:") {
		t.Errorf("der Abdruck %q sagt nicht, dass die Schrift fehlt", abdruck)
	}

	// **Und er ist ein anderer als der einer vorhandenen** — taucht sie auf,
	// ändert sich der Schlüssel.
	vorhanden := schriften.Neue(nil).
		MitPaketen(paketMitSchrift(t, "mplantin.ttf")).
		Abdruck("MPlantin", false)

	if abdruck == vorhanden {
		t.Error("fehlend und vorhanden geben denselben Abdruck")
	}
}

// **Aufrecht und fett werden getrennt gefragt.**
//
// **Weil es zwei Dateien sein können** — und weil beide zeichnen. Ein
// gemeinsamer Abdruck ließe einen ausgetauschten fetten Schnitt durch.
func TestAufrechtUndFettWerdenGetrenntAbgedruckt(t *testing.T) {
	suche := schriften.Neue(nil).MitPaketen(paketMitSchrift(t, "mplantin.ttf"))

	aufrecht := suche.Abdruck("MPlantin", false)
	fett := suche.Abdruck("MPlantin", true)

	// **Sie dürfen gleich sein**, wenn dieselbe Datei beide Schnitte trägt —
	// gemessen wird, dass **gefragt** wurde und eine Antwort kam.
	if aufrecht == "" || fett == "" {
		t.Errorf("aufrecht %q, fett %q", aufrecht, fett)
	}
}
