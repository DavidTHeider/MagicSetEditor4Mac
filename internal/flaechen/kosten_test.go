package flaechen_test

// **Dieselbe teure Operation kostet über jede Fläche dasselbe** — Folgebefund
// F1 des Lastlaufs QA-08
// (#2003).
//
// # Wie gemessen wird
//
// **Am gebauten Programm und am Guthaben, nicht an einer Liste.** Ein Ruf,
// der an seinen Argumenten scheitert, ist schon gebucht — das Guthaben steht
// vor der Arbeit (#1429). Wer eine Karte zeichnen lässt, ohne eine zu nennen,
// zahlt also, was das Zeichnen kostet, und bekommt sofort eine Absage. So
// viele gehen durch, bis das Guthaben leer ist: 120 Einheiten geteilt durch
// die Kosten, und was in der kurzen Zeit nachläuft, ändert daran nichts
// Ganzzahliges.
//
// **Die Kommandozeile bucht nicht** — sie ist der Mensch an seinem Rechner.
// Sie **sieht** die Kosten in `mse grenzen`, und dort steht dieselbe Tafel.
//
// # Was der Lauf nicht hält
//
// **Jede teure Operation einzeln** — das hält `internal/rest` für jeden Weg
// der Tafel und `internal/mcpnaehte` für jedes Werkzeug. Hier geht es um die
// Naht: dass die gebauten Programme die Tafel wirklich fragen.

