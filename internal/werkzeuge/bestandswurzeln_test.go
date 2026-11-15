package werkzeuge_test

// **Jede Fläche findet ihren Bestand auf demselben Weg**
// (#1455).
//
// # Wofür es diesen Wächter gibt
//
// **Am 2026-09-14 hat eine Sichtprüfung die Peeridentität des echten Profils
// abgelöst.** Die Startkonfiguration `wails-sichtpruefung` setzt `MSE_BESTAND`
// und `MSE_SCHLUESSELBUND` auf einen Ort unter `$TMPDIR` und versprach damit
// „eigener Bestand, eigener Schlüsselbund". Der Bund griff — der Bestand
// nicht: `main.go` rief `app.Datenbankpfad()` unmittelbar, und nur die drei
// Flächen ohne Fenster gingen über [app.BestandsortFuerFlaechen].
//
// **Eine Zusage, die zur Hälfte hält, ist schlimmer als keine** (§10): sie
// wird geglaubt. In diesem Fall von dem, der die Sichtprüfung fährt.
//
// # Was er hält
//
// **Dass kein Composition Root den Umgebungsschalter übergeht.** Wer
// `Datenbankpfad()` unmittelbar ruft, umgeht ihn — die Funktion ist die
// **Vorgabe** und nicht der Weg dorthin.
//
// # Was er nicht hält
//
// **Dass der Schalter das Richtige tut.** Das prüft
// `internal/app/bestandsort_test.go`. Hier steht nur, dass jeder Einstieg ihn
// überhaupt fragt.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **`ohneKommentare` steht in `datenschutz_test.go`** und wird hier
// mitbenutzt — eine zweite Fassung liefe beim ersten Sonderfall auseinander
// (§10). Ohne sie zählte die Begründung als Beleg: in `main.go` steht
// `app.Datenbankpfad()` im Kommentar über der Zeile, die ihn ersetzt hat.

// wurzeln sind die Einstiege, die einen Bestand öffnen.
//
// **Eine Aufzählung, und das ist Absicht**: ein fünfter Composition Root ist
// eine Entscheidung (§2), und er soll hier auffallen statt still den falschen
// Weg zu nehmen.
var wurzeln = []string{
	"main.go",
	"cmd/mse/main.go",
	"cmd/mse-mcp/server.go",
	"cmd/mse-rest/dienst.go",
}

func TestJedeWurzelFragtDenUmgebungsschalter(t *testing.T) {
	t.Parallel()

	for _, datei := range wurzeln {
		t.Run(datei, func(t *testing.T) {
			t.Parallel()

			// **Der Pfad kommt aus einer Aufzählung in dieser Datei**,
			// nicht von außen — gosec sieht nur die Veränderliche.
			roh, fehler := os.ReadFile( //nolint:gosec // siehe darüber
				filepath.Join("..", "..", datei))
			if fehler != nil {
				t.Fatalf("%s lesen: %v", datei, fehler)
			}

			quelle := ohneKommentare(string(roh))

			if !strings.Contains(quelle, "BestandsortFuerFlaechen()") {
				t.Errorf(
					"%s fragt MSE_BESTAND nicht — es ruft den Vorgabeort "+
						"unmittelbar, und eine Umlenkung greift dort nicht",
					datei)
			}

			if strings.Contains(quelle, "app.Datenbankpfad()") ||
				strings.Contains(quelle, "= Datenbankpfad()") {
				t.Errorf(
					"%s ruft Datenbankpfad() unmittelbar — das ist die "+
						"Vorgabe und nicht der Weg dorthin", datei)
			}
		})
	}
}

// vorgabeortErlaubt sind die Stellen, die den Vorgabeort fragen dürfen — mit
// dem Grund, warum sie nicht den Bestand fragen, der offen ist.
var vorgabeortErlaubt = map[string]string{
	"internal/app/orte.go":            "hier steht er",
	"internal/app/bestandsort.go":     "die Vorgabe, wenn die Umgebung keinen Ort nennt",
	"internal/app/scryfalldrossel.go": "die Drossel gilt je Rechner: zwei Instanzen gehen über eine Adresse hinaus",
	"internal/app/paketbezug.go":      "Pakete und ihre Schriften gehören dem Rechner (docs/AUTORISIERUNG.md)",
	"internal/app/scryfall.go":        "der Zwischenspeicher öffentlicher Daten gehört dem Rechner wie die Pakete",
	"internal/app/restepflege.go":     "gefragt wird, **ob** der offene Bestand am Vorgabeort liegt — nur dann sind die Pakete daneben eigene (Zweitblick über Phase 4, Befund 1)",
}

// **Nur benannte Stellen fragen den Vorgabeort** — QA-10.NAT-01
// (#1741).
//
// Bis dahin fragten ihn auch die Bilder und die Diagnose: ein Bestand mit
// eigenem Ort legte seine Bilder neben einen anderen, und die Diagnose nannte
// eine Datei, in der nichts stand. **Was dem Rechner gehört, darf ihn fragen,
// und steht deshalb hier mit Grund** — wer eine weitere Stelle braucht, trägt
// sie ein und begründet sie; still an den Vorgabeort zu schreiben ist der
// Befund.
func TestNurBenannteStellenFragenDenVorgabeort(t *testing.T) {
	t.Parallel()

	for _, wurzel := range []string{"internal", "cmd"} {
		for _, pfad := range quelldateien(t, filepath.Join("..", "..", wurzel)) {
			roh, fehler := os.ReadFile(pfad) //nolint:gosec // aus dem eigenen Baum
			if fehler != nil {
				t.Fatalf("%s lesen: %v", pfad, fehler)
			}

			relativ := strings.TrimPrefix(filepath.ToSlash(pfad), "../../")
			quelle := ohneKommentare(string(roh))

			if _, erlaubt := vorgabeortErlaubt[relativ]; erlaubt {
				continue
			}

			if strings.Contains(quelle, "Datenbankpfad()") ||
				strings.Contains(quelle, "os.UserConfigDir()") {
				t.Errorf("%s fragt den Vorgabeort und nicht den Bestand, der "+
					"offen ist — BestandsortFuerFlaechen() nehmen oder hier "+
					"begründen", relativ)
			}
		}
	}
}
