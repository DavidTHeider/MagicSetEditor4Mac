// **Welches Bedienelement eine Einstellung bekommt** — UXR-F2
// (#2008),
// Regel R7 aus `docs/UIREVIEW.md`.
//
// Hier steht die Regel an Beispielen; **dass sie für jede Einstellung des
// Katalogs gilt**, hält `wahlkatalog.test.tsx`.
import { describe, expect, it } from 'vitest'

import { bedienung } from './wahlart'

describe('das Bedienelement nach der Zahl der Werte', () => {
  it.each([
    ['ohne Werteliste ein Textfeld', [], 'text'],
    ['ein Wert eine Auswahl', ['scryfall'], 'auswahl'],
    ['an und aus ein Schalter', ['an', 'aus'], 'schalter'],
    ['aus und an ein Schalter', ['aus', 'an'], 'schalter'],
    [
      'sichtbar und verborgen ein Schalter',
      ['sichtbar', 'verborgen'],
      'schalter',
    ],
    [
      'verborgen und sichtbar ein Schalter',
      ['verborgen', 'sichtbar'],
      'schalter',
    ],
    [
      'zwei Werte ohne Gegenpaar Segmente',
      ['senkrecht', 'waagerecht'],
      'segmente',
    ],
    [
      'zwei Werte mit nur einem Wort des Paares Segmente',
      ['an', 'nein'],
      'segmente',
    ],
    ['zwei Werte aus zwei Paaren Segmente', ['an', 'verborgen'], 'segmente'],
    ['drei Werte Segmente', ['aus', 'melden', 'holen'], 'segmente'],
    ['drei Werte mit an und aus Segmente', ['an', 'aus', 'melden'], 'segmente'],
    ['vier Werte Segmente', ['aus', 'klein', 'normal', 'gross'], 'segmente'],
    [
      'fünf Werte eine Auswahl',
      ['erben', 'klein', 'normal', 'gross', 'sehr'],
      'auswahl',
    ],
    [
      'acht Werte eine Auswahl',
      ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'],
      'auswahl',
    ],
  ] as const)('%s', (_name, werte, soll) => {
    expect(bedienung(werte).art).toBe(soll)
  })

  // **`erben` ist ein Wert unter anderen** — wer es wählen kann, soll es
  // wählen können. Drei Werte mit `erben` sind darum drei Segmente und kein
  // Schalter mit einem versteckten dritten Zustand.
  it('zählt den Erbwert mit', () => {
    expect(bedienung(['erben', 'sichtbar', 'verborgen']).art).toBe('segmente')
    expect(bedienung(['erben', 'system', 'hell', 'dunkel']).art).toBe(
      'segmente',
    )
  })
})

describe('das Paar eines Schalters', () => {
  it('nennt den Wert, der angeschaltet heißt — gleich in welcher Reihe', () => {
    expect(bedienung(['aus', 'an'])).toEqual({
      art: 'schalter',
      paar: { an: 'an', aus: 'aus' },
    })
    expect(bedienung(['verborgen', 'sichtbar'])).toEqual({
      art: 'schalter',
      paar: { an: 'sichtbar', aus: 'verborgen' },
    })
  })
})
