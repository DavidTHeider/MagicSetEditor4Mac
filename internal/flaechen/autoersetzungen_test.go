// **Die Autoersetzungen verwalten — an jeder Fläche** — Werkstück E von Phase
// 4.11 (#1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier geprüft wird
//
// Die Saat `probe.mse-game` bringt drei `auto replace:`-Blöcke mit — `(C)`,
// `--` und `->`. Daran hängt der Satz, um den es geht: eine Regel des Pakets
// steht in der Liste, ist **nicht eigen** und lässt sich nicht löschen.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// **Jede Fläche nennt dieselben Regeln.**
func TestDieAutoersetzungenStehenUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		pruefeErsetzungen(t, mussKommandozeile(t, umgebung,
			"autoersetzung", setID, "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		roh, istFehler := mcpRuf(t, umgebung, "lesen.autoersetzungen",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen.autoersetzungen: %s", roh)
		}

		pruefeErsetzungen(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/autoersetzungen", "")
		if status != http.StatusOK {
			t.Fatalf("GET /sets/%s/autoersetzungen: %d (%s)", setID, status, rumpf)
		}

		pruefeErsetzungen(t, rumpf)
	})
}

// **Und jede Fläche legt dieselbe eigene Regel an.**
func TestEineEigeneAutoersetzungEntstehtUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		mussKommandozeile(t, umgebung,
			"autoersetzung", setID, "setzen", "(W)", "weiss")

		pruefeEigene(t, mussKommandozeile(t, umgebung,
			"autoersetzung", setID, "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.autoersetzung",
			map[string]any{
				"setID": zahl(t, setID), "muster": "(W)", "ersatz": "weiss",
				"ganzesWort": true, "aktiv": true,
			})
		if istFehler {
			t.Fatalf("schreiben.autoersetzung: %s", roh)
		}

		gelesen, _ := mcpRuf(t, umgebung, "lesen.autoersetzungen",
			map[string]any{"setID": zahl(t, setID)})

		pruefeEigene(t, gelesen)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/autoersetzungen/(W)",
			`{"ersatz":"weiss","ganzesWort":true,"aktiv":true}`)
		if status != http.StatusNoContent {
			t.Fatalf("PUT: %d (%s)", status, rumpf)
		}

		_, gelesen := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/autoersetzungen", "")

		pruefeEigene(t, gelesen)
	})
}

// **Und eine Regel des Pakets lässt sich nirgends löschen.**
//
// MSEs `remove->Enable(ar && ar->custom)` (`src/gui/auto_replace_window.cpp`,
// Zeile 258, Fassung f9d9356 vom 2021-07-06) — hier als fachliches Nein an
// allen drei Flächen.
func TestEinePaketregelLaesstSichNirgendsLoeschen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		if _, _, code := kommandozeile(t, umgebung,
			"autoersetzung", setID, "loeschen", "(C)"); code != 1 {
			t.Errorf("Rückgabewert %d, erwartet war 1 (fachlich)", code)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		if _, istFehler := mcpRuf(t, umgebung, "schreiben.autoersetzung.loeschen",
			map[string]any{"setID": zahl(t, setID), "muster": "(C)"}); !istFehler {
			t.Error("MCP löschte eine Regel des Pakets")
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodDelete,
			"/sets/"+setID+"/autoersetzungen/(C)", "")
		if status != http.StatusBadRequest {
			t.Errorf("Status %d, erwartet war 400 (%s)", status, rumpf)
		}
	})
}

