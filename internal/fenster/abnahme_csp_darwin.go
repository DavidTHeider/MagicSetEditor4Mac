//go:build abnahme && darwin && !server

package fenster

// **Die Schutzrichtlinie im Abnahmebau** — SEC-17
// (#2164).

import (
	"os"
	"strings"
)

// cspZusatz lässt den Rückkanal der Hüllenprobe zu — **und nur ihn**. Wails'
// MCP-Server schickt die Ergebnisse von `js_eval` per `fetch` an
// `http://<WAILS_MCP_HOST>:<WAILS_MCP_PORT>/eval-result`, Vorgabe
// `127.0.0.1:9099`, eine Bindung an alle Schnittstellen wird `localhost`
// (`pkg/application/mcp_enabled.go`, Zeilen 10–33, und
// `mcp_eval_enabled.go`, `callbackURL`, v3.0.0-beta.23). Ohne diese Quelle
// erreichte keine Antwort die Probe, und die Abnahme maße eine andere
// Richtlinie als die des Freigabebaus — so misst sie dieselbe bis auf diese
// eine Quelle.
func cspZusatz() string {
	wirt := os.Getenv("WAILS_MCP_HOST")
	switch wirt {
	case "":
		wirt = "127.0.0.1"
	case "0.0.0.0", "::":
		wirt = "localhost"
	}

	hafen := os.Getenv("WAILS_MCP_PORT")
	if hafen == "" {
		hafen = "9099"
	}

	// **Ohne `net`** — `internal/schichten` hält, wer das Netz kennt; hier
	// wird nur eine Adresse geschrieben, wie `net.JoinHostPort` sie schriebe.
	if strings.Contains(wirt, ":") {
		wirt = "[" + wirt + "]"
	}

	return "http://" + wirt + ":" + hafen
}

// cspBeobachten stellt die Richtlinie auf „nur melden" — für den Lauf, der
// misst, was sie abweisen würde (`MSE_CSP_BEOBACHTEN=1`).
func cspBeobachten() bool { return os.Getenv("MSE_CSP_BEOBACHTEN") == "1" }
