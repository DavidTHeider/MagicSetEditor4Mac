// **Die Problemprüfung über ein ganzes Set, über jede Fläche ohne Fenster** —
// PERF02 (#2352).
//
// # Was die Läufe halten
//
// Drei Prozesse prüfen dasselbe Set und nennen **dieselbe Zahl**. Vor einer
// Prüfung sagt jede Fläche »ungeprüft« und **keine sichere Null** — das ist
// der Befund aus dem Issue. Und weil der Zähler im Bestand liegt und nicht im
// Prozess, macht eine Karte, die **eine andere** Fläche anlegt, den Stand des
// Dienstes veraltet.
//
// **Was sie nicht halten**: ob die Befunde einer Karte die richtigen sind —
// das ist der Prüfer, und er ist derselbe wie die Vorschau
// (`internal/app/kartenpruefer.go`). Das Fenster fehlt, bis AP-8
// (#2365) die
// Zahl auf der Übersicht zeigt.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// problemstand ist, was jede Fläche über den Stand sagt.
type problemstand struct {
	Urteil   string `json:"urteil"`
	Zahl     int    `json:"zahl"`
	Sicher   bool   `json:"sicher"`
	Geprueft int    `json:"geprueft"`
	Gesamt   int    `json:"gesamt"`
}

// problemstandAus liest einen Stand aus einer Antwort.
func problemstandAus(t *testing.T, wer, roh string) problemstand {
	t.Helper()

	var stand problemstand
	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("%s gibt keinen Stand: %v\n%s", wer, err, roh)
	}

	return stand
}

func TestDieProblempruefungUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)

	agent := starteMCPProzess(t, umgebung)
	an := starteDienst(t, umgebung)
	argumente := map[string]any{"setID": zahl(t, setID)}

	// **Vor der Prüfung keine sichere Null** — an beiden langlebigen Flächen.
	text, istFehler := agent.ruf("lesen.set.probleme", argumente)
	if vorher := problemstandAus(t, "MCP", text); istFehler || vorher.Urteil != "ungeprueft" || vorher.Sicher {
		t.Errorf("MCP vor der Prüfung: %+v (%s)", vorher, text)
	}

	status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/probleme", "")
	if vorher := problemstandAus(t, "der Dienst", rumpf); status != http.StatusOK || vorher.Urteil != "ungeprueft" || vorher.Sicher {
		t.Errorf("der Dienst vor der Prüfung: %d %+v", status, vorher)
	}

	// **Dreimal geprüft, dreimal dasselbe.**
	cli := problemstandAus(t, "die Kommandozeile", mussKommandozeile(t, umgebung, "probleme", setID, "--json"))

	text, istFehler = agent.ruf("lesen.set.probleme.pruefen", argumente)
	mcp := problemstandAus(t, "MCP", text)

	status, rumpf = an.ruf(t, http.MethodPost, "/sets/"+setID+"/probleme", "")
	rest := problemstandAus(t, "der Dienst", rumpf)

	if istFehler || status != http.StatusOK {
		t.Fatalf("die Prüfung scheiterte: MCP %s, Dienst %d", text, status)
	}

	for wer, stand := range map[string]problemstand{"Kommandozeile": cli, "MCP": mcp, "Dienst": rest} {
		if stand.Urteil != "aktuell" || !stand.Sicher || stand.Gesamt == 0 || stand.Geprueft != stand.Gesamt {
			t.Errorf("%s nach der Prüfung: %+v", wer, stand)
		}

		if stand.Zahl != cli.Zahl || stand.Gesamt != cli.Gesamt {
			t.Errorf("%s zählt %d von %d, die Kommandozeile %d von %d", wer, stand.Zahl, stand.Gesamt, cli.Zahl, cli.Gesamt)
		}
	}

	// **Der Stand bleibt am Dienst stehen** — und veraltet, wenn eine andere
	// Fläche eine Karte anlegt. Ein Zähler im Prozess sähe das nicht.
	mussKommandozeile(t, umgebung, "karte", setID, "anlegen")

	status, rumpf = an.ruf(t, http.MethodGet, "/sets/"+setID+"/probleme", "")
	if danach := problemstandAus(t, "der Dienst", rumpf); status != http.StatusOK || danach.Urteil != "veraltet" || danach.Sicher {
		t.Errorf("der Dienst nach der neuen Karte: %d %+v", status, danach)
	}
}

// **Ein Set, das es nicht gibt, wird überall gleich abgewiesen** — der
// Fehlerfall neben dem Erfolg (§16).
func TestDieProblempruefungWeistEinFehlendesSetAb(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "die Prüfung eines fehlenden Sets", abweisung{
		kern:   "gibt es nicht",
		klasse: klasseNichtGefunden,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "probleme", "999")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "lesen.set.probleme.pruefen", map[string]any{"setID": 999})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/sets/999/probleme", "")
		},
		ohneFenster: "die Bindung kommt mit AP-8 (#2365), das die Zahl auf der Übersicht zeigt",
	})
}
