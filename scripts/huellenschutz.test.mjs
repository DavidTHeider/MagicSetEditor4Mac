// Prüfungen von scripts/huellenschutz.mjs — SEC-17
// (#2164).
//
// **Jede Hüllenprobe liest nach ihrem Lauf die Protokolle ihrer Orte** und
// streicht jeden Verstoß gegen die Schutzrichtlinie an, den das Drehbuch
// nicht ausdrücklich erwartet: so belegt jedes Drehbuch mit, dass die
// Richtlinie der Anwendung nichts wegnimmt, was es benutzt.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { cspVerstoesse } from './huellenschutz.mjs'

const zeile = (anweisung, abgewiesen) =>
  `time=2026-10-07T09:35:47+02:00 level=WARN msg=CSP-Verstoß anweisung=${anweisung} abgewiesen=${abgewiesen} stelle=index.js:15`

function ortMit(...zeilen) {
  const ort = mkdtempSync(join(tmpdir(), 'huellenschutz-'))
  writeFileSync(join(ort, 'mse.log'), `${['level=INFO msg="der Bestand ist offen"', ...zeilen].join('\n')}\n`)

  return { ort, bestand: join(ort, 'bestand.db') }
}

test('ein Ort ohne Verstoß gibt nichts', () => {
  const ort = ortMit()
  assert.deepEqual(cspVerstoesse([ort]), [])
  rmSync(ort.ort, { recursive: true })
})

test('jeder Verstoß jedes Ortes wird genannt', () => {
  const eins = ortMit(zeile('img-src', 'data'))
  const zwei = ortMit(zeile('worker-src', 'blob'), zeile('style-src-elem', 'inline'))
  const gefunden = cspVerstoesse([eins, zwei])
  assert.equal(gefunden.length, 3)
  assert.ok(gefunden.some((z) => z.includes('worker-src')))
  rmSync(eins.ort, { recursive: true })
  rmSync(zwei.ort, { recursive: true })
})

test('was das Drehbuch erwartet, zählt nicht', () => {
  const ort = ortMit(zeile('script-src-elem', 'inline'), zeile('img-src', 'data'))
  const gefunden = cspVerstoesse([ort], /anweisung=script-src-elem abgewiesen=inline/)
  assert.deepEqual(gefunden.map((z) => z.includes('img-src')), [true])
  rmSync(ort.ort, { recursive: true })
})

test('ein Ort ohne Protokoll ist keiner mit Verstoß', () => {
  const ort = mkdtempSync(join(tmpdir(), 'huellenschutz-'))
  assert.deepEqual(cspVerstoesse([{ ort, bestand: join(ort, 'bestand.db') }]), [])
  rmSync(ort, { recursive: true })
})

// **Jeder Kanal, den die Anwendung kennt** — angestrichen vom Zweitblick über
// Phase H: `ops04` stellt Ort **und** Format um, und ein Wächter, der nur
// `mse.log` am Vorgabeort in Textform las, blieb dort grün.
const jsonzeile = (anweisung, abgewiesen) =>
  JSON.stringify({ time: '2026-10-07T09:35:47+02:00', level: 'WARN', msg: 'CSP-Verstoß', anweisung, abgewiesen, stelle: 'index.js:15' })

test('ein Verstoß im JSON-Protokoll wird genannt, in derselben Gestalt', () => {
  const ort = ortMit(jsonzeile('img-src', 'https://cards.scryfall.io'))
  const gefunden = cspVerstoesse([ort])
  assert.equal(gefunden.length, 1)
  assert.match(gefunden[0], /anweisung=img-src abgewiesen=https:\/\/cards\.scryfall\.io/)
  assert.deepEqual(cspVerstoesse([ort], /anweisung=img-src/), [])
  rmSync(ort.ort, { recursive: true })
})

test('ein Protokoll an eigenem Ort und ein gedrehtes werden gelesen', () => {
  const ort = ortMit()
  mkdirSync(join(ort.ort, 'Protokolle'))
  writeFileSync(join(ort.ort, 'Protokolle', 'mse.log'), `${zeile('font-src', 'data')}\n`)
  writeFileSync(join(ort.ort, 'mse.log.1'), `${zeile('media-src', 'blob')}\n`)
  const gefunden = cspVerstoesse([ort])
  assert.equal(gefunden.length, 2, gefunden.join(' | '))
  rmSync(ort.ort, { recursive: true })
})

test('ein unlesbarer Bericht ist ein Befund und kein Schweigen', () => {
  const text = ortMit('time=2026-10-07T09:35:47+02:00 level=WARN msg="CSP-Bericht unlesbar" anzahl=12')
  const json = ortMit(JSON.stringify({ level: 'WARN', msg: 'CSP-Bericht unlesbar', anzahl: 12 }))
  assert.equal(cspVerstoesse([text]).length, 1)
  assert.equal(cspVerstoesse([json]).length, 1)
  rmSync(text.ort, { recursive: true })
  rmSync(json.ort, { recursive: true })
})
