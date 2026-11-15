//go:build fremdpakete

// Was ein Messlauf nicht messen konnte — und warum
// (#616).
//
// **Eine eigene Datei, weil es eine Frage für alle Läufe ist.** Sie stand in
// `referenzfehler_intern_test.go` und galt damit für einen; angestrichen hat
// das der Nachlauf zum Zweitblick über Phase K — nicht den Fall behandeln,
// sondern die Klasse (§10).
package app

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
)

// ausfaelle zählt, welche Karten der Lauf nicht messen konnte.
//
// **Ohne diese Zählung ist der Lauf ein halber Stand** (§10, angestrichen vom
// Zweitblick über Phase K): jede Karte, die an einem fehlenden Namen, einer
// fehlenden Referenz oder einem Vorschaufehler ausfällt, verschwand
// wortlos — und gerade eine Verschlechterung im Renderpfad kann Karten aus
// der Messmenge nehmen und die Zahl damit **verbessern**.
type ausfaelle map[string]*ausfallgrund

// ausfallgrund zählt einen Grund und merkt sich die ersten Namen dazu.
//
// **Die Zahl allein sagt nicht, wo man nachsehen soll** (#616). Wer wissen
// will, ob dieselbe Karte immer wieder ausfällt oder jedes Mal eine andere,
// braucht Namen — aber nicht alle: eine Liste über 25.148 Karten läse
// niemand.
type ausfallgrund struct {
	anzahl int
	namen  []string
}

// GenannteAusfaelle ist die Zahl der Namen, die je Grund dastehen.
//
// **Fünf, weil der Zweck das Nachsehen ist**, nicht die Buchführung: wer eine
// Klasse verstehen will, braucht ein paar Beispiele, keine Aufzählung.
const GenannteAusfaelle = 5

// unsereGruende sind die Ausfälle, die **dieses Programm** zu verantworten
// hat — sie machen den Lauf rot.
//
// **Ein fehlendes Referenzbild gehört nicht dazu.** Die Referenz ist MSE2s
// eigene Ausgabe, und dass ein Kartenname mit Apostroph dort nicht als Datei
// erscheint, ist ihre Eigenheit
// (#535) —
// gemeldet wird es trotzdem, denn eine stille Auslassung bleibt eine.
var unsereGruende = map[string]bool{
	"die Vorschau scheiterte":                              true,
	"der Abstand ließ sich nicht rechnen":                  true,
	"der Abstand des Symbolfeldes ließ sich nicht rechnen": true,
}

// nimm zählt einen Ausfall — mit dem Namen der Karte, wenn einer dasteht.
func (a ausfaelle) nimm(grund string, namen ...string) {
	eintrag, da := a[grund]
	if !da {
		eintrag = &ausfallgrund{}
		a[grund] = eintrag
	}

	eintrag.anzahl++

	if len(namen) > 0 && namen[0] != "" && len(eintrag.namen) < GenannteAusfaelle {
		eintrag.namen = append(eintrag.namen, namen[0])
	}
}

func (a ausfaelle) melde(t *testing.T, gemessen int) {
	t.Helper()

	unsere, fremde := 0, 0

	for grund, eintrag := range a {
		if unsereGruende[grund] {
			unsere += eintrag.anzahl
		} else {
			fremde += eintrag.anzahl
		}
	}

	if unsere+fremde == 0 {
		t.Logf("  keine Karte ausgefallen")

		return
	}

	t.Logf("  %d Karten ausgefallen bei %d gemessenen", unsere+fremde, gemessen)

	for _, grund := range slices.Sorted(maps.Keys(a)) {
		eintrag := a[grund]

		if len(eintrag.namen) == 0 {
			t.Logf("    %4d × %s", eintrag.anzahl, grund)

			continue
		}

		t.Logf("    %4d × %s — etwa %s", eintrag.anzahl, grund,
			strings.Join(eintrag.namen, ", "))
	}

	// **Rot wird der Lauf nur bei unseren Ausfällen.** Ein fehlendes
	// Referenzbild ist eine Eigenschaft der Referenz; eine gescheiterte
	// Vorschau ist ein Fehler dieses Programms — und genau der könnte Karten
	// aus der Messmenge nehmen und die Zahl damit **verbessern**.
	if unsere > 0 {
		t.Errorf("%d der %d Ausfälle gehen auf dieses Programm — die Zahl "+
			"spricht dann über eine Menge, die eine Verschlechterung "+
			"verkleinert hat", unsere, unsere+fremde)
	}
}

// **Ein Grund zählt und merkt sich die ersten Namen**
// (#616).
//
// **Und nicht alle**: eine Aufzählung über 25.148 Karten läse niemand. Wer
// eine Klasse verstehen will, braucht ein paar Beispiele.
func TestEinAusfallgrundZaehltUndNenntBeispiele(t *testing.T) {
	verluste := ausfaelle{}

	for i := range GenannteAusfaelle + 3 {
		verluste.nimm("kein Referenzbild", fmt.Sprintf("Karte %d", i))
	}

	// **Ein Ausfall ohne Namen zählt trotzdem** — „die Karte hat keinen
	// Namen" ist selbst ein Grund.
	verluste.nimm("die Karte hat keinen Namen")

	referenz := verluste["kein Referenzbild"]

	if referenz.anzahl != GenannteAusfaelle+3 {
		t.Errorf("gezählt %d, erwartet %d", referenz.anzahl, GenannteAusfaelle+3)
	}

	if len(referenz.namen) != GenannteAusfaelle {
		t.Errorf("%d Namen, erwartet %d: %v", len(referenz.namen),
			GenannteAusfaelle, referenz.namen)
	}

	if referenz.namen[0] != "Karte 0" {
		t.Errorf("erster Name %q, erwartet `Karte 0`", referenz.namen[0])
	}

	if ohne := verluste["die Karte hat keinen Namen"]; ohne.anzahl != 1 || len(ohne.namen) != 0 {
		t.Errorf("%+v, erwartet einen Ausfall ohne Namen", ohne)
	}
}

// **Ein leerer Name wird nicht mitgeschrieben** — eine Zeile „etwa , , " sagt
// nichts.
func TestEinLeererNameWirdNichtGenannt(t *testing.T) {
	verluste := ausfaelle{}
	verluste.nimm("kein Referenzbild", "")

	if namen := verluste["kein Referenzbild"].namen; len(namen) != 0 {
		t.Errorf("%v, erwartet keine Namen", namen)
	}
}
