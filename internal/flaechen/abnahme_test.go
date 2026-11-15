// **Die Abnahme von Phase T** — derselbe Arbeitsablauf über beide Flächen
// (#886).
//
// # Warum das ein eigenes Paket ist
//
// Der Abnahmesatz der Phase lautet: „Ein festgelegter Arbeitsablauf läuft ohne
// Fenster **sowohl über die Kommandozeile als auch über MCP**, mit denselben
// fachlichen Fehlern."
//
// Beide sind `package main` in verschiedenen Verzeichnissen; ein Test kann sie
// nicht zusammen einbinden. Er kann sie **bauen und fahren** — und das ist
// ohnehin der ehrlichere Weg: geprüft wird, was jemand bekommt, nicht was ein
// Aufruf im selben Prozess täte.
//
// # Was er nicht prüft
//
//   - **Das Fenster.** Es braucht ein echtes; die vierte Fläche steht in der
//     Sichtprüfung (§3).
//   - **Dass die Ausgaben gleich aussehen.** Sie sollen es nicht: eine
//     Kommandozeile schreibt Tabulatorzeilen, MCP schreibt JSON-RPC. Geprüft
//     wird, dass **dieselbe Fachlichkeit** herauskommt.
//   - **Jede Fähigkeit.** Nur der Ablauf, den die Abnahme nennt: einfahren,
//     ändern, zeichnen, sichern.
//   - **Dass beide Flächen denselben Satz sagen.** Sie sollen es nicht — eine
//     Kommandozeile spricht mit einem Menschen, MCP mit einem Modell. Geprüft
//     wird, dass beide bei **demselben Vorgang** scheitern und beide das
//     nennen, woran es lag.
//   - **Byte für Byte dieselbe Datei wie die Vorlage.** Das ist die Zusage von
//     #48 und
//     wird dort gehalten. Hier wird das geschriebene Set in einem **frischen**
//     Bestand wieder aufgemacht und die Änderung darin gesucht — benutzbar,
//     nicht identisch.
//
// # Drei Sätze, die hier bis zum 2026-08-23 standen und zu viel versprachen
//
// Der Zweitblick über Phase T hat sie angestrichen, und alle drei hatten
// dieselbe Gestalt: **geprüft wurde, dass etwas geschah, nicht dass das
// Richtige geschah.**
//
//   - „mit denselben fachlichen Fehlern" war je Fläche einzeln geprüft — und
//     beim unbekannten Set auf zwei **verschiedenen** Vorgängen.
//   - Das geschriebene Set wurde mit `os.Stat` belegt. Eine leere Datei wäre
//     durchgegangen.
//   - Das Bild wurde an seinen ersten acht Bytes erkannt. Eine Datei aus
//     genau diesen acht Bytes wäre durchgegangen.
//
// Jede Verschärfung hat ihre Gegenprobe; sie stehen an den Helfern.
package flaechen_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// gebaut baut alle Programme einmal und gibt ihre Pfade.
//
// **Einmal für alle Fälle**: der Bau kostet unter einer Sekunde bei warmem
// Zwischenspeicher, aber nicht je Fall.
//
// **Drei seit dem 2026-08-23** — der Dienst auf der Rückschleife ist die
// dritte Fläche ohne Fenster
// (#997).
var gebaut struct {
	mse, mcp, rest string
}

// vorrat ist die Kopie von `testdaten/` und `saat/`, die die Läufe an einen
// offenen Server reichen — und zugleich sein Dienstordner.
//
// **Eine Kopie und nicht das Repositorium** (C06,
// #2231):
// der Dienst weist jeden Pfad mit einem verborgenen Glied ab, zu Recht. Lag
// der Arbeitsbaum unter `~/.codex/worktrees/…`, traf das jede Probedatei,
// und fünf REST-Fälle bekamen 400 — die Vorrichtung setzte einen Rechnerpfad
// voraus. Die Schutzregel bleibt; die Proben liegen in einem Ort, den der
// Lauf selbst anlegt und freigibt.
var vorrat string

func TestMain(m *testing.M) {
	// **Kein `defer` vor `os.Exit`** — es liefe nicht. Aufgeräumt wird
	// deshalb ausdrücklich, bevor der Rückgabewert gesetzt wird.
	ort, err := os.MkdirTemp("", "flaechen")
	if err != nil {
		panic(err)
	}

	gebaut.mse = filepath.Join(ort, "mse")
	gebaut.mcp = filepath.Join(ort, "mse-mcp")
	gebaut.rest = filepath.Join(ort, "mse-rest")

	for ziel, paket := range map[string]string{
		gebaut.mse:  "./cmd/mse",
		gebaut.mcp:  "./cmd/mse-mcp",
		gebaut.rest: "./cmd/mse-rest",
	} {
		// **Ohne Frist, und das ist Absicht**: `TestMain` hat keinen `t`, und
		// ein Bau, der an einer erfundenen Uhr scheitert, sagt nichts über
		// den Code. Scheitert er wirklich, sagt `CombinedOutput` es.
		//
		// #nosec G204 -- feste Werte aus dieser Datei
		//
		// **Mit dem Probeanker** (SEC03, #2356): die Programme prüfen eine
		// Freigabe gegen einen Schlüssel, den dieser Lauf erzeugt — nicht
		// gegen einen aus dem Schlüsselbund des Rechners (`probeanker_test.go`).
		bau := exec.CommandContext(context.Background(), "go", "build", "-tags", "probeanker", "-o", ziel, paket)
		bau.Dir = "../.."

		if ausgabe, fehler := bau.CombinedOutput(); fehler != nil {
			panic(string(ausgabe))
		}
	}

	vorrat = vorratAnlegen(ort)
	probeankerAnlegen(ort)

	ausgang := m.Run()

	_ = os.RemoveAll(ort) //nolint:errcheck // ein Verzeichnis in TMPDIR

	os.Exit(ausgang)
}

