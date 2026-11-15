// Prüfungen von scripts/leistungsbericht.mjs — QA-04.AK-03
// (#1735).

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { bilanz, bilanzZeile, belegJson, berichtMarkdown, schritteZusammenfassen, zusammenfassen } from './leistungsbericht.mjs'

const faelle = [
  { aufgabe: 'start', satz: '', programm: 'mse2', wiederholung: 0, platz: 0, wandMs: 30, cpuS: 0.02, rssByte: 10e6, fehler: null, absturz: false },
  { aufgabe: 'start', satz: '', programm: 'mse2', wiederholung: 1, platz: 1, wandMs: 10, cpuS: 0.01, rssByte: 12e6, fehler: null, absturz: false },
  { aufgabe: 'start', satz: '', programm: 'mse2', wiederholung: 2, platz: 0, wandMs: 20, cpuS: 0.03, rssByte: 11e6, fehler: 'Rückgabe 1', absturz: false },
  { aufgabe: 'start', satz: '', programm: 'heute', wiederholung: 0, platz: 1, wandMs: 50, cpuS: 0.05, rssByte: 30e6, fehler: null, absturz: true },
]

test('je Aufgabe, Satz und Programm: Verteilungen, die erste Wiederholung, Fehler und Abstürze', () => {
  const z = zusammenfassen(faelle)
  assert.deepEqual(Object.keys(z), ['start'])
  const mse2 = z.start.mse2
  assert.equal(mse2.faelle, 3)
  assert.equal(mse2.fehler, 1)
  assert.equal(mse2.wand.n, 2, 'die Zeit steht nur für die Fälle ohne Fehler')
  assert.equal(mse2.wand.median, 10)
  assert.equal(mse2.wand.max, 30)
  assert.equal(mse2.ersteWandMs, 30, 'die erste Wiederholung steht gesondert')
  assert.equal(mse2.cpu.median, 0.01)
  assert.equal(mse2.rss.max, 12e6)
  assert.equal(mse2.abstuerze, 0)
  assert.deepEqual(mse2.reihenfolge, { vorn: 30, dahinter: 10, nVorn: 1, nDahinter: 1 })
  assert.equal(z.start.heute.abstuerze, 1)
})

// **Ein Fall, der nicht vollständig war, hat keine Zeit, die man vergleicht.**
// Im Beleg vom 2026-10-03 zeichnete MSE4Mac im Batchfall 24 von 70 Bildern
// (das Kontingent wies den Rest ab) — und die Verteilung stellte seine 2,4 s
// neben die 3,0 s von MSE2 für 70.
test('ein Fall mit Fehler oder Befund fällt aus dem Vergleich der Zeiten heraus', () => {
  const z = zusammenfassen([
    { aufgabe: 'bilder', satz: 'Archenemy', programm: 'heute', wiederholung: 0, platz: 0, wandMs: 2400, fehler: '24 Bilder bei 70 Karten', befund: '24 Bilder bei 70 Karten' },
    { aufgabe: 'bilder', satz: 'Archenemy', programm: 'mse2', wiederholung: 0, platz: 1, wandMs: 3000, fehler: null },
  ])
  const heute = z['bilder · Archenemy'].heute
  assert.equal(heute.wand.n, 0)
  assert.equal(heute.faelle, 1)
  assert.equal(heute.fehler, 1)
  assert.equal(heute.ersteWandMs, null, 'auch die erste Zeit ist die eines unvollständigen Falls')
  assert.equal(z['bilder · Archenemy'].mse2.wand.median, 3000)
})

test('ein Fall allein mit Befund, aber ohne Fehlertext, gilt ebenso als unvollständig', () => {
  const z = zusammenfassen([{ ...faelle[0], befund: 'andere Größe', fehler: null }])
  assert.equal(z.start.mse2.wand.n, 0)
})

test('ein Satz gehört zum Schlüssel, damit klein und groß nicht zusammenfallen', () => {
  const z = zusammenfassen([
    { ...faelle[0], aufgabe: 'oeffnen', satz: 'Archenemy' },
    { ...faelle[0], aufgabe: 'oeffnen', satz: 'Black', wandMs: 9000 },
  ])
  assert.deepEqual(Object.keys(z), ['oeffnen · Archenemy', 'oeffnen · Black'])
})

test('die Schritte einer Sitzung zählen nach Zustand und Schicht', () => {
  const schritte = [
    { programm: 'mse2', zustand: 'kalt', schicht: 'Normal', dauerMs: 900, ok: true },
    { programm: 'mse2', zustand: 'warm', schicht: 'Normal', dauerMs: 40, ok: true },
    { programm: 'mse2', zustand: 'warm', schicht: 'Normal', dauerMs: 60, ok: false },
    { programm: 'mse2', zustand: 'warm', schicht: 'Problemfall', dauerMs: 80, ok: true },
  ]
  const z = schritteZusammenfassen(schritte)
  assert.equal(z['mse2 · kalt · Normal'].dauer.median, 900)
  assert.equal(z['mse2 · warm · Normal'].dauer.n, 2)
  assert.equal(z['mse2 · warm · Normal'].fehler, 1)
  assert.equal(z['mse2 · warm · Problemfall'].dauer.median, 80)
})

