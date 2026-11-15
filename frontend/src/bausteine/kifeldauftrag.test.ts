// **Der Auftrag eines Feldvorschlags** — AI-05.AK-05 und AK-06
// (#2042).
import { expect, it } from 'vitest'

import { auftragsgrenze, istKleinesFeld, kiFeldauftrag } from './kifeldauftrag'

const formen = new Map([
  ['name', { art: 'text' }],
  ['rule text', { art: 'text' }],
  ['image', { art: 'image' }],
])

// **Nur Name und Flavortext** — dieselbe Liste wie `domain.IstKleineTextaufgabe`.
it('kennt die kleinen Felder', () => {
  expect(istKleinesFeld('name')).toBe(true)
  expect(istKleinesFeld('flavor text')).toBe(true)
  expect(istKleinesFeld('rule text')).toBe(false)
})

// **Das Modell sieht die Karte** — sonst schlüge es einen Namen für nichts
// vor. Bilder und leere Felder gehen nicht mit.
it('hängt die Textfelder der Karte an', () => {
  const auftrag = kiFeldauftrag(
    'Schlage einen Namen vor.',
    'Die Karte bisher:',
    [
      { feld: 'name', wert: 'Alter Hirsch' },
      { feld: 'rule text', wert: 'Fliegend' },
      { feld: 'image', wert: 'bild.png' },
      { feld: 'flavor text', wert: '' },
    ],
    formen,
  )

  expect(auftrag).toBe(
    'Schlage einen Namen vor.\n\nDie Karte bisher:\nname: Alter Hirsch\nrule text: Fliegend',
  )
})

// **Der Auftrag bleibt unter der Grenze** — darüber wäre es keine kleine
// Aufgabe mehr, und das Gerätemodell käme nicht dran.
it('kürzt lange Werte und bleibt unter der Grenze', () => {
  const lang = 'x'.repeat(5000)
  const auftrag = kiFeldauftrag(
    'Basis.',
    'Karte:',
    [
      { feld: 'rule text', wert: lang },
      { feld: 'name', wert: lang },
    ],
    formen,
  )

  expect(Array.from(auftrag).length).toBeLessThanOrEqual(auftragsgrenze)
  expect(auftrag.startsWith('Basis.')).toBe(true)
})

// **Ohne Kartenwerte steht nur der Satz da.**
it('lässt den Kartenteil ohne Werte weg', () => {
  expect(kiFeldauftrag('Basis.', 'Karte:', [], formen)).toBe('Basis.')
})
