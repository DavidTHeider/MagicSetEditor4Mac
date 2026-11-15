// **Ein Nachlesestand lädt nach, ohne zu räumen** (#2069).
//
// # Der Befund
//
// Mehrere Stellen trugen einen Nachladezähler — den Stand fremder
// Schreibvorgänge, eine eigene Runde, den Verlaufsstand — in den
// Abhängigkeiten. Für `useLaden` ist das eine neue **Frage**, und eine neue
// Frage räumt (§7, Regel 1). Gemessen am nativen Fenster: die Paketliste
// ersetzte ihre Tabelle für einen Augenblick durch „Wird geladen", die Seite
// sprang um 37 Punkte, und ein Klick traf die Überschrift statt den Knopf.
//
// # Die Trennung
//
// Der **Bezug** steht in den Abhängigkeiten und räumt; der **Nachlesestand**
// steht daneben und lädt nach wie `nachladen` — derselbe Bezug, ein neuer
// Stand.
import { renderHook, waitFor } from '@testing-library/react'
import { expect, it } from 'vitest'

import { useLaden } from './useLaden'

interface Eigenschaften {
  setID: string
  stand: number
}

function useNachlesehaken({ setID, stand }: Eigenschaften) {
  return useLaden(
    () => Promise.resolve(`${setID}@${String(stand)}`),
    [setID],
    false,
    stand,
  )
}

it('lädt bei einem neuen Nachlesestand nach und behält die Daten', async () => {
  const { result, rerender } = renderHook(useNachlesehaken, {
    initialProps: { setID: 'eins', stand: 0 },
  })

  await waitFor(() => {
    expect(result.current.daten).toBe('eins@0')
  })

  rerender({ setID: 'eins', stand: 1 })

  // **Unmittelbar danach**, bevor die Antwort kommt: es lädt, und die Liste
  // steht noch da.
  expect(result.current.laedt).toBe(true)
  expect(result.current.daten).toBe('eins@0')

  await waitFor(() => {
    expect(result.current.daten).toBe('eins@1')
  })
})

it('räumt bei einem Bezugswechsel auch mit Nachlesestand', async () => {
  const { result, rerender } = renderHook(useNachlesehaken, {
    initialProps: { setID: 'eins', stand: 0 },
  })

  await waitFor(() => {
    expect(result.current.daten).toBe('eins@0')
  })

  rerender({ setID: 'zwei', stand: 1 })

  expect(result.current.daten).toBeUndefined()

  await waitFor(() => {
    expect(result.current.daten).toBe('zwei@1')
  })
})
