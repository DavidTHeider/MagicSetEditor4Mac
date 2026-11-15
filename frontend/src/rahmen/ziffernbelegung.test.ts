// **Welche Seite ⌘0 bis ⌘9 öffnen** — UX-02.AK-03
// (#1674).
import { describe, expect, it } from 'vitest'

import {
  alsWert,
  belegungAus,
  belegungVor06,
  vorgabebelegung,
} from './ziffernbelegung'
import { deuten } from './tastatur'

import type { Seitenschluessel } from '../navigation'

describe('die Belegung der Ziffern', () => {
  // **Die Vorgabe sind die Bereiche der Seitenleiste** — dieselben Ziffern,
  // die `navigation.ts` den ersten Seiten der Bereiche gibt.
  it('hat als Vorgabe die sechs Bereiche und vier freie Ziffern', () => {
    expect(vorgabebelegung).toEqual({
      '1': 'start',
      '2': 'sets',
      '3': 'sammlung',
      '4': 'decks',
      '5': 'handeln',
      '6': 'pakete',
    })
  })

  it('liest einen leeren Wert als Vorgabe', () => {
    expect(belegungAus('')).toEqual(vorgabebelegung)
  })

  // **Was genannt ist, übersteuert; was nicht, bleibt** — „3=karten" nimmt
  // der Sammlung ihre Ziffer und lässt die übrigen stehen.
  it('übersteuert nur die genannten Ziffern und räumt mit einem leeren Eintrag', () => {
    expect(belegungAus('3=karten,2=,0=handeln')).toEqual({
      '1': 'start',
      '3': 'karten',
      '4': 'decks',
      '5': 'handeln',
      '6': 'pakete',
      '0': 'handeln',
    })
  })

  // **Eine Seite, die es nicht gibt, fällt weg**, statt eine Ziffer ins
  // Leere zeigen zu lassen — die Domäne prüft die Gestalt, nicht die Seite.
  it('lässt eine unbekannte Seite und einen kaputten Eintrag fallen', () => {
    expect(belegungAus('3=gibtsnicht,7,8=turniere')).toEqual({
      ...vorgabebelegung,
      '8': 'turniere',
    })
  })

  it('schreibt nur, was von der Vorgabe abweicht — und nichts, wenn nichts', () => {
    expect(alsWert(vorgabebelegung)).toBe('')
    expect(alsWert({ ...vorgabebelegung, '3': 'karten', '7': 'decks' })).toBe(
      '3=karten,7=decks',
    )

    const ohneStart: Record<string, Seitenschluessel> = { ...vorgabebelegung }
    delete ohneStart['1']
    expect(alsWert(ohneStart)).toBe('1=')
  })

  it('kommt über Schreiben und Lesen unverändert zurück', () => {
    expect(belegungAus(alsWert(belegungVor06))).toEqual(belegungVor06)
  })

  // **Die Belegung vor 0.6.0** — die Tabelle der Migration in
  // `CHANGELOG.md` (#1753), Spalte „führte auf".
  it('kennt die Belegung vor 0.6.0', () => {
    expect(belegungVor06).toEqual({
      '1': 'start',
      '2': 'sets',
      '3': 'karten',
      '4': 'pakete',
      '5': 'vorlagen',
      '6': 'sammlung',
      '7': 'decks',
      '8': 'turniere',
      '9': 'spielen',
      '0': 'handeln',
    })
  })
})

describe('deuten mit einer Belegung', () => {
  const lage = (taste: string) => ({
    taste,
    code: `Digit${taste}`,
    meta: true,
    alt: false,
    umschalt: false,
    imEingabefeld: false,
  })

  it('führt eine Ziffer dorthin, wohin die Belegung sagt', () => {
    const belegung = belegungAus('3=karten')

    expect(deuten(lage('3'), belegung)).toEqual({
      art: 'seite',
      seite: 'karten',
    })
    expect(deuten(lage('1'), belegung)).toEqual({
      art: 'seite',
      seite: 'start',
    })
  })

  it('tut mit einer freien Ziffer nichts', () => {
    expect(deuten(lage('7'))).toBeNull()
    expect(deuten(lage('3'), belegungAus('3='))).toBeNull()
  })

  // **Nur die Ziffern sind einstellbar** — ⌘, ⌘' und ⌘? bleiben, wo sie
  // sind, gleich was die Belegung sagt.
  it('lässt die übrigen Seitenkürzel stehen', () => {
    expect(deuten(lage(','), belegungAus('1='))).toEqual({
      art: 'seite',
      seite: 'einstellungen',
    })
  })
})
