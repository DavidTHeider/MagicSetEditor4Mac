package mcp

// **Die Antwort als Ereignisstrom** — FMT-03.AK-02 und AK-03
// (#1692).
//
// Nachgelesen am 2026-09-29 über Context7, MCP-Spezifikation Fassung
// 2026-07-28, `basic/transports/streamable-http`: der Server darf auf eine
// Anfrage mit `text/event-stream` antworten, vor der Antwort Meldungen zu
// dieser Anfrage schicken, und die Antwort beendet den Strom. **Schließt der
// Klient ihn, ist das der Abbruch** — der Kontext der Anfrage endet, und der
// Fachkern hält beim nächsten Schritt an.
//
// **Gewählt wird er nur auf Bitte**: der Klient nimmt ihn an (`Accept`) und
// will Meldungen (`progressToken`). Ohne beides bliebe es eine Antwort in
// einer Verpackung, die ihr nichts hinzufügt.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// stromGewuenscht sagt, ob die Antwort als Ereignisstrom gehen soll.
func stromGewuenscht(anfrage *http.Request, roh []byte) bool {
	if !strings.Contains(anfrage.Header.Get("Accept"), "text/event-stream") {
		return false
	}

	var gelesen Anfrage
	if json.Unmarshal(roh, &gelesen) != nil {
		return false
	}

	return len(fortschrittsmarke(gelesen.Parameter)) > 0
}

// ereignisschreiber schreibt Nachrichten als Ereignisse in die Antwort.
type ereignisschreiber struct {
	sperre    sync.Mutex
	schreiben http.ResponseWriter
}

// beginneStrom setzt die Kopfzeilen des Stroms.
func beginneStrom(schreiben http.ResponseWriter) *ereignisschreiber {
	schreiben.Header().Set("Content-Type", "text/event-stream")
	schreiben.Header().Set("Cache-Control", "no-cache")
	schreiben.WriteHeader(http.StatusOK)

	return &ereignisschreiber{schreiben: schreiben}
}

// sende schreibt eine Nachricht als ein Ereignis und gibt es sofort hinaus.
func (e *ereignisschreiber) sende(nachricht any) error {
	e.sperre.Lock()
	defer e.sperre.Unlock()

	// **Ohne Fehlerweg für das Serialisieren** — derselbe Satz wie bei
	// [Server.schreibe].
	roh, _ := json.Marshal(nachricht) //nolint:errcheck // siehe darüber

	if _, fehler := fmt.Fprintf(e.schreiben, "data: %s\n\n", roh); fehler != nil {
		return fmt.Errorf("MCP-Ereignis schreiben: %w", fehler)
	}

	return http.NewResponseController(e.schreiben).Flush()
}
