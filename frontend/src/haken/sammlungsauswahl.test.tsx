// **Ein Ansichtswechsel behält die Haken** — #2016, Folge aus UXR-10
// (#1918).
//
// Das Artboard *Sammlung* sagt es ausdrücklich: Tabelle, Galerie und Gruppen
// sind drei Blicke auf **dieselben** Zeilen, und wer zwischen ihnen wechselt,
// verliert nicht, was er angehakt hat. Bis dahin räumte jeder Teil der Sicht
// die Haken — auch die, die keine Zeile weg- oder hinzunehmen.
//
// **Ein Filterwechsel räumt weiter**: Suche und Ausführung ändern, welche
// Zeilen dastehen, und ein Haken an einer Zeile, die niemand mehr sieht, wäre
// eine Massenänderung an Unsichtbarem.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import type { Sicht } from '../seiten/sammlungssicht'
import { belegeApi, raeumeApi } from '../test/api'
import { bestand, exemplar } from '../test/sammlungsprobe'
import { useSammlung } from './useSammlung'

afterEach(() => {
  raeumeApi()
})

async function mitZweiHaken() {
  belegeApi({
    Sammlungsfenster: (
      _suche: string,
      _ausfuehrung: string,
      _sortierung: string,
      ab: number,
      anzahl: number,
    ) =>
      Promise.resolve({
        exemplare: Array.from(
          { length: Math.min(anzahl, bestand - ab) },
          (_, versatz) =>
            exemplar(ab + versatz, `Karte ${String(ab + versatz)}`),
        ),
        gesamt: bestand,
        insgesamt: bestand,
        stueck: bestand,
        fehler: null,
      }),
  })

  const ansicht = renderHook(() => useSammlung())
  await waitFor(() => {
    expect(ansicht.result.current.gesamt).toBe(bestand)
  })

  const [erste, zweite] = ansicht.result.current.exemplare
  act(() => {
    ansicht.result.current.anhaken(erste?.id ?? -1, true)
    ansicht.result.current.anhaken(zweite?.id ?? -1, true)
  })
  expect(ansicht.result.current.gewaehlt.size).toBe(2)

  return ansicht
}

describe('Die Haken über einen Wechsel der Sicht', () => {
  const blicke: [string, Partial<Sicht>][] = [
    ['die Ansicht', { ansicht: 'galerie' }],
    ['die Gruppierung', { gruppierung: 'set' }],
    ['die Sortierung', { sortierung: 'menge' }],
  ]

  for (const [was, teil] of blicke) {
    it(`bleiben, wenn ${was} wechselt`, async () => {
      const { result } = await mitZweiHaken()

      act(() => {
        result.current.setzeSicht(teil)
      })

      expect(result.current.gewaehlt.size).toBe(2)
    })
  }

  const filter: [string, Partial<Sicht>][] = [
    ['die Suche', { suche: 'Drache' }],
    ['die Ausführung', { ausfuehrung: 'foil' }],
  ]

  for (const [was, teil] of filter) {
    it(`gehen, wenn ${was} wechselt`, async () => {
      const { result } = await mitZweiHaken()

      act(() => {
        result.current.setzeSicht(teil)
      })

      expect(result.current.gewaehlt.size).toBe(0)
    })
  }
})
