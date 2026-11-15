// **Der Stand der Darstellung, rein gerechnet** — UXR-F14
// (#2029).
import { describe, expect, it } from 'vitest'

import {
  anfangsstand,
  fehlt,
  gleicherStand,
  gleichesTeil,
  zusammenfuehren,
  type Darstellungsstand,
} from './darstellungsstand'

const steht = { vorhanden: true, offen: true }
const zu = { vorhanden: true, offen: false }

describe('anfangsstand', () => {
  // **Derselbe Nullwert wie `app.Darstellungsstand` in Go**: nichts ist da,
  // alles ist zu. Ändert einer von beiden seine Vorgabe, ohne den anderen,
  // meldet das Fenster beim Einhängen einen Stand, der Go etwas anderes
  // sagt, als es für „nichts" annimmt.
  it('hat nichts da und alles zu', () => {
    expect(anfangsstand).toEqual({
      seitenleiste: { vorhanden: false, offen: false },
      navigator: { vorhanden: false, offen: false },
      inspector: { vorhanden: false, offen: false },
    })
    expect(fehlt).toEqual({ vorhanden: false, offen: false })
  })
})

describe('zusammenfuehren', () => {
  it('trägt ein Teil ein und lässt die übrigen stehen', () => {
    const stand = zusammenfuehren(anfangsstand, {
      teil: 'navigator',
      wert: steht,
    })

    expect(stand).toEqual({ ...anfangsstand, navigator: steht })
  })

  it('verändert den Stand davor nicht', () => {
    zusammenfuehren(anfangsstand, { teil: 'inspector', wert: zu })

    expect(anfangsstand.inspector).toEqual(fehlt)
  })

  // **Derselbe Stand und nicht ein gleicher**: React zeichnet dann nicht neu,
  // und der Raum sendet nichts.
  it('gibt bei gleichem Wert denselben Stand zurück', () => {
    const stand = zusammenfuehren(anfangsstand, {
      teil: 'seitenleiste',
      wert: steht,
    })

    expect(
      zusammenfuehren(stand, { teil: 'seitenleiste', wert: { ...steht } }),
    ).toBe(stand)
  })

  it('nimmt ein Teil zurück, wenn es „fehlt" meldet', () => {
    const stand = zusammenfuehren(
      zusammenfuehren(anfangsstand, { teil: 'navigator', wert: steht }),
      { teil: 'navigator', wert: fehlt },
    )

    expect(stand).toEqual(anfangsstand)
  })
})

describe('gleichesTeil und gleicherStand', () => {
  it('vergleichen beide Angaben', () => {
    expect(gleichesTeil(steht, { ...steht })).toBe(true)
    expect(gleichesTeil(steht, zu)).toBe(false)
    expect(gleichesTeil(steht, { vorhanden: false, offen: true })).toBe(false)
  })

  // **Jedes der drei Teile zählt** — ein Vergleich, der eines vergisst,
  // hielte einen veränderten Stand für den alten, und die Leiste bliebe
  // stehen.
  it('sehen jedes Teil', () => {
    const stand: Darstellungsstand = {
      seitenleiste: steht,
      navigator: steht,
      inspector: steht,
    }

    expect(gleicherStand(stand, { ...stand })).toBe(true)

    for (const teil of ['seitenleiste', 'navigator', 'inspector'] as const) {
      expect(gleicherStand(stand, { ...stand, [teil]: zu })).toBe(false)
    }
  })
})
