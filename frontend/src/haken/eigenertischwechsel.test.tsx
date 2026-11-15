// **Der eigene Tisch beim Wechsel und beim Nachholen** — N1 und N3 aus dem
// Nachlauf über Phase 6.11 (#1672).
//
// - **N1**: während eines Sitzwechsels stand die alte Partie noch da, und ein
//   Zug für den alten Sitz kam samt privater Ergänzung in die Liste — die
//   danach dem neuen Sitz gehörte.
// - **N3**: zwei nachgeholte Bilder hintereinander, erst das neuere, dann das
//   ältere, verglichen beide gegen die geladene Sicht; das ältere gewann.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { usePartie } from './usePartie'
import {
  ereignisEigenerZug,
  ereignisEigenesTischbild,
} from './useEigenerTischfunk'
import { eineAntwort, einePartie } from '../seiten/partieprobe'

import type { ReactNode } from 'react'
import type { PartieAntwort, Spielereignis } from '../apipartie'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function einZug(folge: string, privat: Record<string, unknown>): Spielereignis {
  return { folge, art: 'objekt.gezogen', oeffentlich: {}, privat, kette: '' }
}

it('nimmt während eines Sitzwechsels keinen Zug des alten Sitzes an', async () => {
  probe = laufzeitprobe()
  const { melde } = probe
  let zweite: (antwort: PartieAntwort) => void = () => undefined
  const antworten = [
    Promise.resolve(eineAntwort({ fuer: 'sitz1' })),
    new Promise<PartieAntwort>((los) => {
      zweite = los
    }),
  ]
  belegeApi({
    Partie: () =>
      antworten.shift() ?? Promise.reject(new Error('keine Antwort mehr')),
  })

  const lauf = renderHook(() => usePartie(), { wrapper: raum })
  act(() => {
    lauf.result.current.ansehen('p1', 'sitz1')
  })
  await waitFor(() => {
    expect(lauf.result.current.partie?.fuer).toBe('sitz1')
  })

  // **Der Sitz ist bestätigt** — ein Bild für sitz1, danach trüge ein Zug
  // seine private Ergänzung.
  act(() => {
    melde(ereignisEigenesTischbild, 'p1', {
      partie: einePartie({ fassung: '2' }),
      erlaubt: [],
    })
  })

  act(() => {
    lauf.result.current.ansehen('p1', 'sitz2')
  })
  await waitFor(() => {
    expect(lauf.result.current.wechselt).toBe(true)
  })

  act(() => {
    melde(ereignisEigenerZug, 'p1', einZug('3', { gezogen: 'Blitz' }))
    melde(ereignisEigenesTischbild, 'p1', {
      partie: einePartie({ fassung: '9' }),
      erlaubt: [],
    })
  })

  // **Auch ein Bild des alten Sitzes ändert nichts**, solange gewechselt wird.
  expect(lauf.result.current.partie?.fassung).toBe('2')

  await act(async () => {
    zweite(eineAntwort({ fuer: 'sitz2', fassung: '3' }))
    await Promise.resolve()
  })
  await waitFor(() => {
    expect(lauf.result.current.partie?.fuer).toBe('sitz2')
  })

  expect(lauf.result.current.ereignisse).toEqual([])
})

it('legt von zwei nachgeholten Bildern das neuere nicht unter das ältere', async () => {
  probe = laufzeitprobe()
  const { melde } = probe
  belegeApi({
    Partie: () => Promise.resolve(eineAntwort({ fuer: 'sitz1', fassung: '3' })),
  })

  const lauf = renderHook(() => usePartie(), { wrapper: raum })
  act(() => {
    lauf.result.current.ansehen('p1', 'sitz1')
  })
  await waitFor(() => {
    expect(lauf.result.current.partie?.fassung).toBe('3')
  })

  act(() => {
    melde(ereignisEigenesTischbild, 'p1', {
      partie: einePartie({ fassung: '5' }),
      erlaubt: ['fuenf'],
    })
    melde(ereignisEigenesTischbild, 'p1', {
      partie: einePartie({ fassung: '4' }),
      erlaubt: ['vier'],
    })
  })

  expect(lauf.result.current.partie?.fassung).toBe('5')
  expect(lauf.result.current.erlaubt).toEqual(['fuenf'])
})