// vorratAnlegen kopiert die Proben unter `ort` und prüft, dass der Dienst sie
// dort annimmt — **ein verborgenes `TMPDIR` ist eine Grenze dieses Laufs**
// und wird genannt, statt in fünf Fällen als 400 zu erscheinen.
func vorratAnlegen(ort string) string {
	ziel := filepath.Join(ort, "vorrat")
	for _, teil := range []string{"testdaten", "saat"} {
		if fehler := os.CopyFS(filepath.Join(ziel, teil), os.DirFS(filepath.Join("..", "..", teil))); fehler != nil {
			panic(fehler)
		}
	}

	for glied := range strings.SplitSeq(ziel, string(filepath.Separator)) {
		if strings.HasPrefix(glied, ".") {
			panic("der Vorrat liegt unter einem verborgenen Ordner (" + ziel + ") — der Dienst nimmt dort keinen Pfad an; TMPDIR auf einen sichtbaren Ort setzen")
		}
	}

	return ziel
}

// eigenerBestand gibt einen Ort, den nur dieser Fall benutzt.
//
// **Über die Umgebung**, denn beide Programme suchen sonst den Bestand der
// Nutzenden — und ein Test, der dort schriebe, wäre kein Test.
func eigenerBestand(t *testing.T) []string {
	t.Helper()

	// **Und ein Schlüsselbund, den es nicht gibt.**
	//
	// **Weil ein Kindprozess kein Testbinärprogramm ist.**
	// `internal/schluesselbund` weist den Anmeldebund einem `*.test` ab;
	// `mse`, `mse-mcp` und `mse-rest` heißen anders, und für sie gilt die
	// Sperre nicht. Hier ist die einzige Stelle, an der jeder von ihnen
	// vorbeikommt.
	//
	// **Gemessen am 2026-09-14**: nach der Sperre blieben **18** Zugriffe auf
	// den Anmeldebund dieses Rechners übrig, alle aus Kindprozessen von
	// `internal/flaechen`. Wer wirklich einen Bund braucht, hängt
	// `probebund(t)` **dahinter** — bei doppeltem Schlüssel gilt der letzte
	// Eintrag.
	// **Und der Vorrat als Dienstordner** (OPS-02.AK-04, #1719): die Läufe
	// reichen `testdaten/` und `saat/` an einen offenen Server, und ein
	// Arbeitsbaum liegt nicht immer im Zuhause. Derselbe Weg, den ein
	// Betrieb für `/srv/karten` nimmt — kein abgeschalteter Vertrag. **Die
	// Kopie und nicht das Repositorium** — siehe [vorrat].

	// **Und die Schleife für den Paketbezug** (#2000): er wählt nur
	// öffentlich gerichtete Adressen, und jede Gegenstelle dieser Läufe ist
	// ein `httptest`-Server auf `127.0.0.1`. Der Schalter öffnet die Schleife
	// und **nur** sie; was ohne ihn gilt, fährt `zielsperrepaket_test.go`,
	// das ihn dahinter wieder leert — bei doppeltem Schlüssel gilt der letzte.
	//
	// **Und ein eigenes Zuhause** (Zweitblick über Phase C, Befund 5):
	// `os.Environ()` trug das echte `HOME`, und jeder Kindprozess bekam es —
	// bei doppeltem Schlüssel gilt der letzte.
	return append(os.Environ(),
		"HOME="+t.TempDir(),
		"MSE_BESTAND="+filepath.Join(t.TempDir(), "bestand.db"),
		"MSE_DIENSTORDNER="+vorrat,
		netz.SchleifeUmgebung+"=an",
		schluesselbund.Bundort+"="+filepath.Join(t.TempDir(), "keiner.keychain"),
		// **Und Apples Gerätemodell gesperrt** (AI-05, #2042) — aus demselben
		// Grund wie der Schlüsselbund: ein Kindprozess ist kein `*.test`.
		applefm.Sperre+"=an",
		// **Und jeder verändernde Ruf an launchd** — aus demselben Grund
		// (Nachlauf über Phase 2 von 0.9.0, Befund 6).
		app.Dienstsperre+"=an")
}

// probeset gibt den Pfad des Probesets aus `testdaten/` — im [vorrat].
func probeset(t *testing.T) string {
	t.Helper()

	return filepath.Join(vorrat, "testdaten", "probe.mse-set")
}
