// **Die Verschlüsselung des Bestands über die Flächen** — Werkstück 14 von
// Phase 4.12 (#1351),
// nach [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// # Was hier gleich sein muss und was nicht
//
// **Gleich ist die Auskunft**: alle drei kopflosen Flächen sagen dasselbe über
// denselben Bestand, und sie sagen es nach dem Umschalten **anders** als davor.
//
// **Nicht gleich ist, wer schalten darf.** Der Dienst und der MCP-Server halten
// eine Verbindung über ihre ganze Laufzeit, an der jede Naht hängt; ein Tausch
// der Bestandsdatei ließe sie alle auf einem Griff sitzen, den es nicht mehr
// gibt. Das ist derselbe Grund, aus dem dort auch das Zurückspielen fehlt
// (`docs/FAEHIGKEITEN.md`), und der Lauf hält ihn fest: **beide dürfen lesen,
// keiner darf schalten.**
package flaechen_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"testing"
)

// TestDieVerschluesselungWirdUeberJedeFlaecheGelesen fährt denselben Ablauf
// über Kommandozeile, MCP und Dienst — vor und nach dem Umschalten.
func TestDieVerschluesselungWirdUeberJedeFlaecheGelesen(t *testing.T) {
	// **Eigener Bestand und eigener Schlüsselbund** (§3). Beides ist hier
	// keine Hygiene, sondern Notwehr: dieser Lauf **verschlüsselt** einen
	// Bestand und legt einen Schlüssel ab. Ohne die Trennung träfe er den
	// Bestand dieses Rechners und den Anmeldebund seines Menschen.
	umgebung := bestandMitBund(t)

	// **Vorher: alle drei sagen Klartext.**
	//
	// **In einem Unterlauf**, weil `starteDienst` den Dienst bis zum Ende des
	// Tests hält — und ein gehaltener Bestand lässt sich nicht tauschen. Der
	// Unterlauf gibt ihn frei, bevor geschaltet wird. Dass das Schalten
	// **daran** scheitert, prüft der Lauf unten ausdrücklich.
	t.Run("vorher im Klartext", func(t *testing.T) {
		pruefeAlleDrei(t, umgebung, false)
	})

	// **Geschaltet wird an der einen Fläche, die es darf.**
	aus := mussKommandozeile(t, umgebung, "bestand-verschluesselung", "an")
	if !strings.Contains(aus, "verschlüsselt") {
		t.Fatalf("die Kommandozeile bestätigt es nicht: %s", aus)
	}

	// **Und der Satz über den Schlüsselbund gehört dazu** — wer hier schaltet
	// und nicht weiß, dass der Schlüssel dorthin geht, verliert seine Daten
	// beim nächsten Rechnerwechsel.
	if !strings.Contains(aus, "Schlüsselbund") {
		t.Errorf("die Bestätigung nennt den Schlüsselbund nicht: %s", aus)
	}

	// **Nachher: alle drei sagen verschlüsselt** — über einen Bestand, den
	// keine von ihnen selbst umgeschrieben hat.
	t.Run("nachher verschlüsselt", func(t *testing.T) {
		pruefeAlleDrei(t, umgebung, true)
	})

	// **Und zurück geht es auch.** Eine Entscheidung, die sich nicht
	// zurücknehmen lässt, ist keine.
	zurueck := mussKommandozeile(t, umgebung, "bestand-verschluesselung", "aus")
	if !strings.Contains(zurueck, "Klartext") {
		t.Fatalf("das Ausschalten bestätigt es nicht: %s", zurueck)
	}

	t.Run("und wieder im Klartext", func(t *testing.T) {
		pruefeAlleDrei(t, umgebung, false)
	})
}

// **Und geschaltet wird nicht, solange jemand anders den Bestand hält.**
//
// # Warum dieser Fall hier steht
//
// **Weil er beim Schreiben des Laufs darüber zugeschlagen hat.** Der erste
// Anlauf prüfte alle drei Flächen und schaltete danach — und `starteDienst`
// hält den Dienst bis zum Ende des Tests. Das Umschalten scheiterte mit
// „bestand.db ist von einem anderen Programm geöffnet".
//
// Das war kein Fehler des Laufs, sondern die Zusage von `AlleinAmBestand`:
// eine Datei umzubenennen, die ein anderer Prozess geöffnet hält, nennt SQLite
// undefiniert und korruptionsgefährlich (#961).
// Ein Zufallsfund, der ungeprüft bliebe, wäre beim nächsten Umbau weg.
func TestGeschaltetWirdNichtWaehrendEinAndererDenBestandHaelt(t *testing.T) {
	umgebung := bestandMitBund(t)

	// **Der Dienst hält ihn ab hier** — und `t.Cleanup` gibt ihn erst am Ende
	// zurück.
	starteDienst(t, umgebung)

	aus, fehler, code := kommandozeile(t, umgebung, "bestand-verschluesselung", "an")
	if code == 0 {
		t.Fatalf("das Umschalten glückte, obwohl der Dienst den Bestand hält: %s", aus)
	}

	gesagt := aus + fehler
	if !strings.Contains(gesagt, "anderen Programm") {
		t.Errorf("die Meldung sagt nicht, was los ist: %s", gesagt)
	}
}

