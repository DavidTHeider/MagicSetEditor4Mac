// Prüfungen von scripts/leistungssitzung.mjs — QA-04
// (#1735).
//
// **Sie startet kein Programm.** Ein Kindprozess ist hier ein Paar aus zwei
// Strömen, und die Antworten schreibt der Test — so, wie MSE2 sie im
// Rohmodus schreibt (`src/cli/text_io_handler.cpp`, `flushRaw`, Fassung
// f9d9356 vom 2021-07-06) und wie `mse-mcp` je Zeile eine JSON-Antwort gibt.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { PassThrough } from 'node:stream'

import { McpSitzung, Mse2Sitzung, Rohleser } from './leistungssitzung.mjs'

test('ein Satz im Rohmodus ist Status, Zeilenzahl und die Zeilen', () => {
  const leser = new Rohleser()
  assert.deepEqual(leser.fuettere('0\n1\n"a.png"\n'), [{ status: 0, zeilen: ['"a.png"'] }])
})

test('ein Satz darf in Stücken kommen und mehrere in einem', () => {
  const leser = new Rohleser()
  assert.deepEqual(leser.fuettere('2\n2\nERROR: x'), [])
  assert.deepEqual(leser.fuettere('\n  y\n0\n0\n1\n'), [
    { status: 2, zeilen: ['ERROR: x', '  y'] },
    { status: 0, zeilen: [] },
  ])
  assert.deepEqual(leser.fuettere('1\nWARNING: z\n'), [{ status: 1, zeilen: ['WARNING: z'] }])
})

test('was vor dem ersten Satz steht, ist kein Satz — die Meldungen beim Laden', () => {
  // **Gemessen am 2026-10-02**: steht das Set vor `--raw`, schreibt MSE2 die
  // Warnungen und Fehler beim Laden als Text, bevor der Rohmodus gilt
  // (`src/main.cpp`, Zeilen 221–231: die Argumente werden der Reihe nach
  // gelesen). Darunter stehen Zeilen wie „On line 28:" — aber nie zwei Zahlen
  // allein untereinander.
  const leser = new Rohleser()
  const vorspann = '\u001b[1;33mWARNING: \u001b[0mWarnings while reading file:\n         On line 28: \tUnexpected key\n' +
    '\u001b[1;31mERROR: \u001b[0mVariable not set: extra_card\n         while updating card value \'vorthos_box\'\n'
  assert.deepEqual(leser.fuettere(vorspann + '0\n1\n70\n'), [{ status: 0, zeilen: ['70'] }])
  assert.deepEqual(leser.fuettere('0\n1\n7\n'), [{ status: 0, zeilen: ['7'] }], 'danach zählt jede Zeile')
})

/**
 * kind baut ein Paar Ströme, das wie ein Kindprozess aussieht — samt einer
 * Warteschlange der Zeilen, die das Programm bekommt. **Ein einziger Hörer**:
 * wer ihn je Zeile an- und abmeldete, verlöre, was dazwischen ankommt.
 */
function kind() {
  const k = { stdin: new PassThrough(), stdout: new PassThrough() }
  const zeilen = []
  const wartende = []
  let puffer = ''
  k.stdin.setEncoding('utf8')
  k.stdin.on('data', (stueck) => {
    puffer += stueck
    let ende
    while ((ende = puffer.indexOf('\n')) >= 0) {
      zeilen.push(puffer.slice(0, ende))
      puffer = puffer.slice(ende + 1)
    }
    while (zeilen.length > 0 && wartende.length > 0) wartende.shift()(zeilen.shift())
  })
  k.zeile = () => new Promise((fertig) => {
    if (zeilen.length > 0) fertig(zeilen.shift())
    else wartende.push(fertig)
  })
  return k
}

test('MSE2: ein Befehl geht als Zeile hinaus, die Antwort kommt mit ihrer Dauer', async () => {
  const k = kind()
  const sitzung = new Mse2Sitzung(k)
  const antwort = sitzung.befehl('1 + 1')
  assert.equal(await k.zeile(), '1 + 1')
  k.stdout.write('0\n1\n2\n')
  const ergebnis = await antwort
  assert.equal(ergebnis.status, 0)
  assert.deepEqual(ergebnis.zeilen, ['2'])
  assert.ok(ergebnis.dauerMs >= 0)
})

