// **Ein Nachlesefehler nimmt die bestätigten Daten nicht weg** — U07
// (#2123).
//
// # Der Befund
//
// Schlug ein Nachlesen desselben Bezugs fehl, trug `useLaden` den Fehler ein
// und **warf die Daten weg** — für beide Werte von `behalten`. Das bekannte
// Arbeitsobjekt verschwand, und an seiner Stelle stand nur die Meldung, obwohl
// sich an ihm nichts geändert hatte außer dem Weg dorthin.
//
// # Der Vertrag
//
// - **Gleicher Bezug**: die bestätigten Daten bleiben, der Fehler steht
//   daneben; ein Wiederholen aktualisiert beides.
// - **Neuer Bezug**: keine fremden alten Daten — es sei denn, die Stelle hat
//   sich ausdrücklich dafür entschieden (`behalten`, die Vorschau).
import { act, renderHook, waitFor } from '@testing-library/react'
import { expect, it } from 'vitest'

import { useLaden } from './useLaden'

import { Bindungsfehler } from '../api'

interface Eigenschaften {
  bezug: string
  stand: number
  scheitert: boolean
  behalten: boolean
}

function useProbe({ bezug, stand, scheitert, behalten }: Eigenschaften) {
  return useLaden(
    () =>
      scheitert
        ? Promise.reject(
            new Bindungsfehler({
              code: 'unbekannt',
              schluessel: 'unbekannt',
              meldung: 'der Bestand klemmt',
              feld: '',
              werte: {},
            }),
          )
        : Promise.resolve(`${bezug}@${String(stand)}`),
    [bezug],
    behalten,
    stand,
  )
}

for (const behalten of [false, true]) {
  it(`behält beim Nachlesefehler die bestätigten Daten (behalten: ${String(behalten)})`, async () => {
    const { result, rerender } = renderHook(useProbe, {
      initialProps: { bezug: 'eins', stand: 0, scheitert: false, behalten },
    })

    await waitFor(() => {
      expect(result.current.daten).toBe('eins@0')
    })

    rerender({ bezug: 'eins', stand: 1, scheitert: true, behalten })

    await waitFor(() => {
      expect(result.current.fehler).toBeDefined()
    })

    // **Daten und Fehler zugleich** — das Objekt ist bekannt, nur das
    // Nachlesen ging schief.
    expect(result.current.daten).toBe('eins@0')

    // **Und ein Wiederholen aktualisiert beides.**
    rerender({ bezug: 'eins', stand: 1, scheitert: false, behalten })
    act(() => {
      result.current.nachladen()
    })

    await waitFor(() => {
      expect(result.current.daten).toBe('eins@1')
    })
    expect(result.current.fehler).toBeUndefined()
  })
}

it('zeigt bei einem neuen Bezug, der scheitert, keine fremden alten Daten', async () => {
  const { result, rerender } = renderHook(useProbe, {
    initialProps: {
      bezug: 'eins',
      stand: 0,
      scheitert: false,
      behalten: false,
    },
  })

  await waitFor(() => {
    expect(result.current.daten).toBe('eins@0')
  })

  rerender({ bezug: 'zwei', stand: 0, scheitert: true, behalten: false })

  await waitFor(() => {
    expect(result.current.fehler).toBeDefined()
  })

  expect(result.current.daten).toBeUndefined()
})
