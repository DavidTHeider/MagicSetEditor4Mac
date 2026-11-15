// **Die Schlüsselliste und die Platzhalterzahl** —
// #1178.
//
// Beide gibt es für den Vertragstest gegen die Kataloge des Fensters
// (`internal/werkzeuge/skriptmeldungen_test.go`). Hier steht, dass sie
// stimmen — dort, dass die Kataloge dazu passen.
package msescript_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestDieSchluessellisteFuehrtJedenSchluessel(t *testing.T) {
	t.Parallel()

	alle := msescript.Meldungsschluesselliste()

	if len(alle) < 20 {
		t.Errorf("%d Schlüssel — der Zerteiler und der Auswerter haben mehr", len(alle))
	}

	// **Stichprobe an beiden Enden**: einer aus dem Lexer, einer aus dem
	// Zerteiler, einer aus dem Auswerter.
	for _, schluessel := range []msescript.Meldungsschluessel{
		msescript.SchluesselFremdesZeichen,
		msescript.SchluesselFehltNach,
		msescript.SchluesselNameUnbekannt,
	} {
		if !slices.Contains(alle, schluessel) {
			t.Errorf("die Liste führt %q nicht", schluessel)
		}
	}
}

func TestDiePlatzhalterzahlStimmt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		schluessel msescript.Meldungsschluessel
		erwartet   int
	}{
		{msescript.SchluesselOhneRumpf, 0},
		{msescript.SchluesselNameUnbekannt, 1},
		{msescript.SchluesselFehltNach, 2},
	} {
		if zahl := msescript.PlatzhalterZahl(fall.schluessel); zahl != fall.erwartet {
			t.Errorf("%q hat %d Platzhalter, erwartet %d",
				fall.schluessel, zahl, fall.erwartet)
		}
	}
}

// **Ein Fehler trägt Schlüssel, Werte und den deutschen Satz.**
//
// Der Satz bleibt (§9: Fehlermeldungen sind deutsch); Kommandozeile, MCP und
// Dienst bekommen ihn so, das Fenster bekommt zusätzlich den Schlüssel.
func TestEinFehlerTraegtSchluesselUndSatz(t *testing.T) {
	t.Parallel()

	fehler := msescript.NeuerFehler(3, 7, msescript.SchluesselFehltNach, "`if`", "`then`")

	if fehler.Meldung != "nach `if` fehlt `then`" {
		t.Errorf("Meldung %q", fehler.Meldung)
	}

	if fehler.Schluessel != msescript.SchluesselFehltNach || len(fehler.Werte) != 2 {
		t.Errorf("Schlüssel %q, Werte %v", fehler.Schluessel, fehler.Werte)
	}

	if fehler.Error() != "Zeile 3, Spalte 7: nach `if` fehlt `then`" {
		t.Errorf("Error() %q", fehler.Error())
	}
}

// **Ein Schlüssel ohne Muster gibt seinen Namen.**
//
// Er kommt nicht vor — Aufzählung und Tabelle stehen nebeneinander —, und ein
// leerer Satz wäre die schlechtere Antwort.
func TestEinSchluesselOhneMusterGibtSeinenNamen(t *testing.T) {
	t.Parallel()

	fehler := msescript.NeuerFehler(1, 1, msescript.Meldungsschluessel("gibtsnicht"))

	if fehler.Meldung != "gibtsnicht" {
		t.Errorf("Meldung %q, erwartet den Namen", fehler.Meldung)
	}
}
