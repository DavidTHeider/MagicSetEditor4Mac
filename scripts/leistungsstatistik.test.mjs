// Prüfungen von scripts/leistungsstatistik.mjs — QA-04
// (#1735).
//
// **Jede Rechnung an Zahlen, die von Hand nachzurechnen sind** — dieselbe
// Regel wie für den Kanonlauf (`docs/MESSMETHODIK.md` §8).

import { test } from 'node:test'
import assert from 'node:assert/strict'

import {
  quantil,
  reihenfolge,
  reihenfolgeEffekt,
  verteilung,
  zeitLesen,
} from './leistungsstatistik.mjs'

test('Quantil nach der Nächste-Rang-Regel, ohne Interpolation', () => {
  const werte = [10, 1, 9, 2, 8, 3, 7, 4, 6, 5]
  assert.equal(quantil(werte, 0.5), 5, 'k = ⌈0,5 · 10⌉ = 5')
  assert.equal(quantil(werte, 0.95), 10, 'k = ⌈9,5⌉ = 10')
  assert.equal(quantil(werte, 0.25), 3, 'k = ⌈2,5⌉ = 3')
  assert.equal(quantil(werte, 0), 1, 'p = 0 ist das Minimum')
  assert.equal(quantil(werte, 1), 10)
  assert.deepEqual(werte, [10, 1, 9, 2, 8, 3, 7, 4, 6, 5], 'die Eingabe bleibt ungeordnet')
})

test('ein Quantil über nichts ist keines', () => {
  assert.equal(quantil([], 0.5), null)
})

test('die Verteilung nennt n, Median, p95, p99 und die Ränder', () => {
  const werte = Array.from({ length: 100 }, (_, i) => i + 1)
  assert.deepEqual(verteilung(werte), {
    n: 100, min: 1, median: 50, p95: 95, p99: 99, max: 100, mittel: 50.5,
  })
})

test('bei fünf Werten sind p95 und p99 das Maximum — und das steht so da', () => {
  const v = verteilung([3, 1, 2, 5, 4])
  assert.equal(v.median, 3)
  assert.equal(v.p95, 5)
  assert.equal(v.p99, 5)
})

test('eine leere Verteilung trägt n = 0 und sonst nichts', () => {
  assert.deepEqual(verteilung([]), { n: 0 })
})

const zeitAusgabe = `        1.52 real         0.31 user         0.07 sys
            59768832  maximum resident set size
                   0  average shared memory size
                  11  involuntary context switches
            46596352  peak memory footprint
`

test('die Ausgabe von /usr/bin/time -l wird gelesen', () => {
  assert.deepEqual(zeitLesen(zeitAusgabe), {
    wand: 1.52, nutzer: 0.31, system: 0.07, rss: 59768832, spitze: 46596352, abgebrochen: false,
  })
})

test('steht vor der Zeitangabe Protokoll des Programms, zählt nur der Schluss', () => {
  const text = 'ERROR: 1.00 real kein Messwert\nirgendwas\n' + zeitAusgabe
  assert.equal(zeitLesen(text).wand, 1.52)
})

test('ein abgebrochener Prozess wird als solcher erkannt', () => {
  const text = 'Command terminated abnormally.\n' + zeitAusgabe
  assert.equal(zeitLesen(text).abgebrochen, true)
})

test('ohne Zeitangabe gibt es keine Zahlen', () => {
  assert.equal(zeitLesen('nur Protokoll\n'), null)
})

test('die Reihenfolge wechselt je Wiederholung — jedes Programm steht gleich oft vorn', () => {
  const namen = ['mse2', 'mse4mac', 'mse050']
  assert.deepEqual(reihenfolge(0, namen), ['mse2', 'mse4mac', 'mse050'])
  assert.deepEqual(reihenfolge(1, namen), ['mse4mac', 'mse050', 'mse2'])
  assert.deepEqual(reihenfolge(2, namen), ['mse050', 'mse2', 'mse4mac'])
  assert.deepEqual(reihenfolge(3, namen), ['mse2', 'mse4mac', 'mse050'])
  assert.deepEqual(namen, ['mse2', 'mse4mac', 'mse050'], 'die Eingabe bleibt')
})

test('der Reihenfolgeeffekt vergleicht den Median vorn mit dem dahinter', () => {
  const laeufe = [
    { programm: 'a', platz: 0, wert: 10 },
    { programm: 'a', platz: 0, wert: 12 },
    { programm: 'a', platz: 1, wert: 20 },
    { programm: 'a', platz: 2, wert: 22 },
    { programm: 'b', platz: 1, wert: 5 },
  ]
  assert.deepEqual(reihenfolgeEffekt(laeufe), {
    a: { vorn: 10, dahinter: 20, nVorn: 2, nDahinter: 2 },
    b: { vorn: null, dahinter: 5, nVorn: 0, nDahinter: 1 },
  })
})
