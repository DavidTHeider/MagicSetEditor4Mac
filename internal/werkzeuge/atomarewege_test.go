// Der Wächter über das Schreiben neben dem Bestand —
// Werkstück 7 von Phase 4.13
// (#1352).
//
// # Die Frage aus dem Issue
//
// *„Was geschieht bei ⌘Q mitten in einem Schreibvorgang, bei einem Absturz,
// beim Ruhezustand mit offener Datenbank?"*
//
// **Für die Datenbank ist die Antwort gemessen** (2026-09-07): `journal_mode =
// wal`, `synchronous = 2` (FULL). Jede festgeschriebene Änderung liegt auf der
// Platte, jede halbe rollt beim nächsten Öffnen zurück.
//
// **Für die Dateien daneben galt sie nicht.** `os.WriteFile` öffnet mit
// `O_TRUNC`: die vorhandene Datei ist ab diesem Augenblick leer, und bricht
// das Schreiben danach ab — ⌘Q, Absturz, volle Platte —, bleibt sie leer oder
// halb zurück. **Ein halbes PNG sieht im Dateibrowser aus wie ein ganzes**, und
// ein halbes `.mse-set` ist die Arbeit von Wochen.
//
// **Die Abhilfe stand schon da** (#853):
// `service.SchreibeAtomar` schreibt in eine Nachbardatei und benennt um.
// Benutzt wurde sie an vier Stellen — an sieben nicht, darunter der
// Setschreiber.
//
// # Und der Wächter sah nur die halbe Klasse
//
// **Befund 2 des Zweitblicks über Phase 4.13** (2026-09-07): er durchsuchte
// **nur `internal/service`**. Vier Ausgabewege in `internal/app` schrieben
// weiter unmittelbar — die Diagnose als Datei, die Auskunft über die eigenen
// Daten, die Ausgabe einer Vorlage und ihre Beigaben. Alle vier schreiben
// dorthin, wohin jemand im Dateidialog gezeigt hat.
//
// **Das ist genau die Gestalt aus §10**: ein Wächter, der die halbe Zusage
// hält, macht die andere unsichtbar. Der Kopf hier sagte „jede dauerhafte
// Datei", und `docs/ADMINISTRATION.md` schrieb ihm diese Zusage zu.
//
// # Was er nicht prüft
//
// **Ob `schreibeAtomar` hält, was es verspricht.** Das prüfen seine eigenen
// Läufe in `internal/service`. Hier wird gezählt, **wer sie benutzt**.
//
// **Und nicht den Absturz selbst.** Ein Rechner, der mitten im `write(2)`
// stehen bleibt, lässt sich in einem Testlauf nicht herstellen; was sich
// prüfen lässt, ist der Weg, der ihn überlebt.
package werkzeuge

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// unmittelbareWege sind die Rufe, die eine Datei an Ort und Stelle
// überschreiben.
//
// **`O_TRUNC` kam mit dem Nachlauf dazu** (2026-09-07). Der Wächter kannte nur
// die zwei bequemen Formen; `os.OpenFile(pfad, …|os.O_TRUNC, …)` trägt genau
// dasselbe Risiko — die Datei ist ab dem Öffnen leer — und ging durch. Die
// Form steht bereits im geprüften Paket, dort für eine Zwischenkopie; wer sie
// eines Tages für ein dauerhaftes Ziel nähme, bliebe grün.
//
// **Das ist die Klasse, nicht der Ruf**: gesucht ist „diese Datei ist ab
// jetzt leer", und dafür gibt es drei Schreibweisen.
var unmittelbareWege = []string{"os.WriteFile(", "os.Create(", "os.O_TRUNC"}

// schreibausnahme ist eine Stelle, die unmittelbar schreiben darf — **mit
// ihrem Grund**.
//
// **Je Stelle und nicht je Datei** (Nachlauf zum Zweitblick, 2026-09-07): eine
// Datei ganz auszunehmen deckte auch jede Zeile, die später dazukommt. Der
// `satzteil` bindet die Ausnahme an den Ruf, um den es geht.
type schreibausnahme struct {
	datei    string
	satzteil string
	grund    string
}

