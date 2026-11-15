// **Eine alte Antwort übermalt keinen neuen Stand — in jeder Reihenfolge**
// (C05, #2230).
//
// Die Fälle daneben prüfen je eine ausgedachte Überholung. Hier gehen
// mehrere Fragen hinaus — ein neuer Nachlesestand nach dem anderen, dann ein
// Bezugswechsel —, und die Antworten kommen in einer Reihenfolge zurück, die
// ein Samen bestimmt. §7, Regel 2: gezeigt wird nie ein älterer Stand als der
// schon gezeigte, nie ein fremder Bezug, und am Ende der jüngste.
import { act, renderHook } from '@testing-library/react'
import { expect, it } from 'vitest'

import { useLaden } from './useLaden'

interface Anfrage {
  wert: string
  erfuellen: (wert: string) => void
}

// zufall ist ein kleiner, gesäter Zahlengeber — damit ein roter Lauf sich
// nachstellen lässt.
function zufall(samen: number): () => number {
  let s = samen >>> 0

  return () => {
    s = (s + 0x6d2b79f5) >>> 0
    let t = s
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)

    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const stufe = (wert: string | undefined) =>
  wert === undefined ? -1 : Number(wert.split('@')[1])

it.each(Array.from({ length: 25 }, (_, samen) => samen))(
  'Samen %i: keine alte Antwort übermalt einen neuen Stand',
  async (samen) => {
    const naechste = zufall(samen)
    const anfragen: Anfrage[] = []

    const { result, rerender } = renderHook(
      ({ bezug, stand }: { bezug: string; stand: number }) =>
        useLaden(
          () =>
            new Promise<string>((erfuellen) => {
              anfragen.push({ wert: `${bezug}@${String(stand)}`, erfuellen })
            }),
          [bezug],
          false,
          stand,
        ),
      { initialProps: { bezug: 'eins', stand: 0 } },
    )

    // **Mehrere Nachlesestände auf demselben Bezug, dann ein Wechsel.**
    const staende = 2 + Math.floor(naechste() * 4)
    for (let stand = 1; stand <= staende; stand++) {
      rerender({ bezug: 'eins', stand })
    }
    const wechseln = naechste() < 0.5
    if (wechseln) rerender({ bezug: 'zwei', stand: staende + 1 })

    const juengste = wechseln
      ? `zwei@${String(staende + 1)}`
      : `eins@${String(staende)}`
    const reihenfolge = [...anfragen].sort(() => naechste() - 0.5)
    let gezeigt = -1

    for (const anfrage of reihenfolge) {
      await act(async () => {
        anfrage.erfuellen(anfrage.wert)
        await Promise.resolve()
      })

      const daten = result.current.daten
      if (wechseln && daten !== undefined)
        expect(daten.startsWith('zwei@')).toBe(true)
      expect(stufe(daten)).toBeGreaterThanOrEqual(gezeigt)
      gezeigt = Math.max(gezeigt, stufe(daten))
    }

    expect(result.current.daten).toBe(juengste)
  },
)
