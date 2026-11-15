// Package mcp bedient das Model Context Protocol über Standardein- und
// -ausgabe.
//
// # Wo es läuft
//
// **Als eigener Prozess, nicht im Fenster**
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). Ein
// Server im Fenster bräuchte einen Port auf `localhost` — und damit wäre die
// Anwendung ein Netzwerkdienst, den `docs/ROADMAP.md` ausdrücklich ausschließt.
// Über stdio gibt es keinen Port, keine Adresse und nichts, was auf etwas
// wartet.
//
// # Warum ohne fremde Bibliothek
//
// Was hier gebraucht wird, ist JSON-RPC 2.0 über zwei Ströme und drei Methoden.
// Das ist weniger Code als die Pinnung, die Stückliste und die Prüfung einer
// Abhängigkeit kosteten (§1) — und es lässt sich vollständig prüfen.
package mcp

import (
	"encoding/json"
	"fmt"
)

// fassung ist die Protokollfassung, die dieser Server spricht.
//
// **Fest und nicht ausgehandelt.** MCP erlaubt, eine andere anzubieten; wer das
// tut, sagt eine Verträglichkeit zu, die er nicht geprüft hat.
const fassung = "2025-06-18"

// Anfrage ist eine eingehende JSON-RPC-Nachricht.
//
// `ID` ist roh, weil JSON-RPC Zahlen **und** Zeichenketten erlaubt und die
// Antwort dieselbe Gestalt zurückgeben muss. Wer sie in einen Typ zwingt,
// antwortet dem einen Gegenüber mit `"1"`, wo es `1` geschickt hat.
type Anfrage struct {
	JSONRPC   string          `json:"jsonrpc"`
	ID        json.RawMessage `json:"id"`
	Methode   string          `json:"method"`
	Parameter json.RawMessage `json:"params"`
}

// istMeldung sagt, ob die Nachricht keine Antwort erwartet.
//
// Eine Meldung (*notification*) hat keine Kennung. Wer ihr trotzdem antwortet,
// schickt eine Antwort auf nichts — manche Gegenstellen brechen daran ab.
func (a Anfrage) istMeldung() bool { return len(a.ID) == 0 || string(a.ID) == "null" }

// Antwort ist eine ausgehende JSON-RPC-Nachricht.
type Antwort struct {
	JSONRPC  string           `json:"jsonrpc"`
	ID       json.RawMessage  `json:"id"`
	Ergebnis any              `json:"result,omitempty"`
	Fehler   *Protokollfehler `json:"error,omitempty"`
}

// Protokollfehler ist ein Fehler nach JSON-RPC.
type Protokollfehler struct {
	Code    int    `json:"code"`
	Meldung string `json:"message"`
}

// Die Codes aus JSON-RPC 2.0, so weit hier gebraucht.
const (
	codeZerteilung    = -32700
	codeUnbekannt     = -32601
	codeParameter     = -32602
	codeInnererFehler = -32603
)

// antwortAuf baut eine Antwort mit der Kennung der Anfrage.
func antwortAuf(anfrage Anfrage, ergebnis any) Antwort {
	return Antwort{JSONRPC: "2.0", ID: anfrage.ID, Ergebnis: ergebnis}
}

// fehlerAuf baut eine Fehlerantwort mit der Kennung der Anfrage.
func fehlerAuf(anfrage Anfrage, code int, vorlage string, werte ...any) Antwort {
	return Antwort{
		JSONRPC: "2.0",
		ID:      anfrage.ID,
		Fehler:  &Protokollfehler{Code: code, Meldung: fmt.Sprintf(vorlage, werte...)},
	}
}
