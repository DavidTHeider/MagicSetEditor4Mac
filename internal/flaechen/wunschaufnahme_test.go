// **Ein Wunsch und ein Lagerort lassen sich an jeder Fläche aufnehmen** —
// Nachlauf über Phase 4.5, Befund 1
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
//
// # Was der Nachlauf sah und was dahinter lag
//
// Angestrichen war eine **Vorrichtung**: die MCP-Probe behauptet, so streng
// wie der Bestand zu sein, verglich aber vier Spalten statt fünf. Der Grund
// dafür lag eine Ebene tiefer — `mcp.Exemplarsicht` und `rest.Exemplarsicht`
// **hatten die fünfte Spalte gar nicht**. Über beide Flächen ließ sich kein
// Wunsch aufnehmen und kein Lagerort setzen; die Vorrichtung konnte gar nicht
// strenger sein als der Verbund, den sie bekam.
//
// §16 sagt: was das Fenster kann, können Kommandozeile und MCP — oder es
// steht begründet, warum nicht. Für einen Wunsch stand dort nichts.
//
// **Der Wächter über `FAEHIGKEITEN.md` sah es nicht**, und das ist genau, was
// neben ihm steht: er sieht, **dass** eine Fähigkeit vorhanden ist, nicht,
// dass sie dasselbe tut.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// **Derselbe Druck als Besitz und als Wunsch — über jede Fläche.**
//
// Der Fall ist der aus 0057: wer eine gespielte Fassung hat und eine schöne
// will, hat beides. Er ist zugleich die Probe auf die Eindeutigkeit, denn
// ohne `wunsch` in ihr wäre die zweite Zeile ein Konflikt.
func TestEinWunschLaesstSichUeberallAufnehmen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := eigenerBestand(t)

		mussKommandozeile(t, umgebung, "exemplar", "probe-druck", "4",
			"nearMint", "en", "nonfoil", "", "", "", "", "", "", "", "Kiste 4", "besitz")
		mussKommandozeile(t, umgebung, "exemplar", "probe-druck", "2",
			"nearMint", "en", "nonfoil", "", "", "", "", "", "", "", "", "wunsch")

		pruefeWunschUndOrt(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := eigenerBestand(t)

		for _, zeile := range []map[string]any{
			{"printing": "probe-druck", "menge": 4.0, "lagerort": "Kiste 4"},
			{"printing": "probe-druck", "menge": 2.0, "wunsch": true},
		} {
			if roh, istFehler := mcpRuf(t, umgebung, "schreiben.exemplar", zeile); istFehler {
				t.Fatalf("schreiben.exemplar %v: %s", zeile, roh)
			}
		}

		roh, istFehler := mcpRuf(t, umgebung, "lesen.sammlung", map[string]any{})
		if istFehler {
			t.Fatalf("lesen.sammlung: %s", roh)
		}

		pruefeWunschUndOrt(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, eigenerBestand(t))

		for _, rumpf := range []string{
			`{"printing":"probe-druck","menge":4,"lagerort":"Kiste 4"}`,
			`{"printing":"probe-druck","menge":2,"wunsch":true}`,
		} {
			if status, antwort := an.ruf(t, http.MethodPost, "/sammlung", rumpf); status != http.StatusCreated {
				t.Fatalf("POST /sammlung %s: %d (%s)", rumpf, status, antwort)
			}
		}

		status, rumpf := an.ruf(t, http.MethodGet, "/sammlung", "")
		if status != http.StatusOK {
			t.Fatalf("GET /sammlung: %d (%s)", status, rumpf)
		}

		pruefeWunschUndOrt(t, rumpf)
	})
}

// pruefeWunschUndOrt hält die fachliche Zusage — an jeder Fläche dieselbe.
//
// **Zwei Zeilen, nicht eine.** Wäre `wunsch` nicht Teil der Eindeutigkeit,
// hätte das zweite Anlegen einen Konflikt gegeben; wäre es zwar Teil der
// Eindeutigkeit, aber an der Fläche nicht setzbar, stünde hier eine Zeile
// mit sechs Stück.
func pruefeWunschUndOrt(t *testing.T, roh string) {
	t.Helper()

	var stand struct {
		Exemplare []struct {
			Menge    int    `json:"menge"`
			Lagerort string `json:"lagerort"`
			Wunsch   bool   `json:"wunsch"`
		} `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v — %s", err, roh)
	}

	if len(stand.Exemplare) != 2 {
		t.Fatalf("%d Zeilen, erwartet waren zwei: %s", len(stand.Exemplare), roh)
	}

	besitz, wunsch := stand.Exemplare[0], stand.Exemplare[1]
	if besitz.Wunsch {
		besitz, wunsch = wunsch, besitz
	}

	if besitz.Menge != 4 || besitz.Lagerort != "Kiste 4" || besitz.Wunsch {
		t.Errorf("die Besitzzeile lautet %+v", besitz)
	}

	if wunsch.Menge != 2 || !wunsch.Wunsch {
		t.Errorf("die Wunschzeile lautet %+v", wunsch)
	}
}
