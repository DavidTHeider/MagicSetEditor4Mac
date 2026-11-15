// **Was ein Blick gezeigt hat** —
// #1836.
//
// # Wogegen hier geprüft wird
//
// **Dass die private Ergänzung so weit gelesen wird, wie sie trägt.** Sie
// kommt unbekannt getypt an; ein Eintrag, der keiner ist, fällt heraus,
// statt die Liste umzuwerfen (§8).
import { describe, expect, test } from 'vitest'

import { gesehene } from './blickkarten'

describe('gesehene', () => {
  test('liest die Karten der Reihe nach, von oben', () => {
    expect(
      gesehene({
        karten: [
          { objekt: 'a/2', karte: 'p1', name: 'Berg', ordnung: 2 },
          { objekt: 'a/1', karte: 'p0', name: 'Wald', ordnung: 1 },
        ],
      }),
    ).toEqual([
      { objekt: 'a/1', name: 'Wald', ordnung: 1 },
      { objekt: 'a/2', name: 'Berg', ordnung: 2 },
    ])
  })

  // **Eine Kennung ist besser als ein leerer Eintrag** — erst die der Karte,
  // dann die des Objekts.
  test('fällt ohne Namen auf die Kennungen zurück', () => {
    expect(
      gesehene({
        karten: [
          { objekt: 'a/1', karte: 'p0-1', name: '', ordnung: 1 },
          { objekt: 'a/2', ordnung: 2 },
        ],
      }).map((eine) => eine.name),
    ).toEqual(['p0-1', 'a/2'])
  })

  test('gibt ohne Karten eine leere Liste', () => {
    expect(gesehene({})).toEqual([])
    expect(gesehene({ karten: 'keine Liste' })).toEqual([])
  })

  // **Was kein Eintrag ist, fällt heraus**, und ein Eintrag ohne Ordnung
  // steht vorn — die Null sagt „unbekannt", und oben ist, wo man zuerst
  // hinsieht.
  test('übergeht, was kein Eintrag ist', () => {
    expect(
      gesehene({
        karten: [
          null,
          'Wald',
          { objekt: 'a/1', name: 'Wald', ordnung: 1 },
          {
            objekt: 'a/0',
            name: 'Berg',
            ordnung: 'oben',
          },
        ],
      }),
    ).toEqual([
      { objekt: 'a/0', name: 'Berg', ordnung: 0 },
      { objekt: 'a/1', name: 'Wald', ordnung: 1 },
    ])
  })
})
