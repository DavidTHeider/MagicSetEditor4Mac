// **Ein geglückter Entwurf ist erledigt, auch wenn das Wertskript rechnet** —
// Gesamtreview über 0.9.0, Befund 1
// (#2382).
//
// Seit #2396 speichert der Bestand den gerechneten Wert: aus „Glut" wird
// „[Glut]", in einem M15-Set aus „U1" die geordneten Manakosten. Der Entwurf
// galt als erledigt, wenn der Bestand **genau** seinen Wert zeigte — das
// geschah nie, und nach geglücktem Speichern stand ein Konflikt da, als hätte
// ein anderer geschrieben.
//
// **Woran „geglückt" erkannt wird, sagt das Schreiben**, nicht der nächste
// Bestand: der könnte von einem anderen stammen, und ein abgewiesener
// Entwurf gälte dann als erledigt.
import { act, renderHook } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { useEntwurf } from './useEntwurf'

function feld(wert: string, gelungen: boolean) {
  const schreiben = vi.fn(() => Promise.resolve(gelungen))
  const griff = renderHook(
    ({ wert: gespeichert }: { wert: string }) =>
      useEntwurf(gespeichert, schreiben),
    { initialProps: { wert } },
  )

  return { ...griff, schreiben }
}

async function schicke(
  result: { current: ReturnType<typeof useEntwurf> },
  neu: string,
) {
  act(() => {
    result.current.aendern(neu)
  })
  await act(async () => {
    result.current.verlassen(neu)
    await Promise.resolve()
  })
}

it('nimmt nach geglücktem Schreiben den gerechneten Wert als seinen', async () => {
  const { result, rerender } = feld('[Funke]', true)
  const vorher = result.current.schluessel

  await schicke(result, 'Glut')
  rerender({ wert: '[Glut]' })

  expect(result.current.konflikt).toBe(false)
  // **Die Eingabe zeigt danach den gerechneten Wert** — sie wird dafür neu
  // aufgebaut, denn sie hielt den getippten.
  expect(result.current.schluessel).not.toBe(vorher)
})

it('bleibt nach abgewiesenem Schreiben im Konflikt mit einem fremden Wert', async () => {
  const { result, rerender } = feld('[Funke]', false)

  await schicke(result, 'Glut')
  rerender({ wert: '[Asche]' })

  expect(result.current.konflikt).toBe(true)
})

// **Wer weitertippt, während geschrieben wird, hat einen neuen Entwurf** —
// die Bestätigung gilt dem geschickten, nicht ihm.
it('bestätigt einen Entwurf nicht, der nach dem Schicken weitergetippt wurde', async () => {
  let antworte: (gelungen: boolean) => void = () => undefined
  const schreiben = vi.fn(
    () =>
      new Promise<boolean>((fertig) => {
        antworte = fertig
      }),
  )
  const { result, rerender } = renderHook(
    ({ wert: gespeichert }: { wert: string }) =>
      useEntwurf(gespeichert, schreiben),
    { initialProps: { wert: '[Funke]' } },
  )

  act(() => {
    result.current.aendern('Glut')
  })
  act(() => {
    result.current.verlassen('Glut')
  })
  act(() => {
    result.current.aendern('Glut weiter')
  })
  await act(async () => {
    antworte(true)
    await Promise.resolve()
  })
  rerender({ wert: '[Glut]' })

  expect(result.current.konflikt).toBe(true)
})
