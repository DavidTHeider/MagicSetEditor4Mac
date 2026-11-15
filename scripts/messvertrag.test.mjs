// Prüfungen des Messvertrags — PERF01
// (#2351).
//
// **Die Positivkontrollen stehen hier**: eine absichtlich überhöhte Probe —
// Zeit wie Speicher — muss rot werden, und eine Reihe ohne P95 auch. Sonst
// stünde eine Grenze im Vertrag, die niemand hält.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { pruefe } from './belegplan.mjs'
import { MESSVERTRAG, fussabdrucksumme, planeintraege, tafel, urteil } from './messvertrag.mjs'

const wurzel = resolve(import.meta.dirname, '..')
const plan = JSON.parse(readFileSync(join(wurzel, 'docs/belege/BELEGPLAN.json'), 'utf8'))
const heute = new Date('2026-10-15T12:00:00Z')

const kopf = (name) => `# Hüllenprobe: ${name}

**bestanden** — Commit \`a77c95866\`,
2026-10-15 07:00 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus \`scripts/bauen.sh abnahme\`, gestartet über LaunchServices.
`

// reihe gibt n Werte um `mitte`, damit Median und P95 dort liegen.
const reihe = (mitte, n = 20) => Array.from({ length: n }, () => mitte)

test('der Vertrag nennt jeden Kernweg aus PERF01 mit Grenze', () => {
  const reihen = MESSVERTRAG.map((e) => e.reihe)
  for (const weg of ['Start kalt', 'Start warm', 'Set öffnen, 3.788 Karten', 'Feld → Vorschaubild', 'Sammlung durchsuchen, 25.000 Zeilen',
    'Set ausgeben als Bilder', 'Druckbogen', 'Zurückspielen', 'Fußabdruck nach 120 Minuten', 'Dienst nach Rendern und Ruhe', 'Bündel, logisch',
    'Leerlauf sichtbar', 'Leerlauf ausgeblendet', 'Leerlauf minimiert']) {
    assert.ok(reihen.some((r) => r.startsWith(weg)), `»${weg}« fehlt im Vertrag`)
  }
  for (const e of MESSVERTRAG) assert.ok(e.grenze > 0 && e.beleg && ['P95', 'max'].includes(e.mass), `${e.reihe}: Grenze, Maß und Beleg`)
})

test('die vorläufigen Grenzen aus dem Issue stehen so im Vertrag', () => {
  const grenze = (anfang) => MESSVERTRAG.find((e) => e.reihe.startsWith(anfang))
  assert.deepEqual([grenze('Start kalt').grenze, grenze('Start kalt').mass], [2000, 'P95'])
  assert.deepEqual([grenze('Set öffnen').grenze, grenze('Set öffnen').mass], [1000, 'P95'])
  assert.deepEqual([grenze('Feld → Vorschaubild').grenze, grenze('Feld → Vorschaubild').mass], [300, 'P95'])
  assert.deepEqual([grenze('Sammlung durchsuchen').grenze, grenze('Sammlung durchsuchen').mass], [500, 'P95'])
  for (const weg of ['Set ausgeben als Bilder', 'Druckbogen', 'Zurückspielen']) assert.equal(grenze(weg).grenze, 60000)
  assert.equal(grenze('Fußabdruck nach 120 Minuten').grenze, 1536)
  assert.equal(grenze('Dienst nach Rendern und Ruhe').grenze, 512)
  assert.equal(grenze('Bündel, logisch').grenze, 500)
  assert.equal(grenze('Leerlauf sichtbar').grenze, 1)
})

