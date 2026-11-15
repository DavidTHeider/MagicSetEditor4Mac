// **Der eigene Tisch erfährt, was ein Gast tut** — Befund aus dem Zweitblick
// über Phase 6.11 (#1672, zu Z1).
//
// # Der Befund
//
// Seit #1627 meldet Go dem Fenster des Wirts `partie:tischbild` und
// `partie:zug`, wenn ein Gast zieht (`MitFenster`). Gehört hat darauf nur der
// Haken des **Gastes** — der eigene Tisch zeigte den Stand von vorhin, bis
// der Wirt selbst etwas tat, und der Gastplatz bekam einen Tisch, an dem
// niemand Gast war.
//
// # Was er hier hält
//
// **Nur die eigene Partie, nur die eigene Sicht**: ein Bild für einen anderen
// Sitz bliebe sonst als fremde Hand im eigenen Fenster stehen — dieselbe
// Sorge wie bei `ansehen` (`sichtwechsel.test.tsx`).
import { act, renderHook, waitFor } from '@testing-library/react'
import { Activity } from 'react'
import { afterEach, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { usePartie } from './usePartie'
import {
  ereignisEigenerZug as ereignisZug,
  ereignisEigenesTischbild as ereignisTischbild,
} from './useEigenerTischfunk'
import { Nachholgrenze } from '../nachholen'
import { eineAntwort, einePartie } from '../seiten/partieprobe'

import type { ReactNode } from 'react'
import type { Spielereignis } from '../apipartie'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function einEreignis(
  folge: string,
  privat: Record<string, unknown> = {},
): Spielereignis {
  return { folge, art: 'objekt.getappt', oeffentlich: {}, privat, kette: '' }
}

// geoeffnet öffnet `p1` aus der Sicht von `sitz1`.
async function geoeffnet() {
  probe = laufzeitprobe()
  belegeApi({ Partie: () => Promise.resolve(eineAntwort({ fuer: 'sitz1' })) })

  const lauf = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    lauf.result.current.ansehen('p1', 'sitz1')
  })

  await waitFor(() => {
    expect(lauf.result.current.partie?.fuer).toBe('sitz1')
  })

  return { lauf, melde: probe.melde }
}

it('nimmt das Tischbild der eigenen Partie für den eigenen Sitz', async () => {
  const { lauf, melde } = await geoeffnet()

  act(() => {
    melde(ereignisTischbild, 'p1', {
      partie: einePartie({ fassung: '5', fuer: 'sitz1' }),
      erlaubt: ['objekt.ziehen'],
      handlungen: [
        {
          objekt: 'sitz1/k1/0',
          handlungen: ['objekt.ziehen'],
          primaer: '',
          zweit: '',
        },
      ],
    })
  })

  expect(lauf.result.current.partie?.fassung).toBe('5')
  expect(lauf.result.current.erlaubt).toEqual(['objekt.ziehen'])
  // **Und die Auskunft je Objekt aus demselben Bild** (P3, #2371).
  expect(lauf.result.current.handlungen).toEqual([
    {
      objekt: 'sitz1/k1/0',
      handlungen: ['objekt.ziehen'],
      primaer: '',
      zweit: '',
    },
  ])
})

it('lässt ein Bild einer anderen Partie oder eines anderen Sitzes liegen', async () => {
  const { lauf, melde } = await geoeffnet()

  act(() => {
    melde(ereignisTischbild, 'p2', {
      partie: einePartie({ id: 'p2', fassung: '7' }),
      erlaubt: [],
    })
    melde(ereignisTischbild, 'p1', {
      partie: einePartie({ fassung: '8', fuer: 'sitz2' }),
      erlaubt: [],
    })
  })

  expect(lauf.result.current.partie?.fassung).toBe('1')

  // **Und es hört jemand zu** (§3): dasselbe Ereignis für den eigenen Sitz
  // wirkt.
  act(() => {
    melde(ereignisTischbild, 'p1', {
      partie: einePartie({ fassung: '9' }),
      erlaubt: [],
    })
  })

  expect(lauf.result.current.partie?.fassung).toBe('9')
})

