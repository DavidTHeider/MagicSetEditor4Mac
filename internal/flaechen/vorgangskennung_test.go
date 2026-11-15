package flaechen_test

// **Die Vorgangskennung an den Flächen ohne Fenster** — Werkstück O9 von
// Phase 5.2 (#1433).
//
// # Was er hält
//
// Die Abnahme aus dem Issue: *„Ein Fehler aus GUI/MCP/REST lässt sich per
// Kennung auf erlaubte Details beziehen."* Geprüft wird an den **gebauten**
// Prozessen, denn die Kennung entsteht in ihrer Wurzel — ein Test am Paket
// ginge daran vorbei.
//
// # Was er nicht hält
//
// **Das Fenster.** Dieser Lauf startet Prozesse; die Wails-Bindung läuft in
// keinem. Sie prüft `internal/app/vorgangskennung_test.go`.

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// achtHexziffern ist die Gestalt einer Kennung.
var achtHexziffern = regexp.MustCompile(`^[0-9a-f]{8}$`)

// **Der Dienst trägt sie im Kopf — an jeder Antwort, auch der geglückten.**
func TestJedeDienstantwortTraegtIhreVorgangskennung(t *testing.T) {
	t.Parallel()

	an := starteDienst(t, deckumgebung(t))

	code, koepfe := an.kopfruf(t, http.MethodGet, "/gesundheit", "")
	if code != http.StatusOK {
		t.Fatalf("/gesundheit gab %d", code)
	}

	erste := koepfe.Get(vorgangskopf)
	if !achtHexziffern.MatchString(erste) {
		t.Fatalf("die Kopfzeile %s trug %q", vorgangskopf, erste)
	}

	// **Und die gescheiterte trägt eine andere** — sonst wäre sie kein Griff
	// auf **eine** Anfrage.
	code, koepfe = an.kopfruf(t, http.MethodGet, "/sets/999999/karten", "")

	zweite := koepfe.Get(vorgangskopf)

	switch {
	case code == http.StatusOK:
		t.Fatal("die Anfrage auf ein Set, das es nicht gibt, gelang")
	case !achtHexziffern.MatchString(zweite):
		t.Fatalf("die gescheiterte Antwort trug %q", zweite)
	case erste == zweite:
		t.Fatalf("zwei Anfragen teilen die Kennung %q", erste)
	}
}

// vorgangskopf ist der Name der Kopfzeile — **aus dem Quelltext und nicht
// abgeschrieben**: eine zweite Zeichenkette liefe auseinander.
const vorgangskopf = rest.Vorgangskopf

// **Und MCP hängt sie an den Fehlertext.**
func TestEinMCPFehlerNenntSeinenVorgang(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	// **Ein fachlicher Fehler und kein Protokollfehler**: eine Karte, die es
	// nicht gibt. Der Ruf gelingt, das Werkzeug sagt nein — und genau dieser
	// Text soll den Griff tragen.
	antwort, istFehler := mcpRuf(t, umgebung, "schreiben.karte.loeschen",
		map[string]any{"setID": 999999, "karteID": 999999})
	if !istFehler {
		t.Fatalf("der Ruf auf eine Karte, die es nicht gibt, gelang: %s", antwort)
	}

	stelle := strings.Index(antwort, "(Vorgang: ")
	if stelle < 0 {
		t.Fatalf("Antwort %q trägt keine Kennung", antwort)
	}

	kennung := antwort[stelle+len("(Vorgang: ") : stelle+len("(Vorgang: ")+8]
	if !achtHexziffern.MatchString(kennung) {
		t.Fatalf("die Kennung %q hat nicht die Gestalt", kennung)
	}
}
