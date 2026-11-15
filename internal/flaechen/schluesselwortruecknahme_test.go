// **Die Rücknahme greift auf Schlüsselwörter — über jede Fläche**
// (#1077).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Warum der Lauf hierhergehört
//
// ⌘Z ist eine Fähigkeit des Fensters; kopflos heißt sie `mse zuruecknehmen`,
// `POST /sets/{id}/zuruecknehmen` und `schreiben.zuruecknehmen`. Eine
// Rücknahme, die eine Schlüsselwortänderung an einer Fläche umkehrt und an
// einer anderen nicht, wäre genau der Fall, den §16 verbietet — und er ließe
// sich nur so finden: der Verlauf steht im Bestand, den alle vier teilen.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Angelegt hier, zurückgenommen dort — und die dritte Fläche sieht es.**
func TestDieRuecknahmeEinerWortaenderungUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)

	// **Angelegt über die Kommandozeile.**
	wortID := strings.TrimSpace(
		mussKommandozeile(t, umgebung, "schluesselwort", setID, "anlegen"))
	if wortID == "" {
		t.Fatal("das Anlegen gab keine Kennung")
	}

	mussKommandozeile(t, umgebung,
		"schluesselwort", setID, "setzen", wortID, "wort", "Flügelschlag")

	t.Run("zurückgenommen über REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/zuruecknehmen", "")
		if status != http.StatusOK {
			t.Fatalf("POST zuruecknehmen: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, `"bewegt":true`) {
			t.Fatalf("nichts bewegt: %s", rumpf)
		}
	})

	t.Run("und die Kommandozeile sieht es", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "schluesselwoerter", setID)

		if strings.Contains(aus, "Flügelschlag") {
			t.Errorf("der Name steht noch in\n%s", aus)
		}
	})

	t.Run("wiederhergestellt über MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "schreiben.wiederherstellen",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("wiederherstellen: %s", text)
		}

		if !strings.Contains(text, `"bewegt": true`) &&
			!strings.Contains(text, `"bewegt":true`) {
			t.Fatalf("nichts wiederhergestellt: %s", text)
		}
	})

	t.Run("und der Dienst sieht es", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/schluesselwoerter", "")
		if status != http.StatusOK {
			t.Fatalf("GET: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, "Flügelschlag") {
			t.Errorf("der Name fehlt in\n%s", rumpf)
		}
	})
}

// **Ein gelöschtes Schlüsselwort kommt an seine Stelle zurück — über jede
// Fläche gesehen.**
//
// **An seine Stelle und nicht ans Ende**: die Reihenfolge entscheidet, welches
// Schlüsselwort zuerst auf einen Text passt (`src/data/keyword.cpp`, Zeilen
// 148–208, Fassung f9d9356 vom 2021-07-06).
func TestEinGeloeschtesWortKommtUeberJedeFlaecheAnSeineStelle(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)

	erstes := strings.TrimSpace(
		mussKommandozeile(t, umgebung, "schluesselwort", setID, "anlegen"))
	mussKommandozeile(t, umgebung,
		"schluesselwort", setID, "setzen", erstes, "wort", "Erstes")

	zweites := strings.TrimSpace(
		mussKommandozeile(t, umgebung, "schluesselwort", setID, "anlegen"))
	mussKommandozeile(t, umgebung,
		"schluesselwort", setID, "setzen", zweites, "wort", "Zweites")

	// **Gelöscht wird das erste** — ein gelöschtes letztes käme auch ans Ende
	// zurück, und der Test sagte nichts.
	mussKommandozeile(t, umgebung, "schluesselwort", setID, "loeschen", erstes)

	if aus := mussKommandozeile(t, umgebung, "schluesselwoerter", setID); strings.Contains(
		aus, "Erstes") {
		t.Fatalf("das gelöschte steht noch in\n%s", aus)
	}

	t.Run("zurückgenommen über MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "schreiben.zuruecknehmen",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("zuruecknehmen: %s", text)
		}
	})

	t.Run("und es steht wieder vorn", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "schluesselwoerter", setID)

		ersteStelle := strings.Index(aus, "Erstes")
		zweiteStelle := strings.Index(aus, "Zweites")

		if ersteStelle < 0 {
			t.Fatalf("das zurückgebrachte fehlt in\n%s", aus)
		}

		if ersteStelle > zweiteStelle {
			t.Errorf("es steht hinter dem zweiten:\n%s", aus)
		}
	})
}
