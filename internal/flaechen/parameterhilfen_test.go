// **Die Parameterhilfen der Schlüsselwörter — an jeder Fläche** — Werkstück D
// von Phase 4.11
// (#1342,
// #1349).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier geprüft wird
//
// Die Saat `probe.mse-game` bringt seit #1342 zwei `keyword parameter type:`
// mit — eine **ohne** Verweisskripte und eine **mit** zweien. Daran hängt die
// Unterscheidung, die MSE macht: ohne gibt es `{paramN}`, mit je Skript einen
// Eintrag.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Jede Fläche nennt dieselben Hilfen.**
func TestDieParameterhilfenStehenUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID, wortID := wortMitParameter(t, umgebung)

		pruefeHilfen(t, mussKommandozeile(t, umgebung,
			"schluesselwort", setID, "parameter", wortID, "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID, wortID := wortMitParameter(t, umgebung)

		roh, istFehler := mcpRuf(t, umgebung, "lesen.parameterhilfen",
			map[string]any{"setID": zahl(t, setID), "wortID": zahl(t, wortID)})
		if istFehler {
			t.Fatalf("lesen.parameterhilfen: %s", roh)
		}

		pruefeHilfen(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID, wortID := wortMitParameter(t, umgebung)

		an := starteDienst(t, umgebung)

		weg := "/sets/" + setID + "/schluesselwoerter/" + wortID + "/parameter"

		status, rumpf := an.ruf(t, http.MethodGet, weg, "")
		if status != http.StatusOK {
			t.Fatalf("GET %s: %d (%s)", weg, status, rumpf)
		}

		pruefeHilfen(t, rumpf)
	})
}

// wortMitParameter legt ein Set an und darin ein Schlüsselwort mit einem
// Parameter der Art `farbwort` — der mit den Verweisskripten.
func wortMitParameter(t *testing.T, umgebung []string) (setID, wortID string) {
	t.Helper()

	setID = neuesSet(t, umgebung, "Wortset")

	// **Die Kennung kommt als blanke Zahl** — so gibt `schluesselwort
	// anlegen` sie aus, mit `--json` wie ohne.
	wortID = strings.TrimSpace(mussKommandozeile(t, umgebung,
		"schluesselwort", setID, "anlegen"))

	mussKommandozeile(t, umgebung, "schluesselwort", setID, "setzen",
		wortID, "muster", "Kicker <atom-param>farbwort</atom-param>")

	return setID, wortID
}

// pruefeHilfen hält die fachliche Zusage — an jeder Fläche dieselbe.
func pruefeHilfen(t *testing.T, roh string) {
	t.Helper()

	var stand struct {
		Arten []struct {
			Name      string `json:"name"`
			Einfuegen string `json:"einfuegen"`
		} `json:"arten"`
		Verweise []struct {
			Nummer    int    `json:"nummer"`
			Art       string `json:"art"`
			Name      string `json:"name"`
			Einfuegen string `json:"einfuegen"`
		} `json:"verweise"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, roh)
	}

	// **Beide Arten des Spiels** — unabhängig davon, was im Muster steht.
	if len(stand.Arten) != 2 {
		t.Fatalf("%d Arten: %s", len(stand.Arten), roh)
	}

	if stand.Arten[0].Einfuegen != "<atom-param>zahl</atom-param>" {
		t.Errorf("die erste Art fügt %q ein", stand.Arten[0].Einfuegen)
	}

	// **Und zwei Verweise** — die Art `farbwort` bringt zwei Skripte mit, und
	// beide rechnen mit `input = "param1"`.
	if len(stand.Verweise) != 2 {
		t.Fatalf("%d Verweise: %s", len(stand.Verweise), roh)
	}

	if stand.Verweise[0].Nummer != 1 || stand.Verweise[0].Art != "farbwort" ||
		stand.Verweise[0].Einfuegen != "{param1}" {
		t.Errorf("der erste Verweis lautet %+v", stand.Verweise[0])
	}

	if stand.Verweise[1].Einfuegen != "({param1})" {
		t.Errorf("der zweite Verweis lautet %+v", stand.Verweise[1])
	}
}
