// **Die Ausgleichsgerade von `dauerteile`** — B04 (#2223).
import { strict as assert } from 'node:assert'
import { test } from 'node:test'

import { EINSCHWINGEN_MIN, phasenAus, phasensteigung, steigung } from './huellenprobe/dauerteile.mjs'

test('eine Gerade gibt ihre Steigung', () => {
  assert.equal(steigung([[0, 100], [1, 132], [2, 164], [3, 196]]), 32)
})

test('Rauschen um eine flache Linie gibt fast null', () => {
  const s = steigung([[0, 100], [1, 103], [2, 98], [3, 101], [4, 99]])
  assert.ok(Math.abs(s) < 1, `Steigung ${String(s)}`)
})

test('zu wenige Punkte oder nur ein Zeitpunkt geben keine Steigung', () => {
  assert.equal(steigung([[0, 1], [1, 2]]), null)
  assert.equal(steigung([[1, 1], [1, 2], [1, 3]]), null)
})

// **Die Phasen und die Heap-Probe kommen aus der Umgebung** — E06 (#2278).
// Die Heap-Probe vergrößert den Puffer der Ressourcenzeiten auf eine
// Million Einträge, und jeder IPC-Ruf der Hülle ist ein `fetch`: ein Lauf
// mit ihr misst sich mit. Ohne sie bleibt der Leerlauf, wie ihn die
// Anwendung hat.
test('ohne Angabe: Leerlauf, fünf Teile, Heap-Probe an', () => {
  const { phasen, heap } = phasenAus({})
  assert.deepEqual(phasen.map((x) => [x.name, x.minuten]), [
    ['Leerlauf', 6], ['bearbeiten', 6], ['navigation', 6], ['sammlung', 6], ['tisch', 6], ['ausgabe', 6]])
  assert.equal(heap, true)
})

test('eine leere Teileliste heißt: nur Leerlauf', () => {
  const { phasen } = phasenAus({ MSE_DAUER_TEILE: '', MSE_DAUER_LEER_MIN: '20' })
  assert.deepEqual(phasen.map((x) => [x.name, x.minuten, x.teile]), [['Leerlauf', 20, []]])
})

test('Längen und Teile wie genannt, Teile mit + als eine Phase', () => {
  const { phasen } = phasenAus({ MSE_DAUER_PHASE_MIN: '15', MSE_DAUER_LEER_MIN: '30', MSE_DAUER_TEILE: 'sammlung,tisch+ausgabe' })
  assert.deepEqual(phasen.map((x) => [x.name, x.minuten, x.teile]), [
    ['Leerlauf', 30, []], ['sammlung', 15, ['sammlung']], ['tisch+ausgabe', 15, ['tisch', 'ausgabe']]])
})

test('MSE_DAUER_HEAP=aus schaltet die Heap-Probe ab', () => {
  assert.equal(phasenAus({ MSE_DAUER_HEAP: 'aus' }).heap, false)
  assert.equal(phasenAus({ MSE_DAUER_HEAP: 'an' }).heap, true)
})

// **Die Steigung nach dem Einschwingen** — Befund 9 des Zweitblicks über E/F.
// Der Leerlauf beginnt mit dem Start der Anwendung, und der Webprozess gibt
// nach knapp zwei Minuten über 100 MB zurück. Eine Gerade über alles zieht
// diesen Einbruch in die Steigung und unterschätzt das Wachstum danach.
test('ein Einbruch am Anfang zählt nach dem Einschwingen nicht mit', () => {
  const proben = []
  for (let minute = 0; minute <= 30; minute += 0.25) {
    const web = minute < 1.75 ? 229 : 109 + 2 * (minute - 1.75)
    proben.push({ minute, web, app: 140 })
  }

  const { eingeschwungen, gesamt } = phasensteigung(proben, [], 0)
  assert.equal(EINSCHWINGEN_MIN, 3)
  assert.ok(Math.abs(eingeschwungen.web - 2) < 0.01, `nach dem Einschwingen ${String(eingeschwungen.web)}`)
  assert.ok(gesamt.web < 1.5, `gesamt ${String(gesamt.web)}`)
  assert.equal(eingeschwungen.app, 0)
})

test('der Go-Heap zählt nach demselben Anfang', () => {
  const haufen = [{ minute: 1, heap: 500 }, { minute: 4, heap: 10 }, { minute: 5, heap: 11 }, { minute: 6, heap: 12 }]
  const { eingeschwungen } = phasensteigung([], haufen, 0)
  assert.equal(eingeschwungen.goHeap, 1)
})