test('der Bericht nennt jede Zeile mit ihren Zahlen', () => {
  const md = berichtMarkdown({
    commit: 'abc1234', anfang: '2026-10-02T00:00:00Z',
    zusammenfassung: zusammenfassen(faelle),
    schritte: schritteZusammenfassen([{ programm: 'heute', zustand: 'warm', schicht: 'Normal', dauerMs: 12.5, ok: true }]),
  })
  assert.match(md, /abc1234/)
  assert.match(md, /\| start \| mse2 \| 2 \| 10 \| 30 \| 30 \| 30 \| 0,01 \| 10,0 \| 1 \| 0 \| 30 \/ 10 \|/)
  assert.match(md, /\| heute · warm · Normal \| 1 \| 12,5 \|/)
})

// **Der Kopf nennt die Maschine** — der Belegplan (B06) las den Bericht und
// fand keine: die Angaben standen nur im JSON daneben.
test('der Bericht nennt im Kopf die Maschine, auf der er lief', () => {
  const md = berichtMarkdown({
    commit: 'abc1234', anfang: '2026-10-02T00:00:00Z',
    maschine: { chip: 'Apple M5', kerne: 10, speicherByte: 34359738368, system: 'macOS 27.0.1 (26A434)', node: 'v24.15.0' },
    zusammenfassung: zusammenfassen(faelle), schritte: {},
  })
  assert.match(md.split('\n').slice(0, 6).join('\n'), /Apple M5, 10 Kerne, macOS 27\.0\.1 \(26A434\)/)
})

test('der Beleg steht eine Zeile je Fall — lesbar und ohne zehntausend Zeilen Einrückung', () => {
  const text = belegJson({
    commit: 'abc', faelle: [{ a: 1 }, { a: 2 }], zusammenfassung: { start: { mse2: { n: 1 } } }, leer: [],
  })
  assert.deepEqual(JSON.parse(text), {
    commit: 'abc', faelle: [{ a: 1 }, { a: 2 }], zusammenfassung: { start: { mse2: { n: 1 } } }, leer: [],
  })
  assert.equal(text, [
    '{',
    '  "commit": "abc",',
    '  "faelle": [',
    '    {"a":1},',
    '    {"a":2}',
    '  ],',
    '  "zusammenfassung": {',
    '    "start": {"mse2":{"n":1}}',
    '  },',
    '  "leer": []',
    '}',
    '',
  ].join('\n'))
})

test('die Bilanz zählt Befunde aus dem Lauf und aus den Fällen — und sagt nicht null, wenn einer da ist', () => {
  const b = bilanz({
    faelle: [
      { aufgabe: 'bilder', satz: 'Archenemy', programm: 'heute', wiederholung: 0, fehler: '24 Bilder bei 70 Karten', befund: '24 Bilder bei 70 Karten' },
      { aufgabe: 'start', satz: '', programm: 'mse2', wiederholung: 1, fehler: 'Rückgabe 1', befund: undefined },
      { aufgabe: 'start', satz: '', programm: 'mse2', wiederholung: 2, fehler: null },
    ],
    schritte: [{ ok: true }, { ok: false }, { ok: false }],
    befunde: ['Kreuzprobe Black: heute liest aus mse2 69 Karten'],
  })
  assert.equal(b.befunde, 2, 'ein Befund des Laufs und ein Fall, dessen Ergebnis nicht stimmt')
  assert.equal(b.faelleMitFehler, 1, 'ein Fall mit Fehler, aber ohne Befund, bleibt ein Fehler')
  assert.equal(b.schritteMitFehler, 2)
  assert.deepEqual([b.faelle, b.schritte], [3, 3])
  assert.equal(b.beispiele.length, 2)
  assert.match(b.beispiele[1], /bilder Archenemy heute #0: 24 Bilder bei 70 Karten/)
  const zeile = bilanzZeile(b)
  assert.match(zeile, /2 Befunde/)
  assert.doesNotMatch(zeile, /null Befunde|0 Befunde/)
})

test('ohne Befund steht 0 da, und die Fehler stehen daneben', () => {
  const b = bilanz({ faelle: [], schritte: [{ ok: false }], befunde: [] })
  assert.equal(b.befunde, 0)
  assert.match(bilanzZeile(b), /0 Befunde.*mit Fehler: 1 Schritt/)
})

test('der Bericht nennt den Dauerbetrieb mit Versuchen, Bildern und Ablehnungen getrennt', () => {
  const md = berichtMarkdown({
    commit: 'abc', anfang: 'x', zusammenfassung: {}, schritte: {},
    langlauf: {
      mse2: {
        sekunden: 120.1, versuche: 2537, gelungen: 2411, fehler: 126,
        dauer: { n: 2411, median: 44.6, p95: 91.4, max: 100 }, ablehnung: { n: 126, median: 3.2, p95: 4, max: 5 },
        anfang: { n: 482, median: 44.5 }, schluss: { n: 482, median: 45 },
        speicher: [{ byte: 269e6 }, { byte: 9350e6 }, { byte: 8630e6 }],
      },
    },
    bilanz: { befunde: 0, faelle: 1, schritte: 2, faelleMitFehler: 0, schritteMitFehler: 0, beispiele: [] },
  })
  assert.match(md, /## Dauerbetrieb/)
  assert.match(md, /\| mse2 \| 2\.537 \| 2\.411 \| 126 \| 44,6 \| 91,4 \| 3,2 \| 44,5 → 45 \| 269 → 9\.350 → 8\.630 \|/)
  assert.match(md, /Bilanz/)
})
