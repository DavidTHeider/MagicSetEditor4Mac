// **Ein Entwurf und ein fremder Wert im selben Feld** — U01 bis U06
// (#2119).
//
// Der Vertrag, den jedes Formular mit Entwurf teilt: Karteneditor,
// Setangaben, Notiz, Marke, Schlüsselworteditor. Was er an der Seite heißt,
// prüfen deren eigene Läufe; hier steht die Regel allein.
import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { useEntwurf } from './useEntwurf'

function feld(wert: string) {
  const schreiben = vi.fn<(neu: string, bekannt: readonly string[]) => void>()
  const getippt = vi.fn<(neu: string | undefined) => void>()
  const griff = renderHook(
    ({ wert: gespeichert }: { wert: string }) =>
      useEntwurf(gespeichert, schreiben, getippt),
    { initialProps: { wert } },
  )

  // anfang ist der Schlüssel der ersten Eingabe — ein anderer heißt Neuaufbau.
  return {
    ...griff,
    schreiben,
    getippt,
    anfang: griff.result.current.schluessel,
  }
}

describe('useEntwurf', () => {
  it('hält den Schlüssel, solange getippt wird, und schreibt beim Verlassen', () => {
    const { result, rerender, schreiben, getippt, anfang } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    expect(result.current.schluessel).toBe(anfang)
    expect(getippt).toHaveBeenLastCalledWith('Feuerball groß')

    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    expect(schreiben).toHaveBeenCalledWith('Feuerball groß', expect.any(Array))

    // **Der Bestand zeigt ihn** — erst dann ist der Entwurf erledigt.
    rerender({ wert: 'Feuerball groß' })
    expect(result.current.schluessel).toBe(anfang)
    expect(result.current.konflikt).toBe(false)
  })

  it('sieht einen fremden Wert im selben Feld und schreibt beim Verlassen nicht', () => {
    const { result, rerender, schreiben, anfang } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    rerender({ wert: 'Eisball' })

    expect(result.current.konflikt).toBe(true)
    expect(result.current.schluessel).toBe(anfang)

    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    expect(schreiben).not.toHaveBeenCalled()
  })

  it('schreibt auf Wunsch den Entwurf über den fremden Wert', () => {
    const { result, rerender, schreiben } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    rerender({ wert: 'Eisball' })

    act(() => {
      result.current.behalten()
    })

    expect(schreiben).toHaveBeenCalledWith('Feuerball groß', expect.any(Array))
    expect(result.current.konflikt).toBe(false)
  })

  it('übernimmt auf Wunsch den fremden Wert', () => {
    const { result, rerender, schreiben, getippt, anfang } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    rerender({ wert: 'Eisball' })

    act(() => {
      result.current.verwerfen()
    })

    expect(result.current.konflikt).toBe(false)
    expect(result.current.schluessel).not.toBe(anfang)
    expect(getippt).toHaveBeenLastCalledWith(undefined)
    expect(schreiben).not.toHaveBeenCalled()
  })

  // **Ein Schreibfehler schreibt keinen veralteten Wert zurück**: der
  // gescheiterte Entwurf bleibt einer, und ändert danach ein anderer das
  // Feld, ist das wieder ein Konflikt — nicht ein stilles Überschreiben beim
  // nächsten Verlassen.
  it('behält nach einem gescheiterten Schreiben den Entwurf samt seinem Ausgangswert', () => {
    const { result, rerender, schreiben, anfang } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    expect(schreiben).toHaveBeenCalledTimes(1)

    // Gescheitert: der Bestand bleibt, wie er war.
    rerender({ wert: 'Feuerball' })
    expect(result.current.schluessel).toBe(anfang)
    expect(result.current.konflikt).toBe(false)

    rerender({ wert: 'Eisball' })
    expect(result.current.konflikt).toBe(true)

    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    expect(schreiben).toHaveBeenCalledTimes(1)
  })

  it('sieht keinen Konflikt, wenn der fremde Wert der Entwurf ist, und behält den Schlüssel', () => {
    const { result, rerender, schreiben, anfang } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    rerender({ wert: 'Feuerball groß' })

    expect(result.current.konflikt).toBe(false)
    expect(result.current.schluessel).toBe(anfang)

    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    expect(schreiben).not.toHaveBeenCalled()
  })

  it('vergisst einen geschriebenen Entwurf, sodass eine spätere fremde Änderung keinen Konflikt macht', () => {
    const { result, rerender, anfang } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    rerender({ wert: 'Feuerball groß' })
    rerender({ wert: 'Eisball' })

    expect(result.current.konflikt).toBe(false)
    expect(result.current.schluessel).not.toBe(anfang)
  })

  // **Das eigene erste Schreiben ist kein fremder Wert** — wer ein Feld
  // verlässt, weitertippt und es wieder verlässt, bevor der Bestand den
  // ersten Text zeigt, sieht danach diesen ersten Text dort ankommen.
  it('sieht das eigene frühere Schreiben nicht als Konflikt', () => {
    const { result, rerender, schreiben, anfang } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    act(() => {
      result.current.aendern('Feuerball größer')
    })

    rerender({ wert: 'Feuerball groß' })
    expect(result.current.konflikt).toBe(false)
    expect(result.current.schluessel).toBe(anfang)

    act(() => {
      result.current.verlassen('Feuerball größer')
    })
    expect(schreiben).toHaveBeenLastCalledWith(
      'Feuerball größer',
      expect.any(Array),
    )
  })

  // **Zurück zum Ausgangswert ist eine Änderung, solange ein Schreiben
  // unterwegs ist** — der Bestand bekommt gleich den ersten Text, und ein
  // Verlassen ohne Schreiben ließe ihn dort stehen.
  it('schreibt den Ausgangswert, wenn ein früheres Schreiben noch aussteht', () => {
    const { result, schreiben } = feld('Feuerball')

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    act(() => {
      result.current.verlassen('Feuerball groß')
    })
    act(() => {
      result.current.aendern('Feuerball')
    })
    act(() => {
      result.current.verlassen('Feuerball')
    })

    expect(schreiben).toHaveBeenLastCalledWith('Feuerball', expect.any(Array))
    expect(schreiben).toHaveBeenCalledTimes(2)
  })

  // **Ein Einschub aus einem Menü ist Tippen und Verlassen in einem** — der
  // Schlüsselworteditor setzt einen Parameter an die Textmarke und schreibt
  // sofort. Was er schreibt, ist danach kein fremder Wert.
  it('übernimmt einen Einschub als geschickten Entwurf', () => {
    const { result, rerender, schreiben, anfang } = feld('Beben')

    act(() => {
      result.current.uebernehmen('Beben <atom-param>')
    })
    expect(schreiben).toHaveBeenCalledWith(
      'Beben <atom-param>',
      expect.any(Array),
    )
    expect(result.current.schluessel).toBe(anfang)

    rerender({ wert: 'Beben <atom-param>' })
    expect(result.current.konflikt).toBe(false)
    expect(result.current.schluessel).toBe(anfang)
  })

  it('schreibt einen Einschub im Konflikt nicht, sondern hält ihn als Entwurf', () => {
    const { result, rerender, schreiben } = feld('Beben')

    act(() => {
      result.current.aendern('Beben groß')
    })
    rerender({ wert: 'Erdbeben' })

    act(() => {
      result.current.uebernehmen('Beben groß <atom-param>')
    })

    expect(schreiben).not.toHaveBeenCalled()
    expect(result.current.konflikt).toBe(true)

    act(() => {
      result.current.behalten()
    })
    expect(schreiben).toHaveBeenCalledWith(
      'Beben groß <atom-param>',
      expect.any(Array),
    )
  })

  it('schreibt einen unveränderten Wert nicht', () => {
    const { result, schreiben, anfang } = feld('Feuerball')

    act(() => {
      result.current.verlassen('Feuerball')
    })

    expect(schreiben).not.toHaveBeenCalled()
    expect(result.current.schluessel).toBe(anfang)
  })

  it('kommt ohne Vorschau aus', () => {
    const schreiben = vi.fn<(neu: string, bekannt: readonly string[]) => void>()
    const { result } = renderHook(() => useEntwurf('Feuerball', schreiben))
    const anfang = result.current.schluessel

    act(() => {
      result.current.aendern('Feuerball groß')
    })
    act(() => {
      result.current.verwerfen()
    })

    expect(result.current.schluessel).not.toBe(anfang)
  })
})
