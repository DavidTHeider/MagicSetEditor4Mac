// Prüfungen von scripts/vertragsargumente.mjs — #2528.
//
// **Der Vertragsrundgang ruft jede Bindung mit der Zahl Argumente, die sie
// hat.** Zwei Rufe folgten ihren neuen Argumenten nicht (`HandelSchritt` seit
// #2496, `PartieBeitreten` seit #2377): Wails wies sie am Tor ab, und der
// Rundgang zählte das als fachliche Absage — rot wurde er erst, als jemand
// die Absagen las (QA01, #2382).
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { resolve } from 'node:path'

import { abweichungen, stellenImQuelltext, stellvertreter } from './vertragsargumente.mjs'
import { gruppen, kohorten, schluss } from './huellenprobe/vertrag/gruppen.mjs'

const wurzel = resolve(import.meta.dirname, '..')

const signaturen = [
  { name: 'HandelSchritt', parameter: [{ name: 'kennung' }, { name: 'was' }, { name: 'wer' }, { name: 'gesehen' }] },
  { name: 'Sets', parameter: [] },
]

test('eine fehlende Stelle fällt auf, eine passende nicht', async () => {
  const rundgang = [{
    name: 'Handel',
    rufe: [
      { name: 'HandelSchritt', argumente: (k) => [k.handel.id, 'zustimmen', k.ich] },
      { name: 'Sets' },
    ],
  }]

  assert.deepEqual(await abweichungen(rundgang, signaturen), [
    'HandelSchritt: 3 Argumente, die Bindung hat 4',
  ])
})

test('ein Ruf auf eine Bindung, die es nicht gibt, fällt auf', async () => {
  const rundgang = [{ name: 'Alt', rufe: [{ name: 'GibtEsNicht', argumente: () => [] }] }]

  assert.deepEqual(await abweichungen(rundgang, signaturen), [
    'GibtEsNicht: keine Bindung dieses Namens',
  ])
})

// **Der Stellvertreter trägt jede Rechnung** — Eigenschaften, Aufrufe,
// Vergleiche und Text —, damit `argumente` ohne echte Anwendung läuft.
test('der Stellvertreter trägt Eigenschaften, Aufrufe und Text', async () => {
  const k = stellvertreter()

  assert.equal(typeof `${k.deck.id}`, 'string')
  assert.ok(k.liste.find((x) => x.id === k.set.id))
  assert.equal([k.a, k.b(1), 'x'].length, 3)
})

// **Was beim Auswerten wirklich rechnet**, zählt der Quelltext: ein Bild in
// den Bestandsordner, eine Unterschrift mit echtem Schlüssel — der
// Stellvertreter trägt beides nicht, und so schreibt die Prüfung nichts.
test('der Quelltext zählt die Stellen der Liste, auch mit Text, Liste und Objekt darin', () => {
  const geschachtelt = (k) => [k.f(k, 'x,y'), [1, 2], { a: 1, b: 2 }, `t${k.x}, u`, "a]b"]

  assert.equal(stellenImQuelltext(geschachtelt), 5)
  assert.equal(stellenImQuelltext(() => []), 0)
  assert.equal(stellenImQuelltext((k) => [k.a, ]), 1)
  assert.equal(stellenImQuelltext((k) => k.liste), undefined)
})

test('ein Ruf, der beim Auswerten rechnet, wird am Quelltext gezählt', async () => {
  const rundgang = [{
    name: 'Handel',
    rufe: [{ name: 'HandelSchritt', argumente: (k) => [k.handel.id, 'zustimmen', k.ich, JSON.parse(k.kaputt)] }],
  }]

  assert.deepEqual(await abweichungen(rundgang, signaturen), [])
})

test('jeder Ruf des Rundgangs hat so viele Argumente wie seine Bindung', async () => {
  const echte = JSON.parse(execFileSync('go', ['run', 'scripts/bindungssignaturen.go', 'internal/app'], { cwd: wurzel, encoding: 'utf8' }))

  assert.deepEqual(await abweichungen([...kohorten, ...gruppen, ...schluss], echte), [])
})
