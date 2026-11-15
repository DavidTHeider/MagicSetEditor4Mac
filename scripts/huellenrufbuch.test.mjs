// Prüfungen von scripts/huellenrufbuch.mjs — B01 (#2220).
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { mischen, rufbuchHolen } from './huellenrufbuch.mjs'

test('zwei Abzüge werden je Bindung zusammengezählt', () => {
  const buch = {}
  mischen(buch, { Sets: { anzahl: 2, ausgaenge: { gelungen: 2 }, dauerMs: 3 } })
  mischen(buch, { Sets: { anzahl: 1, ausgaenge: { fachfehler: 1 }, dauerMs: 1.5 }, Datenstand: { anzahl: 1, ausgaenge: { gelungen: 1 }, dauerMs: 1 } })
  assert.deepEqual(buch.Sets, { anzahl: 3, ausgaenge: { gelungen: 2, fachfehler: 1 }, dauerMs: 4.5 })
  assert.equal(buch.Datenstand.anzahl, 1)
})

test('ein leerer Abzug ändert nichts', () => {
  assert.deepEqual(mischen({}, undefined), {})
})

test('eine geschlossene Leitung kostet nur die Mitschrift', async () => {
  const probe = { leitung: { frage: () => Promise.reject(new Error('zu')) } }
  assert.deepEqual(await rufbuchHolen(probe), {})
})

test('der Abzug wird geholt, geleert und gemischt', async () => {
  const gefragt = []
  const probe = { leitung: { frage: (befehl) => { gefragt.push(befehl); return Promise.resolve({ ok: true, daten: { Sets: { anzahl: 1, ausgaenge: { gelungen: 1 }, dauerMs: 2 } } }) } } }
  await rufbuchHolen(probe)
  await rufbuchHolen(probe)
  assert.deepEqual(gefragt[0], { was: 'rufbuch', leeren: true })
  assert.equal(probe.rufbuch.Sets.anzahl, 2)
})