var schreibausnahmen = []schreibausnahme{
	{
		"atomarschreiben.go", "",
		"hier steht der atomare Weg selbst",
	},
	{
		"bestandersetzen.go", "os.OpenFile(ziel, os.O_WRONLY",
		"das Ziel **ist** die Zwischendatei `…eingehend`, die danach umbenannt " +
			"wird — also selbst die atomare Gestalt",
	},
	{
		"kennungsumzug.go", "os.WriteFile(filepath.Join(zwischen, umzugsdatei)",
		"die Marke liegt im **Zwischenordner** `…umzug`, und erst das " +
			"Umbenennen des ganzen Ordners gibt ihr den Namen — eine halbe " +
			"Marke steht nie unter dem neuen Ordner (R02, #2334)",
	},
	{
		"kennungsumzugvorgaben.go", "os.WriteFile(vereinigt, vereint",
		"eine **Zwischendatei** im Arbeitsordner dieses Laufs, die `defaults import` " +
			"gleich liest und der Umzug danach mit dem Ordner löscht; eine halbe davon " +
			"scheitert beim Import, und die Marke kommt erst mit ihm (#2424, Zweitblick " +
			"über Phase 4, Befund 5)",
	},
}

func TestJedeDauerhafteDateiWirdAtomarGeschrieben(t *testing.T) {
	t.Parallel()

	var unmittelbar []string

	// **Und die drei Flächen ohne Fenster** — FMT-01.AK-03
	// (#1690):
	// „keine unabhängigen Dateischreibimplementierungen je Oberfläche". Sie
	// schreiben heute über dieselben Nähte wie das Fenster; ohne sie hier
	// hielte das niemand fest, und der erste eigene Schreibweg an einer
	// Fläche fiele nicht auf.
	wurzeln := [][]string{
		gowurzeln(t, "../service"), gowurzeln(t, "../app"),
		gowurzeln(t, "../mcp"), gowurzeln(t, "../rest"), gowurzeln(t, "../../cmd"),
	}

	for _, datei := range slices.Concat(wurzeln...) {
		roh, err := os.ReadFile(datei) //nolint:gosec // Pfad aus dem Baum selbst
		if err != nil {
			t.Fatalf("%s: %v", datei, err)
		}

		for nummer, zeile := range strings.Split(string(roh), "\n") {
			// **Kommentare zählen nicht.** Mehrere Stellen erklären, *warum*
			// sie den unmittelbaren Weg nicht mehr gehen — und das ist die
			// Erklärung, nicht der Ruf.
			if istKommentar(zeile) || !einUnmittelbarerWeg(zeile) ||
				ausgenommen(filepath.Base(datei), zeile) {
				continue
			}

			unmittelbar = append(unmittelbar,
				filepath.Base(datei)+":"+itoa(nummer+1)+": "+strings.TrimSpace(zeile))
		}
	}

	if len(unmittelbar) > 0 {
		t.Errorf("diese Stellen schreiben an Ort und Stelle — ein Abbruch "+
			"lässt eine halbe Datei unter dem richtigen Namen zurück; "+
			"`schreibeAtomar` nebenan tut es nicht:\n%s",
			strings.Join(unmittelbar, "\n"))
	}
}

// ausgenommen sagt, ob eine Stelle mit ihrem Grund geführt wird.
//
// **Ein leerer `satzteil` nimmt die ganze Datei aus** — das gilt allein für
// `atomarschreiben.go`, wo der atomare Weg selbst steht.
func ausgenommen(name, zeile string) bool {
	for _, ausnahme := range schreibausnahmen {
		if ausnahme.datei != name || ausnahme.grund == "" {
			continue
		}

		if ausnahme.satzteil == "" || strings.Contains(zeile, ausnahme.satzteil) {
			return true
		}
	}

	return false
}

func istKommentar(zeile string) bool {
	return strings.HasPrefix(strings.TrimSpace(zeile), "//")
}

func einUnmittelbarerWeg(zeile string) bool {
	for _, weg := range unmittelbareWege {
		if strings.Contains(zeile, weg) {
			return true
		}
	}

	return false
}

// itoa hält den Wächter frei von `strconv` — eine Zeilennummer ist klein.
func itoa(zahl int) string {
	if zahl < 10 {
		return string(rune('0' + zahl))
	}

	return itoa(zahl/10) + string(rune('0'+zahl%10))
}
