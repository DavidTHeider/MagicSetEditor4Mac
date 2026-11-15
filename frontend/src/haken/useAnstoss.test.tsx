// **Der Anstoß für ein anderes Objekt** — AP-5
// (#2362):
// ein Rechtsklick oder ⇧F10 auf einer Zeile, die nicht gewählt ist, wählt sie
// zuerst; das Menü gehört dann ihr. Der Haken hält die Stelle, bis sie
// gewählt ist, und lässt sie verfallen, wenn stattdessen eine dritte kommt.
import { act, renderHook } from '@testing-library/react'
import type { KeyboardEvent, MouseEvent } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { anstossstelle, useAnstoss } from './useAnstoss'

function rechtsklick(x: number, y: number) {
  return {
    type: 'contextmenu',
    clientX: x,
    clientY: y,
    target: document.body,
    currentTarget: document.body,
  } as unknown as MouseEvent<HTMLElement>
}

function taste(key: string, shiftKey: boolean) {
  const ziel = document.createElement('div')
  ziel.getBoundingClientRect = () =>
    ({ left: 7, bottom: 30 }) as unknown as DOMRect

  return {
    type: 'keydown',
    key,
    shiftKey,
    currentTarget: ziel,
    target: ziel,
  } as unknown as KeyboardEvent<HTMLElement>
}

describe('anstossstelle', () => {
  it('liest den Zeiger beim Rechtsklick', () => {
    expect(anstossstelle(rechtsklick(3, 4))).toEqual({ x: 3, y: 4 })
  })

  it('setzt ⇧F10 und die Menütaste unter die Zeile', () => {
    expect(anstossstelle(taste('F10', true))).toEqual({ x: 7, y: 30 })
    expect(anstossstelle(taste('ContextMenu', false))).toEqual({ x: 7, y: 30 })
  })

  it('meint keine andere Taste und keinen Rechtsklick in einem Feld', () => {
    expect(anstossstelle(taste('F10', false))).toBeUndefined()
    expect(anstossstelle(taste('a', true))).toBeUndefined()

    const feld = document.createElement('input')
    expect(
      anstossstelle({
        type: 'contextmenu',
        clientX: 1,
        clientY: 1,
        target: feld,
      } as unknown as MouseEvent<HTMLElement>),
    ).toBeUndefined()
  })
})

describe('useAnstoss', () => {
  function zeige(gewaehlt: string | undefined) {
    const waehlen = vi.fn()
    const haken = renderHook(
      ({ gewaehlt }: { gewaehlt: string | undefined }) =>
        useAnstoss<string>(gewaehlt, waehlen),
      { initialProps: { gewaehlt } },
    )

    return { waehlen, ...haken }
  }

  it('öffnet am gewählten Objekt sofort, ohne neu zu wählen', () => {
    const { result, waehlen } = zeige('1')
    const ereignis = {
      ...rechtsklick(1, 2),
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
    }

    act(() => {
      result.current.anstossen(ereignis, '1', '1')
    })

    expect(ereignis.preventDefault).toHaveBeenCalled()
    expect(waehlen).not.toHaveBeenCalled()
    expect(result.current.anstoss).toMatchObject({ x: 1, y: 2 })
  })

  it('wählt ein anderes und gibt die Stelle erst, wenn es gewählt ist', () => {
    const { result, waehlen, rerender } = zeige('1')
    const ereignis = {
      ...rechtsklick(5, 6),
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
    }

    act(() => {
      result.current.anstossen(ereignis, '2', '2')
    })
    expect(ereignis.preventDefault).toHaveBeenCalled()
    expect(ereignis.stopPropagation).toHaveBeenCalled()
    expect(waehlen).toHaveBeenCalledWith('2')
    expect(result.current.anstoss).toBeUndefined()

    rerender({ gewaehlt: '2' })
    expect(result.current.anstoss).toMatchObject({ x: 5, y: 6 })

    act(() => {
      result.current.verbraucht()
    })
    expect(result.current.anstoss).toBeUndefined()
  })

  it('verfällt, wenn stattdessen ein drittes gewählt wird', () => {
    const { result, rerender } = zeige('1')
    act(() => {
      result.current.anstossen(
        {
          ...rechtsklick(5, 6),
          preventDefault: vi.fn(),
          stopPropagation: vi.fn(),
        },
        '2',
        '2',
      )
    })

    rerender({ gewaehlt: '3' })
    rerender({ gewaehlt: '2' })

    expect(result.current.anstoss).toBeUndefined()
  })

  it('übergeht ein Ereignis, das kein Menü meint', () => {
    const { result, waehlen } = zeige('1')

    act(() => {
      result.current.anstossen(taste('a', false), '2', '2')
    })

    expect(waehlen).not.toHaveBeenCalled()
  })

  // **Eine weitere Eingabe hebt ihn auf** — hält die Seite die Wahl an (die
  // Werkstatt fragt bei Ungesichertem nach), ginge das Menü sonst nach der
  // Antwort an einer Stelle auf, an der niemand mehr geklickt hat.
  it('verfällt mit der nächsten Eingabe, bevor gewählt ist', () => {
    const { result, rerender } = zeige('1')
    act(() => {
      result.current.anstossen(
        {
          ...rechtsklick(5, 6),
          preventDefault: vi.fn(),
          stopPropagation: vi.fn(),
        },
        '2',
        '2',
      )
    })

    act(() => {
      document.body.dispatchEvent(new Event('pointerdown', { bubbles: true }))
    })
    rerender({ gewaehlt: '2' })

    expect(result.current.anstoss).toBeUndefined()
  })

  // **Nichts gewählt ist kein drittes Objekt** — gemessen am nativen Fenster
  // mit `ap6decks09`: ein anderes Deck räumt zuerst den offenen Kontext
  // (`useDecks`, §7 Regel 1), und der Anstoß verfiel im Übergang.
  it('übersteht den Übergang über »nichts gewählt«', () => {
    const { result, rerender } = zeige('1')
    act(() => {
      result.current.anstossen(
        {
          ...rechtsklick(5, 6),
          preventDefault: vi.fn(),
          stopPropagation: vi.fn(),
        },
        '2',
        '2',
      )
    })

    rerender({ gewaehlt: undefined })
    rerender({ gewaehlt: '2' })

    expect(result.current.anstoss).toMatchObject({ x: 5, y: 6 })
  })

  // **Der Anstoß merkt sich seine Zeile** — Befund 5: geht das Menü erst auf,
  // wenn die Zeile schon woanders steht, folgt es ihr.
  it('merkt die Zeile und die Stelle des Zeigers in ihr', () => {
    const { result, rerender } = zeige('1')
    const zeile = document.createElement('div')
    zeile.getBoundingClientRect = () => new DOMRect(100, 200, 80, 30)

    act(() => {
      result.current.anstossen(
        {
          ...rechtsklick(110, 215),
          currentTarget: zeile,
          preventDefault: vi.fn(),
          stopPropagation: vi.fn(),
        },
        '2',
        '2',
      )
    })
    rerender({ gewaehlt: '2' })

    expect(result.current.anstoss).toMatchObject({
      anker: zeile,
      versatz: { x: 10, y: 15 },
    })
  })
})