import (
	"encoding/json"
	"math"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// kostentafelzeile ist eine Zeile aus `grenzen`.
type kostentafelzeile struct {
	Faehigkeit string   `json:"faehigkeit"`
	Kosten     int      `json:"kosten"`
	CLI        []string `json:"cli"`
	MCP        []string `json:"mcp"`
	REST       []string `json:"rest"`
}

// kostenLautKommandozeile liest, was `mse grenzen` für einen Befehl nennt.
func kostenLautKommandozeile(t *testing.T, umgebung []string, befehl string) (zeile kostentafelzeile, burst int) {
	t.Helper()

	var grenzen struct {
		Principal struct {
			Burst int `json:"burst"`
		} `json:"principal"`
		Kostentafel []kostentafelzeile `json:"kostentafel"`
	}

	if fehler := json.Unmarshal([]byte(mussKommandozeile(t, umgebung, "grenzen", "--json")), &grenzen); fehler != nil {
		t.Fatalf("mse grenzen: %v", fehler)
	}

	for _, gelesen := range grenzen.Kostentafel {
		if slices.Contains(gelesen.CLI, befehl) {
			return gelesen, grenzen.Principal.Burst
		}
	}

	t.Fatalf("mse grenzen nennt %q nicht", befehl)

	return kostentafelzeile{}, 0
}

// gebuchteKosten rechnet aus der Zahl der Rufe, die durchgingen, zurück.
func gebuchteKosten(burst, durch int) int {
	if durch == 0 {
		return 0
	}

	return int(math.Round(float64(burst) / float64(durch)))
}

func TestDieTeureOperationKostetUeberJedeFlaecheDasselbe(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	zeile, burst := kostenLautKommandozeile(t, umgebung, "mse zeichnen")
	if zeile.Kosten != 5 || !slices.Contains(zeile.MCP, "schreiben.karte.zeichnen") ||
		!slices.Contains(zeile.REST, "POST /karten/{}/bild") {
		t.Fatalf("die Tafel sagt %+v", zeile)
	}

	// **REST**: zeichnen ohne Karte, bis das Guthaben leer ist.
	an := starteDienst(t, umgebung)
	durchREST := 0

	for range 2 * burst {
		status, antwort := an.ruf(t, http.MethodPost, "/karten/0/bild", `{`)
		if status == http.StatusTooManyRequests {
			break
		}

		if status < http.StatusBadRequest {
			t.Fatalf("ein Ruf ohne Karte gelang: %d %s", status, antwort)
		}

		durchREST++
	}

	// **MCP über stdio**: derselbe Ruf an einem Prozess — er ist ein Principal.
	prozess := starteMCPProzess(t, umgebung)
	durchMCP := 0

	for range 2 * burst {
		text, istFehler := prozess.ruf("schreiben.karte.zeichnen", map[string]any{})
		if strings.Contains(text, "zu viele Rufe") {
			break
		}

		if !istFehler {
			t.Fatalf("ein Ruf ohne Karte gelang: %s", text)
		}

		durchMCP++
	}

	t.Logf("durch, bis das Guthaben von %d leer war: REST %d, MCP %d; mse grenzen sagt %d je Ruf",
		burst, durchREST, durchMCP, zeile.Kosten)

	if gebucht := gebuchteKosten(burst, durchREST); gebucht != zeile.Kosten {
		t.Errorf("REST buchte %d je Ruf (%d durch), mse grenzen sagt %d", gebucht, durchREST, zeile.Kosten)
	}

	if gebucht := gebuchteKosten(burst, durchMCP); gebucht != zeile.Kosten {
		t.Errorf("MCP buchte %d je Ruf (%d durch), mse grenzen sagt %d", gebucht, durchMCP, zeile.Kosten)
	}
}

// **Ein Abbruch wird am Dienst nie gedrosselt** — auch nicht mit leerem
// Guthaben. Im Lastlauf bekam `DELETE /auftraege/{id}` ein 429, und der
// Auftrag lief zu Ende.
func TestEinAbbruchWirdAmDienstNieGedrosselt(t *testing.T) {
	an := starteDienst(t, mitEigenerSaat(t))

	leer := false

	for range 200 {
		if status, _ := an.ruf(t, http.MethodPost, "/karten/0/bild", `{`); status == http.StatusTooManyRequests {
			leer = true

			break
		}
	}

	if !leer {
		t.Fatal("das Guthaben wurde nie leer — der Lauf sagte nichts")
	}

	for _, weg := range []string{"/auftraege/unbekannt", "/index/aufbau"} {
		if status, antwort := an.ruf(t, http.MethodDelete, weg, ""); status == http.StatusTooManyRequests {
			t.Errorf("DELETE %s wurde gedrosselt: %s", weg, antwort)
		}
	}
}

// **`grenzen` sagt an jeder Fläche dasselbe** — Feld für Feld, samt
// Auftrags- und Verbindungsgrenze und der Kostentafel. Bis zum 2026-10-02
// nannte keine Fläche die vier Aufträge und die sechzehn Verbindungen.
func TestDieGrenzenSagenAnJederFlaecheDasselbe(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	gelesen := map[string]map[string]any{}

	for flaeche, roh := range map[string]string{
		"Kommandozeile": mussKommandozeile(t, umgebung, "grenzen", "--json"),
		"MCP":           mussMCP(t, umgebung, "lesen.grenzen", map[string]any{}),
		"REST":          mussREST(t, an, http.MethodGet, "/grenzen", ""),
	} {
		var auskunft map[string]any
		if fehler := json.Unmarshal([]byte(roh), &auskunft); fehler != nil {
			t.Fatalf("%s: %v — %s", flaeche, fehler, roh)
		}

		gelesen[flaeche] = auskunft
	}

	kommandozeile := gelesen["Kommandozeile"]
	for _, schluessel := range []string{"verbindungen", "auftraege", "kostentafel"} {
		if _, steht := kommandozeile[schluessel]; !steht {
			t.Errorf("die Auskunft nennt %q nicht", schluessel)
		}
	}

	for _, flaeche := range []string{"MCP", "REST"} {
		if !reflect.DeepEqual(gelesen[flaeche], kommandozeile) {
			t.Errorf("%s sagt etwas anderes als die Kommandozeile:\n%v\n%v", flaeche, gelesen[flaeche], kommandozeile)
		}
	}
}