test('MSE2: zwei Befehle nacheinander bekommen jeder ihren Satz', async () => {
  const k = kind()
  const sitzung = new Mse2Sitzung(k)
  const erste = sitzung.befehl('a')
  k.stdout.write('0\n1\n"a"\n')
  assert.deepEqual((await erste).zeilen, ['"a"'])
  const zweite = sitzung.befehl('b')
  k.stdout.write('2\n1\nERROR: b\n')
  assert.equal((await zweite).status, 2)
})

test('MSE2: ein Bruch des Protokolls nach dem ersten Satz lässt den Befehl scheitern', async () => {
  const k = kind()
  const sitzung = new Mse2Sitzung(k)
  const erste = sitzung.befehl('a')
  k.stdout.write('0\n0\n')
  await erste
  const zweite = sitzung.befehl('b')
  k.stdout.write('Speicherzugriffsfehler\nirgendwas\n')
  await assert.rejects(zweite, /kein Satz des Rohmodus/)
})

test('MSE2: endet der Prozess vor der Antwort, scheitert der Befehl', async () => {
  const k = kind()
  const sitzung = new Mse2Sitzung(k)
  const antwort = sitzung.befehl('x')
  k.stdout.end()
  await assert.rejects(antwort, /beendet/)
})

/** antworte liest eine JSON-Zeile und gibt die Antwort dazu. */
async function antworte(k, ergebnis) {
  const anfrage = JSON.parse(await k.zeile())
  k.stdout.write(JSON.stringify({ jsonrpc: '2.0', id: anfrage.id, result: ergebnis(anfrage) }) + '\n')
  return anfrage
}

test('MCP: ein Werkzeugruf geht als tools/call hinaus und kommt als JSON zurück', async () => {
  const k = kind()
  const sitzung = new McpSitzung(k)
  const antwort = sitzung.werkzeug('lesen.karten', { setID: 1 })
  const anfrage = await antworte(k, () => ({
    content: [{ type: 'text', text: '[{"id":7}]' }], isError: false,
  }))
  assert.equal(anfrage.method, 'tools/call')
  assert.deepEqual(anfrage.params, { name: 'lesen.karten', arguments: { setID: 1 } })
  const ergebnis = await antwort
  assert.deepEqual(ergebnis.wert, [{ id: 7 }])
  assert.ok(ergebnis.dauerMs >= 0)
})

test('MCP: ein Ergebnis mit isError ist ein Fehler, und sein Text steht darin', async () => {
  const k = kind()
  const sitzung = new McpSitzung(k)
  const antwort = sitzung.werkzeug('schreiben.karte.zeichnen', { karteID: 1, ziel: '/x.png' })
  await antworte(k, () => ({ content: [{ type: 'text', text: 'Ziel liegt schon da' }], isError: true }))
  await assert.rejects(antwort, /Ziel liegt schon da/)
})

test('MCP: eine Antwort, die kein JSON ist, bleibt Text', async () => {
  const k = kind()
  const sitzung = new McpSitzung(k)
  const antwort = sitzung.werkzeug('lesen.x', {})
  await antworte(k, () => ({ content: [{ type: 'text', text: 'nur Text' }], isError: false }))
  assert.equal((await antwort).wert, 'nur Text')
})

test('MCP: der Anfang ist initialize und danach die Meldung initialized', async () => {
  const k = kind()
  const sitzung = new McpSitzung(k)
  const fertig = sitzung.beginne()
  const anfrage = await antworte(k, () => ({ protocolVersion: '2025-06-18' }))
  assert.equal(anfrage.method, 'initialize')
  assert.equal(await k.zeile(), JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized' }))
  await fertig
})

test('MCP: Antworten werden nach ihrer Kennung zugeordnet, nicht nach der Reihe', async () => {
  const k = kind()
  const sitzung = new McpSitzung(k)
  const a = sitzung.rufe('ping', {})
  const b = sitzung.rufe('ping', {})
  k.stdout.write(JSON.stringify({ jsonrpc: '2.0', id: 2, result: { b: true } }) + '\n')
  k.stdout.write(JSON.stringify({ jsonrpc: '2.0', id: 1, result: { a: true } }) + '\n')
  assert.deepEqual(await a, { a: true })
  assert.deepEqual(await b, { b: true })
})

test('MCP: ein Protokollfehler scheitert mit seiner Meldung', async () => {
  const k = kind()
  const sitzung = new McpSitzung(k)
  const a = sitzung.rufe('unbekannt', {})
  k.stdout.write(JSON.stringify({ jsonrpc: '2.0', id: 1, error: { code: -32601, message: 'unbekannte Methode' } }) + '\n')
  await assert.rejects(a, /unbekannte Methode/)
})
