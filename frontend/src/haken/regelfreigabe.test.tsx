// **Was die Regelfreigabe behält, während sie neu fragt** — P1a
// (#2335).
//
// # Wogegen hier geprüft wird
//
// **In beide Richtungen** (§7, Regel 1 und 2): eine neue Abwahl am selben
// Spiel lässt die Antwort stehen, bis die neue da ist — sonst verschwindet
// die Tafel, an der gerade gewählt wird. **Eine andere Sitzzahl dagegen ist
// ein anderer Bezug**, und dort zeigt der Haken nichts Altes: die Antwort
// gehörte zu einem Stand, den es nicht mehr gibt.
import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useRegelfreigabe } from './useRegelfreigabe'

afterEach(() => {
  raeumeApi()
})

// eineFreigabe ist eine Antwort, an deren Kennung sich der Lauf erkennt.
function eineFreigabe(regelwerk: string) {
  return {
    freigabe: {
      erlaubt: true,
      regelwerk,
      fassung: 1,
      wirksam: '2026-08-19',
      hoechste: 'begleitet',
      matrix: [],
      luecken: [],
      hindernisse: [],
      modus: 'frei',
      baum: [],
      widersprueche: [],
    },
    fehler: null,
  }
}

// mitZweiAntworten antwortet beim ersten Mal sofort und lässt jede weitere
// Frage offen — so ist der Augenblick dazwischen zu sehen.
function mitZweiAntworten() {
  let gefragt = 0

  belegeApi({
    Regelfreigabe: () => {
      gefragt += 1

      return gefragt === 1
        ? Promise.resolve(eineFreigabe('erste'))
        : new Promise(() => undefined)
    },
  })
}

it('behält die Antwort desselben Spiels, solange eine neue Abwahl unterwegs ist', async () => {
  mitZweiAntworten()

  const { result, rerender } = renderHook(
    ({ abgewaehlt }: { abgewaehlt: string[] }) =>
      useRegelfreigabe('magic', 'sandkasten', 2, 'frei', abgewaehlt),
    { initialProps: { abgewaehlt: [] as string[] } },
  )

  await waitFor(() => {
    expect(result.current.freigabe?.regelwerk).toBe('erste')
  })

  rerender({ abgewaehlt: ['zugstruktur'] })

  expect(result.current.laeuft).toBe(true)
  expect(result.current.freigabe?.regelwerk).toBe('erste')
})

it('zeigt nach einer anderen Sitzzahl nichts Altes', async () => {
  mitZweiAntworten()

  const { result, rerender } = renderHook(
    ({ sitze }: { sitze: number }) =>
      useRegelfreigabe('magic', 'sandkasten', sitze, 'frei', []),
    { initialProps: { sitze: 2 } },
  )

  await waitFor(() => {
    expect(result.current.freigabe?.regelwerk).toBe('erste')
  })

  rerender({ sitze: 3 })

  expect(result.current.laeuft).toBe(true)
  expect(result.current.freigabe).toBeUndefined()
})
