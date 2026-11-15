// Prüfungen von scripts/lastlaufprofile.mjs — QA-LAST01
// (#2092).
//
// **Eine 400 ist keine Wiedererkennung.** Der Befund der Einfuhr nannte jeden
// Ruf, der kein Set anlegte, eine Erkennung der schon offenen Quelle — auch
// eine Abweisung und eine Drosselung.

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { einfuhrBefund } from './lastlaufprofile.mjs'

test('der Befund trennt angelegt, erkannt, abgewiesen und gedrosselt', () => {
  const proben = [
    { status: 201, text: JSON.stringify({ angelegt: [{ id: 7 }] }) },
    { status: 201, text: JSON.stringify({ angelegt: [] }) },
    { status: 200, text: JSON.stringify({ angelegt: [] }) },
    { status: 400, text: '{"fehler":"verborgen"}' },
    { status: 400, text: '{"fehler":"verborgen"}' },
    { status: 429, text: '' },
  ]
  assert.equal(einfuhrBefund(proben), '6 Rufe: 1 legte ein Set an, 2 erkannten die Quelle als schon offen, 2 abgewiesen (400 × 2), 1 gedrosselt')
})
