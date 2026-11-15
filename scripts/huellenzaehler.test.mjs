// Prüfungen von scripts/huellenzaehler.mjs — G01 (#2248).
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { sichtbarerKnopf, zaehler } from './huellenzaehler.mjs'

// probe schreibt mit, was der Zähler an der Probe ruft.
function probe() {
  const rufe = []
  const geprueft = []
  const merke = (name) => async (...werte) => { rufe.push([name, ...werte]) }

  return {
    rufe,
    geprueft,
    seite: merke('seite'),
    klickeAuf: merke('klickeAuf'),
    taste: merke('taste'),
    tippeInsFeld: merke('tippeInsFeld'),
    klick: merke('klick'),
    klickeText: merke('klickeText'),
    dialogantwort: merke('dialogantwort'),
    pruefe: (satz, gilt) => { geprueft.push([satz, gilt]) },
  }
}

test('jede Handlung zählt einmal, Rollen nicht', async () => {
  const p = probe()
  let jetzt = 1000
  const z = zaehler(p, () => jetzt)

  z.beginne()
  await z.klick('Karte anlegen')
  await z.klickAuf('document.body', 'den Körper')
  await z.taste('Tab')
  await z.tippe('Name')
  await z.menue('Datei', 'Sichern')
  await z.wahl('return true')
  await z.dialog('/tmp/x')
  await z.text('Diagnose')
  jetzt = 3250

  assert.deepEqual(z.ende('Weg'), { handlungen: 8, dauer: 2250 })
  assert.deepEqual(p.geprueft, [['Weg: 8 Handlungen, 0 Fehlversuche, 2,3 s Maschinenzeit', true]])
  assert.deepEqual(p.rufe.at(-1), ['klickeText', 'Diagnose'])
  // **Zwei Rufe an der Seite sind Rollen** und keine Handlung.
  assert.equal(p.rufe.filter(([name]) => name === 'seite').length, 3)
})

test('ein neuer Anfang zählt von vorn', async () => {
  const p = probe()
  const z = zaehler(p, () => 0)

  z.beginne()
  await z.taste('a')
  z.beginne()

  assert.equal(z.ende('leer').handlungen, 0)
})

test('der Knopf wird am Wort gefunden, auch mit Zeichen davor', () => {
  const ausdruck = sichtbarerKnopf('Karte anlegen')

  assert.match(ausdruck, /"Karte anlegen"/)
  assert.match(ausdruck, /\\p\{L\}/)
})