// **Die Kernwege nach Phase 3 haben eine Reihe** (#2532, Gesamtreview über
// 0.9.0, Befund 6) — je Zeit und Fußabdruck, die Grenze vor der Messung.
test('Markttisch, Handelszug und Kennzahlleisten stehen mit Grenze im Vertrag', () => {
  const wege = MESSVERTRAG.filter((e) => e.beleg === 'huellenprobe/budgetswege09.md')
  const grenze = (anfang) => wege.find((e) => e.reihe.startsWith(anfang))
  assert.deepEqual([grenze('Markttisch aufschlagen, 25.000 eigene Exemplare').grenze, grenze('Markttisch aufschlagen').mass], [1000, 'P95'])
  assert.deepEqual([grenze('Anbieten bis zum neuen Stand').grenze, grenze('Anbieten bis zum neuen Stand').mass], [2000, 'P95'])
  for (const leiste of ['Set, 3.788 Karten', 'Sammlung, 25.000 Zeilen', 'Deck', 'Pakete']) {
    assert.deepEqual([grenze(`Kennzahlleiste ${leiste} `)?.grenze, grenze(`Kennzahlleiste ${leiste} `)?.mass], [500, 'P95'], leiste)
  }
  assert.deepEqual([grenze('Fußabdruck nach Markttisch, Handel und Leisten').grenze, grenze('Fußabdruck nach Markttisch').mass], [1536, 'max'])
  assert.equal(wege.length, 7)
  assert.ok(wege.every((e) => e.paket === '#2532'))
})

test('der Belegplan trägt jede Grenze des Vertrags', () => {
  // **Ein leerer Vertrag hielte jede Schleife hier** — gezählt wird zuerst.
  assert.equal(planeintraege().reduce((s, e) => s + e.werte.length, 0), MESSVERTRAG.length)
  assert.ok(MESSVERTRAG.length >= 14)
  for (const soll of planeintraege()) {
    const ist = plan.eintraege.find((e) => e.beleg === soll.beleg)
    assert.ok(ist, `kein Eintrag für ${soll.beleg}`)
    for (const wert of soll.werte) {
      assert.ok((ist.werte ?? []).some((w) => w.muster === wert.muster && w.hoechstens === wert.hoechstens), `${soll.beleg}: ${wert.name} fehlt oder hat eine andere Grenze`)
    }
  }
})

test('der Schritt im Plan nennt das Paket seiner Reihen', () => {
  const schritte = Object.fromEntries(planeintraege().map((e) => [e.beleg, e.schritt]))
  assert.equal(schritte['huellenprobe/budgets09.md'], 'PERF01: huellenprobe/budgets09.md')
  assert.equal(schritte['huellenprobe/kennzahlen09.md'], 'PERF02: huellenprobe/kennzahlen09.md')
})

test('eine Reihe im Rahmen hält, im Beleg wie im Plan', () => {
  const zeilen = MESSVERTRAG.filter((e) => e.beleg === 'huellenprobe/budgets09.md').map((e) => tafel(e, reihe(e.grenze / 2)))
  const roh = `${kopf('budgets09')}\n${zeilen.join('\n')}\n`
  const eintrag = planeintraege().find((e) => e.beleg === 'huellenprobe/budgets09.md')
  assert.deepEqual(pruefe(eintrag, roh, heute), [])
  for (const e of MESSVERTRAG) assert.equal(urteil(e, reihe(e.grenze / 2), { sekunden: 600 }), 'gehalten')
})

test('eine überhöhte Probe reißt — Zeit wie Speicher', () => {
  assert.ok(MESSVERTRAG.length >= 14)
  for (const e of MESSVERTRAG) {
    assert.equal(urteil(e, reihe(e.grenze * 2), { sekunden: 600 }), 'gerissen', e.reihe)
    const eintrag = planeintraege().find((p) => p.beleg === e.beleg)
    const roh = `${kopf('probe')}\n${tafel(e, reihe(e.grenze * 2), { sekunden: 600 })}\n`
    const befunde = pruefe({ ...eintrag, werte: eintrag.werte.filter((w) => w.name === e.reihe) }, roh, heute)
    assert.match(befunde.join(' '), /über der Schwelle/, e.reihe)
  }
})

test('ein einziger Ausreißer reißt das P95 nicht, aber das Höchstmaß', () => {
  const start = MESSVERTRAG.find((e) => e.reihe.startsWith('Start kalt'))
  const druck = MESSVERTRAG.find((e) => e.reihe.startsWith('Druckbogen'))
  assert.equal(urteil(start, [...reihe(500, 39), 5000]), 'gehalten')
  assert.equal(urteil(druck, [1000, 1000, 70000]), 'gerissen')
})

