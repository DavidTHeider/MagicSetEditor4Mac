// **Angelegt und gewechselt ist geschehen** — §7, Regel 3
// (#1970).
//
// # Wogegen hier geprüft wird
//
// Go meldet nach dem Anlegen oder einem Standwechsel, dass die Sicht danach
// nicht zu lesen war — `nachladefehler` neben einer leeren Partie, die nur
// Kennung und Sicht trägt. Der Lauf nimmt den Schritt als geschehen und
// liest über den gewöhnlichen Weg nach; scheitert auch der, steht seine
// Meldung da — als Lesefehler und nicht als gescheiterter Schritt.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { usePartie } from './usePartie'

import { eineAntwort } from '../seiten/partieprobe'
import { einEntwurf } from '../test/partieentwurf'

import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// ohneSicht ist, was Go nach dem Schreiben ohne Sicht schickt.
function ohneSicht() {
  const leer = eineAntwort({ id: 'p1', fuer: 'sitz1' })

  return {
    ...leer,
    nachladefehler: {
      code: 'validierung' as const,
      schluessel: '',
      meldung: 'Die Sicht klemmt',
    },
  }
}

it('liest nach dem Anlegen ohne Sicht selbst nach', async () => {
  const gelesen = vi.fn(() => Promise.resolve(eineAntwort({ id: 'p1' })))

  belegeApi({
    PartieAnlegen: () => Promise.resolve(ohneSicht()),
    Partie: gelesen,
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.partie?.id).toBe('p1')
  })

  expect(gelesen).toHaveBeenCalledWith('p1', 'sitz1')
  expect(result.current.fehler).toBeUndefined()
})

it('meldet, wenn auch das Nachlesen scheitert — als Lesefehler', async () => {
  belegeApi({
    PartieAnlegen: () => Promise.resolve(ohneSicht()),
    Partie: () =>
      Promise.resolve({
        ...eineAntwort(),
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'Die Sicht klemmt noch',
        },
      }),
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.fehler).toBe('Die Sicht klemmt noch')
  })

  expect(result.current.fehlerVorgang).toBe('partie-ansehen-p1')
})

it('liest nach einem Standwechsel ohne Sicht selbst nach', async () => {
  const gelesen = vi.fn(() =>
    Promise.resolve(eineAntwort({ id: 'p1', stand: 'pausiert' })),
  )

  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort({ id: 'p1' })),
    PartiePausieren: () => Promise.resolve(ohneSicht()),
    Partie: gelesen,
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.partie).toBeDefined()
  })

  act(() => {
    result.current.wechseln('pausieren')
  })

  await waitFor(() => {
    expect(result.current.partie?.stand).toBe('pausiert')
  })

  expect(gelesen).toHaveBeenCalledTimes(1)
  expect(result.current.fehler).toBeUndefined()
})
