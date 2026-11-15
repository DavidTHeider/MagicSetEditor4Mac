// **Wie groß ein Schild auf dem Tisch steht — oder ob überhaupt** — die
// Einstellung `tisch.beschriftung`, seit dem 2026-09-24.
import { describe, expect, it } from 'vitest'

import { Kopfhoehe, Namenshoehe } from './masse'
import { schriftgroesse } from './schildmass'
import { Sitzschildgroesse } from './sitzschilder'
import { Schildgroesse } from './zonenschilder'

import type { Zonenschild } from './zonenschilder'

const name: Zonenschild = {
  kennung: 'sitz1:hand:name',
  x: 0,
  y: 0,
  text: 'HAND',
  groesse: Schildgroesse,
  ankerY: 0.5,
  art: 'name',
}
const zahl: Zonenschild = { ...name, kennung: 'sitz1:hand:zahl', art: 'zahl' }
const kopf: Zonenschild = {
  ...name,
  kennung: 'sitz1:sitz:kopf',
  groesse: Sitzschildgroesse,
  art: 'kopf',
}

describe('die Schriftgröße eines Schildes', () => {
  it('schreibt mittel in der Größe des Schildes', () => {
    expect(schriftgroesse(name, 'normal')).toBe(Schildgroesse)
    expect(schriftgroesse(kopf, 'normal')).toBe(Sitzschildgroesse)
  })

  it('schreibt klein ein Fünftel kleiner und groß ein Fünftel größer', () => {
    expect(schriftgroesse(kopf, 'klein')).toBeCloseTo(Sitzschildgroesse * 0.8)
    expect(schriftgroesse(kopf, 'gross')).toBeCloseTo(Sitzschildgroesse * 1.2)
  })

  // **Aus nimmt, was anderswo auch steht** — der Kopf steht in der
  // Spielerleiste, der Name einer Zone in jedem Kartenknopf. Die Anzahl
  // steht nirgends sonst.
  it('nimmt bei »aus« Namen und Köpfe weg und lässt die Anzahl stehen', () => {
    expect(schriftgroesse(name, 'aus')).toBeUndefined()
    expect(schriftgroesse({ ...name, art: 'kurzname' }, 'aus')).toBeUndefined()
    expect(schriftgroesse(kopf, 'aus')).toBeUndefined()
    expect(schriftgroesse(zahl, 'aus')).toBe(Schildgroesse)
  })

  // **Ein Wert, den dieses Fenster nicht kennt, ist kein Fehler** — eine
  // spätere Fassung kann eine Stufe mehr haben.
  it('liest einen unbekannten Wert wie »mittel«', () => {
    expect(schriftgroesse(name, 'riesig')).toBe(Schildgroesse)
    expect(schriftgroesse(name, '')).toBe(Schildgroesse)
  })

  // **Auch groß passt in sein Band** — sonst läge der Kopf wieder auf dem
  // Namen der ersten Zone, und der Name auf der ersten Karte.
  it('passt auch in groß in sein Band', () => {
    expect(schriftgroesse(kopf, 'gross')).toBeLessThanOrEqual(Kopfhoehe * 0.85)
    expect(schriftgroesse(name, 'gross')).toBeLessThanOrEqual(
      Namenshoehe * 0.85,
    )
  })
})
