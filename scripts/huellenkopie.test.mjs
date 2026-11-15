// **Ein Inhalt wird kopiert, nicht sein Ordner** — PERF01
// (#2351).
//
// `join(pack, 'Magic - Fonts/.')` gibt `…/Magic - Fonts` zurück: Node streicht
// das `/.`. `cp -cR` kopierte damit den **Ordner** nach `~/Library/Fonts`, und
// die Schriftsuche, die dort nicht in Unterordner sieht, fand keine einzige
// Schrift des Packs — die Vorschau trug in `budgets09` und in der Dichtemessung
// AP-13 (`uirevue09zustaende`) keinen Text. Die Pakete fielen nicht auf, weil
// ihre Suche in die Tiefe geht.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const muster = /join\([^)]*\/\.['"`]\)/

test('kein Drehbuch und kein Skript kopiert einen Inhalt über join(…, "x/.")', () => {
  const funde = []
  for (const ordner of ['', 'huellenprobe']) {
    const ort = join(import.meta.dirname, ordner)
    for (const name of readdirSync(ort).filter((n) => n.endsWith('.mjs') && n !== 'huellenkopie.test.mjs')) {
      readFileSync(join(ort, name), 'utf8').split('\n').forEach((zeile, i) => {
        if (muster.test(zeile)) funde.push(`${join(ordner, name)}:${String(i + 1)}`)
      })
    }
  }
  assert.deepEqual(funde, [])
})

test('Node streicht das /. — der Grund für den Wächter', () => {
  assert.equal(join('/pack', 'Magic - Fonts/.'), '/pack/Magic - Fonts')
})
