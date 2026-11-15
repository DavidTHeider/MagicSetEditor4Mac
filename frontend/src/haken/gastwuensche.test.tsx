// **Drei schnelle Wünsche eines Gastes, keiner verloren** — P5
// (#2373),
// Befunde C-21 und C-24 am Gastplatz: ein zweiter Wunsch, während der erste
// zum Wirt unterwegs war, prallte an `useVorgang` ab — still. Jetzt warten
// sie der Reihe nach; **gedeutet wird beim Wirt**, der Gast hält keine
// eigene Regelkopie.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { usePartiegast } from './usePartiegast'

import type { ReactNode } from 'react'
import type { PartiebeitrittAntwort } from '../apipartie'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

const stand = (partie: string, laeuft = true): PartiebeitrittAntwort => ({
  laeuft,
  partie,
  gegenueber: 'ab:cd',
  pausiert: false,
  absage: '',
  fehler: null,
})

// amTisch setzt einen Gast an `p1`, dessen Wünsche erst auf `frei` antworten
// — mit einem Stand am Tisch oder, mit `scheitern`, einer Ablehnung.
async function amTisch(scheitern = false) {
  probe = laufzeitprobe()
  const offen: (() => void)[] = []
  const wunsch = vi.fn(
    () =>
      new Promise<PartiebeitrittAntwort>((fertig, abgelehnt) => {
        offen.push(() => {
          if (scheitern) abgelehnt(new Error('der Wirt lehnt ab'))
          else fertig(stand('p1'))
        })
      }),
  )
  const beenden = vi.fn(() => Promise.resolve(stand('', false)))
  belegeApi({
    PartieBeitreten: (_gegenueber: string, partie: string) =>
      Promise.resolve(stand(partie)),
    PartieBeitrittBeenden: beenden,
    PartieWuenschen: wunsch,
  })
  const lauf = renderHook(() => usePartiegast(), { wrapper: raum })
  act(() => {
    lauf.result.current.beitreten('ab:cd', 'p1')
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.partie).toBe('p1')
  })

  const frei = async () => {
    await act(async () => {
      offen.shift()?.()
      await Promise.resolve()
    })
  }

  return { lauf, wunsch, frei, beenden }
}

it('schickt drei schnelle Wünsche nacheinander, keiner geht verloren', async () => {
  const { lauf, wunsch, frei } = await amTisch()

  act(() => {
    lauf.result.current.wuenschen('objekt.tappen', { objekt: 'a' })
    lauf.result.current.wuenschen('objekt.tappen', { objekt: 'b' })
    lauf.result.current.wuenschen('objekt.tappen', { objekt: 'c' })
  })

  expect(wunsch).toHaveBeenCalledTimes(1)
  expect(lauf.result.current.wartend).toBe(3)

  for (const zahl of [2, 3]) {
    await frei()
    await waitFor(() => {
      expect(wunsch).toHaveBeenCalledTimes(zahl)
    })
  }
  await frei()
  await waitFor(() => {
    expect(lauf.result.current.wartend).toBe(0)
  })

  expect(wunsch.mock.calls.map((ruf) => ruf.slice(1))).toEqual([
    ['objekt.tappen', { objekt: 'a' }],
    ['objekt.tappen', { objekt: 'b' }],
    ['objekt.tappen', { objekt: 'c' }],
  ])
  expect(lauf.result.current.fehler).toBeUndefined()
})

// **Wer aufsteht, nimmt seine wartenden Wünsche nicht mit** (§7, Regel 1),
// und ein Wunsch unterwegs hält ihn nicht auf. **Die Antwort darauf gehört
// zur alten Sitzung** (Regel 2): sie setzte ihn sonst an den Tisch zurück.
it('schickt nach dem Aufstehen nichts mehr und bleibt aufgestanden', async () => {
  const { lauf, wunsch, frei, beenden } = await amTisch()

  act(() => {
    lauf.result.current.wuenschen('objekt.tappen', { objekt: 'a' })
    lauf.result.current.wuenschen('objekt.tappen', { objekt: 'b' })
  })
  act(() => {
    lauf.result.current.aufstehen()
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.laeuft).toBe(false)
  })
  expect(beenden).toHaveBeenCalledTimes(1)
  await frei()

  await waitFor(() => {
    expect(lauf.result.current.wartend).toBe(0)
  })
  expect(wunsch).toHaveBeenCalledTimes(1)
  expect(lauf.result.current.stand?.laeuft).toBe(false)
})

it('zeigt die Ablehnung eines Wunsches — nach dem Aufstehen nicht mehr', async () => {
  const { lauf, frei } = await amTisch(true)

  act(() => {
    lauf.result.current.wuenschen('objekt.tappen', { objekt: 'a' })
  })
  await frei()
  await waitFor(() => {
    expect(lauf.result.current.fehler).toBe('der Wirt lehnt ab')
  })

  act(() => {
    lauf.result.current.wuenschen('objekt.tappen', { objekt: 'b' })
  })
  act(() => {
    lauf.result.current.aufstehen()
  })
  await frei()
  await waitFor(() => {
    expect(lauf.result.current.wartend).toBe(0)
  })
  expect(lauf.result.current.fehler).toBeUndefined()
})
