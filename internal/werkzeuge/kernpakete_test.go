// **Jedes Paket unter `internal/` ist entschieden** —
// #1270.
//
// # Wofür dieser Wächter da ist
//
// `scripts/ci-abdeckung.sh` führt in `KERNPAKETE` die Pakete mit harter
// 100-Prozent-Schwelle. Die Liste ist von Hand geschrieben, und ein neues Paket
// steht nicht darin — es fällt in die Gesamtschwelle von 90 % und ist damit
// **still** schwächer geprüft als seine Nachbarn.
//
// **Das ist nicht theoretisch.** `internal/sicherung` entstand am 2026-08-26
// und stand zwei Tage lang nicht in der Liste; aufgefallen ist es erst, als
// `internal/vorschauweg` denselben Weg nahm. Zwei Fälle derselben Klasse — und
// die Klasse ist: *eine Aufzählung, an die jemand denken muss.*
//
// # Wie er entscheidet
//
// Jedes Verzeichnis unter `internal/` steht entweder in `KERNPAKETE` **oder**
// in [ohneHarteSchwelle] mit einem Grund. Ein drittes gibt es nicht.
//
// # Was er nicht prüft
//
//   - **Ob die Schwelle eingehalten wird.** Das tut `scripts/ci-abdeckung.sh`
//     selbst, bei jedem Lauf.
//   - **Ob der Grund der richtige ist.** Er sieht, dass einer dasteht — wie
//     `covignore_pruefen` und aus demselben Grund.
package werkzeuge_test

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// ohneHarteSchwelle sind die Pakete ohne 100-Prozent-Schwelle — jedes mit
// seinem Grund. Ein leerer Grund ist keiner.
var ohneHarteSchwelle = map[string]string{
	"mcpnaehte": "die 42 Adapter des MCP-Servers, bis U12 in `cmd/mse-mcp`: dünne Nähte auf " +
		"`app`, geprüft dort, wo sie wirken — in `internal/flaechen` gegen das gebaute " +
		"Programm und in `internal/app` an der Fachlogik dahinter",

	"schriften": "sucht Schriftdateien auf **diesem** Rechner; welche Zweige " +
		"laufen, hängt davon ab, was installiert ist — die Ausnahmen stehen " +
		"einzeln in .covignore",

	"dateisperre": "ein Aufsatz auf `flock`, in drei Fassungen je Plattform. " +
		"Er hat keine eigenen Tests: geprüft wird er dort, wo er wirkt — in " +
		"`storage`, `netz`, `msepaket`, `service` und `scryfall`, jeweils am " +
		"Wettlauf, den er verhindert",

	"flaechen": "reines Abnahmepaket ohne Produktivcode: es fährt denselben " +
		"Arbeitsablauf über Fenster, Kommandozeile und MCP (§16). Es hat " +
		"keine eigenen Anweisungen",

	"werkzeuge": "die Wächter selbst — nichts davon wird ausgeliefert",

	"schichten": "der Wächter über die Schichtenregel, ebenso",
}

// kernpaketzeile findet `KERNPAKETE="…"` in `scripts/ci-abdeckung.sh`.
var kernpaketzeile = regexp.MustCompile(`(?m)^KERNPAKETE="([^"]*)"`)

func TestJedesPaketHatEineSchwelleOderEinenGrund(t *testing.T) {
	t.Parallel()

	treffer := kernpaketzeile.FindStringSubmatch(lies(t, "scripts/ci-abdeckung.sh"))
	if treffer == nil {
		t.Fatal("scripts/ci-abdeckung.sh nennt kein KERNPAKETE — heißt die " +
			"Veränderliche noch so?")
	}

	kern := map[string]bool{}
	for _, name := range strings.Fields(treffer[1]) {
		kern[name] = true
	}

	if len(kern) == 0 {
		t.Fatal("KERNPAKETE ist leer — dann prüft die Abdeckungsstufe nichts")
	}

	var ohne []string

	for _, paket := range paketeUnterInternal(t) {
		grund, ausgenommen := ohneHarteSchwelle[paket]

		switch {
		case kern[paket] && ausgenommen:
			t.Errorf("internal/%s steht in KERNPAKETE **und** in "+
				"ohneHarteSchwelle — eines von beiden ist falsch", paket)
		case kern[paket]:
		case ausgenommen && strings.TrimSpace(grund) == "":
			t.Errorf("internal/%s ist ausgenommen, ohne dass ein Grund "+
				"dasteht — eine Ausnahme ohne Grund ist eine stille Absenkung", paket)
		case ausgenommen:
		default:
			ohne = append(ohne, paket)
		}
	}

	if len(ohne) > 0 {
		sort.Strings(ohne)
		t.Errorf("diese Pakete stehen weder in KERNPAKETE noch in "+
			"ohneHarteSchwelle: %s — jedes fällt damit still in die "+
			"Gesamtschwelle von 90 %%", strings.Join(ohne, ", "))
	}

	// **Und die Ausnahmeliste führt kein Gespenst.** Ein Eintrag für ein
	// Paket, das es nicht mehr gibt, sieht aus wie eine Entscheidung und ist
	// eine Karteileiche.
	vorhanden := map[string]bool{}
	for _, paket := range paketeUnterInternal(t) {
		vorhanden[paket] = true
	}

	for paket := range ohneHarteSchwelle {
		if !vorhanden[paket] {
			t.Errorf("ohneHarteSchwelle nennt internal/%s — das gibt es nicht "+
				"(mehr)", paket)
		}
	}

	for paket := range kern {
		if !vorhanden[paket] {
			t.Errorf("KERNPAKETE nennt internal/%s — das gibt es nicht (mehr)", paket)
		}
	}
}

// paketeUnterInternal gibt die Verzeichnisnamen unter `internal/`.
func paketeUnterInternal(t *testing.T) []string {
	t.Helper()

	eintraege, err := os.ReadDir("../../internal")
	if err != nil {
		t.Fatalf("internal/ lesen: %v", err)
	}

	var pakete []string

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			pakete = append(pakete, eintrag.Name())
		}
	}

	if len(pakete) == 0 {
		t.Fatal("kein einziges Paket unter internal/ — liest der Wächter das " +
			"richtige Verzeichnis?")
	}

	return pakete
}
