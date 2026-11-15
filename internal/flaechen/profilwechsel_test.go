// **Der Profilwechsel trennt wirklich — an jeder Fläche** — Werkstück 3 von
// Phase 4.12
// (#1351).
//
// # Was hier geprüft wird und was nicht
//
// **Nicht, dass es die Fähigkeit gibt** — das hält
// `internal/werkzeuge/faehigkeiten_test.go`. Hier läuft derselbe
// Arbeitsablauf über jede Fläche und muss **dasselbe** ergeben (§16, letzter
// Absatz): ein Set anlegen, wechseln, und es ist weg; zurückwechseln, und es
// ist wieder da.
//
// # Und warum es **zwei Profile desselben Nutzers** sind
//
// **Weil dieser Lauf den Wechsel prüft und nicht die Erlaubnis.** Seit
// Werkstück 11 (#1351) verlangt ein Wechsel in ein **fremdes** Profil eine
// Freigabe; bis dahin standen hier zwei Nutzer, und der Lauf wurde rot — nicht
// an einem Fehler, sondern an der Zusage, die das Werkstück einlöst.
//
// **Dass die Erlaubnis wirkt, prüft `freigaben_test.go`** — an derselben
// Naht, mit demselben Ablauf, und mit der Ablehnung als Ergebnis. Zwei
// Aussagen, zwei Läufe: sonst sagt ein rotes Ergebnis nicht, welche der
// beiden gebrochen ist.
//
// **Das ist zugleich die Abnahme aus dem Issue** — „was das eine anlegt, sieht
// das andere nicht" —, allerdings nur ihre erste Hälfte. Die zweite („auch
// dann nicht, wenn man die Bestandsdatei mit einem fremden Werkzeug öffnet")
// leistet nur die Verschlüsselung, und Vermerk 0066 sagt das ausdrücklich.
package flaechen_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestDerProfilwechselTrenntAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	mussKommandozeile(t, umgebung, "set-neu", "Erstes", "probe", "probe-schlicht")

	if aus := mussKommandozeile(t, umgebung, "sets"); !strings.Contains(aus, "Erstes") {
		t.Fatalf("das Set fehlt vor dem Wechsel: %q", aus)
	}

	erstes := profilkennung(t, mussKommandozeile(t, umgebung, "profil", "--json"))
	zweites := zweitesProfil(t, umgebung)

	mussKommandozeile(t, umgebung, "profil", zweites)

	// **Und das ist die Trennung**: dieselbe Abfrage, ein anderes Ergebnis.
	if aus := mussKommandozeile(t, umgebung, "sets"); strings.Contains(aus, "Erstes") {
		t.Errorf("das Set des anderen Profils steht da: %q", aus)
	}

	// **Ein Wechsel löscht nicht.**
	mussKommandozeile(t, umgebung, "profil", erstes)

	if aus := mussKommandozeile(t, umgebung, "sets"); !strings.Contains(aus, "Erstes") {
		t.Errorf("das Set kam nicht zurück: %q", aus)
	}

	// **Und `--profil` gilt für genau einen Aufruf.**
	if aus := mussKommandozeile(t, umgebung, "--profil="+zweites, "sets"); strings.Contains(aus, "Erstes") {
		t.Errorf("--profil wirkte nicht: %q", aus)
	}

	if aus := mussKommandozeile(t, umgebung, "sets"); !strings.Contains(aus, "Erstes") {
		t.Errorf("--profil hat den Zeiger verstellt: %q", aus)
	}
}

// **Derselbe Weg über MCP.**
func TestDerProfilwechselTrenntUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	mussMCP(t, umgebung, "schreiben.set.neu",
		map[string]any{"name": "Erstes", "spiel": "probe", "stil": "probe-schlicht"})

	if aus := mussMCP(t, umgebung, "lesen.sets", map[string]any{}); !strings.Contains(aus, "Erstes") {
		t.Fatalf("das Set fehlt vor dem Wechsel: %s", aus)
	}

	zweites := zweitesProfilUeberMCP(t, umgebung)

	mussMCP(t, umgebung, "schreiben.profil.wechseln",
		map[string]any{"profilID": zweites})

	if aus := mussMCP(t, umgebung, "lesen.sets", map[string]any{}); strings.Contains(aus, "Erstes") {
		t.Errorf("das Set des anderen Profils steht da: %s", aus)
	}

	// **Und `lesen.profil` sagt, wo man steht.**
	if aus := mussMCP(t, umgebung, "lesen.profil", map[string]any{}); !strings.Contains(aus, zweites) {
		t.Errorf("lesen.profil nennt das gewechselte Profil nicht: %s", aus)
	}
}

