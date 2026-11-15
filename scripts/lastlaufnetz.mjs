// **Die Leitungen des Lastlaufs** — QA-08.AK-01
// (#1739).
//
// Eine eigene Datei, weil `lastlauf.mjs` mit ihr über die Zeilengrenze aus §2
// ginge. Der Schnitt folgt der Frage: hier steht, **wie** ein Ruf hinausgeht
// und gemessen wird — HTTP, MCP über HTTP, ein langsamer Klient, der Speicher
// eines Prozesses —, dort, **was** gerufen wird und wer die Prozesse startet.
// **Diese Datei startet keinen**: sie ist ein Helfer und kein Lauf
// (`internal/werkzeuge/nahtlauf_test.go`).
//
// # Warum `node:http` und nicht `fetch`
//
// **Weil die Messung den Weg kennen soll.** `fetch` legt einen eigenen Pool
// mit eigenen Regeln dazwischen; hier hält ein `Agent` mit `keepAlive` je
// Dienst die Verbindungen, und jede gemessene Zeit ist die eines Rufs vom
// Schreiben der Anfrage bis zum letzten Byte der Antwort.

import { execFileSync } from 'node:child_process'
import http from 'node:http'
import net from 'node:net'
import { performance } from 'node:perf_hooks'

// ruf schickt eine Anfrage und misst sie: Status, Wartezeit aus
// `Retry-After`, Dauer in Millisekunden und der Rumpf als Text.
export function ruf(dienst, methode, weg, { rumpf, nachweis, koepfe = {}, frist = 120_000 } = {}) {
  const ziel = new URL(dienst.adresse + weg)
  const daten = rumpf === undefined ? undefined : Buffer.from(typeof rumpf === 'string' ? rumpf : JSON.stringify(rumpf))
  const beginn = performance.now()

  return new Promise((gut) => {
    const anfrage = http.request({
      host: ziel.hostname, port: ziel.port, path: ziel.pathname + ziel.search, method: methode, agent: dienst.agent,
      headers: {
        Authorization: `Bearer ${nachweis ?? dienst.marke}`,
        ...(daten ? { 'Content-Type': 'application/json', 'Content-Length': daten.length } : {}),
        ...koepfe,
      },
      timeout: frist,
    }, (antwort) => {
      const teile = []
      antwort.on('data', (teil) => teile.push(teil))
      antwort.on('end', () => gut({
        status: antwort.statusCode, ms: performance.now() - beginn, koepfe: antwort.headers,
        warten: Number(antwort.headers['retry-after'] ?? 0), text: Buffer.concat(teile).toString('utf8'),
      }))
      antwort.on('error', (fehler) => gut({ status: 0, ms: performance.now() - beginn, fehler: fehler.message }))
    })

    anfrage.on('timeout', () => anfrage.destroy(new Error('Frist')))
    anfrage.on('error', (fehler) => gut({ status: 0, ms: performance.now() - beginn, fehler: fehler.message }))
    anfrage.end(daten)
  })
}

// mcpUeberHTTP ruft ein Werkzeug über den HTTP-Transport — eine Nachricht
// je Anfrage, ohne Sitzungskennung, wie `internal/mcp/ueberhttp.go` es sagt.
let mcpKennung = 0

export async function mcpUeberHTTP(dienst, werkzeug, argumente) {
  mcpKennung += 1
  const antwort = await ruf(dienst, 'POST', dienst.pfad, {
    rumpf: { jsonrpc: '2.0', id: mcpKennung, method: 'tools/call', params: { name: werkzeug, arguments: argumente } },
    koepfe: { 'MCP-Protocol-Version': '2025-06-18', Accept: 'application/json' },
  })

  return { ...antwort, ...mcpErgebnis(antwort.text) }
}

// mcpErgebnis liest `isError` und die Wartezeit aus dem Text eines
// abgewiesenen Rufs — MCP kennt kein 429, nur den Satz „in N s wieder".
export function mcpErgebnis(text) {
  try {
    const gelesen = JSON.parse(text ?? '')
    const inhalt = gelesen.result?.content?.[0]?.text ?? ''
    const gedrosselt = /zu viele Rufe — in (\d+) s wieder/.exec(inhalt)

    return {
      mcpFehler: gelesen.result?.isError === true || gelesen.error !== undefined,
      gedrosselt: gedrosselt !== null, warten: gedrosselt ? Number(gedrosselt[1]) : 0, inhalt,
    }
  } catch {
    return { mcpFehler: true, gedrosselt: false, warten: 0, inhalt: text }
  }
}

// langsamerKlient öffnet eine Verbindung und schickt einen halben Kopf — und
// dann nichts mehr. Gibt eine Funktion, die sie schließt.
export function langsamerKlient(dienst) {
  const ziel = new URL(dienst.adresse)
  const leitung = net.connect({ host: ziel.hostname, port: Number(ziel.port) })
  leitung.on('error', () => {})
  leitung.on('connect', () => leitung.write(`GET /grenzen HTTP/1.1\r\nHost: ${ziel.host}\r\n`))

  return () => leitung.destroy()
}

// speicher nimmt alle 250 ms den belegten Arbeitsspeicher (RSS) der Prozesse
// auf, in MiB und gelesen über `ps`: je Prozess den Wert am Anfang, den
// höchsten und den letzten.
export function speicher(pids) {
  const stand = new Map()
  const messen = () => {
    try {
      const zeilen = execFileSync('ps', ['-o', 'pid=,rss=', '-p', pids.join(',')], { encoding: 'utf8' })
      for (const zeile of zeilen.trim().split('\n')) {
        const [pid, kib] = zeile.trim().split(/\s+/).map(Number)
        const mib = kib / 1024
        const bisher = stand.get(pid) ?? { anfang: mib, hoechst: 0 }
        stand.set(pid, { ...bisher, hoechst: Math.max(bisher.hoechst, mib), zuletzt: mib })
      }
    } catch { /* ein beendeter Prozess hat keinen Speicher mehr */ }
  }

  messen()
  const takt = setInterval(messen, 250)

  return () => { clearInterval(takt); messen(); return Object.fromEntries(stand) }
}
