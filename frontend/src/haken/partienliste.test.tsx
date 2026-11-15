// **Der Haken über die eigenen Partien** —
// #1600.
//
// # Wogegen hier geprüft wird
//
//   - Die Liste kommt an, und **leer heißt leer**, nicht `undefined`.
//   - Ein Fehler wird **gemeldet und nicht zur leeren Liste**: „ich konnte
//     nicht nachsehen“ und „du hast keine Partie“ sind zwei Auskünfte mit
//     zwei nächsten Schritten (§8).
import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { usePartienliste } from './usePartienliste'

import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

it('trägt die eigenen Partien ein', async () => {
  belegeApi({
    Partien: () =>
      Promise.resolve({
        partien: [
          {
            id: 'p1',
            fassung: '1',
            spiel: 'magic',
            stand: 'laeuft',
            sitze: [{ kennung: 'sitz-1', name: 'David' }],
            objekte: 6,
            zug: 0,
            eigenerSitz: '',
            dran: false,
          },
        ],
        fehler: null,
      }),
  })

  const { result } = renderHook(() => usePartienliste(), { wrapper: raum })

  await waitFor(() => {
    expect(result.current.partien).toHaveLength(1)
  })

  expect(result.current.partien[0]?.id).toBe('p1')
})

// **Ein Fehler steht als Fehler da.**
it('meldet, was beim Lesen schiefging', async () => {
  belegeApi({
    Partien: () =>
      Promise.resolve({
        partien: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'Der Bestand ist zu.',
          feld: '',
        },
      }),
  })

  const { result } = renderHook(() => usePartienliste(), { wrapper: raum })

  await waitFor(() => {
    expect(result.current.fehler).not.toBeUndefined()
  })

  // **Und die Liste bleibt leer statt `undefined`** — ein Aufrufer prüft eine
  // Länge und nicht zwei Dinge.
  expect(result.current.partien).toHaveLength(0)
})

// **Nachladen fragt noch einmal.**
//
// **Für den Aufrufer, der gerade eine Partie angefangen hat**: die Liste ist
// dann von vorhin, und §7 Regel 3 sagt, dass ein Nachladen Bestenfalls-Arbeit
// ist — es steht bereit, und wer es braucht, ruft es.
it('holt die Liste auf Zuruf erneut', async () => {
  const gefragt = vi.fn(() => Promise.resolve({ partien: [], fehler: null }))

  belegeApi({ Partien: gefragt })

  const { result } = renderHook(() => usePartienliste(), { wrapper: raum })

  await waitFor(() => {
    expect(gefragt).toHaveBeenCalledTimes(1)
  })

  result.current.nachladen()

  await waitFor(() => {
    expect(gefragt).toHaveBeenCalledTimes(2)
  })
})
