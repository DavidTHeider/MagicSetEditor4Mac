// **Wie ein Feldwert aussieht und wie eine Karte heißt** —
// #1623.
import { describe, expect, test } from 'vitest'

import { alsListe, alsText, kartenwort, zielVon } from './feldwerte'
import { einObjekt } from '../seiten/partieprobe'

describe('alsText', () => {
  test('gibt eine Zeichenkette unverändert', () => {
    expect(alsText('Soldat')).toBe('Soldat')
  })

  test('und fügt eine Liste an den Trennzeichen zusammen', () => {
    expect(alsText(['Soldat', 'Creature'])).toBe('Soldat|Creature')
  })
})

describe('alsListe', () => {
  test('gibt eine Liste unverändert', () => {
    expect(alsListe(['a', 'b'])).toEqual(['a', 'b'])
  })

  test('macht aus einer Zeichenkette eine Liste mit einem Eintrag', () => {
    expect(alsListe('a')).toEqual(['a'])
  })

  test('und aus einer leeren keine', () => {
    expect(alsListe('')).toEqual([])
  })
})

describe('kartenwort', () => {
  test('nennt den Namen aus den Eigenschaften', () => {
    expect(
      kartenwort(
        einObjekt({
          name: 'am Objekt',
          eigenschaften: {
            name: 'Wald',
            farben: [],
            manawert: 0,
            supertypen: [],
            typen: [],
            untertypen: [],
            staerke: { steht: false, stern: false, zahl: 0 },
            widerstandskraft: { steht: false, stern: false, zahl: 0 },
            loyalitaet: { steht: false, stern: false, zahl: 0 },
            verteidigung: { steht: false, stern: false, zahl: 0 },
            unbekannteTypen: [],
          },
        }),
      ),
    ).toBe('Wald')
  })

  test('dann den am Objekt', () => {
    expect(kartenwort(einObjekt({ name: 'Soldat' }))).toBe('Soldat')
  })

  test('dann die Karte', () => {
    expect(kartenwort(einObjekt({ name: '', karte: 'k7' }))).toBe('k7')
  })

  test('und zuletzt die Kennung', () => {
    expect(
      kartenwort(einObjekt({ kennung: 'sitz1/x/0', name: '', karte: '' })),
    ).toBe('sitz1/x/0')
  })
})

describe('zielVon', () => {
  test('gibt die zweite Hälfte einer Zeile', () => {
    expect(zielVon(['a|sitz2'], 'a')).toBe('sitz2')
  })

  test('leer, wo keine Zeile steht', () => {
    expect(zielVon([], 'a')).toBe('')
  })

  test('und leer, wo die zweite Hälfte leer ist', () => {
    expect(zielVon(['a|'], 'a')).toBe('')
  })

  test('eine Kennung mit Trennzeichen bleibt ganz', () => {
    // **Ein Ziel kann selbst eines tragen** — Objektkennungen sehen aus wie
    // `sitz1/k1/0`, und niemand verbietet ein `|`. Abgeschnitten statt
    // zerteilt bleibt es heil.
    expect(zielVon(['a|b|c'], 'a')).toBe('b|c')
  })
})
