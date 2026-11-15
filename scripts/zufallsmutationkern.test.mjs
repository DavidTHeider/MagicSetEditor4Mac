// Prüfungen des Kerns der Zufallsmutation (B05, #2224).
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { chmodSync, existsSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { geschichtet, mitAktivierung, wegraeumen, wilson, zufall } from './zufallsmutationkern.mjs'

const stellen = ['a', 'b', 'c'].flatMap((paket) => Array.from({ length: 50 }, (_, i) => ({ paket, nummer: i })))

test('derselbe Samen zieht dieselben Mutanten — und ein anderer andere', () => {
  assert.deepEqual(geschichtet(stellen, 7, 10), geschichtet(stellen, 7, 10))
  assert.notDeepEqual(geschichtet(stellen, 7, 10), geschichtet(stellen, 8, 10))
})

test('je Paket gleich viele, ohne Zurücklegen', () => {
  const gezogen = geschichtet(stellen, 1, 10)
  for (const paket of ['a', 'b', 'c']) {
    const je = gezogen.filter((s) => s.paket === paket)
    assert.equal(je.length, 10)
    assert.equal(new Set(je.map((s) => s.nummer)).size, 10)
  }
})

test('ein Paket mit weniger Stellen gibt alle her', () => {
  assert.equal(geschichtet([{ paket: 'x', nummer: 1 }], 1, 10).length, 1)
})

test('der Zufall liegt in [0, 1)', () => {
  const naechste = zufall(42)
  for (let i = 0; i < 1000; i++) {
    const wert = naechste()
    assert.ok(wert >= 0 && wert < 1)
  }
})

test('das Wilson-Intervall umschließt die Quote und bleibt in [0, 1]', () => {
  const [unten, oben] = wilson(80, 100)
  assert.ok(unten < 0.8 && 0.8 < oben)
  assert.equal(Math.round(unten * 1000), 711)
  assert.deepEqual(wilson(0, 0), [0, 1])
  assert.ok(wilson(10, 10)[1] <= 1)
})

// **Die Aktivierung nimmt dem Lauf sein Zuhause nicht** — Zweitblick über
// Phase B, Befund 1: `activate.sh` lief mit der Umgebung des Rufenden, und
// ihre Ausgabe überschrieb das Wegwerf-`HOME` mit dem echten.
test('die Aktivierung behält das isolierte Zuhause und die Go-Schalter', () => {
  const isoliert = { HOME: '/tmp/heim', GOFLAGS: '-mod=readonly', PATH: '/usr/bin' }
  const aktiviert = 'HOME=/Users/echt\nGOFLAGS=\nGOPATH=/baum/.go\nPATH=/baum/.go/bin:/usr/bin\nkeine Zeile\n'

  const umgebung = mitAktivierung(isoliert, aktiviert)

  assert.equal(umgebung.HOME, '/tmp/heim')
  assert.equal(umgebung.GOFLAGS, '-mod=readonly')
  assert.equal(umgebung.GOPATH, '/baum/.go')
  assert.equal(umgebung.PATH, '/baum/.go/bin:/usr/bin')
  assert.equal(isoliert.GOPATH, undefined, 'die Vorlage bleibt unverändert')
})

// **Der Arbeitsbaum geht weg, auch mit seinem Modulcache** — Go legt ihn
// schreibgeschützt an, und `rmSync` scheiterte daran mit ENOTEMPTY; zwei
// Läufe ließen je 338 MB in `$TMPDIR` liegen.
test('wegraeumen entfernt auch schreibgeschützte Verzeichnisse', () => {
  const ort = mkdtempSync(join(tmpdir(), 'wegraeumen-'))
  const geschuetzt = join(ort, '.go', 'mod', 'paket@v1')
  mkdirSync(geschuetzt, { recursive: true })
  writeFileSync(join(geschuetzt, 'go.mod'), 'module paket\n')
  chmodSync(join(geschuetzt, 'go.mod'), 0o444)
  chmodSync(geschuetzt, 0o555)

  wegraeumen(ort)

  assert.equal(existsSync(ort), false)
})
