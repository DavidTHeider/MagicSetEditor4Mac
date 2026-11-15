// **Ein fremder Setstilwechsel bei offenem Editor** — #2314 (E2314).
//
// Eine Karte, die den Stil des Sets erbt, schreibt `''`. Wechselt ein
// anderer Prozess den Setstil, zeichnet die Vorschau mit dem neuen Paket —
// und das Formular ordnete weiter unter `''`, also nach der Lage des alten.
// **Gelernt und geordnet wird jetzt unter dem Paket, mit dem gezeichnet
// wurde**; der geschriebene Wert sagt nur, welches das gerade ist.
import { act, renderHook } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import type { FeldflaecheDTO } from '../formen'
import { useFeldordnung } from './formularruhe'
import { grundansicht } from './vorschauansicht'

function flaeche(feld: string, oben: number): FeldflaecheDTO {
  return { feld, links: 0, oben, breite: 10, hoehe: 2 }
}

interface Gezeigt {
  flaechen: FeldflaecheDTO[]
  stil: string
  gezeichnet: string
}

function ordnung(gezeigt: Gezeigt, stil: string) {
  return renderHook(
    ({ gezeigt: jetzt, stil: jetzigerStil }) =>
      useFeldordnung(useState(grundansicht), jetzt, false, jetzigerStil),
    { initialProps: { gezeigt, stil } },
  )
}

const altesPaket: Gezeigt = {
  flaechen: [flaeche('name', 0), flaeche('kosten', 5)],
  stil: '',
  gezeichnet: 'magic-alt.mse-style',
}

const neuesPaket: Gezeigt = {
  flaechen: [flaeche('kosten', 0), flaeche('name', 5)],
  stil: '',
  gezeichnet: 'magic-neu.mse-style',
}

describe('ein fremder Setstilwechsel', () => {
  it('ordnet nach dem Paket, mit dem jetzt gezeichnet wird', () => {
    const { result, rerender } = ordnung(altesPaket, '')

    expect(result.current.reihenfolge).toEqual(['name', 'kosten'])

    act(() => {
      rerender({ gezeigt: neuesPaket, stil: '' })
    })

    expect(result.current.reihenfolge).toEqual(['kosten', 'name'])
  })

  it('kehrt zur Ordnung des alten Pakets zurück, wenn es wiederkommt', () => {
    const { result, rerender } = ordnung(altesPaket, '')

    act(() => {
      rerender({ gezeigt: neuesPaket, stil: '' })
    })
    act(() => {
      rerender({ gezeigt: altesPaket, stil: '' })
    })

    expect(result.current.reihenfolge).toEqual(['name', 'kosten'])
  })
})
