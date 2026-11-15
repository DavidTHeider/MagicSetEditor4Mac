// **Die Schrift über die ganze Oberfläche** — UX-09.AK-02
// (#1681):
// *„Schriftwahl wirkt wahlweise auf Überschriften oder gesamte UI; fehlende
// Glyphen erhalten geeignete Fallbacks, insbesondere für japanische und
// chinesische Texte."*
import { beforeEach, describe, expect, it } from 'vitest'

import {
  anzeigeschriften,
  schriftbereichGilt,
  setzeSchriftbereich,
  textschriften,
} from './erscheinungsbild'

let wurzel: HTMLElement

beforeEach(() => {
  wurzel = document.createElement('html')
})

describe('Die Schriftketten', () => {
  // **Japanisch vor Chinesisch**: Kana gibt es nur in japanischen Schriften,
  // und eine chinesische zuerst gäbe einem japanischen Namen chinesische
  // Zeichenformen. Beide stehen **vor** der allgemeinen Familie — sonst
  // entschiede der Rückfall des Systems, und der nähme für eine Serifenschrift
  // womöglich eine serifenlose.
  it.each(Object.entries(anzeigeschriften))(
    '%s trägt japanische und chinesische Rückfälle vor der allgemeinen Familie',
    (_wahl, kette) => {
      const teile = kette.split(',').map((teil) => teil.trim())
      const japanisch = teile.findIndex((teil) => teil.includes('Hiragino'))
      const chinesisch = teile.findIndex(
        (teil) => teil.includes('PingFang') || teil.includes('Songti'),
      )

      expect(japanisch).toBeGreaterThanOrEqual(0)
      expect(chinesisch).toBeGreaterThan(japanisch)
      expect(chinesisch).toBeLessThan(teile.length - 1)
    },
  )

  it('kennt vier Textschriften und drei Schmuckschriften', () => {
    expect([...textschriften].sort()).toEqual(
      ['buch', 'pergament', 'rund', 'serifenlos'].sort(),
    )
    expect(Object.keys(anzeigeschriften)).toHaveLength(7)
  })
})

describe('setzeSchriftbereich', () => {
  it('legt eine Textschrift über die ganze Oberfläche', () => {
    expect(setzeSchriftbereich(wurzel, 'ganz', 'buch')).toBe(true)
    expect(wurzel.style.getPropertyValue('--font-text')).toBe(
      anzeigeschriften.buch,
    )
  })

  // **Eine Schmuckschrift bleibt an den Überschriften** (AK-03, Lesbarkeit):
  // Kalligrafie über einem Absatz ist keine Gestaltung, sondern eine Zumutung.
  it('lässt eine Schmuckschrift an den Überschriften', () => {
    wurzel.style.setProperty('--font-text', 'irgendwas')

    expect(setzeSchriftbereich(wurzel, 'ganz', 'kalligrafie')).toBe(false)
    expect(wurzel.style.getPropertyValue('--font-text')).toBe('')
  })

  it('nimmt die Schrift des Textes zurück, wenn nur Überschriften gemeint sind', () => {
    setzeSchriftbereich(wurzel, 'ganz', 'rund')

    expect(setzeSchriftbereich(wurzel, 'ueberschriften', 'rund')).toBe(true)
    expect(wurzel.style.getPropertyValue('--font-text')).toBe('')
  })
})

// **Ob die Wahl gilt, sagt eine reine Rechnung** — der Gestaltungsraum
// braucht die Antwort beim Zeichnen, nicht erst nach dem Setzen; und sie
// darf nie etwas anderes sagen als das, was `setzeSchriftbereich` tut.
describe('schriftbereichGilt', () => {
  const bereiche = ['ganz', 'ueberschriften', '']
  const wahlen = [...Object.keys(anzeigeschriften), 'unbekannt', '']

  it.each(bereiche.flatMap((bereich) => wahlen.map((wahl) => [bereich, wahl])))(
    'sagt für %s und %s dasselbe wie das Setzen',
    (bereich, wahl) => {
      expect(schriftbereichGilt(bereich, wahl)).toBe(
        setzeSchriftbereich(wurzel, bereich, wahl),
      )
    },
  )

  it('lässt nur eine Schmuckschrift über der ganzen Oberfläche nicht gelten', () => {
    expect(schriftbereichGilt('ganz', 'kalligrafie')).toBe(false)
    expect(schriftbereichGilt('ganz', 'buch')).toBe(true)
    expect(schriftbereichGilt('ueberschriften', 'kalligrafie')).toBe(true)
  })
})
