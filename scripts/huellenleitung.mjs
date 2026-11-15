// **Die Leitung der Hüllenprobe** — zeilenweise JSON über den Unix-Socket,
// den der Abnahmebau öffnet (`internal/fenster/abnahme.go`). Getrennt von
// `huellenprobe.mjs`, weil sie eine andere Frage beantwortet: wie eine
// Nachricht hin- und zurückkommt, nicht was geprüft wird.
import { createConnection } from 'node:net'

export const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

// Leitung spricht zeilenweise JSON über den Unix-Socket der Anwendung.
export class Leitung {
  constructor(ort) {
    this.ort = ort
    this.puffer = ''
    this.wartend = []
  }

  async verbinde(frist = 30000) {
    const ende = Date.now() + frist
    for (;;) {
      try {
        this.socket = await new Promise((los, weg) => {
          const s = createConnection(this.ort, () => los(s))
          s.once('error', weg)
        })
        break
      } catch (grund) {
        if (Date.now() > ende) throw new Error(`keine Abnahmeleitung unter ${this.ort}: ${grund}`)
        await schlafe(200)
      }
    }
    this.socket.setEncoding('utf8')
    // **Endet die Anwendung, endet die Leitung** — wer dann noch wartet,
    // bekommt das als Antwort und nicht nie eine.
    this.socket.on('close', () => {
      for (const los of this.wartend.splice(0)) los({ ok: false, fehler: 'die Leitung ist zu', zu: true })
    })
    this.socket.on('data', (teil) => {
      this.puffer += teil
      let stelle
      while ((stelle = this.puffer.indexOf('\n')) >= 0) {
        const zeile = this.puffer.slice(0, stelle)
        this.puffer = this.puffer.slice(stelle + 1)
        this.wartend.shift()?.(JSON.parse(zeile))
      }
    })
  }

  // frage schickt einen Befehl und wartet auf seine Antwort — **höchstens
  // `frist` Millisekunden** (T02, #2096). Bis zum 2026-10-06 wartete sie ohne
  // Grenze, und eine hängende Anwendung hielt die ganze Probe fest.
  //
  // **Nach einer Frist ist die Leitung zu.** Antworten kommen der Reihe nach;
  // eine verspätete käme sonst der nächsten Frage zu.
  frage(befehl, frist = 30000) {
    if (this.socket.destroyed) return Promise.resolve({ ok: false, fehler: 'die Leitung ist zu', zu: true })

    return new Promise((los) => {
      const fertig = (antwort) => {
        clearTimeout(uhr)
        los(antwort)
      }
      const uhr = setTimeout(() => {
        this.wartend.splice(this.wartend.indexOf(fertig), 1)
        los({ ok: false, fehler: `${String(befehl.was ?? 'Befehl')}: keine Antwort nach ${String(frist)} ms`, zu: true })
        this.socket.destroy()
      }, frist)
      this.wartend.push(fertig)
      this.socket.write(`${JSON.stringify(befehl)}\n`)
    })
  }

  schliesse() {
    this.socket?.destroy()
  }
}

// mcpRuf ruft ein Werkzeug über Wails' MCP und gibt den Text des Ergebnisses.
//
// **Mit Frist, und die Frist bricht die Anfrage ab** — T02 (#2096). Ein
// `Promise.race` allein ließe den Abruf offen; `AbortSignal.timeout` schließt
// ihn. Dreißig Sekunden, weil eine Messung wie `bildrate` fünf braucht.
export async function mcpRuf({ port, schluessel, id }, name, argumente, frist = 30000) {
  const antwort = await fetch(`http://127.0.0.1:${String(port)}/mcp`, {
    method: 'POST',
    headers: { 'content-type': 'application/json', authorization: `Bearer ${schluessel}` },
    body: JSON.stringify({ jsonrpc: '2.0', id, method: 'tools/call', params: { name, arguments: argumente } }),
    signal: AbortSignal.timeout(frist),
  }).then((r) => r.json()).catch((grund) => ({
    error: { message: grund?.name === 'TimeoutError' ? `${name}: keine Antwort nach ${String(frist)} ms` : String(grund) },
  }))
  if (antwort.error) throw new Error(`MCP: ${antwort.error.message}`)
  const inhalt = antwort.result.content[0].text
  if (antwort.result.isError) throw new Error(`${name}: ${inhalt}`)

  return inhalt
}
