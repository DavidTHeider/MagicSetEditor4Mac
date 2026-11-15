// **Der Vorhang am gemeinsamen Gerät** — P7
// (#2375),
// Befund C-05, Artboard `Vorhang09`.
//
// Entscheidet ein anderer Sitz an **diesem** Gerät, geht der Vorhang zu —
// statt der Sichtwahl von Hand (bis P5) oder einer Sicht, die ohne Nachfrage
// folgt und dabei die nächste Hand aufdeckt (P6). Die Sicht wechselt erst,
// wenn der neue Sitz bestätigt; Esc lässt den bisherigen weiterhandeln.
// **Nie für den Zuschauer, nie zu einem Gast**, und ohne Vorhang nur, wo
// „Offene Hände“ ausdrücklich gewählt ist.
import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { Partie } from '../apipartie'
import { einePartie, einSitz, einZug } from '../seiten/partieprobe'
import { useVorhang } from './useVorhang'

function beginn(fuer: string, eins: string, zwei: string): Partie {
  return einePartie({
    fuer,
    sitze: [
      einSitz({ kennung: 'sitz1', name: 'Anna', eroeffnung: eins as never }),
      einSitz({ kennung: 'sitz2', name: 'Bert', eroeffnung: zwei as never }),
    ],
  })
}

function prioritaet(fuer: string, wer: string): Partie {
  return einePartie({
    fuer,
    zug: einZug({ nummer: 2, aktiv: 'sitz1', prioritaet: wer }),
  })
}

interface Lage {
  partie: Partie
  gastsitze: readonly string[] | undefined
  offen: boolean
}

// **Ohne Vorgabewerte** — ein `undefined` darf nicht still „keine Gäste“
// heißen (dieselbe Falle wie in P6).
function zeige(anfang: Lage) {
  const ansehen = vi.fn()
  const haken = renderHook((lage: Lage) => useVorhang({ ...lage, ansehen }), {
    initialProps: anfang,
  })

  return {
    ansehen,
    haken,
    neu: (teile: Partial<Lage>) => {
      haken.rerender({ ...anfang, ...teile })
    },
  }
}

const amGeraet = { gastsitze: [] as readonly string[], offen: false }

