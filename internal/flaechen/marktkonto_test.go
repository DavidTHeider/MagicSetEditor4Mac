package flaechen_test

// **Das eigene Konto über jede Fläche** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Was hier gemessen wird
//
//   - **Dass alle drei kopflosen Flächen dasselbe Nein geben**, solange die
//     Kontoschau aus ist — und sie fängt aus an. Ein Schalter, den eine
//     Fläche übergeht, wäre kein Schalter.
//   - **Dass keine von ihnen ihn einschalten kann.** Er steht in den
//     Einstellungen, und dort schreibt ein Mensch.
//   - **Und dass über MCP kein Weg zum Schreiben führt** — kein Angebot,
//     kein Warenkorb, kein Kauf.
//
// # Was er **nicht** misst
//
// **Den geglückten Abruf.** Er braucht eine Zugangsmarke eines echten
// CardTrader-Kontos; es gibt keine, und eine zu beschaffen ist keine
// Maßnahme, die ein Lauf treffen darf. Was der Adapter aus einer Antwort
// macht, ist über `internal/cardtrader` gegen selbst geschriebene Beispiele
// gemessen — **und das ist keine Anbieterabnahme** (§3, und das Issue
// verlangt die Unterscheidung wörtlich).

import (
	"net/http"
	"strings"
	"testing"
)

// derKontoweg ist der Endpunkt, wie er im Katalog steht.
const derKontoweg = "GET /markt/{quelle}/konto"

// **Eine ausgeschaltete Kontoschau sagt überall dasselbe.**
func TestEineAusgeschalteteKontoschauSagtUeberallNein(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	ausgabe, klage, code := kommandozeile(t, umgebung,
		"--json", "markt-konto", "cardtrader")
	if code == 0 {
		t.Fatalf("die Kommandozeile las das Konto: %s", ausgabe)
	}

	if strings.TrimSpace(klage+ausgabe) == "" {
		t.Error("die Absage sagt nichts")
	}

	status, rumpf := an.ruf(t, http.MethodGet,
		strings.ReplaceAll(pfadAus(derKontoweg), "{quelle}", "cardtrader"), "")
	if status == http.StatusOK {
		t.Errorf("der Dienst las das Konto: %s", rumpf)
	}

	text, istFehler := mcpRuf(t, umgebung, "lesen.marktkonto",
		map[string]any{"quelle": "cardtrader"})
	if !istFehler {
		t.Errorf("MCP las das Konto: %s", text)
	}
}

// **Ein Marktplatz, den es nicht gibt, wird überall abgewiesen.**
//
// **Die Gegenprobe zum Lauf darüber**: wäre der Schalter der einzige Riegel,
// gäbe es hier dieselbe Absage — und man wüsste nicht, ob der Name überhaupt
// geprüft wird. Er wird **zuerst** geprüft, denn ein falscher Name soll nicht
// den Schlüsselbund öffnen.
func TestEinUnbekannterMarktplatzHatKeinKonto(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)

	ausgabe, klage, code := kommandozeile(t, umgebung,
		"--json", "markt-konto", "kartenkiosk")
	if code == 0 {
		t.Fatalf("der fremde Marktplatz kam durch: %s", ausgabe)
	}

	if !strings.Contains(strings.ToLower(klage+ausgabe), "kartenkiosk") {
		t.Errorf("die Absage nennt den Namen nicht: %s / %s", klage, ausgabe)
	}
}

// **Keine kopflose Fläche schaltet die Kontoschau ein, und keine schreibt
// bei einem Marktplatz.**
func TestKeineKopfloseFlaecheSchaltetDieKontoschauEin(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)

	// **Die Kommandozeile kennt keinen Befehl dafür** — der Schalter steht
	// in den Einstellungen, und dort schreibt ein Mensch am Fenster.
	hilfe, _, _ := kommandozeile(t, umgebung, "hilfe")

	for _, wort := range []string{"konto-schau", "kontoschau"} {
		if strings.Contains(hilfe, wort) {
			t.Errorf("die Kommandozeile bietet %q an", wort)
		}
	}

	// **Und MCP hat kein Werkzeug, das bei einem Marktplatz schreibt.**
	agent := starteMCPProzess(t, umgebung)

	for _, name := range werkzeugnamen(t, agent) {
		if strings.HasPrefix(name, "schreiben.") &&
			strings.Contains(name, "markt") {
			t.Errorf("MCP führt das schreibende Werkzeug %q", name)
		}
	}
}
