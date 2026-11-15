// **Ein MCP-Prozess, der zwischen zwei Rufen am Leben bleibt** — Werkstück
// A14 von Phase 5.3
// (#1453).
//
// # Warum es [mcpRuf] und [mcpSitzung] nicht tun
//
// **Weil beide die ganze Eingabe vorher schreiben.** `mcpRuf` fährt einen
// Ruf je Prozess, `mcpSitzung` mehrere über einen — aber sie stehen alle
// schon fest, bevor der Prozess startet. Zwischen zwei Rufen kann damit
// **nichts geschehen**.
//
// **Und genau das ist die Abnahme dieses Werkstücks**: „Cloudscope-Entzug
// greift im nächsten Aufruf vor Request". Ein Entzug, der zwischen zwei
// Rufen stattfindet, braucht einen Prozess, der zwischen ihnen wartet — und
// einen Test, der ihn genau dann anfasst.
//
// # Was dieser Helfer damit prüfbar macht
//
// **Die Ausstattung je Ruf.** `internal/mcp` baut sie für jeden `tools/call`
// neu ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md));
// ob das für **jede** Frage gilt — Rolle, Richtlinie, Schalter —, sagt nur
// ein Lauf, der den Bestand mitten in der Sitzung ändert.
package flaechen_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// mcpprozess ist ein laufender `mse-mcp`, den ein Fall Schritt für Schritt
// fährt.
type mcpprozess struct {
	t       *testing.T
	hinein  io.WriteCloser
	heraus  *bufio.Reader
	lauf    *exec.Cmd
	naechst int

	// abbrechen erschlägt ihn, wenn er nach dem Dateiende nicht geht.
	abbrechen context.CancelFunc
}

// starteMCPProzess startet einen und meldet ihn an.
//
// **Die Anmeldung gehört dazu und nicht in den ersten Ruf**: ein Server, der
// vor `initialize` Werkzeuge ausführte, wäre selbst ein Befund.
func starteMCPProzess(t *testing.T, umgebung []string) *mcpprozess {
	t.Helper()

	// **Ein eigener Kontext und nicht der des Falls, und das ist der
	// Unterschied zu [mcpRuf]**: `t.Context()` endet **vor** den Aufräumern,
	// und `exec.CommandContext` erschlüge den Server damit jedes Mal. Ein
	// Lauf, der mit `t.Fatal` endet, sähe dann „signal: killed" statt des
	// Grundes — und ein Server, der beim Dateiende nicht von selbst geht,
	// fiele gar nicht auf.
	//
	// **Der Riegel bleibt**: [context.CancelFunc] wird als **erster**
	// Aufräumer angemeldet und läuft damit als letzter (LIFO), nach
	// [mcpprozess.beenden].
	ctx, abbrechen := context.WithCancel(context.Background())
	t.Cleanup(abbrechen)

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, gebaut.mcp)
	lauf.Env = umgebung

	hinein, fehler := lauf.StdinPipe()
	if fehler != nil {
		t.Fatalf("die Standardeingabe: %v", fehler)
	}

	heraus, fehler := lauf.StdoutPipe()
	if fehler != nil {
		t.Fatalf("die Standardausgabe: %v", fehler)
	}

	if fehler := lauf.Start(); fehler != nil {
		t.Fatalf("mse-mcp starten: %v", fehler)
	}

	prozess := &mcpprozess{
		t: t, hinein: hinein, heraus: bufio.NewReader(heraus),
		lauf: lauf, abbrechen: abbrechen, naechst: 1,
	}

	t.Cleanup(prozess.beenden)

	prozess.schicke(map[string]any{
		"jsonrpc": "2.0", "id": prozess.kennung(), "method": "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "abnahme", "version": "1"},
		},
	})

	prozess.zeile()

	return prozess
}

// kennung gibt die nächste Rufnummer.
func (p *mcpprozess) kennung() int {
	p.naechst++

	return p.naechst - 1
}

// schicke schreibt einen Ruf und leert den Puffer.
func (p *mcpprozess) schicke(ruf map[string]any) {
	p.t.Helper()

	if _, fehler := p.hinein.Write(append(mussJSON(p.t, ruf), '\n')); fehler != nil {
		p.t.Fatalf("den Ruf schreiben: %v", fehler)
	}
}

