// **Ein Entwurf schickt mit, welche Werte er kennt** — U02b
// (#2179).
//
// `useEntwurf` erkannte einen fremden Wert nur, wenn das Fenster ihn schon
// **gezeigt** hatte. Schrieb ein anderer Klient unmittelbar vor dem
// Verlassen, schrieb der Entwurf darüber. Jetzt geht mit jedem Schreiben die
// Liste der bekannten Werte an den Bestand, und der prüft sie in derselben
// Transaktion.
import { act, renderHook } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { useEntwurf } from './useEntwurf'

function feld(wert: string) {
  const schreiben = vi.fn<(neu: string, bekannt: readonly string[]) => void>()
  const griff = renderHook(
    ({ wert: gespeichert }: { wert: string }) =>
      useEntwurf(gespeichert, schreiben),
    { initialProps: { wert } },
  )

  return { ...griff, schreiben }
}

it('schickt den Ausgangswert und den Entwurf mit, jeden einmal', () => {
  const { result, schreiben } = feld('Feuerball')

  act(() => {
    result.current.aendern('Feuerball groß')
  })
  act(() => {
    result.current.verlassen('Feuerball groß')
  })

  expect(schreiben).toHaveBeenCalledWith('Feuerball groß', [
    'Feuerball',
    'Feuerball groß',
  ])
})

it('kennt nach »behalten« den fremden Wert', () => {
  const { result, rerender, schreiben } = feld('Feuerball')

  act(() => {
    result.current.aendern('Feuerball groß')
  })
  rerender({ wert: 'Eisball' })
  expect(result.current.konflikt).toBe(true)

  act(() => {
    result.current.behalten()
  })

  expect(schreiben).toHaveBeenCalledWith('Feuerball groß', [
    'Feuerball',
    'Eisball',
    'Feuerball groß',
  ])
})

it('kennt nach einem zweiten Verlassen auch den ersten Entwurf', () => {
  const { result, schreiben } = feld('Feuerball')

  act(() => {
    result.current.verlassen('Erster')
  })
  act(() => {
    result.current.aendern('Zweiter')
  })
  act(() => {
    result.current.verlassen('Zweiter')
  })

  expect(schreiben).toHaveBeenLastCalledWith('Zweiter', [
    'Feuerball',
    'Erster',
    'Zweiter',
  ])
})