test('ohne zwanzig Werte gibt es kein P95 — und keinen Beleg', () => {
  const vorschau = MESSVERTRAG.find((e) => e.reihe.startsWith('Feld → Vorschaubild'))
  assert.equal(urteil(vorschau, reihe(10, 19)), 'zu wenige Werte')
  const eintrag = planeintraege().find((e) => e.beleg === vorschau.beleg)
  const roh = `${kopf('budgets09')}\n${tafel(vorschau, reihe(10, 19))}\n`
  assert.match(pruefe({ ...eintrag, werte: eintrag.werte.filter((w) => w.name === vorschau.reihe) }, roh, heute).join(' '), /Messwert nicht gefunden/)
})

// **Eine gescheiterte Messung ist kein Wert** — im ersten Lauf von budgets09
// standen 23 Mal −1 in der Reihe der Vorschau, und das P95 von −1 „hielt".
test('eine gescheiterte Messung lässt die Reihe nicht halten', () => {
  const vorschau = MESSVERTRAG.find((e) => e.reihe.startsWith('Feld → Vorschaubild'))
  const druck = MESSVERTRAG.find((e) => e.reihe.startsWith('Druckbogen'))
  assert.equal(urteil(vorschau, [...reihe(20, 23), -1]), 'gescheitert')
  assert.equal(urteil(druck, [1000, -1, 1000]), 'gescheitert')
  assert.match(tafel(vorschau, [...reihe(20, 23), -1]), /\| 24 \| 20 \| 20 \| 20 \| 20 \| P95 ≤ 300 \| gescheitert \|$/)
})

test('eine fehlende Messung steht als nicht gemessen da und hält nicht', () => {
  const dienst = MESSVERTRAG.find((e) => e.reihe.startsWith('Dienst nach Rendern und Ruhe'))
  assert.equal(urteil(dienst, []), 'nicht gemessen')
  assert.match(tafel(dienst, []), /nicht gemessen/)
})

// **Der Plan liest das Urteil mit, nicht nur die Zahl** — Befund 12 des
// Zweitblicks über Phase 3 von 0.9.0: eine Reihe mit einer gescheiterten
// Messung zeigte das Höchstmaß der übrigen, und der Plan las es als gehalten.
test('eine gescheiterte Reihe hält auch im Plan nicht', () => {
  const druck = MESSVERTRAG.find((e) => e.reihe.startsWith('Druckbogen'))
  const eintrag = planeintraege().find((e) => e.beleg === druck.beleg)
  const roh = `${kopf('budgets09')}\n${tafel(druck, [1000, -1, 1000])}\n`
  assert.notDeepEqual(pruefe({ ...eintrag, werte: eintrag.werte.filter((w) => w.name === druck.reihe) }, roh, heute), [])
})

// **Ein verkürzter Leerlauf ist kein Beleg** — Befund 13: `MSE_LEERLAUF_MIN`
// kürzte die Messung, und ihr Urteil blieb „gehalten".
test('ein verkürzter Leerlauf hält nicht, im Beleg wie im Plan', () => {
  const leerlauf = MESSVERTRAG.find((e) => e.reihe.startsWith('Leerlauf sichtbar, Anwendung'))
  assert.ok(leerlauf.mindestSekunden >= 600, 'der Vertrag nennt die Dauer')
  assert.equal(urteil(leerlauf, [0.1], { sekunden: 60 }), 'zu kurz')
  assert.equal(urteil(leerlauf, [0.1], { sekunden: 600 }), 'gehalten')
  const eintrag = planeintraege().find((e) => e.beleg === leerlauf.beleg)
  const roh = `${kopf('leerlauf09')}\n${tafel(leerlauf, [0.1], { sekunden: 60 })}\n`
  assert.notDeepEqual(pruefe({ ...eintrag, werte: eintrag.werte.filter((w) => w.name === leerlauf.reihe) }, roh, heute), [])
})

// **Ein fehlender Fußabdruck ist keine Null** — Befund 12: der Summierer
// ersetzte jeden fehlenden Prozess durch 0, und `budgets09` hielt die Grenze.
test('ein Fußabdruck mit fehlendem Prozess ist gescheitert', () => {
  assert.equal(fussabdrucksumme(new Map([['1', 300], ['2', 200]]), [1, 2]), 500)
  assert.equal(fussabdrucksumme(new Map([['1', 300]]), [1, 2]), -1)
  assert.equal(fussabdrucksumme(new Map(), [1]), -1)
})
