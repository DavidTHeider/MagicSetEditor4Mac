// Prüfungen des Laufs über die Drehbücher — PERF01
// (#2351).
//
// **Ein Modul ohne Drehbuch ist kein roter Lauf**: Teile, die mehrere
// Drehbücher teilen, liegen im selben Verzeichnis (`gastlauf`,
// `aufstiegsteile`, `budgetsorte09`), und ein voller Lauf rief sie bis
// dahin als Drehbuch auf — mit einem `TypeError` als Befund.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { join } from 'node:path'

import { drehbuchAus } from './huellenlauf.mjs'

test('ein Modul mit Vorgabefunktion ist ein Drehbuch', () => {
  const drehbuch = async () => {}
  assert.equal(drehbuchAus({ default: drehbuch }), drehbuch)
})

test('ein Modul ohne Vorgabefunktion ist keines', () => {
  assert.equal(drehbuchAus({ fahre() {} }), undefined)
  assert.equal(drehbuchAus({ default: 'kein Ruf' }), undefined)
})

// **Nur genannte Module werden geladen** — ein Import führt den Kopf eines
// Moduls aus, und nicht jedes Skript dieses Projekts ist dafür gebaut.
test('die Teile im Verzeichnis werden als Teile erkannt, ein Drehbuch als Drehbuch', async () => {
  const ordner = join(import.meta.dirname, 'huellenprobe')
  for (const teil of ['gastlauf.mjs', 'aufstiegsteile.mjs', 'budgetsorte09.mjs']) {
    assert.equal(drehbuchAus(await import(join(ordner, teil))), undefined, `${teil} galt als Drehbuch`)
  }
  assert.equal(typeof drehbuchAus(await import(join(ordner, 'budgets09.mjs'))), 'function')
})
