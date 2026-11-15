// **Eine Kartenliste an jeder Fläche einlesen** — Werkstück 2 von Phase 4.5
// (#1292).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was dieser Lauf zeigt und was nicht
//
// **Er zeigt, dass alle drei Flächen dieselbe fachliche Antwort geben** — eine
// Zeile mit Setangabe ist eindeutig, eine ohne nicht, und was eingelesen wird,
// steht danach in der Sammlung.
//
// **Er zeigt nicht, dass Scryfall das auch täte.** Die Gegenstelle stellt es
// nach; was gegen den echten Dienst nachgelesen ist, steht mit Abrufdatum im
// Kopf von `internal/service/listeneinfuhr.go` (§16, Regel 1).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// probeliste ist eine Liste mit einer eindeutigen und einer offenen Zeile.
//
// **Die erste nennt ein Set, die zweite nicht.** Das ist der ganze
// Unterschied, um den es hier geht: die Gegenstelle gibt mit Setangabe einen
// Druck und ohne sie zwei.
const probeliste = "2 Erste Probe (probe) 1\n3 Zweite Probe"

// **Deuten sagt an jeder Fläche dasselbe: eine eindeutig, eine offen.**
func TestEineListeWirdUeberallGleichGedeutet(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		roh := mitEingabe(t, umgebung, probeliste, "liste-deuten", "--json")

		var deutung struct {
			Zeilen []struct {
				Eindeutig bool `json:"eindeutig"`
				Drucke    int  `json:"drucke"`
			} `json:"zeilen"`
		}

		if err := json.Unmarshal([]byte(roh), &deutung); err != nil {
			t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
		}

		pruefeDeutung(t, len(deutung.Zeilen),
			deutung.Zeilen[0].Eindeutig, deutung.Zeilen[1].Eindeutig,
			deutung.Zeilen[1].Drucke)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		roh, istFehler := mcpRuf(t, umgebung, "lesen.liste.deuten",
			map[string]any{"text": probeliste})
		if istFehler {
			t.Fatalf("lesen.liste.deuten: %s", roh)
		}

		var deutung struct {
			Zeilen []struct {
				Eindeutig bool `json:"eindeutig"`
				Drucke    int  `json:"drucke"`
			} `json:"zeilen"`
		}

		if err := json.Unmarshal([]byte(roh), &deutung); err != nil {
			t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
		}

		pruefeDeutung(t, len(deutung.Zeilen),
			deutung.Zeilen[0].Eindeutig, deutung.Zeilen[1].Eindeutig,
			deutung.Zeilen[1].Drucke)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sammlung/liste/deuten",
			string(mussJSON(t, map[string]any{"text": probeliste})))
		if status != http.StatusOK {
			t.Fatalf("POST /sammlung/liste/deuten: %d (%s)", status, rumpf)
		}

		var deutung struct {
			Zeilen []struct {
				Eindeutig bool `json:"eindeutig"`
				Drucke    int  `json:"drucke"`
			} `json:"zeilen"`
		}

		if err := json.Unmarshal([]byte(rumpf), &deutung); err != nil {
			t.Fatalf("die Antwort lesen: %v (%s)", err, rumpf)
		}

		pruefeDeutung(t, len(deutung.Zeilen),
			deutung.Zeilen[0].Eindeutig, deutung.Zeilen[1].Eindeutig,
			deutung.Zeilen[1].Drucke)
	})
}

// pruefeDeutung hält die fachliche Zusage — an jeder Fläche dieselbe.
func pruefeDeutung(t *testing.T, zeilen int, ersteEindeutig, zweiteEindeutig bool, drucke int) {
	t.Helper()

	if zeilen != 2 {
		t.Fatalf("%d Zeilen, erwartet 2", zeilen)
	}

	if !ersteEindeutig {
		t.Error("die Zeile mit Setangabe gilt als nicht eindeutig")
	}

	if zweiteEindeutig {
		t.Error("die Zeile ohne Setangabe gilt als eindeutig — sie passt auf zwei Drucke")
	}

	if drucke != 2 {
		t.Errorf("die offene Zeile nennt %d Drucke, erwartet 2", drucke)
	}
}

// **Was eingelesen wird, steht danach in der Sammlung — an jeder Fläche.**
func TestEineListeWirdUeberallGleichEingelesen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		// **Die Kommandozeile nimmt selbst, was eindeutig ist** — und sagt,
		// wie viel offen bleibt.
		roh := mitEingabe(t, umgebung, probeliste, "liste-einlesen", "--json")

		var bericht struct {
			Angelegt []int64 `json:"angelegt"`
			Stueck   int     `json:"stueck"`
			Offen    int     `json:"offen"`
		}

		if err := json.Unmarshal([]byte(roh), &bericht); err != nil {
			t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
		}

		if len(bericht.Angelegt) != 1 || bericht.Stueck != 2 || bericht.Offen != 1 {
			t.Fatalf("%d angelegt, %d Stück, %d offen — erwartet 1, 2, 1",
				len(bericht.Angelegt), bericht.Stueck, bericht.Offen)
		}

		pruefeSammlung(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		// **MCP verlangt die Druckkennung je Zeile** — es entscheidet nicht
		// selbst, welcher von mehreren Drucken gemeint ist (0044).
		roh, istFehler := mcpRuf(t, umgebung, "schreiben.liste.einlesen",
			map[string]any{"zeilen": []any{map[string]any{
				"printing": "00000000-0000-4000-8000-000000000001",
				"menge":    2, "zustand": "nearMint", "sprache": "en",
				"name": "Erste Probe", "setcode": "probe", "nummer": "1",
			}}})
		if istFehler {
			t.Fatalf("schreiben.liste.einlesen: %s", roh)
		}

		pruefeSammlung(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sammlung/liste/einlesen",
			string(mussJSON(t, map[string]any{"zeilen": []any{map[string]any{
				"printing": "00000000-0000-4000-8000-000000000001",
				"menge":    2, "zustand": "nearMint", "sprache": "en",
				"name": "Erste Probe", "setcode": "probe", "nummer": "1",
			}}})))
		if status != http.StatusOK {
			t.Fatalf("POST /sammlung/liste/einlesen: %d (%s)", status, rumpf)
		}

		pruefeSammlung(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})
}

// pruefeSammlung hält, was nach dem Einlesen im Bestand steht.
//
// **Über die Kommandozeile gelesen, gleich welche Fläche geschrieben hat.**
// Das ist der Punkt: eine Fläche, die in einen eigenen Bestand schriebe, fiele
// hier auf — die drei Läufe teilen ihre Umgebung je Fall, aber nicht
// untereinander.
func pruefeSammlung(t *testing.T, roh string) {
	t.Helper()

	var sammlung struct {
		Exemplare []struct {
			Printing string `json:"printing"`
			Menge    int    `json:"menge"`
			Name     string `json:"name"`
		} `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &sammlung); err != nil {
		t.Fatalf("die Sammlung lesen: %v (%s)", err, roh)
	}

	if len(sammlung.Exemplare) != 1 {
		t.Fatalf("%d Exemplare, erwartet 1: %s", len(sammlung.Exemplare), roh)
	}

	eines := sammlung.Exemplare[0]

	if eines.Menge != 2 {
		t.Errorf("Menge %d, erwartet 2", eines.Menge)
	}

	// **Der lesbare Bezug ist mitgekommen** (§16, Regel 6: was übernommen
	// wird, trägt seine Herkunft). Ohne ihn stünde dort eine UUID.
	if !strings.Contains(eines.Name, "Probe") {
		t.Errorf("Name %q — der lesbare Bezug ist nicht mitgekommen", eines.Name)
	}
}