// pruefeAlleDrei fragt denselben Stand an drei Flächen ab.
func pruefeAlleDrei(t *testing.T, umgebung []string, erwartet bool) {
	t.Helper()

	// Kommandozeile — ohne Wort zeigt sie nur.
	aus := mussKommandozeile(t, umgebung, "bestand-verschluesselung")

	if strings.Contains(aus, "ist verschlüsselt") != erwartet {
		t.Errorf("die Kommandozeile sagt %q, erwartet war verschlüsselt=%v",
			strings.TrimSpace(aus), erwartet)
	}

	// MCP.
	roh := mussMCP(t, umgebung, "lesen.bestand.verschluesselung", map[string]any{})

	var ausMCP struct {
		Verschluesselt bool `json:"verschluesselt"`
	}

	if fehler := json.Unmarshal([]byte(roh), &ausMCP); fehler != nil {
		t.Fatalf("MCP-Antwort lesen: %v (%s)", fehler, roh)
	}

	if ausMCP.Verschluesselt != erwartet {
		t.Errorf("MCP sagt %v, erwartet war %v", ausMCP.Verschluesselt, erwartet)
	}

	// Dienst.
	an := starteDienst(t, umgebung)

	antwort := mussREST(t, an, http.MethodGet, "/bestand/verschluesselung", "")

	var ausREST struct {
		Verschluesselt bool `json:"verschluesselt"`
	}

	if fehler := json.Unmarshal([]byte(antwort), &ausREST); fehler != nil {
		t.Fatalf("REST-Antwort lesen: %v (%s)", fehler, antwort)
	}

	if ausREST.Verschluesselt != erwartet {
		t.Errorf("der Dienst sagt %v, erwartet war %v", ausREST.Verschluesselt, erwartet)
	}
}

// **Und geschaltet wird an den beiden langlebigen Flächen nicht** — es gibt
// dort kein Werkzeug und keinen Weg dafür.
//
// **Ein Lauf über die Abwesenheit**, und der ist selten richtig. Hier ist er
// es: die Begründung im Katalog ist eine **Zusage**, und ohne diesen Lauf
// stünde sie da, während ein späterer Handgriff sie stillschweigend aufhebt.
func TestGeschaltetWirdWederUeberMCPNochUeberDenDienst(t *testing.T) {
	umgebung := bestandMitBund(t)

	// **Nicht über `mcpRuf`**, denn das ruft ein Werkzeug und erwartet eine
	// Werkzeugantwort. Ein Werkzeug, das es **nicht gibt**, beantwortet der
	// Server eine Ebene tiefer — als Protokollfehler ohne Inhalt.
	// **Der Name steht zusammengesetzt da und nicht als Literal.**
	//
	// `internal/werkzeuge/flaechenlauf_test.go` liest jede Werkzeugkennung im
	// Quelltext dieses Verzeichnisses und verlangt, dass sie im
	// Fähigkeitenkatalog steht. Hier steht sie gerade **nicht** dort — und
	// genau das ist der Gegenstand dieses Laufs. Ein Literal machte den
	// Wächter rot und sagte damit das Gegenteil dessen, was hier geprüft wird.
	fehlend := "schreiben.bestand." + "verschluesselung"

	roh := mcpRoh(t, umgebung,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":`+
			`{"name":"`+fehlend+`","arguments":{"an":true}}}`)

	if !strings.Contains(roh, "unbekanntes Werkzeug") {
		t.Errorf("MCP kennt ein Schaltwerkzeug: %s", roh)
	}

	// **Und der Beleg, dass die Probe überhaupt etwas trifft** (§10): dasselbe
	// mit einem Werkzeug, das es **gibt**, darf diese Meldung nicht bringen.
	vorhanden := mcpRoh(t, umgebung,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":`+
			`{"name":"lesen.bestand.verschluesselung","arguments":{}}}`)

	if strings.Contains(vorhanden, "unbekanntes Werkzeug") {
		t.Fatalf("auch das lesende Werkzeug fehlt — die Probe misst nichts: %s", vorhanden)
	}

	an := starteDienst(t, umgebung)

	for _, methode := range []string{http.MethodPut, http.MethodPost} {
		status, antwort := an.ruf(t, methode, "/bestand/verschluesselung",
			`{"verschluesselt":true}`)
		if status < 400 {
			t.Errorf("%s /bestand/verschluesselung: %d (%s)", methode, status, antwort)
		}
	}
}

// mcpRoh fährt einen einzelnen Ruf gegen den MCP-Server und gibt, was
// herauskommt — **auch einen Protokollfehler**.
func mcpRoh(t *testing.T, umgebung []string, ruf string) string {
	t.Helper()

	const anmeldung = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":` +
		`{"protocolVersion":"2025-06-18","capabilities":{},` +
		`"clientInfo":{"name":"abnahme","version":"1"}}}` + "\n"

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(t.Context(), gebaut.mcp)
	lauf.Env = umgebung
	lauf.Stdin = strings.NewReader(anmeldung + ruf + "\n")

	var aus bytes.Buffer
	lauf.Stdout = &aus

	if err := lauf.Run(); err != nil {
		t.Fatalf("mse-mcp: %v", err)
	}

	zeilen := strings.Split(strings.TrimSpace(aus.String()), "\n")

	return zeilen[len(zeilen)-1]
}
