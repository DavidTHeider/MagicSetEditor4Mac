package vorschauweg_test

// **Das Blatt macht keine Verbindung auf** —
// #1270.
//
// # Wofür dieser Wächter da ist
//
// `internal/schichten/netz_test.go` führt, wer das Netz erreichen darf, und
// dieses Paket steht dort. Der Eintrag ist eine **Entscheidung** (§16) und
// begründet damit, dass hier keine Verbindung entsteht: es formt Bytes zu einer
// HTTP-Antwort für die Assetleitung, die Wails für das eigene Fenster ohnehin
// betreibt.
//
// **Eine Begründung, die niemand nachprüft, ist eine Ausrede** (§10). Also
// steht sie hier als Lauf: wer einen Klienten, einen Lauscher oder eine Adresse
// aufnimmt, macht die Pipeline rot — und schreibt dann entweder den Eintrag
// drüben fort oder lässt es.
//
// # Was er nicht prüft
//
//   - **Keinen Weg unter der Bibliothek durch.** `syscall.Socket` käme hier
//     durch; dieselbe Lücke wie beim Wächter drüben, aus demselben Grund.
//   - **Keine fremde Bibliothek.** Dieses Paket hat keine, und ein Wächter über
//     `go.mod` gehört nicht hierher.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// verbindungsformen sind die Aufrufe, die eine Verbindung aufmachen oder auf
// eine warten.
//
// **`http.NotFound` und `http.StatusNotFound` stehen nicht dabei** — sie
// schreiben eine Antwort und öffnen nichts. Die Liste nennt Erzeuger, nicht das
// ganze Paket: sonst stünde `net/http` selbst darin, und der Wächter wäre von
// der ersten Zeile an rot.
var verbindungsformen = []string{
	"http.Client", "http.DefaultClient", "http.Get", "http.Post", "http.Head",
	"http.NewRequest", "http.ListenAndServe", "http.Serve", "http.Server",
	"net.Listen", "net.Dial", "os/exec", "syscall.Socket",
}

func TestDasBlattMachtKeineVerbindungAuf(t *testing.T) {
	t.Parallel()

	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Verzeichnis lesen: %v", err)
	}

	gelesen := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		gelesen++

		quelle, fehler := os.ReadFile(filepath.Clean(name))
		if fehler != nil {
			t.Fatalf("%s lesen: %v", name, fehler)
		}

		for _, form := range verbindungsformen {
			if strings.Contains(string(quelle), form) {
				t.Errorf("%s nennt %s — dieses Blatt steht in der "+
					"Ausnahmeliste von internal/schichten/netz_test.go mit der "+
					"Begründung, dass es keine Verbindung aufmacht. Wer das "+
					"ändert, schreibt den Eintrag dort fort (§16)", name, form)
			}
		}
	}

	// **Und es hat überhaupt etwas gelesen.** Ein Wächter über null Dateien
	// ist grün und sagt nichts — dieselbe Gestalt wie die leere Aufzählung.
	if gelesen == 0 {
		t.Fatal("keine einzige Quelldatei gelesen — läuft der Wächter im richtigen Verzeichnis?")
	}

	t.Logf("%d Quelldateien geprüft", gelesen)
}
