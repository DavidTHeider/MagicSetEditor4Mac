// Prüfungen der Leerlaufrechnung — PERF01
// (#2351),
// Befund A-26: CPU im Mittel je Prozess aus zwei Abfragen von `ps`.
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { anteile, cpuSekunden, zeitenAus } from './leerlauf.mjs'

test('ps nennt die CPU-Zeit in Minuten, Stunden und Tagen', () => {
  assert.equal(cpuSekunden('0:00.25'), 0.25)
  assert.equal(cpuSekunden('2:03.50'), 123.5)
  assert.equal(cpuSekunden('1:02:03.00'), 3723)
  assert.equal(cpuSekunden('2-01:00:00'), 176400)
})

test('die Ausgabe von ps wird je Kennung gelesen', () => {
  const zeiten = zeitenAus('  101   0:01.50\n 202   1:00.00\n\n')
  assert.deepEqual([...zeiten.entries()], [['101', 1.5], ['202', 60]])
})

test('der Anteil ist die CPU-Zeit je Wanduhrzeit, je Art summiert', () => {
  const arten = new Map([['101', 'Anwendung'], ['202', 'Webinhalt'], ['203', 'Webinhalt']])
  const vorher = new Map([['101', 10], ['202', 5], ['203', 1]])
  const nachher = new Map([['101', 16], ['202', 5.6], ['203', 1.6]])
  assert.deepEqual(anteile(arten, vorher, nachher, 600), { Anwendung: 1, Webinhalt: 0.2 })
})

test('ein Prozess, der verschwand, fehlt und wird nicht als null gezählt', () => {
  const arten = new Map([['101', 'Anwendung'], ['202', 'GPU']])
  const anteil = anteile(arten, new Map([['101', 1], ['202', 1]]), new Map([['101', 1.6]]), 60)
  assert.deepEqual(anteil, { Anwendung: 1 })
})
