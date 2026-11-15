// **Die Ordnung des Formulars gehört zu dem Stil, mit dem gezeichnet wurde**
// — Zweitblick über die Folgen vor 0.8.0, Befunde 8 und 9 (#2309).
import { act, renderHook } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import type { FeldflaecheDTO } from '../formen'
import { useFeldordnung } from './formularruhe'
import { grundansicht, type Vorschauansicht } from './vorschauansicht'

function flaeche(
  feld: string,
  oben: number,
  links: number,
  hoehe = 2,
): FeldflaecheDTO {
  return { feld, links, oben, breite: 10, hoehe }
}

function ordnung(
  gezeigt: { flaechen: FeldflaecheDTO[]; stil: string } | undefined,
  stil: string,
  anfang: Vorschauansicht = grundansicht,
) {
  return renderHook(
    ({ gezeigt: jetzt, stil: jetzigerStil }) => {
      const ansicht = useState(anfang)

      return {
        ordnung: useFeldordnung(ansicht, jetzt, false, jetzigerStil),
        ansicht: ansicht[0],
      }
    },
    { initialProps: { gezeigt, stil } },
  )
}

describe('die Ordnung je Stil', () => {
  // **Befund 8**: beim Stilwechsel derselben Karte steht noch die vorige
  // Vorschau da. Ihre Flächen gehören dem vorigen Stil, nicht dem, der jetzt
  // geschrieben steht.
  it('legt die Flächen unter dem Stil ab, mit dem die Vorschau gezeichnet ist', () => {
    const { result } = ordnung(
      { flaechen: [flaeche('name', 0, 0)], stil: 'alt' },
      'neu',
    )

    expect(result.current.ansicht.flaechenJeStil?.alt).toHaveLength(1)
    expect(result.current.ansicht.flaechenJeStil?.neu).toBeUndefined()
  })

  it('ordnet nicht nach einer Vorschau eines anderen Stils', () => {
    const { result } = ordnung(
      { flaechen: [flaeche('name', 0, 0)], stil: 'alt' },
      'neu',
    )

    expect(result.current.ordnung.reihenfolge).toBeUndefined()
  })

  // **Befund 9**: A und B stehen in zwei Zeilen — A oben rechts, B darunter
  // links. Kommt ein hohes Feld dazu, das beide in eine Zeile fasst, rechnete
  // die Leseordnung B vor A. Was einmal stand, bleibt stehen.
  it('behält die Reihenfolge, wenn ein neues Feld zwei Zeilen zusammenfasst', () => {
    const a = flaeche('a', 1, 100)
    const b = flaeche('b', 3, 0)
    const { result, rerender } = ordnung({ flaechen: [a, b], stil: '' }, '')

    expect(result.current.ordnung.reihenfolge).toEqual(['a', 'b'])

    act(() => {
      rerender({
        gezeigt: { flaechen: [a, b, flaeche('hoch', 0, 50, 10)], stil: '' },
        stil: '',
      })
    })

    const reihe = result.current.ordnung.reihenfolge ?? []
    expect(reihe.indexOf('a')).toBeLessThan(reihe.indexOf('b'))
    expect(reihe).toContain('hoch')
  })
})
