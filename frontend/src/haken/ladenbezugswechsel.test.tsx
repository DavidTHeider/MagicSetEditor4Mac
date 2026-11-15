// **Ein Wechsel des Bezugsobjekts räumt den geladenen Stand** — §7 Regel 1,
// aus dem Zweitblick über Phase L (2026-08-22).
//
// # Der Befund
//
// `useLaden` gab bei einem **neuen** Schlüssel weiterhin die alten Daten
// zurück. Für die Kartenvorschau ist das ausdrücklich gewollt — ein
// abgebrochener Lauf soll das Bild stehen lassen, statt es blinken zu lassen.
//
// **Für eine Liste ist es der Bruch der Regel**: wer das Set wechselt, sah die
// Karten des vorigen unter dem Namen des neuen, bis die Antwort kam. §7
// verlangt, dass ein Wechsel des Bezugsobjekts auch die geladenen Nebenlisten
// schließt.
//
// # Warum es eine Wahl ist und keine Regel
//
// Ein blanker „immer räumen"-Haken ließe die Vorschau bei jedem Tastendruck
// verschwinden und wiederkommen; ihre Abhängigkeiten ändern sich mit jedem
// Zeichen. Der Haken räumt deshalb **von sich aus** und behält nur, wo jemand
// es ausdrücklich verlangt.
import { renderHook, waitFor } from '@testing-library/react'
import { expect, it } from 'vitest'

import { useLaden } from './useLaden'

it('räumt die Daten, wenn sich das Bezugsobjekt ändert', async () => {
  const { result, rerender } = renderHook(
    ({ setID }: { setID: string }) =>
      useLaden(() => Promise.resolve('Karten von ' + setID), [setID]),
    { initialProps: { setID: 'eins' } },
  )

  await waitFor(() => {
    expect(result.current.daten).toBe('Karten von eins')
  })

  rerender({ setID: 'zwei' })

  // **Unmittelbar nach dem Wechsel**, bevor die Antwort kommt.
  expect(result.current.laedt).toBe(true)
  expect(result.current.daten).toBeUndefined()

  await waitFor(() => {
    expect(result.current.daten).toBe('Karten von zwei')
  })
})

// **Und beim Nachladen desselben Bezugsobjekts bleibt er stehen** — sonst
// blinkte jede Liste bei jeder Änderung.
it('behält die Daten, solange das Bezugsobjekt dasselbe ist', async () => {
  const { result } = renderHook(() =>
    useLaden(() => Promise.resolve('unverändert'), ['eins']),
  )

  await waitFor(() => {
    expect(result.current.daten).toBe('unverändert')
  })

  result.current.nachladen()

  expect(result.current.daten).toBe('unverändert')
})
