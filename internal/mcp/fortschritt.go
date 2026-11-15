package mcp

// **Fortschritt eines Werkzeugrufs** — FMT-03.AK-02
// (#1692).
//
// Nachgelesen am 2026-09-29 über Context7, MCP-Spezifikation Fassung
// 2026-07-28, `basic/patterns/progress`: der Aufrufer bittet mit
// `params._meta.progressToken` um Meldungen; der Server schickt
// `notifications/progress` mit `progressToken`, `progress`, und wo bekannt
// `total` und `message`. **`progress` muss mit jeder Meldung steigen**, beide
// Seiten sollen drosseln, und nach dem Ende kommt keine mehr.
//
// # Wer meldet und wer schickt
//
// **Gemeldet wird im Fachkern** (`service.Schritt`), und wie der Melder in
// den Kontext kommt, weiß `service` — dieses Blatt kennt es nicht (§2). Die
// Wurzel reicht die Naht herein ([Server.MitFortschrittsnaht]), dieselbe
// Gestalt wie die Vorgangskennung. **Geschickt wird über den Transport**:
// über stdio in denselben Strom wie die Antworten, über HTTP in den
// Ereignisstrom der einen Antwort (`ueberhttp.go`).

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fortschrittspause ist, wie lange zwischen zwei Meldungen mindestens liegt.
//
// **Zehn in der Sekunde** — genug, dass ein Balken sich bewegt, und wenig
// genug, dass ein Bogen mit dreihundert Karten nicht dreihundert Zeilen in
// den Strom legt. Die erste und die letzte gehen immer hinaus.
const fortschrittspause = 100 * time.Millisecond

// Meldung ist eine JSON-RPC-Benachrichtigung — ohne Kennung.
type Meldung struct {
	JSONRPC   string `json:"jsonrpc"`
	Methode   string `json:"method"`
	Parameter any    `json:"params"`
}

// MitFortschrittsnaht sagt dem Server, wie ein Melder in den Kontext kommt.
// Ohne sie meldet niemand — auch wenn ein Aufrufer darum bittet.
func (s *Server) MitFortschrittsnaht(
	naht func(context.Context, domain.Fortschrittsmelder) context.Context,
) *Server {
	s.fortschritt = naht

	return s
}

type senderschluessel struct{}

// mitSender legt in den Kontext, wie eine Meldung zum Gegenüber kommt.
func mitSender(ctx context.Context, senden func(any) error) context.Context {
	return context.WithValue(ctx, senderschluessel{}, senden)
}

// fortschrittsmarke liest `_meta.progressToken` — leer, wenn keine genannt
// ist. **Roh**, denn sie darf Zahl oder Zeichenkette sein und kommt in
// ihrer Gestalt zurück.
func fortschrittsmarke(parameter json.RawMessage) json.RawMessage {
	var gelesen struct {
		Meta struct {
			Marke json.RawMessage `json:"progressToken"`
		} `json:"_meta"`
	}

	if json.Unmarshal(parameter, &gelesen) != nil || string(gelesen.Meta.Marke) == "null" {
		return nil
	}

	return gelesen.Meta.Marke
}

// mitMelder hängt einen Melder an, wenn der Aufrufer darum bat und der
// Transport schicken kann.
func (s *Server) mitMelder(ctx context.Context, marke json.RawMessage) context.Context {
	senden, kann := ctx.Value(senderschluessel{}).(func(any) error)
	if len(marke) == 0 || !kann || s.fortschritt == nil {
		return ctx
	}

	return s.fortschritt(ctx, drosselnderMelder(marke, senden, time.Now))
}

// drosselnderMelder schickt steigende Stände, höchstens alle
// [fortschrittspause] — den ersten und den letzten immer.
func drosselnderMelder(
	marke json.RawMessage, senden func(any) error, jetzt func() time.Time,
) domain.Fortschrittsmelder {
	var (
		sperre   sync.Mutex
		zuletzt  = -1
		gesendet time.Time
	)

	return func(stand domain.Fortschritt) {
		sperre.Lock()
		defer sperre.Unlock()

		letzter := stand.Gesamt > 0 && stand.Erledigt >= stand.Gesamt
		if stand.Erledigt <= zuletzt || (zuletzt >= 0 && !letzter && jetzt().Sub(gesendet) < fortschrittspause) {
			return
		}

		parameter := map[string]any{
			"progressToken": marke, "progress": stand.Erledigt, "message": stand.Satz(),
		}
		if stand.Gesamt > 0 {
			parameter["total"] = stand.Gesamt
		}

		// **Eine Meldung, die nicht ankommt, hält den Lauf nicht an** — die
		// Antwort geht denselben Weg und meldet den toten Strom.
		_ = senden(Meldung{JSONRPC: "2.0", Methode: "notifications/progress", Parameter: parameter}) //nolint:errcheck // siehe darüber

		zuletzt, gesendet = stand.Erledigt, jetzt()
	}
}
