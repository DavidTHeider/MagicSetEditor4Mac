// **Ein Katalog, flach und wieder gefaltet** — LOC-02
// (#1688).
import { describe, expect, it } from 'vitest'

import { flach, ueberlagere } from './flach'

const katalog = {
  titel: 'Sets',
  zaehlwoerter: {
    karten: { eine: 'Karte', viele: 'Karten' },
    seiten: { eine: 'Seite', viele: 'Seiten' },
  },
}

describe('flach', () => {
  it('gibt jeden Satz unter seinem Pfad', () => {
    expect(flach(katalog)).toEqual({
      titel: 'Sets',
      'zaehlwoerter.karten.eine': 'Karte',
      'zaehlwoerter.karten.viele': 'Karten',
      'zaehlwoerter.seiten.eine': 'Seite',
      'zaehlwoerter.seiten.viele': 'Seiten',
    })
  })
})

describe('ueberlagere', () => {
  it('setzt die Sätze einer Übersetzung auf einen Katalog', () => {
    const ergebnis = ueberlagere(katalog, {
      titel: 'Conjuntos',
      'zaehlwoerter.karten.wenige': 'карты',
    })

    expect(ergebnis.titel).toBe('Conjuntos')
    expect(ergebnis.zaehlwoerter.karten).toEqual({
      eine: 'Karte',
      wenige: 'карты',
      viele: 'Karten',
    })
  })

  // **Was die Vorlage nicht kennt, bleibt draußen** — ein Pfad aus einer
  // neueren Fassung legt keinen Zweig an, den niemand liest, und einer,
  // der auf einen Satz statt auf einen Zweig trifft, zerstört ihn nicht.
  it('lässt unbekannte Pfade und Pfade durch einen Satz fallen', () => {
    const ergebnis = ueberlagere(katalog, {
      'gibt.es.nicht': 'x',
      'titel.unter': 'y',
      'zaehlwoerter.karten.eine.mehr': 'z',
    })

    expect(ergebnis).toEqual(katalog)
  })

  // **Die Regel reist mit der Form** — Befund N5 des Nachlaufs über Phase
  // 6.5: nur eine Zählform, deren Wörter die Übersetzung setzt, zählt nach
  // ihrer Sprache; eine aus der Basis zählt nach der Basis.
  it('gibt einer gesetzten Zählform die Regel der Übersetzung', () => {
    const ergebnis = ueberlagere(
      katalog,
      { 'zaehlwoerter.karten.viele': 'kart' },
      'pl',
    )

    expect(ergebnis.zaehlwoerter.karten).toHaveProperty('regel', 'pl')
    expect(ergebnis.zaehlwoerter.seiten).not.toHaveProperty('regel')
    expect(Object.keys(ergebnis.zaehlwoerter.karten)).toEqual(['eine', 'viele'])
    expect(
      ueberlagere(katalog, { titel: 'x' }).zaehlwoerter.karten,
    ).not.toHaveProperty('regel')
  })

  it('ändert den Grund nicht', () => {
    ueberlagere(katalog, { titel: 'Conjuntos' })

    expect(katalog.titel).toBe('Sets')
  })
})
