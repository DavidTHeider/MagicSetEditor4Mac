// Prüfungen des Berichts der Zufallsmutation (B05, #2224).
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { auswerten, bericht } from './zufallsmutationbericht.mjs'

const lauf = {
  commit: 'abc1234', samen: 7, je: 2, stellen: 100, ende: '2026-10-09T12:00:00Z', maschine: 'Mac17,3, macOS 27.0.1', befehl: 'node scripts/zufallsmutation.mjs --samen 7 --je 2',
  ergebnisse: [
    { nummer: 0, paket: 'internal/domain', datei: 'a.go', zeile: 1, art: 'vergleich', urtext: '==', ersatz: '!=', urteil: 'gefangen' },
    { nummer: 1, paket: 'internal/domain', datei: 'a.go', zeile: 2, art: 'bedingung', urtext: 'x', ersatz: '!(x)', urteil: 'überlebt' },
    { nummer: 2, paket: 'internal/service', datei: 'b.go', zeile: 3, art: 'rechnung', urtext: '+', ersatz: '-', urteil: 'nicht übersetzt' },
    { nummer: 3, paket: 'internal/service', datei: 'b.go', zeile: 4, art: 'wahrheitswert', urtext: 'true', ersatz: 'false', urteil: 'überlebt' },
  ],
}

test('äquivalente Überlebende zählen nicht, Lücken gegen die Quote', () => {
  const a = auswerten(lauf, { 1: { urteil: 'äquivalent', grund: 'nur Formatierung' } })
  assert.equal(a.gefangen, 1)
  assert.equal(a.luecken, 1)
  assert.equal(a.aequivalent, 1)
  assert.equal(a.nichtUebersetzt, 1)
  assert.equal(a.geeignet, 2)
  assert.equal(a.quote, 0.5)
})

test('ein Überlebender ohne Durchsicht ist eine Lücke und nicht äquivalent', () => {
  assert.equal(auswerten(lauf, {}).luecken, 2)
})

// **Ein Urteil gilt nur für die Stelle, an der es gefällt wurde** — die
// Durchsicht ist nach der Nummer des Mutanten verschlüsselt, und eine neue
// Ziehung kann unter derselben Nummer einen anderen ziehen.
test('eine Durchsicht für eine andere Stelle zählt nicht', () => {
  const passt = { 1: { urteil: 'äquivalent', grund: 'g', stelle: 'a.go:2' } }
  const passtNicht = { 1: { urteil: 'äquivalent', grund: 'g', stelle: 'a.go:9' } }
  assert.equal(auswerten(lauf, passt).aequivalent, 1)
  assert.equal(auswerten(lauf, passtNicht).aequivalent, 0)
  assert.match(bericht(lauf, passtNicht), /gilt einer anderen Stelle/)
})

test('der Bericht nennt Commit, Datum, Maschine und Befehl im Kopf — für den Belegplan', () => {
  const text = bericht(lauf, {})
  assert.match(text, /Commit `abc1234`/)
  assert.match(text, /2026-10-09 12:00 UTC/)
  assert.match(text, /Mac17,3/)
  assert.match(text, /`node scripts\/zufallsmutation\.mjs --samen 7 --je 2`/)
  assert.match(text, /gefangen ÷ geeignet/)
})