// **Und der Schalter und das Zurücksetzen greifen überall.**
func TestSchalterUndZuruecksetzenGreifenUeberall(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		mussKommandozeile(t, umgebung, "autoersetzung", setID, "aus")

		if !ersetzungAus(t, mussKommandozeile(t, umgebung,
			"autoersetzung", setID, "--json")) {
			t.Error("der Schalter blieb an")
		}

		mussKommandozeile(t, umgebung, "autoersetzung", setID, "zuruecksetzen")

		if ersetzungAus(t, mussKommandozeile(t, umgebung,
			"autoersetzung", setID, "--json")) {
			t.Error("das Zurücksetzen ließ den Schalter aus")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		if _, istFehler := mcpRuf(t, umgebung, "schreiben.autoersetzung.schalter",
			map[string]any{"setID": zahl(t, setID), "an": false}); istFehler {
			t.Fatal("der Schalter ließ sich nicht umlegen")
		}

		schalterIst(t, umgebung, setID, true)

		if _, istFehler := mcpRuf(t, umgebung,
			"schreiben.autoersetzungen.zuruecksetzen",
			map[string]any{"setID": zahl(t, setID)}); istFehler {
			t.Fatal("das Zurücksetzen scheiterte")
		}

		schalterIst(t, umgebung, setID, false)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Ersetzset")

		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/autoersetzungen", `{"an":false}`); status != http.StatusNoContent {
			t.Fatalf("PUT: %d (%s)", status, rumpf)
		}

		schalterIst(t, umgebung, setID, true)

		if status, rumpf := an.ruf(t, http.MethodDelete,
			"/sets/"+setID+"/autoersetzungen", ""); status != http.StatusNoContent {
			t.Fatalf("DELETE: %d (%s)", status, rumpf)
		}

		schalterIst(t, umgebung, setID, false)
	})
}

// schalterIst liest den Schalter an derselben Stelle wie der Fall der
// Kommandozeile — **die Wirkung an jeder Fläche**: MCP und Dienst prüften nur
// ihr Ja (Zweitblick über Phase C, Befund 3).
func schalterIst(t *testing.T, umgebung []string, setID string, aus bool) {
	t.Helper()

	if ersetzungAus(t, mussKommandozeile(t, umgebung, "autoersetzung", setID, "--json")) != aus {
		t.Errorf("der Schalter steht nicht auf aus=%v", aus)
	}
}

// ersetzungsstand ist die Antwort auf »welche Regeln gelten hier?«.
type ersetzungsstand struct {
	Regeln []struct {
		Muster     string `json:"muster"`
		Ersatz     string `json:"ersatz"`
		GanzesWort bool   `json:"ganzesWort"`
		Aktiv      bool   `json:"aktiv"`
		Eigen      bool   `json:"eigen"`
	} `json:"regeln"`
	An bool `json:"an"`
}

// gelesenerStand liest die Antwort.
func gelesenerStand(t *testing.T, roh string) ersetzungsstand {
	t.Helper()

	var stand ersetzungsstand

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, roh)
	}

	return stand
}

// pruefeErsetzungen hält die fachliche Zusage — an jeder Fläche dieselbe.
func pruefeErsetzungen(t *testing.T, roh string) {
	t.Helper()

	stand := gelesenerStand(t, roh)

	// **Die drei der Saat** — als Gleichheit und nicht als Untergrenze.
	if len(stand.Regeln) != 3 {
		t.Fatalf("%d Regeln: %s", len(stand.Regeln), roh)
	}

	if !stand.An {
		t.Error("ohne Zeile im Bestand soll der Schalter an sein")
	}

	for _, regel := range stand.Regeln {
		if regel.Eigen {
			t.Errorf("die Regel %q gilt als eigen", regel.Muster)
		}

		if !regel.Aktiv {
			t.Errorf("die Regel %q gilt als abgeschaltet", regel.Muster)
		}
	}
}

// pruefeEigene hält die Zusage über die eigene Regel.
func pruefeEigene(t *testing.T, roh string) {
	t.Helper()

	stand := gelesenerStand(t, roh)

	// **Vier Regeln**: die drei der Saat und die eigene.
	if len(stand.Regeln) != 4 {
		t.Fatalf("%d Regeln: %s", len(stand.Regeln), roh)
	}

	gefunden := false

	for _, regel := range stand.Regeln {
		if regel.Muster == "(W)" {
			gefunden = true

			if !regel.Eigen || regel.Ersatz != "weiss" {
				t.Errorf("die eigene Regel lautet %+v", regel)
			}
		}
	}

	if !gefunden {
		t.Errorf("die eigene Regel fehlt: %s", roh)
	}
}

// ersetzungAus sagt, ob der Schalter aus ist.
func ersetzungAus(t *testing.T, roh string) bool {
	t.Helper()

	return !gelesenerStand(t, roh).An
}