it('hängt einen Zug der eigenen Partie an, ohne fremde Ergänzung vor dem ersten Bild', async () => {
  const { lauf, melde } = await geoeffnet()

  act(() => {
    melde(ereignisZug, 'p2', einEreignis('1'))
    melde(ereignisZug, 'p1', einEreignis('2', { gezogen: 'Blitz' }))
  })

  // **Vor dem ersten Bild ist der Sitz der Meldung ungewiss** — die Ergänzung
  // gehört womöglich einem anderen, also bleibt sie weg.
  expect(lauf.result.current.ereignisse).toEqual([einEreignis('2')])

  act(() => {
    melde(ereignisTischbild, 'p1', {
      partie: einePartie({ fassung: '3' }),
      erlaubt: [],
    })
    melde(ereignisZug, 'p1', einEreignis('3', { gezogen: 'Insel' }))
  })

  expect(lauf.result.current.ereignisse.at(-1)).toEqual(
    einEreignis('3', { gezogen: 'Insel' }),
  )
})

it('hört nicht mehr, wenn die Partie weggeräumt ist', async () => {
  const { lauf, melde } = await geoeffnet()

  act(() => {
    lauf.result.current.raeumen()
  })
  act(() => {
    melde(ereignisTischbild, 'p1', {
      partie: einePartie({ fassung: '4' }),
      erlaubt: [],
    })
    melde(ereignisZug, 'p1', einEreignis('4'))
  })

  expect(lauf.result.current.partie).toBeUndefined()
  expect(lauf.result.current.ereignisse).toEqual([])
})

// **Kein älteres Bild über ein neueres** — wer den Tisch wieder öffnet,
// bekommt nachgeholt, was dazwischen kam, und das ist älter als die Sicht,
// die `ansehen` gerade geholt hat.
it('legt ein älteres Bild nicht über die geholte Sicht', async () => {
  const { lauf, melde } = await geoeffnet()

  act(() => {
    melde(ereignisTischbild, 'p1', {
      partie: einePartie({ fassung: '0' }),
      erlaubt: [],
    })
  })

  expect(lauf.result.current.partie?.fassung).toBe('1')
  expect(lauf.result.current.erlaubt).toEqual([
    'objekt.tappen',
    'objekt.schieben',
  ])
})

// **Und eine Lücke über der Nachholgrenze wird genannt** — Z3 aus dem
// Zweitblick über Phase 6.11. Verborgen steht der Tisch in einem `<Activity>`.
it('meldet eine Lücke, wenn verborgen mehr kam, als nachgeholt wird', async () => {
  probe = laufzeitprobe()
  const { melde } = probe
  belegeApi({ Partie: () => Promise.resolve(eineAntwort({ fuer: 'sitz1' })) })

  let sichtbar = true
  function verborgen({ children }: { children: ReactNode }) {
    return (
      <Sprachraum sprache="de">
        <Activity mode={sichtbar ? 'visible' : 'hidden'}>{children}</Activity>
      </Sprachraum>
    )
  }

  const lauf = renderHook(() => usePartie(), { wrapper: verborgen })
  act(() => {
    lauf.result.current.ansehen('p1', 'sitz1')
  })
  await waitFor(() => {
    expect(lauf.result.current.partie?.fuer).toBe('sitz1')
  })
  expect(lauf.result.current.ereignisluecke).toBe(false)

  sichtbar = false
  lauf.rerender()
  act(() => {
    for (let nr = 0; nr <= Nachholgrenze; nr++) {
      melde(ereignisZug, 'p1', einEreignis(String(nr + 10)))
    }
  })
  sichtbar = true
  lauf.rerender()

  await waitFor(() => {
    expect(lauf.result.current.ereignisluecke).toBe(true)
  })
  expect(lauf.result.current.ereignisse).toHaveLength(Nachholgrenze)
})
