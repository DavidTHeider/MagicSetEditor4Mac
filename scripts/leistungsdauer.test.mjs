// Prüfungen von scripts/leistungsdauer.mjs — der Langlauf aus QA-04.AK-03
// (#1735,
// #2005).
//
// **Eine Abweisung ist kein Schritt.** Am 2026-10-03 griff im Langlauf das
// Kontingent von `mse-mcp`, und die Schleife fragte ohne Pause weiter:
// 931.119 Abweisungen in 120 s, 71 gezeichnete Karten. Wer „in 1 s wieder"
// liest, wartet eine Sekunde.

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { wartezeitAus, zaehleVersuche, zeichneReihum } from './leistungsdauer.mjs'

test('liest die Wartezeit aus der Abweisung des Werkzeugservers', () => {
  assert.equal(wartezeitAus('schreiben.karte.zeichnen: zu viele Rufe — in 1 s wieder (Vorgang: f70b1637)'), 1)
  assert.equal(wartezeitAus('zu viele Rufe — in 12 s wieder'), 12)
})

test('wartet nach einem anderen Fehler nicht', () => {
  assert.equal(wartezeitAus('ERROR: Variable not set: exists_in_package'), 0)
  assert.equal(wartezeitAus(''), 0)
})

// **Ein Versuch wird einmal gezählt** — Z6 des Zweitblicks über Phase 6.11.
//
// Im Beleg vom 2026-10-03 standen bei MSE2 `dauer.n = 2537`, `fehler = 126` und
// `schritte = 2663`: die Schleife hatte 2537 Karten angefragt, und die 126
// gescheiterten trugen ihre Dauer **und** zählten als Fehler — `schritte` war
// ihre Summe. Die Zeitverteilung mischte dazu Bilder und Ablehnungen.

test('ein gescheiterter Versuch zählt als Versuch und als Fehler — und nicht doppelt', () => {
  const z = zaehleVersuche([
    { ok: true, dauerMs: 40 }, { ok: true, dauerMs: 50 }, { ok: false, dauerMs: 2 }, { ok: true, dauerMs: 60 },
  ])
  assert.equal(z.versuche, 4)
  assert.equal(z.gelungen, 3)
  assert.equal(z.fehler, 1)
  assert.equal(z.gelungen + z.fehler, z.versuche)
})

test('die Dauer der gelungenen Bilder und die der Ablehnungen stehen getrennt', () => {
  const z = zaehleVersuche([
    { ok: true, dauerMs: 40 }, { ok: true, dauerMs: 60 }, { ok: false, dauerMs: 2 }, { ok: false, dauerMs: 4 },
  ])
  assert.equal(z.dauer.n, 2)
  assert.equal(z.dauer.median, 40)
  assert.equal(z.dauer.max, 60, 'eine Ablehnung gehört nicht in die Verteilung der Bilder')
  assert.equal(z.ablehnung.n, 2)
  assert.equal(z.ablehnung.max, 4)
})

test('ein Versuch ohne Dauer zählt, geht aber in keine Verteilung', () => {
  const z = zaehleVersuche([{ ok: true, dauerMs: 40 }, { ok: false, dauerMs: null }])
  assert.equal(z.versuche, 2)
  assert.equal(z.fehler, 1)
  assert.equal(z.ablehnung.n, 0)
})

test('ohne Versuche steht überall null', () => {
  const z = zaehleVersuche([])
  assert.deepEqual([z.versuche, z.gelungen, z.fehler, z.dauer.n, z.ablehnung.n], [0, 0, 0, 0, 0])
})

/** sitzung gibt eine Sitzung, die der Reihe nach antwortet und ihre Ziele festhält. */
function sitzung(antworten) {
  const ziele = []
  return {
    ziele,
    zeichne: async (stelle, ziel) => {
      ziele.push([stelle, ziel])
      return antworten[(ziele.length - 1) % antworten.length]
    },
  }
}

test('die Schleife zeichnet reihum, zählt jeden Versuch einmal und wartet nach einer Abweisung', async () => {
  const s = sitzung([
    { ok: true, dauerMs: 40, text: '' },
    { ok: false, dauerMs: null, text: 'schreiben.karte.zeichnen: zu viele Rufe — in 3 s wieder (Vorgang: x)' },
  ])
  let uhr = 0
  const gewartet = []
  const lauf = await zeichneReihum(s, [{ stelle: 7 }, { stelle: 9 }], '/aus', 'heute', {
    jetzt: () => uhr++,
    genug: () => uhr >= 5,
    schlafe: async (ms) => { gewartet.push(ms) },
    loesche: () => {},
  })
  assert.deepEqual(s.ziele.map(([stelle]) => stelle), [7, 9, 7, 9, 7].slice(0, s.ziele.length), 'reihum')
  assert.equal(new Set(s.ziele.map(([, ziel]) => ziel)).size, s.ziele.length, 'jeder Versuch hat seinen eigenen Dateinamen')
  assert.ok(gewartet.length >= 1)
  assert.ok(gewartet.every((ms) => ms === 3000), 'wer „in 3 s wieder“ liest, wartet drei Sekunden')
  assert.equal(lauf.versuche.length, s.ziele.length)
  assert.equal(lauf.fehlertexte.length, 1, 'derselbe Fehler steht einmal da')
})

test('ein Fehler ohne eigene Dauer bekommt die gemessene', async () => {
  const s = sitzung([{ ok: false, dauerMs: null, text: 'ERROR: kaputt' }])
  let uhr = 0
  const lauf = await zeichneReihum(s, [{ stelle: 1 }], '/aus', 'heute', {
    jetzt: () => (uhr += 12), genug: () => uhr >= 24, schlafe: async () => {}, loesche: () => {},
  })
  assert.equal(lauf.versuche[0].dauerMs, 12)
  assert.equal(lauf.versuche.length, 1)
})
