// Prüfungen von scripts/leistungsauswahl.mjs — QA-04.AK-04
// (#1735).

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { gleichmaessig, kartenvorlagen, waehleKarten } from './leistungsauswahl.mjs'

const settext = [
  '\ufeffmse version: 2.0.2',
  'game: magic',
  'stylesheet: m15-altered',
  'set info:',
  '\ttitle: Probe',
  'card:',
  '\tname: Erste',
  '\trule text:',
  '\t\tname: keine Karte',
  'card:',
  '\tstylesheet: m15-aftermath',
  '\tname: Zweite',
  'card:',
  '\tname: "Dritte, mit Komma"',
  'keyword:',
  '\tname: kein Kartenname',
  '',
].join('\n')

test('je Karte Name und Vorlage, und ohne eigene gilt die des Sets', () => {
  assert.deepEqual(kartenvorlagen(settext), [
    { name: 'Erste', vorlage: 'm15-altered' },
    { name: 'Zweite', vorlage: 'm15-aftermath' },
    { name: '"Dritte, mit Komma"', vorlage: 'm15-altered' },
  ])
})

test('gleichmäßig trifft die erste und die letzte Stelle und verteilt den Rest', () => {
  assert.deepEqual(gleichmaessig([10, 11, 12, 13, 14, 15, 16, 17, 18, 19], 4), [10, 13, 16, 19])
  assert.deepEqual(gleichmaessig([5, 6], 4), [5, 6], 'weniger da als verlangt: alles')
  assert.deepEqual(gleichmaessig([7, 8, 9], 1), [7])
})

test('die Auswahl folgt den Regeln, in der Reihenfolge des Plans, mit Schicht', () => {
  const plan = {
    schichten: [
      { satz: 'Klein', stichprobe: [0, 1, 2, 3, 4] },
      { satz: 'Gross', stichprobe: [0, 2, 3, 5, 6, 8] },
    ],
  }
  const vorlagen = {
    Klein: ['a', 'a', 'a', 'a', 'a'].map((vorlage, i) => ({ name: `k${i}`, vorlage })),
    Gross: ['n', 'x', 'p', 'n', 'x', 'p', 'n', 'x', 'p'].map((vorlage, i) => ({ name: `g${i}`, vorlage })),
  }
  const regeln = [
    { schicht: 'Sonderformat', satz: 'Klein', anzahl: 2 },
    { schicht: 'Normal', satz: 'Gross', vorlage: 'n', anzahl: 5 },
    { schicht: 'Problemfall', satz: 'Gross', vorlage: 'p', anzahl: 1 },
  ]
  assert.deepEqual(waehleKarten(plan, vorlagen, regeln), [
    { schicht: 'Sonderformat', satz: 'Klein', stelle: 0, name: 'k0', vorlage: 'a' },
    { schicht: 'Sonderformat', satz: 'Klein', stelle: 4, name: 'k4', vorlage: 'a' },
    { schicht: 'Normal', satz: 'Gross', stelle: 0, name: 'g0', vorlage: 'n' },
    { schicht: 'Normal', satz: 'Gross', stelle: 3, name: 'g3', vorlage: 'n' },
    { schicht: 'Normal', satz: 'Gross', stelle: 6, name: 'g6', vorlage: 'n' },
    { schicht: 'Problemfall', satz: 'Gross', stelle: 2, name: 'g2', vorlage: 'p' },
  ])
})

test('ein Satz, den der Plan nicht kennt, ist ein Fehler und keine leere Schicht', () => {
  assert.throws(() => waehleKarten({ schichten: [] }, {}, [{ schicht: 'x', satz: 'Fehlt', anzahl: 1 }]), /Fehlt/)
})

test('passt die Kartenliste nicht zum Plan, bricht die Auswahl ab', () => {
  const plan = { schichten: [{ satz: 'Klein', stichprobe: [0, 7] }] }
  const vorlagen = { Klein: [{ name: 'k0', vorlage: 'a' }] }
  assert.throws(() => waehleKarten(plan, vorlagen, [{ schicht: 'x', satz: 'Klein', anzahl: 2 }]), /Stelle 7/)
})