// **Und beides in **einer** Sitzung** — Werkstück 3 (#1351).
//
// # Warum dieser Lauf eigens dasteht
//
// **Weil `mcpRuf` je Ruf einen eigenen Prozess startet.** Damit liest jeder
// Ruf das Profil beim Anlaufen — und der Lauf darüber wäre grün, auch wenn
// der Server es **einmal** beim Start läse. Genau so war es: der Wechsel
// wirkte, sobald der nächste Prozess anlief, und in einer echten Sitzung nie.
//
// **Eine Vorrichtung, die großzügiger ist als der Betrieb** (§3) — und dies
// ist ihr Gegenstück.
func TestEinProfilwechselWirktInDerselbenMCPSitzung(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	mussMCP(t, umgebung, "schreiben.set.neu",
		map[string]any{"name": "Erstes", "spiel": "probe", "stil": "probe-schlicht"})

	zweites := zweitesProfilUeberMCP(t, umgebung)

	antworten := mcpSitzung(t, umgebung,
		mcpschritt{"schreiben.profil.wechseln", map[string]any{"profilID": zweites}},
		mcpschritt{"lesen.sets", map[string]any{}},
	)

	if strings.Contains(antworten[1], "Erstes") {
		t.Errorf("das Set des anderen Profils steht da: %s", antworten[1])
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerProfilwechselTrenntUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	mussREST(t, an, http.MethodPost, "/sets/neu", `{"name":"Erstes","spiel":"probe","stil":"probe-schlicht"}`)

	if aus := mussREST(t, an, http.MethodGet, "/sets", ""); !strings.Contains(aus, "Erstes") {
		t.Fatalf("das Set fehlt vor dem Wechsel: %s", aus)
	}

	// **Ein zweites Profil DESSELBEN Nutzers** — siehe den Kopf dieser Datei.
	eigner := eignerkennung(t, mussREST(t, an, http.MethodGet, "/profil", ""))
	angelegt := mussREST(t, an, http.MethodPost,
		"/nutzer/"+eigner+"/profile", `{"name":"Zweites"}`)
	zweites := angelegtesProfilAus(t, angelegt)

	mussREST(t, an, http.MethodPut, "/profil", `{"profilID":"`+zweites+`"}`)

	if aus := mussREST(t, an, http.MethodGet, "/sets", ""); strings.Contains(aus, "Erstes") {
		t.Errorf("das Set des anderen Profils steht da: %s", aus)
	}

	if aus := mussREST(t, an, http.MethodGet, "/profil", ""); !strings.Contains(aus, zweites) {
		t.Errorf("GET /profil nennt das gewechselte Profil nicht: %s", aus)
	}
}

// profilkennung liest die Kennung aus einer Profilantwort.
func profilkennung(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Profil struct {
			ID string `json:"id"`
		} `json:"profil"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Profilantwort lesen: %v — %s", err, roh)
	}

	return gelesen.Profil.ID
}

// zweitesProfil legt an der Kommandozeile ein zweites Profil **desselben**
// Nutzers an.
func zweitesProfil(t *testing.T, umgebung []string) string {
	t.Helper()

	eigner := eignerkennung(t, mussKommandozeile(t, umgebung, "profil", "--json"))

	return angelegtesProfilAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", eigner, "Zweites", "--json"))
}

// zweitesProfilUeberMCP legt über MCP ein zweites Profil desselben Nutzers an.
func zweitesProfilUeberMCP(t *testing.T, umgebung []string) string {
	t.Helper()

	eigner := eignerkennung(t, mussMCP(t, umgebung, "lesen.profil", map[string]any{}))

	return angelegtesProfilAus(t, mussMCP(t, umgebung, "schreiben.profil.neu",
		map[string]any{"nutzerID": eigner, "name": "Zweites"}))
}

// eignerkennung liest aus einer Profilantwort, **wem** das Profil gehört.
func eignerkennung(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Profil struct {
			NutzerID string `json:"nutzerID"`
		} `json:"profil"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Profilantwort lesen: %v — %s", err, roh)
	}

	if gelesen.Profil.NutzerID == "" {
		t.Fatalf("die Profilantwort nennt keinen Nutzer: %s", roh)
	}

	return gelesen.Profil.NutzerID
}

// angelegtesProfilAus liest die Kennung aus der Antwort auf eine Profilanlage.
func angelegtesProfilAus(t *testing.T, roh string) string {
	t.Helper()

	kennung := profilkennung(t, roh)
	if kennung == "" {
		t.Fatalf("die Anlage nennt kein Profil: %s", roh)
	}

	return kennung
}

// mcpschritt ist ein Ruf innerhalb einer Sitzung.
type mcpschritt struct {
	werkzeug  string
	argumente map[string]any
}

// mcpSitzung fährt mehrere Rufe über **einen** Prozess.
//
// **Der Unterschied zu [mcpRuf] ist der ganze Zweck**: dort ist jeder Ruf ein
// eigener Prozess, hier teilen sich alle einen — so wie bei einem
// Sprachmodell, das eine Sitzung offen hält.
func mcpSitzung(t *testing.T, umgebung []string, schritte ...mcpschritt) []string {
	t.Helper()

	const anmeldung = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":` +
		`{"protocolVersion":"2025-06-18","capabilities":{},` +
		`"clientInfo":{"name":"abnahme","version":"1"}}}` + "\n"

	eingabe := anmeldung

	for nummer, schritt := range schritte {
		ruf := mussJSON(t, map[string]any{
			"jsonrpc": "2.0", "id": nummer + 2, "method": "tools/call",
			"params": map[string]any{
				"name": schritt.werkzeug, "arguments": schritt.argumente,
			},
		})

		eingabe += string(ruf) + "\n"
	}

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(t.Context(), gebaut.mcp)
	lauf.Env = umgebung
	lauf.Stdin = strings.NewReader(eingabe)

	var aus bytes.Buffer
	lauf.Stdout = &aus

	if err := lauf.Run(); err != nil {
		t.Fatalf("mse-mcp: %v", err)
	}

	zeilen := strings.Split(strings.TrimSpace(aus.String()), "\n")
	if len(zeilen) != len(schritte)+1 {
		t.Fatalf("%d Antworten auf %d Rufe: %s", len(zeilen), len(schritte), aus.String())
	}

	// **Die erste Zeile ist die Anmeldung** — sie gehört zu keinem Schritt.
	return zeilen[1:]
}
