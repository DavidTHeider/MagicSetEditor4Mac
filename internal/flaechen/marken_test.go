// **Marken an Karten über jede Fläche** — Werkstück 3 von Phase 3.5
// (#1208).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier die Zusage ist
//
// **Eine Marke, an einer Fläche gesetzt, steht an jeder anderen — und in
// derselben Gestalt.** Sie geht überall durch dieselbe Normalform: „Fertig"
// und „fertig" sind dieselbe Marke, gleich wo jemand sie schreibt. Ohne diese
// Läufe hätte jede Fläche ihre eigene Antwort auf „was ist eine gültige
// Marke", und ein Filter fiele still auseinander (§16).
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Setzen an einer Fläche, sehen an allen — und wieder abnehmen.**
func TestEineMarkeErreichtJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	// **Geschrieben wird auf der Kommandozeile, und zwar groß.** Was
	// ankommt, ist die Normalform — die Zusage, die Kommandozeile, MCP und
	// Dienst hier teilen (hier stand „alle vier"; das Fenster fährt der Lauf
	// nicht).
	gesetzt := mussKommandozeile(t, umgebung, "marke", setID, karteID, "Fertig", "--json")
	if !strings.Contains(gesetzt, `"fertig"`) {
		t.Fatalf("setzen: %s", gesetzt)
	}

	t.Run("Kommandozeile nennt sie an der Karte und im Set", func(t *testing.T) {
		anKarte := mussKommandozeile(t, umgebung, "marke", setID, karteID, "--json")
		if !strings.Contains(anKarte, `"fertig"`) {
			t.Fatalf("an der Karte: %s", anKarte)
		}

		imSet := mussKommandozeile(t, umgebung, "marken", setID, "--json")
		if !strings.Contains(imSet, `"fertig"`) {
			t.Fatalf("im Set: %s", imSet)
		}
	})

	t.Run("MCP nennt sie", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.kartenmarken",
			map[string]any{"setID": zahl(t, setID), "karteID": zahl(t, karteID)})
		if istFehler || !strings.Contains(text, `"fertig"`) {
			t.Fatalf("lesen.kartenmarken: %s", text)
		}

		imSet, istFehler := mcpRuf(t, umgebung, "lesen.marken",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler || !strings.Contains(imSet, `"fertig"`) {
			t.Fatalf("lesen.marken: %s", imSet)
		}
	})

	t.Run("REST nennt sie", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/karten/"+karteID+"/marken", "")
		if status != http.StatusOK || !strings.Contains(rumpf, `"fertig"`) {
			t.Fatalf("GET kartenmarken: %d (%s)", status, rumpf)
		}

		status, rumpf = an.ruf(t, http.MethodGet, "/sets/"+setID+"/marken", "")
		if status != http.StatusOK || !strings.Contains(rumpf, `"fertig"`) {
			t.Fatalf("GET marken: %d (%s)", status, rumpf)
		}
	})

	// **Und REST antwortet mit der Normalform, nicht mit dem Rohwert.**
	//
	// # Der Befund, für den es diesen Lauf gibt
	//
	// Der Lauf oben schrieb **nur** über die Kommandozeile und las danach über
	// REST. Ein **Schreibweg über REST mit nicht normaler Eingabe** kam darin
	// nicht vor — und genau dort wich REST ab: es reichte den Rohwert an den
	// Dienst weiter (richtig) und antwortete danach mit demselben Rohwert
	// (falsch). `PUT …/marken/Fertig` legte `fertig` ab und meldete `Fertig`.
	//
	// Ein Klient, der die Schreibantwort in seinen Zustand übernimmt, zeigt
	// damit etwas anderes als der nächste Leseaufruf. Angestrichen vom
	// Zweitblick über die Nacharbeit
	// (#1278,
	// Befund 3).
	t.Run("REST antwortet mit der Normalform", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/karten/"+karteID+"/marken/Gross%20Geschrieben", "")
		if status != http.StatusOK {
			t.Fatalf("PUT marke: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, `"gross geschrieben"`) {
			t.Errorf("PUT antwortet mit %s — erwartet die Normalform, in der "+
				"sie gespeichert wird", rumpf)
		}

		// **Und das Abnehmen ebenso.** Dieselbe Naht, derselbe Fehler.
		status, rumpf = an.ruf(t, http.MethodDelete,
			"/sets/"+setID+"/karten/"+karteID+"/marken/Gross%20Geschrieben", "")
		if status != http.StatusOK {
			t.Fatalf("DELETE marke: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, `"gross geschrieben"`) {
			t.Errorf("DELETE antwortet mit %s — erwartet die Normalform", rumpf)
		}
	})

	t.Run("und sie lässt sich über REST wieder abnehmen", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodDelete,
			"/sets/"+setID+"/karten/"+karteID+"/marken/fertig", "")
		if status != http.StatusOK {
			t.Fatalf("DELETE marke: %d (%s)", status, rumpf)
		}

		danach := mussKommandozeile(t, umgebung, "marken", setID, "--json")
		if strings.Contains(danach, `"fertig"`) {
			t.Fatalf("nach dem Abnehmen: %s", danach)
		}
	})
}

// **Eine leere Marke wird überall abgewiesen — und hinterlässt nichts.**
//
// # Warum gerade dieser Fall
//
// Er ist die **Grenze** der Marke: sie ist ein Schlüssel und kein Nichts.
// Stünde die Regel nur an einer Fläche, legte eine andere eine leere Zeile in
// den Bestand, nach der niemand filtern kann.
func TestEineLeereMarkeWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	t.Run("Kommandozeile", func(t *testing.T) {
		aus, _, code := kommandozeile(t, umgebung, "marke", setID, karteID, "   ")
		if code == 0 {
			t.Fatalf("die leere Marke kam durch: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		if _, istFehler := mcpRuf(t, umgebung, "schreiben.marke", map[string]any{
			"setID": zahl(t, setID), "karteID": zahl(t, karteID), "marke": "   ",
		}); !istFehler {
			t.Fatal("die leere Marke kam durch")
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		// **`%20` und nicht der leere Weg**: ein leerer Wegabschnitt träfe die
		// Route gar nicht, und dann prüfte dieser Lauf die Wegetafel statt der
		// Regel.
		status, _ := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/karten/"+karteID+"/marken/%20", "")
		if status == http.StatusOK {
			t.Fatal("die leere Marke kam durch")
		}
	})

	// **Und danach steht nichts da.**
	danach := mussKommandozeile(t, umgebung, "marken", setID, "--json")
	if strings.Contains(danach, `" "`) {
		t.Fatalf("es blieb etwas zurück: %s", danach)
	}
}
