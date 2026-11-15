// Prüfungen der Fristen der Hüllenprobe — T02
// (#2096).
//
// **Eine Anwendung, die nicht mehr antwortet, ist ein roter Fall und kein
// stehender Lauf.** Bis zum 2026-10-06 warteten die Leitung, der Ruf an Wails'
// MCP und `warte` ohne Grenze: ein hängendes WebView hielt die ganze Probe
// fest, und kein Drehbuch danach lief. Geprüft gegen einen schweigenden
// Socket und einen schweigenden HTTP-Dienst — und gegen langsame, aber
// gültige Antworten, die weiter gelten.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { createServer as httpDienst } from 'node:http'
import { createServer } from 'node:net'
import { mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { Leitung, schlafe } from './huellenleitung.mjs'
import { Probe } from './huellenprobe.mjs'

// socketdienst antwortet auf jede Zeile nach `verzug` ms — oder nie.
async function socketdienst(verzug) {
  const ort = join(mkdtempSync(join(tmpdir(), 'leitung-')), 's.sock')
  const dienst = createServer((s) => {
    s.setEncoding('utf8')
    s.on('data', () => {
      if (verzug !== undefined) setTimeout(() => s.write('{"ok":true}\n'), verzug)
    })
    s.on('error', () => {})
  })
  await new Promise((los) => dienst.listen(ort, los))

  return { ort, dienst }
}

test('eine schweigende Leitung endet nach ihrer Frist, nennt den Befehl und schließt', async () => {
  const { ort, dienst } = await socketdienst(undefined)
  const leitung = new Leitung(ort)
  await leitung.verbinde()
  const anfang = Date.now()
  const antwort = await leitung.frage({ was: 'klick' }, 200)
  assert.ok(Date.now() - anfang < 2000)
  assert.equal(antwort.ok, false)
  assert.match(antwort.fehler, /klick.*200 ms/)
  assert.equal(leitung.socket.destroyed, true)
  assert.deepEqual(await leitung.frage({ was: 'danach' }, 200), { ok: false, fehler: 'die Leitung ist zu', zu: true })
  dienst.close()
})

test('eine langsame, aber gültige Antwort gilt weiter', async () => {
  const { ort, dienst } = await socketdienst(100)
  const leitung = new Leitung(ort)
  await leitung.verbinde()
  assert.deepEqual(await leitung.frage({ was: 'klick' }, 2000), { ok: true })
  leitung.schliesse()
  dienst.close()
})

// mcpdienst antwortet nach `verzug` ms mit einem Ergebnis — oder nie.
async function mcpdienst(verzug) {
  const offen = []
  const dienst = httpDienst((anfrage, antwort) => {
    offen.push(antwort)
    if (verzug === undefined) return
    setTimeout(() => {
      antwort.setHeader('content-type', 'application/json')
      antwort.end(JSON.stringify({ jsonrpc: '2.0', id: 1, result: { content: [{ type: 'text', text: '"ok"' }] } }))
    }, verzug)
  })
  await new Promise((los) => dienst.listen(0, '127.0.0.1', los))

  return { dienst, port: dienst.address().port, offen }
}

test('ein schweigender MCP-Ruf endet nach seiner Frist und nennt das Werkzeug', async () => {
  const { dienst, port, offen } = await mcpdienst(undefined)
  const probe = new Probe('fristen-test')
  probe.port = port
  const anfang = Date.now()
  await assert.rejects(probe.mcp('js_eval', { js: 'return 1' }, 200), /js_eval.*200 ms/)
  assert.ok(Date.now() - anfang < 2000)
  for (const antwort of offen) antwort.destroy()
  dienst.close()
})

test('ein langsamer, gültiger MCP-Ruf gilt weiter', async () => {
  const { dienst, port } = await mcpdienst(100)
  const probe = new Probe('fristen-test')
  probe.port = port
  assert.equal(await probe.mcp('js_eval', { js: 'return 1' }, 2000), '"ok"')
  dienst.close()
})

test('warte endet nach ihrer Frist, auch wenn die Bedingung nie zurückkehrt', async () => {
  const probe = new Probe('fristen-test')
  const anfang = Date.now()
  await assert.rejects(probe.warte('die Seite antwortet', () => new Promise(() => {}), 300), /abgebrochen: die Seite antwortet/)
  assert.ok(Date.now() - anfang < 2000)
  assert.equal(probe.fehler, 1)
  assert.match(probe.zeilen.at(-1), /die Seite antwortet \(nach 300 ms\)/)
})

test('warte nimmt eine Bedingung, die erst spät wahr wird', async () => {
  const probe = new Probe('fristen-test')
  let n = 0
  assert.equal(await probe.warte('zählt', async () => { await schlafe(20); n += 1; return n >= 3 }, 2000), true)
})
