// **Zurück und Vorwärts zwischen den Seiten** — UX-03.AK-02
// (#1675).
//
// # Der Befund
//
// **Es gab keinen Gang.** Eine Seite war ein bloßer Zustand; wer von den
// Decks über ⌘K in die Sammlung ging, kam nur über denselben Weg zurück. Und
// ein Kürzel baute die Seite ab, ohne das offene Feld zu verlassen — ein
// Kartenfeld schreibt erst beim Verlassen (#1290).
import { act, render, renderHook, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import {
  gehe,
  hoechstensGemerkt,
  ohneSetseiten,
  schreite,
  useSeitengang,
  type Gangslage,
  type Seitengang,
} from './useSeitengang'

import type { Seitenschluessel } from '../navigation'

// lage schreibt einen Gang als Liste mit einer Stelle — so liest er sich.
function lage(
  seiten: Seitenschluessel[],
  stelle = seiten.length - 1,
): Gangslage {
  return {
    zurueck: seiten.slice(0, stelle),
    aktuell: seiten[stelle] ?? 'start',
    vor: seiten.slice(stelle + 1),
  }
}

describe('gehe', () => {
  it('hängt die Seite an', () => {
    expect(gehe(lage(['start']), 'decks')).toEqual(lage(['start', 'decks']))
  })

  it('bleibt, wenn es dieselbe Seite ist', () => {
    const vorher = lage(['start', 'decks'])

    expect(gehe(vorher, 'decks')).toBe(vorher)
  })

  // **Wie ein Browser**: wer zurückgeht und dann woandershin, verwirft den
  // Weg vor sich.
  it('verwirft, was vor der Stelle lag', () => {
    expect(gehe(lage(['start', 'decks', 'sammlung'], 0), 'handeln')).toEqual(
      lage(['start', 'handeln']),
    )
  })

  it('merkt sich höchstens fünfzig Seiten', () => {
    let gang = lage(['start'])
    for (let runde = 0; runde < hoechstensGemerkt; runde += 1) {
      gang = gehe(gang, runde % 2 === 0 ? 'decks' : 'sammlung')
    }

    expect([...gang.zurueck, gang.aktuell]).toHaveLength(hoechstensGemerkt)
    expect(gang.zurueck[0]).toBe('decks')
  })
})

describe('schreite', () => {
  it('geht zurück und wieder vor', () => {
    const zurueck = schreite(lage(['start', 'decks']), -1)

    expect(zurueck).toEqual(lage(['start', 'decks'], 0))
    expect(schreite(zurueck, 1)).toEqual(lage(['start', 'decks']))
  })

  it('bleibt an beiden Rändern stehen', () => {
    const anfang = lage(['start', 'decks'], 0)
    const ende = lage(['start', 'decks'])

    expect(schreite(anfang, -1)).toBe(anfang)
    expect(schreite(ende, 1)).toBe(ende)
  })
})

describe('ohneSetseiten', () => {
  // **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (§7,
  // Regel 1): *Karten* hieß vorhin „die Karten von Set A".
  it('räumt die Seiten des vorigen Sets, bis auf die aktuelle', () => {
    expect(
      ohneSetseiten(lage(['start', 'karten', 'stil', 'decks', 'karten'])),
    ).toEqual(lage(['start', 'decks', 'karten']))
  })

  it('hält die Stelle, wenn sie mitten im Gang steht', () => {
    expect(
      ohneSetseiten(lage(['start', 'karten', 'decks', 'stil', 'sammlung'], 2)),
    ).toEqual(lage(['start', 'decks', 'sammlung'], 1))
  })

  // **Zwei gleiche Nachbarn fallen zusammen** — sonst führte ein Schritt
  // auf die Seite, auf der man schon steht.
  it('legt gleiche Nachbarn zusammen', () => {
    expect(
      ohneSetseiten(lage(['start', 'decks', 'karten', 'decks'], 1)),
    ).toEqual(lage(['start', 'decks'], 1))
    expect(ohneSetseiten(lage(['decks', 'karten', 'decks']))).toEqual(
      lage(['decks']),
    )
  })
})

describe('useSeitengang', () => {
  it('beginnt am Start und geht zurück und vor', () => {
    const { result } = renderHook(() => useSeitengang(undefined))

    expect(result.current.aktuell).toBe('start')

    act(() => {
      result.current.wechseln('decks')
    })
    act(() => {
      result.current.wechseln('sammlung')
    })
    act(() => {
      result.current.zurueck()
    })

    expect(result.current.aktuell).toBe('decks')

    act(() => {
      result.current.vor()
    })

    expect(result.current.aktuell).toBe('sammlung')
  })

  it('vergisst die Setseiten, wenn das Set wechselt', () => {
    const anfang: { setID: string | undefined } = { setID: 'a' }
    const { result, rerender } = renderHook(
      ({ setID }: { setID: string | undefined }) => useSeitengang(setID),
      { initialProps: anfang },
    )

    act(() => {
      result.current.wechseln('karten')
    })
    act(() => {
      result.current.wechseln('decks')
    })
    expect(result.current.aktuell).toBe('decks')

    rerender({ setID: 'b' })
    act(() => {
      result.current.zurueck()
    })

    expect(result.current.aktuell).toBe('start')
  })

  // **Ein Schritt verlässt das offene Feld zuerst** (#1290): ein Kürzel
  // wechselt die Seite, und ein programmgesteuerter Wechsel erzeugt kein
  // `blur`.
  it.each([
    [
      'wechseln',
      (gang: Seitengang) => {
        gang.wechseln('sammlung')
      },
    ],
    [
      'zurueck',
      (gang: Seitengang) => {
        gang.zurueck()
      },
    ],
    [
      'vor',
      (gang: Seitengang) => {
        gang.vor()
      },
    ],
  ])('verlässt beim %s das offene Feld', async (_, schritt) => {
    const verlassen = vi.fn()
    const { result } = renderHook(() => useSeitengang(undefined))
    // **Mitten im Gang**: hinter den Decks der Start, davor die Sammlung.
    act(() => {
      result.current.wechseln('decks')
    })
    act(() => {
      result.current.wechseln('sammlung')
    })
    act(() => {
      result.current.zurueck()
    })
    render(<input aria-label="Name" onBlur={verlassen} />)

    await userEvent.click(screen.getByLabelText('Name'))
    act(() => {
      schritt(result.current)
    })

    expect(verlassen).toHaveBeenCalledTimes(1)
  })

  // **Und nur dann.** Eine Rechnung kommt aus der Anwendung — die Kartenliste
  // meldet ihre Wahl, während jemand in einer Schublade tippt —, und ein
  // Schritt ins Leere ist keiner.
  it.each([
    [
      'dieselbe Seite',
      (gang: Seitengang) => {
        gang.wechseln('start')
      },
    ],
    [
      'eine Rechnung',
      (gang: Seitengang) => {
        gang.wechseln(() => 'decks')
      },
    ],
    [
      'zurück am Anfang',
      (gang: Seitengang) => {
        gang.zurueck()
      },
    ],
    [
      'vor am Ende',
      (gang: Seitengang) => {
        gang.vor()
      },
    ],
  ])('lässt das Feld stehen bei %s', async (_, schritt) => {
    const verlassen = vi.fn()
    const { result } = renderHook(() => useSeitengang(undefined))
    render(<input aria-label="Name" onBlur={verlassen} />)

    await userEvent.click(screen.getByLabelText('Name'))
    act(() => {
      schritt(result.current)
    })

    expect(verlassen).not.toHaveBeenCalled()
  })

  it('rechnet über die aktuelle Seite', () => {
    const { result } = renderHook(() => useSeitengang(undefined))

    act(() => {
      result.current.wechseln((jetzt) => (jetzt === 'start' ? 'decks' : jetzt))
    })

    expect(result.current.aktuell).toBe('decks')
  })
})