// zeile liest eine Antwortzeile.
func (p *mcpprozess) zeile() string {
	p.t.Helper()

	gelesen, fehler := p.heraus.ReadString('\n')
	if fehler != nil {
		p.t.Fatalf("die Antwort lesen: %v — bisher %q", fehler, gelesen)
	}

	return gelesen
}

// ruf ruft ein Werkzeug und gibt seinen Text.
func (p *mcpprozess) ruf(werkzeug string, argumente map[string]any) (
	text string, istFehler bool,
) {
	p.t.Helper()

	p.schicke(map[string]any{
		"jsonrpc": "2.0", "id": p.kennung(), "method": "tools/call",
		"params": map[string]any{"name": werkzeug, "arguments": argumente},
	})

	return p.antwort(werkzeug, p.zeile())
}

// antwort schält den Werkzeugtext aus einer Antwortzeile.
func (p *mcpprozess) antwort(werkzeug, roh string) (text string, istFehler bool) {
	p.t.Helper()

	var gelesen struct {
		Ergebnis struct {
			Inhalt []struct {
				Text string `json:"text"`
			} `json:"content"`
			IstFehler bool `json:"isError"`
		} `json:"result"`
	}

	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		p.t.Fatalf("%s: die Antwort ist kein JSON: %v — %s", werkzeug, fehler, roh)
	}

	if len(gelesen.Ergebnis.Inhalt) == 0 {
		p.t.Fatalf("%s: die Antwort hat keinen Inhalt: %s", werkzeug, roh)
	}

	return gelesen.Ergebnis.Inhalt[0].Text, gelesen.Ergebnis.IstFehler
}

// muss ruft ein Werkzeug und bricht ab, wenn es scheitert.
func (p *mcpprozess) muss(werkzeug string, argumente map[string]any) string {
	p.t.Helper()

	text, istFehler := p.ruf(werkzeug, argumente)
	if istFehler {
		p.t.Fatalf("%s scheiterte: %s", werkzeug, text)
	}

	return text
}

// scheitert ruft ein Werkzeug und verlangt, dass es abweist.
//
// **Mit dem Kern der Meldung**, denn „irgendein Fehler" ist keine Zusage:
// ein Ruf, der an der Kartenkennung scheitert, sähe sonst aus wie einer, den
// die Erlaubnis aufgehalten hat (§10).
func (p *mcpprozess) scheitert(
	werkzeug string, argumente map[string]any, kern string,
) string {
	p.t.Helper()

	text, istFehler := p.ruf(werkzeug, argumente)
	if !istFehler {
		p.t.Fatalf("%s ging durch: %s", werkzeug, text)
	}

	if !strings.Contains(text, kern) {
		p.t.Errorf("%s scheiterte an etwas anderem als %q: %s", werkzeug, kern, text)
	}

	return text
}

// nachfrist ist, wie lange der Server nach dem Dateiende bekommt.
const nachfrist = 30 * time.Second

// werkzeugliste gibt die rohe Antwort auf `tools/list`.
func (p *mcpprozess) werkzeugliste() string {
	p.t.Helper()

	p.schicke(map[string]any{
		"jsonrpc": "2.0", "id": p.kennung(), "method": "tools/list",
		"params": map[string]any{},
	})

	return p.zeile()
}

// beenden schließt die Eingabe und wartet auf das Ende.
//
// **Über das Schließen und nicht über ein Signal**: der Server liest bis zum
// Dateiende und geht dann von selbst — wer ihn erschlüge, sähe nie, ob er
// das tut. Deshalb steht hier eine Frist und kein `Kill` davor.
func (p *mcpprozess) beenden() {
	if fehler := p.hinein.Close(); fehler != nil {
		p.t.Errorf("die Standardeingabe schließen: %v", fehler)
	}

	fertig := make(chan error, 1)
	go func() { fertig <- p.lauf.Wait() }()

	select {
	case fehler := <-fertig:
		if fehler != nil {
			p.t.Errorf("mse-mcp endete mit: %v", fehler)
		}
	case <-time.After(nachfrist):
		p.abbrechen()

		p.t.Errorf("mse-mcp ging nach dem Dateiende nicht von selbst")
	}
}
