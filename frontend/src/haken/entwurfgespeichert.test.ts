// **Der Entwurf bestätigt genau den gespeicherten Wert** — #2535 (Nachlauf
// über das Gesamtreview über 0.9.0, Befund 1).
//
// Bis dahin galt ein geglückter Entwurf als erledigt, sobald sich der Bestand
// **irgendwie** änderte: auch ein fremder Wert, der zwischen Speichern und
// Nachladen kam, wurde so als eigener genommen. Und rechnete der Bestand den
// Wert auf den Ausgangswert zurück, änderte sich nichts, und der Entwurf blieb
// stehen. Jetzt sagt das Schreiben, was es gespeichert hat.
import { act, renderHook } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { useEntwurf } from './useEntwurf'

function feld(wert: string, gespeichert: string) {
  const schreiben = vi.fn(() => Promise.resolve({ gespeichert }))
  const griff = renderHook(
    ({ wert: bestand }: { wert: string }) => useEntwurf(bestand, schreiben),
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

it('ist erledigt, wenn der Bestand den gespeicherten Wert zeigt', async () => {
  const { result, rerender } = feld('[Funke]', '[Glut]')
  const vorher = result.current.schluessel

  await schicke(result, 'Glut')
  rerender({ wert: '[Glut]' })

  expect(result.current.konflikt).toBe(false)
  expect(result.current.schluessel).not.toBe(vorher)
})

it('nimmt einen fremden Wert nach dem Speichern nicht als seinen', async () => {
  const { result, rerender } = feld('[Funke]', '[Glut]')

  await schicke(result, 'Glut')
  rerender({ wert: '[Asche]' })

  expect(result.current.konflikt).toBe(true)
})

// **Gerechnet auf den Ausgangswert** — der Bestand ändert sich nicht, und
// trotzdem ist der Entwurf angekommen.
it('ist erledigt, wenn das Gespeicherte der Ausgangswert ist', async () => {
  const { result, rerender } = feld('[Glut]', '[Glut]')
  const vorher = result.current.schluessel

  await schicke(result, 'Glut')
  rerender({ wert: '[Glut]' })

  expect(result.current.konflikt).toBe(false)
  expect(result.current.schluessel).not.toBe(vorher)

  // **Danach folgt die Eingabe wieder dem Bestand** — ohne Entwurf ist eine
  // spätere Änderung kein Konflikt.
  rerender({ wert: '[Asche]' })
  expect(result.current.konflikt).toBe(false)
})

// **Ein Gespeichertes, das kein Text ist, bestätigt nichts** — dann bleibt
// der wörtliche Vergleich, und ein fremder Wert ist ein Konflikt.
it('nimmt eine Antwort ohne gespeicherten Text nicht als Bestätigung', async () => {
  const schreiben = vi.fn(() => Promise.resolve({ gespeichert: null }))
  const { result, rerender } = renderHook(
    ({ wert: bestand }: { wert: string }) => useEntwurf(bestand, schreiben),
    { initialProps: { wert: '[Funke]' } },
  )

  await schicke(result, 'Glut')
  rerender({ wert: '[Asche]' })

  expect(result.current.konflikt).toBe(true)
})
