// Prüfungen von scripts/leistungsprozess.mjs — QA-04.AK-03
// (#1735).
//
// **Ein Prozess ist hier ein Ereignisträger mit drei Strömen**; was
// `/usr/bin/time -l` am Ende schreiben würde, schreibt der Test.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { EventEmitter } from 'node:events'
import { PassThrough } from 'node:stream'

import { prozesse } from './leistungsprozess.mjs'

const zeitAusgabe = '        0.40 real         0.25 user         0.05 sys\n' +
  '            52428800  maximum resident set size\n' +
  '            41943040  peak memory footprint\n'

/** gefaelscht gibt eine Erzeugung, die Aufrufe festhält und ein Ende vorgibt. */
function gefaelscht({ aus = '', fehler = zeitAusgabe, rueckgabe = 0, signal = null } = {}) {
  const aufrufe = []
  const erzeuge = (befehl, argumente, optionen) => {
    const kind = new EventEmitter()
    kind.pid = 4242
    kind.stdin = new PassThrough()
    kind.stdout = new PassThrough()
    kind.stderr = new PassThrough()
    const eingabe = []
    kind.stdin.on('data', (d) => eingabe.push(String(d)))
    aufrufe.push({ befehl, argumente, optionen, eingabe })
    kind.stdin.on('finish', () => {
      kind.stdout.end(aus)
      kind.stderr.end(fehler)
      setImmediate(() => kind.emit('close', rueckgabe, signal))
    })
    return kind
  }
  return { erzeuge, aufrufe }
}

test('messe läuft unter /usr/bin/time -l und liest CPU und Speicher daraus', async () => {
  const { erzeuge, aufrufe } = gefaelscht({ aus: '{"ok":true}' })
  const { messe } = prozesse(erzeuge)
  const m = await messe('/bin/mse', ['sets', '--json'], { umgebung: { HOME: '/z' }, ort: '/o', eingabe: ':quit\n' })
  assert.equal(aufrufe[0].befehl, '/usr/bin/time')
  assert.deepEqual(aufrufe[0].argumente, ['-l', '/bin/mse', 'sets', '--json'])
  assert.deepEqual(aufrufe[0].optionen, { env: { HOME: '/z' }, cwd: '/o' })
  assert.deepEqual(aufrufe[0].eingabe, [':quit\n'])
  assert.equal(m.cpuS, 0.3)
  assert.equal(m.rssByte, 52428800)
  assert.equal(m.spitzeByte, 41943040)
  assert.equal(m.aus, '{"ok":true}')
  assert.equal(m.fehler, null)
  assert.equal(m.absturz, false)
  assert.ok(m.wandMs >= 0)
  assert.equal(typeof m.last, 'number', 'die Last des Rechners steht bei jedem Fall')
})

test('ein Rückgabewert ungleich null ist ein Fehler, aber kein Absturz', async () => {
  const { messe } = prozesse(gefaelscht({ rueckgabe: 2 }).erzeuge)
  const m = await messe('/bin/x', [])
  assert.equal(m.fehler, 'Rückgabe 2')
  assert.equal(m.absturz, false)
})

test('ein Signal ist ein Absturz', async () => {
  const { messe } = prozesse(gefaelscht({ fehler: 'Command terminated abnormally.\n' + zeitAusgabe, rueckgabe: 139 }).erzeuge)
  const m = await messe('/bin/x', [])
  assert.equal(m.absturz, true)
  assert.match(m.fehler, /abgebrochen/)
})

test('fehlt die Zeitangabe, gilt der Lauf als abgebrochen — es gibt keine Zahlen', async () => {
  const { messe } = prozesse(gefaelscht({ fehler: 'kaputt\n' }).erzeuge)
  const m = await messe('/bin/x', [])
  assert.equal(m.absturz, true)
  assert.equal(m.cpuS, null)
})

test('starte hält den Prozess offen, bis die Sitzung schließt', async () => {
  const { erzeuge, aufrufe } = gefaelscht()
  const { starte } = prozesse(erzeuge)
  const lauf = starte('/bin/mse-mcp', [], { umgebung: { A: '1' } })
  lauf.kind.stdin.write('{"jsonrpc":"2.0"}\n')
  const ende = await lauf.schliessen()
  assert.deepEqual(aufrufe[0].argumente, ['-l', '/bin/mse-mcp'])
  assert.deepEqual(aufrufe[0].eingabe, ['{"jsonrpc":"2.0"}\n'])
  assert.equal(ende.rssByte, 52428800)
  assert.equal(ende.fehler, null)
})
