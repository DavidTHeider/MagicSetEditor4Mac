// **Die alte Sicht geht, sobald eine andere gewählt ist** — Befund 1 des
// Nachlaufs über Phase 6.9 (PLAY-05.AK-02,
// #1713).
//
// # Der Befund
//
// `ansehen` räumte die privaten Ereignisse und ließ die Partie stehen, bis
// die neue Sicht ankam. Wer von Adas Sitz zur Zuschauersicht wechselte, sah
// in dieser Zeit weiter Adas Hand und konnte an ihr handeln — und schlug der
// Wechsel fehl, blieb sie ganz stehen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { usePartie } from './usePartie'
import { eineAntwort } from '../seiten/partieprobe'

import type { ReactNode } from 'react'
import type { PartieAntwort } from '../apipartie'

afterEach(() => {
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// offen gibt eine Antwort, die erst kommt, wenn der Lauf es sagt.
function offen() {
  let erfuellen: (antwort: PartieAntwort) => void = () => undefined
  let scheitern: (grund: Error) => void = () => undefined
  const versprechen = new Promise<PartieAntwort>((ja, nein) => {
    erfuellen = ja
    scheitern = nein
  })

  return { versprechen, erfuellen, scheitern }
}

async function geoeffnet() {
  const antworten = [Promise.resolve(eineAntwort({ fuer: 'sitz1' }))]
  const naechste: ReturnType<typeof offen>[] = []

  belegeApi({
    Partie: () => antworten.shift() ?? (naechste[0]?.versprechen as never),
  })

  const lauf = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    lauf.result.current.ansehen('p1', 'sitz1')
  })

  await waitFor(() => {
    expect(lauf.result.current.partie?.fuer).toBe('sitz1')
  })

  return { lauf, naechste }
}

it('blendet die alte Sicht aus, bis die neue da ist', async () => {
  const { lauf, naechste } = await geoeffnet()
  const neue = offen()
  naechste.push(neue)

  expect(lauf.result.current.wechselt).toBe(false)

  act(() => {
    lauf.result.current.ansehen('p1', '')
  })

  expect(lauf.result.current.wechselt).toBe(true)

  await act(async () => {
    neue.erfuellen(eineAntwort({ fuer: '' }))
    await neue.versprechen
  })

  await waitFor(() => {
    expect(lauf.result.current.wechselt).toBe(false)
  })
  expect(lauf.result.current.partie?.fuer).toBe('')
})

it('lässt sie auch nach einem gescheiterten Wechsel weg', async () => {
  const { lauf, naechste } = await geoeffnet()
  const neue = offen()
  naechste.push(neue)

  act(() => {
    lauf.result.current.ansehen('p1', '')
  })

  await act(async () => {
    neue.scheitern(new Error('Leitung weg'))
    await neue.versprechen.catch(() => undefined)
  })

  await waitFor(() => {
    expect(lauf.result.current.fehler).toBeDefined()
  })

  // **Die Hand von vorhin kommt nicht zurück** — wer weiter will, wählt
  // die Sicht noch einmal.
  expect(lauf.result.current.wechselt).toBe(true)
})
