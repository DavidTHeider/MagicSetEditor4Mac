// Prüfungen von scripts/ausgangswege.mjs — P0
// (#2336).
//
// **Ein Messwächter, der einen fehlenden Weg durchlässt, misst nichts** —
// geprüft wird an einem vollständigen Datensatz, an einem mit fehlendem
// Pflichtweg und an einem, dessen Weg keine Wirkung hatte.

import { readFileSync } from 'node:fs'
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { pflichtwege, pruefeDatensatz, vereinige } from './ausgangswege.mjs'

// vollstaendig baut einen Datensatz, in dem jeder Pflichtweg gemessen ist.
function vollstaendig() {
  return {
    bauabdruck: 'abcdef012345',
    wege: pflichtwege.flatMap((p) => p.arten.map((art) => ({
      weg: p.weg, eingang: p.eingang, art,
      handlungen: 3, sichtwechsel: 1, tabstellen: 0, fehlversuche: 0,
      maschinenzeitMs: 800, wirkung: true, bedingt: [],
    }))),
  }
}

test('ein vollständiger Datensatz ist ohne Befund', () => {
  assert.deepEqual(pruefeDatensatz(vollstaendig()), [])
})

test('ein fehlender Pflichtweg wird genannt', () => {
  const d = vollstaendig()
  d.wege = d.wege.filter((z) => !(z.weg === 'S1' && z.eingang === 'deck'))

  assert.deepEqual(pruefeDatensatz(d), ['S1/deck/maus: fehlt im Datensatz'])
})

test('ein Weg ohne Wirkung ist nicht gemessen', () => {
  const d = vollstaendig()
  d.wege.find((z) => z.weg === 'H1' && z.art === 'tastatur').wirkung = false

  assert.deepEqual(pruefeDatensatz(d), ['H1/bereich/tastatur: gegangen, aber keine Wirkung beobachtet'])
})

// **Ein Griff, der vom Stand abhängt, steht in der Zeile** (Befund 3 des
// Nachlaufs über Phase 0): ohne die Spalte sah niemand, ob eine Messung
// einen Zusatzgriff enthielt oder einen ausließ.
test('eine Zeile ohne Spalte für bedingte Griffe ist nicht gemessen', () => {
  const d = vollstaendig()
  delete d.wege.find((z) => z.weg === 'H5').bedingt

  assert.deepEqual(pruefeDatensatz(d), ['H5/handel/maus: nennt keine bedingten Griffe — der Zähler ist älter als der Vertrag'])
})

test('NOT_RUN und FEHLT verlangen einen Grund', () => {
  const d = vollstaendig()
  const s11 = d.wege.find((z) => z.weg === 'S11')
  Object.assign(s11, { status: 'FEHLT', grund: '' })

  assert.deepEqual(pruefeDatensatz(d), ['S11/tisch/maus: FEHLT ohne Grund'])

  s11.grund = 'kein Befehl, der einen Zug zurücknimmt'
  assert.deepEqual(pruefeDatensatz(d), [])
})

test('ohne Bauabdruck ist der Datensatz keinem Bündel zuzuordnen', () => {
  const d = vollstaendig()
  delete d.bauabdruck

  assert.equal(pruefeDatensatz(d).length, 1)
})

test('drei Teildatensätze aus einem Bündel werden einer', () => {
  const ganz = vollstaendig()
  const teile = [['S1', 'S2', 'S3', 'S4', 'S5', 'S6', 'S7', 'S8', 'S9', 'S10', 'S11'], ['S12', 'S13'], ['H1', 'H2', 'H3', 'H4', 'H5', 'H6']]
    .map((wege) => ({ bauabdruck: ganz.bauabdruck, wege: ganz.wege.filter((z) => wege.includes(z.weg)) }))

  const { datensatz, fehler } = vereinige(teile)

  assert.deepEqual(fehler, [])
  assert.deepEqual(pruefeDatensatz(datensatz), [])
})

test('Teildatensätze aus verschiedenen Bündeln sind kein Vergleich', () => {
  const { fehler } = vereinige([{ bauabdruck: 'abcdef012345', wege: [] }, { bauabdruck: '0123456789ab', wege: [] }])

  assert.deepEqual(fehler, ['die Teildatensätze stammen aus verschiedenen Bündeln: abcdef012345, 0123456789ab'])
})

// **Der Wächter über die abgelegten Ausgangswerte** — ohne ihn stünde der
// Vertrag neben Dateien, die er nie gesehen hat. Was er nicht hält: ob die
// Zahlen stimmen; das belegt der Lauf mit seinen Bildern.
test('die abgelegten Ausgangswerte 0.9.0 sind vollständig', () => {
  const ort = new URL('../docs/belege/', import.meta.url)
  const teile = ['spielen', 'netz', 'handeln'].map((teil) => JSON.parse(readFileSync(new URL(`ausgangswerte09-${teil}.json`, ort), 'utf8')))
  const { datensatz, fehler } = vereinige(teile)

  assert.deepEqual([...fehler, ...pruefeDatensatz(datensatz)], [])
})
