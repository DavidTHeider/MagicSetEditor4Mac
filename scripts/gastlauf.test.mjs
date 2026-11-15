// **Wer am Wirt laufen darf, sagt der Aufruf ausdrücklich** — Befund 1 des
// Zweitblicks über Phase I: `node scripts/bindunggast.mjs ''` umging die
// Sperre »nie am Wirt«, denn ein leerer Name überstand `??`, und `fahre()`
// las »keine VM« als »am Wirt«. Die Bindungsprobe hätte Herausgeber am Wirt
// eingerichtet, die Quick-Look-Probe dort den Anwendungsordner gelöscht.
import { spawn } from 'node:child_process'
import { chmodSync, existsSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { fahre, gastziel, imGeteilten } from './huellenprobe/gastlauf.mjs'

test('ohne Angabe der vorgegebene Gast', () => {
  assert.deepEqual(gastziel(undefined, { wirtErlaubt: false }), { vm: 'mse080-152', amWirt: false })
})

test('ein genannter Gast', () => {
  assert.deepEqual(gastziel('mse080-14', { wirtErlaubt: false }), { vm: 'mse080-14', amWirt: false })
})

test('--wirt nur, wo es erlaubt ist', () => {
  assert.deepEqual(gastziel('--wirt', { wirtErlaubt: true }), { vm: null, amWirt: true })
  assert.throws(() => gastziel('--wirt', { wirtErlaubt: false }), /nie am Wirt/)
})

test('ein leerer oder fremder Name ist kein Gast', () => {
  for (const name of ['', ' ', '-x', '--', 'a b', '../vm']) {
    assert.throws(() => gastziel(name, { wirtErlaubt: true }), /kein Gast/, JSON.stringify(name))
  }
})

test('fahre läuft am Wirt nur auf ausdrückliches amWirt', async () => {
  await assert.rejects(fahre({ vm: '', geteilt: '/gibt/es/nicht', name: 'x', treiber: 'x.pl' }), /weder Gast noch Wirt/)
  await assert.rejects(fahre({ vm: null, geteilt: '/gibt/es/nicht', name: 'x', treiber: 'x.pl', amWirt: false }), /weder Gast noch Wirt/)
})

// **Das geteilte Verzeichnis geht in jedem Fall weg** — samt Vorbereitung
// (Befund 1 des Nachlaufs über Phase I): ein gescheiterter Compileraufruf
// ließ es liegen, und ein Signal übersprang jedes `finally`.
test('imGeteilten nimmt das Verzeichnis nach Erfolg weg', async () => {
  let ort
  const wert = await imGeteilten('probe', (geteilt) => {
    ort = geteilt
    assert.ok(existsSync(geteilt))
    return 7
  })
  assert.equal(wert, 7)
  assert.equal(existsSync(ort), false)
})

test('imGeteilten nimmt es auch weg, wenn die Vorbereitung scheitert', async () => {
  let ort
  await assert.rejects(imGeteilten('probe', (geteilt) => {
    ort = geteilt
    throw new Error('swiftc')
  }), /swiftc/)
  assert.equal(existsSync(ort), false)
})

test('imGeteilten nimmt es nach einem Signal weg', async () => {
  const modul = join(import.meta.dirname, 'huellenprobe/gastlauf.mjs')
  const kind = spawn(process.execPath, ['--input-type=module', '-e',
    `import { imGeteilten } from ${JSON.stringify(modul)}; await imGeteilten('probe', (g) => { console.log(g); return new Promise(() => { setInterval(() => {}, 1000) }) })`])
  const ort = await new Promise((da) => kind.stdout.once('data', (d) => da(d.toString().trim())))
  assert.ok(existsSync(ort))
  kind.kill('SIGTERM')
  const code = await new Promise((da) => kind.once('exit', (c, s) => da(c ?? s)))
  assert.equal(existsSync(ort), false, `noch da nach ${String(code)}`)
})

test('imGeteilten nimmt es weg, wenn der Läufer selbst endet', async () => {
  const modul = join(import.meta.dirname, 'huellenprobe/gastlauf.mjs')
  const kind = spawn(process.execPath, ['--input-type=module', '-e',
    `import { imGeteilten } from ${JSON.stringify(modul)}; await imGeteilten('probe', (g) => { console.log(g); process.exit(1) })`])
  const ort = await new Promise((da) => kind.stdout.once('data', (d) => da(d.toString().trim())))
  await new Promise((da) => kind.once('exit', da))
  assert.equal(existsSync(ort), false)
})

// **Was der Treiber zuletzt schreibt, steht auf der Platte, bevor die VM
// stoppt** — gemessen am 2026-10-11: `fahre` beendet die VM mit SIGTERM an
// `lume run`, und eine Datei, die Sekunden davor entstand, war nach dem
// Neustart weg; mit `sync` davor blieb sie. Verloren ging so das Aufräumen
// jeder Gastprobe — Bünde in der Suchliste, das Probebündel von Quick Look.
test('fahre schreibt die Platte des Gastes fest, bevor es ihn stoppt', async () => {
  const ort = mkdtempSync(join(tmpdir(), 'lume-attrappe-'))
  const protokoll = join(ort, 'protokoll')
  writeFileSync(join(ort, 'lume'), '#!/bin/bash\necho "$*" >> "$LUME_PROTOKOLL"\n[ "$1" = run ] && exec sleep 60\nexit 0\n')
  chmodSync(join(ort, 'lume'), 0o755)
  const modul = join(import.meta.dirname, 'huellenprobe/gastlauf.mjs')
  const kind = spawn(process.execPath, ['--input-type=module', '-e',
    `import { fahre } from ${JSON.stringify(modul)}; await fahre({ vm: 'mse080-152', geteilt: ${JSON.stringify(ort)}, name: 'p', treiber: 't.pl' })`],
  { env: { ...process.env, PATH: `${ort}:${process.env.PATH}`, LUME_PROTOKOLL: protokoll } })
  const code = await new Promise((da) => kind.once('exit', da))
  assert.equal(code, 0)
  const rufe = readFileSync(protokoll, 'utf8').trim().split('\n')
  const treiber = rufe.findIndex((r) => r.includes("perl '"))
  const sync = rufe.findIndex((r) => /^ssh mse080-152 .*\bsync\b/.test(r))
  assert.ok(treiber >= 0, rufe.join(' | '))
  assert.ok(sync > treiber, `kein sync nach dem Treiber: ${rufe.join(' | ')}`)
})
