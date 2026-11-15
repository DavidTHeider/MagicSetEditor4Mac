// **Die Verteilung nennt die Karten ihrer Werte — an jeder Fläche** — UXR-F6
// (#2013).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier geprüft wird
//
// Die Statistik am Fenster filtert die Kartenliste auf die Karten eines
// Balkens, und dafür nennt jeder Wert der Verteilung seine Karten. Dieselbe
// Auskunft gehört an jede Fläche (§16): ein Modell, das fragt, welche Karten
// „gruen" tragen, bekommt keine Antwort, wenn nur das Fenster sie kennt.
//
// Das Probeset führt drei Karten; `farbe` trägt auf mindestens einer „gruen".
// Geprüft wird, was an jeder Fläche gelten muss: jeder Wert nennt **genau so
// viele Karten, wie er zählt**, und keine Karte steht bei zwei Werten, denn
// eine Karte trägt ein Feld einmal.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestDieVerteilungNenntDieKartenUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := eingefahren(t, umgebung)

		pruefeWertkarten(t, mussKommandozeile(t, umgebung,
			"statistik", setID, "farbe", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := eingefahren(t, umgebung)

		roh, istFehler := mcpRuf(t, umgebung, "lesen.verteilung",
			map[string]any{"setID": zahl(t, setID), "feld": "farbe"})
		if istFehler {
			t.Fatalf("lesen.verteilung: %s", roh)
		}

		pruefeWertkarten(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := eingefahren(t, umgebung)

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/statistik/farbe", "")
		if status != http.StatusOK {
			t.Fatalf("GET verteilung: %d (%s)", status, rumpf)
		}

		pruefeWertkarten(t, rumpf)
	})
}

// pruefeWertkarten hält die Zusage über die Verteilung — an jeder Fläche
// dieselbe.
func pruefeWertkarten(t *testing.T, roh string) {
	t.Helper()

	var stand struct {
		Verteilung []struct {
			Name      string   `json:"name"`
			Karten    int      `json:"karten"`
			KartenIDs []string `json:"kartenIDs"`
		} `json:"verteilung"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, roh)
	}

	if len(stand.Verteilung) == 0 {
		t.Fatalf("die Verteilung ist leer: %s", roh)
	}

	gesehen := map[string]string{}

	for _, wert := range stand.Verteilung {
		if wert.Karten == 0 || len(wert.KartenIDs) != wert.Karten {
			t.Errorf("der Wert %q zählt %d Karten und nennt %v",
				wert.Name, wert.Karten, wert.KartenIDs)
		}

		for _, kennung := range wert.KartenIDs {
			if davor, da := gesehen[kennung]; da {
				t.Errorf("die Karte %s steht bei %q und bei %q", kennung, davor, wert.Name)
			}

			gesehen[kennung] = wert.Name
		}
	}
}