describe('der Vorhang', () => {
  it('geht zu, wenn nach dem Behalten der nächste Sitz entscheidet', () => {
    const { ansehen, haken, neu } = zeige({
      partie: beginn('sitz1', 'offen', 'offen'),
      ...amGeraet,
    })

    expect(haken.result.current.vorhang).toBeUndefined()

    neu({ partie: beginn('sitz1', 'behalten', 'offen') })

    expect(haken.result.current.vorhang).toBe('sitz2')
    expect(ansehen).not.toHaveBeenCalled()
  })

  it('wechselt die Sicht erst, wenn der neue Sitz bestätigt', () => {
    const { ansehen, haken, neu } = zeige({
      partie: beginn('sitz1', 'offen', 'offen'),
      ...amGeraet,
    })
    neu({ partie: beginn('sitz1', 'behalten', 'offen') })

    act(() => {
      haken.result.current.bestaetigen()
    })

    expect(ansehen).toHaveBeenCalledExactlyOnceWith('p1', 'sitz2')
    expect(haken.result.current.vorhang).toBeUndefined()
  })

  it('lässt mit „bleiben“ den bisherigen Sitz weiterhandeln', () => {
    const { ansehen, haken, neu } = zeige({
      partie: beginn('sitz1', 'offen', 'offen'),
      ...amGeraet,
    })
    neu({ partie: beginn('sitz1', 'behalten', 'offen') })

    act(() => {
      haken.result.current.bleiben()
    })
    neu({ partie: { ...beginn('sitz1', 'behalten', 'offen'), fassung: '2' } })

    expect(haken.result.current.vorhang).toBeUndefined()
    expect(ansehen).not.toHaveBeenCalled()
  })

  it('geht zu, wenn die Priorität an einen anderen Sitz am Gerät geht', () => {
    const { haken, neu } = zeige({
      partie: prioritaet('sitz1', 'sitz1'),
      ...amGeraet,
    })

    neu({ partie: prioritaet('sitz1', 'sitz2') })

    expect(haken.result.current.vorhang).toBe('sitz2')
  })

  // **Der Zuschauer bleibt Zuschauer** — keine Sitzübernahme, kein Vorhang.
  it('bleibt für den Zuschauer offen und übernimmt keinen Sitz', () => {
    const { ansehen, haken, neu } = zeige({
      partie: prioritaet('', 'sitz1'),
      ...amGeraet,
      offen: true,
    })

    neu({ partie: prioritaet('', 'sitz2') })

    expect(haken.result.current.vorhang).toBeUndefined()
    expect(ansehen).not.toHaveBeenCalled()
  })

  it('geht nicht zu für einen Sitz, den ein Gast spielt', () => {
    const { haken, neu } = zeige({
      partie: prioritaet('sitz1', 'sitz1'),
      gastsitze: ['sitz2'],
      offen: false,
    })

    neu({ partie: prioritaet('sitz1', 'sitz2') })

    expect(haken.result.current.vorhang).toBeUndefined()
  })

  it('wartet, bis bekannt ist, wer Gast ist', () => {
    const { haken, neu } = zeige({
      partie: prioritaet('sitz1', 'sitz1'),
      gastsitze: undefined,
      offen: false,
    })

    neu({ partie: prioritaet('sitz1', 'sitz2') })
    expect(haken.result.current.vorhang).toBeUndefined()

    neu({ partie: prioritaet('sitz1', 'sitz2'), gastsitze: [] })
    expect(haken.result.current.vorhang).toBe('sitz2')
  })

  it('folgt mit „Offenen Händen“ ohne Vorhang', () => {
    const { ansehen, haken, neu } = zeige({
      partie: prioritaet('sitz1', 'sitz1'),
      gastsitze: [],
      offen: true,
    })

    neu({ partie: prioritaet('sitz1', 'sitz2') })

    expect(haken.result.current.vorhang).toBeUndefined()
    expect(ansehen).toHaveBeenCalledExactlyOnceWith('p1', 'sitz2')
  })

  it('geht wieder auf, wenn die Frage zum gezeigten Sitz zurückkommt', () => {
    const { haken, neu } = zeige({
      partie: prioritaet('sitz1', 'sitz1'),
      ...amGeraet,
    })

    neu({ partie: prioritaet('sitz1', 'sitz2') })
    neu({ partie: prioritaet('sitz1', 'sitz1') })

    expect(haken.result.current.vorhang).toBeUndefined()
  })

  // **Eine Wahl von Hand ist keine Übergabe** — wer in der Sichtwahl einen
  // Sitz nimmt, hat entschieden, wessen Sicht er sehen will.
  it('geht nicht zu, wenn die Sicht von Hand wechselt', () => {
    const { haken, neu } = zeige({
      partie: prioritaet('sitz1', 'sitz1'),
      ...amGeraet,
    })

    neu({ partie: prioritaet('sitz2', 'sitz1') })

    expect(haken.result.current.vorhang).toBeUndefined()
  })

  it('geht aus der Sicht eines Gastsitzes nicht zu', () => {
    const { haken, neu } = zeige({
      partie: prioritaet('sitz2', 'sitz2'),
      gastsitze: ['sitz2'],
      offen: false,
    })

    neu({ partie: prioritaet('sitz2', 'sitz1') })

    expect(haken.result.current.vorhang).toBeUndefined()
  })
})

// **Einmal je Übergang** — `ansehen` ist ein Rückruf des Laufs und darf neu
// gebaut werden, ohne dass die Sicht ein zweites Mal geladen wird.
it('folgt mit „Offenen Händen“ nur einmal, auch mit neuem Rückruf', () => {
  const erster = vi.fn()
  const zweiter = vi.fn()
  const haken = renderHook(
    ({ partie, ansehen }: { partie: Partie; ansehen: typeof erster }) =>
      useVorhang({ partie, ansehen, gastsitze: [], offen: true }),
    { initialProps: { partie: prioritaet('sitz1', 'sitz1'), ansehen: erster } },
  )

  haken.rerender({ partie: prioritaet('sitz1', 'sitz2'), ansehen: erster })
  haken.rerender({ partie: prioritaet('sitz1', 'sitz2'), ansehen: zweiter })

  expect(erster).toHaveBeenCalledExactlyOnceWith('p1', 'sitz2')
  expect(zweiter).not.toHaveBeenCalled()
})

// **Ohne Vorhang bestätigt niemand** — ein verspäteter Druck auf „Ich bin …“
// wechselt keine Sicht.
it('bestätigt ohne Vorhang nichts', () => {
  const { ansehen, haken } = zeige({
    partie: prioritaet('sitz1', 'sitz1'),
    ...amGeraet,
  })

  act(() => {
    haken.result.current.bestaetigen()
  })

  expect(ansehen).not.toHaveBeenCalled()
})
