package app_test

// **Fertige Konfigurationsbruchstücke für Agenten** — Werkstück A14 von
// Phase 5.3 (#1453).
//
// # Was hier geprüft wird
//
// **Dass das Bruchstück wirklich ausführbar ist und nichts Geheimes trägt.**
// Das Issue verlangt beides wörtlich: „echte ausführbare
// Konfigurationsfragmente für aktuelle Claude-Code-/Codex-Konvention,
// kopierbar und **secretsfrei**" — und in der Abnahme: „enthält weder
// argv-Secret noch rohen gespeicherten Token".

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Für beide Agenten steht ein Bruchstück da.**
func TestFuerBeideAgentenStehtEinBruchstueckDa(t *testing.T) {
	t.Parallel()

	agenten := map[string]bool{}

	for _, stueck := range app.Agentenbruchstuecke("/Anwendungen/MSE.app/Contents/MacOS") {
		agenten[stueck.Agent] = true

		if stueck.Datei == "" || stueck.Inhalt == "" || stueck.Befehl == "" {
			t.Errorf("%s ist unvollständig: %+v", stueck.Agent, stueck)
		}

		// **Die Fundstelle steht dabei** (§10, §16): eine Konvention, die
		// jemand aus dem Gedächtnis aufgeschrieben hat, ist kein
		// Prüfergebnis — und wer sie morgen nachsieht, soll wissen, wo.
		if !strings.Contains(stueck.Nachgelesen, "2026-09-13") {
			t.Errorf("%s nennt kein Abrufdatum: %q", stueck.Agent, stueck.Nachgelesen)
		}
	}

	for _, name := range []string{"claude", "codex"} {
		if !agenten[name] {
			t.Errorf("für %q steht kein Bruchstück da", name)
		}
	}
}

// **Das Bruchstück für Claude Code ist gültiges JSON mit `mcpServers`.**
//
// **Geparst und nicht gelesen** (§10): eine Zeichenkette, die wie JSON
// aussieht, ist keines — und was jemand einfügt, muss der Agent lesen können.
func TestDasBruchstueckFuerClaudeIstGueltigesJSON(t *testing.T) {
	t.Parallel()

	const ort = "/Anwendungen/MSE.app/Contents/MacOS"

	stueck := bruchstueckVon(t, ort, "claude")

	var gelesen struct {
		Server map[string]struct {
			Befehl string   `json:"command"`
			Args   []string `json:"args"`
		} `json:"mcpServers"`
	}

	if fehler := json.Unmarshal([]byte(stueck.Inhalt), &gelesen); fehler != nil {
		t.Fatalf("das Bruchstück ist kein JSON: %v\n%s", fehler, stueck.Inhalt)
	}

	eintrag, da := gelesen.Server["magicseteditor4mac"]
	if !da {
		t.Fatalf("der Eintrag fehlt: %s", stueck.Inhalt)
	}

	if eintrag.Befehl != ort+"/mse-mcp" {
		t.Errorf("der Befehl ist %q", eintrag.Befehl)
	}
}

// **Und das für Codex nennt die Tabelle, die Codex erwartet.**
func TestDasBruchstueckFuerCodexNenntSeineTabelle(t *testing.T) {
	t.Parallel()

	stueck := bruchstueckVon(t, "/Anwendungen/MSE.app/Contents/MacOS", "codex")

	for _, teil := range []string{
		"[mcp_servers.magicseteditor4mac]",
		`command = "/Anwendungen/MSE.app/Contents/MacOS/mse-mcp"`,
	} {
		if !strings.Contains(stueck.Inhalt, teil) {
			t.Errorf("das Bruchstück nennt %q nicht:\n%s", teil, stueck.Inhalt)
		}
	}
}

// **Kein Bruchstück trägt ein Geheimnis.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil ein Schlüssel in einer Konfigurationsdatei dort für immer liegt** —
// im Klartext, in einer Datei, die jemand in ein Repositorium legt. Dieses
// Programm hält Zugangsdaten im Schlüsselbund von macOS, und `mse-mcp` liest
// sie **selbst**; ein `env`-Block mit einem Wert darin wäre der eine Weg,
// diese Entscheidung zu unterlaufen.
func TestKeinBruchstueckTraegtEinGeheimnis(t *testing.T) {
	t.Parallel()

	verdaechtig := []string{
		"env", "API_KEY", "TOKEN", "SCHLUESSEL", "sk-", "Bearer", "secret",
	}

	for _, stueck := range app.Agentenbruchstuecke("/Anwendungen/MSE.app/Contents/MacOS") {
		roh := strings.ToUpper(stueck.Inhalt + " " + stueck.Befehl)

		for _, wort := range verdaechtig {
			if strings.Contains(roh, strings.ToUpper(wort)) {
				t.Errorf("%s trägt %q:\n%s", stueck.Agent, wort, stueck.Inhalt)
			}
		}
	}
}

// **Und die Technikangaben führen sie mit.**
func TestDieTechnikangabenFuehrenDieBruchstuecke(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Technikangaben(probekontext(t), "")
	if antwort.Fehler != nil {
		t.Fatalf("Technikangaben: %+v", antwort.Fehler)
	}

	if len(antwort.Agenten) != 2 {
		t.Fatalf("es stehen %d Bruchstücke da: %+v",
			len(antwort.Agenten), antwort.Agenten)
	}
}

// bruchstueckVon sucht das eines Agenten.
func bruchstueckVon(t *testing.T, ort, agent string) app.AgentenbruchstueckDTO {
	t.Helper()

	for _, stueck := range app.Agentenbruchstuecke(ort) {
		if stueck.Agent == agent {
			return stueck
		}
	}

	t.Fatalf("für %q steht kein Bruchstück da", agent)

	return app.AgentenbruchstueckDTO{}
}
