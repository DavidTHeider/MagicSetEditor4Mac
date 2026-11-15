// Prüfungen des Mutationsläufers — T05
// (#2108).
//
// **Rot ist nicht gefangen.** Bis zum 2026-10-06 zählte jeder Lauf mit
// Status ungleich null als gefangen — auch eine Frist, ein Signal, ein roter
// Grundlauf oder ein anderer Test, der zufällig fiel. Gefangen ist jetzt nur,
// was der **erwartete** Test nach **genau einer** Mutation fängt, bei grünem
// Grundlauf.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { mutiere, urteil, waehle } from './mutationskern.mjs'

const m = { kennung: 'M1', datei: 'a.go', alt: 'if wache {', neu: 'if false {', paket: './a', erwartet: 'TestDieWache' }
const gruen = { status: 0, ausgabe: 'ok  \ta\t0.1s\n' }
const rot = (name) => ({ status: 1, ausgabe: `--- FAIL: ${name} (0.00s)\nFAIL\ta\t0.1s\n` })

test('gefangen ist nur der erwartete Test bei grünem Grundlauf', () => {
  assert.deepEqual(urteil(m, gruen, rot('TestDieWache')), { art: 'gefangen', von: 'TestDieWache' })
})

test('ein anderer roter Test ist nicht gefangen', () => {
  assert.deepEqual(urteil(m, gruen, rot('TestEtwasAnderes')), { art: 'anders rot', von: 'TestEtwasAnderes' })
})

test('ein grüner Lauf nach der Mutation überlebt', () => {
  assert.deepEqual(urteil(m, gruen, gruen), { art: 'überlebt' })
})

test('ein roter Grundlauf entscheidet nichts', () => {
  assert.deepEqual(urteil(m, rot('TestDieWache'), rot('TestDieWache')), { art: 'unentschieden', grund: 'der Grundlauf ist rot: TestDieWache' })
})

test('eine Frist und ein Signal entscheiden nichts', () => {
  assert.deepEqual(urteil(m, gruen, { status: null, fehler: 'ETIMEDOUT', signal: 'SIGTERM', ausgabe: '' }), { art: 'unentschieden', grund: 'Frist' })
  assert.deepEqual(urteil(m, gruen, { status: null, signal: 'SIGKILL', ausgabe: '' }), { art: 'unentschieden', grund: 'Signal SIGKILL' })
})

test('ein Baufehler ist »nicht übersetzt«', () => {
  assert.deepEqual(urteil(m, gruen, { status: 1, ausgabe: 'a.go:3:2: undefined: wache\nFAIL\ta [build failed]\n' }), { art: 'nicht übersetzt' })
})

test('eine unbekannte Kennung bricht ab, statt nichts zu fahren', () => {
  assert.throws(() => waehle([m], ['M1', 'M99']), /M99/)
  assert.deepEqual(waehle([m], []), [m])
})

test('die Mutation wird zurückgesetzt, auch wenn der Lauf wirft', async () => {
  const wurzel = mkdtempSync(join(tmpdir(), 'mutation-'))
  writeFileSync(join(wurzel, 'a.go'), 'package a\n\nif wache {\n')
  await assert.rejects(mutiere(m, wurzel, () => {
    assert.match(readFileSync(join(wurzel, 'a.go'), 'utf8'), /if false \{/)
    throw new Error('abgebrochen')
  }), /abgebrochen/)
  assert.equal(readFileSync(join(wurzel, 'a.go'), 'utf8'), 'package a\n\nif wache {\n')
})

test('eine Fundstelle, die nicht genau einmal passt, wird nicht gefahren', async () => {
  const wurzel = mkdtempSync(join(tmpdir(), 'mutation-'))
  writeFileSync(join(wurzel, 'a.go'), 'if wache {\nif wache {\n')
  await assert.rejects(mutiere(m, wurzel, () => gruen), /2-mal/)
})
