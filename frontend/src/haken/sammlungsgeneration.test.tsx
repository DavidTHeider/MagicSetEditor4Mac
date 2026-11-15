// **Eine Antwort gehört zu ihrer Generation, auch wenn die Sicht gleich
// heißt** — Befund 2 des Nachlaufs über Phase G, zu U09
// (#2157).
//
// Nach A → B → A heißt die Sicht wieder wie vorher. Die **Daten** einer
// Antwort aus dem ersten A prüfte der Haken gegen die Generation seines
// Registers, die **Fehlermeldung** nur gegen den Namen der Sicht: ein alter
// Fehler stand in der neuen Sicht, ein alter Erfolg löschte ihren Fehler.
// Jede Antwort wird hier **einzeln** freigegeben — gemeinsam freigegeben
// trügen alte und neue dasselbe, und der Test sähe den Unterschied nicht.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import type { ExemplarDTO } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { bestand, exemplar } from '../test/sammlungsprobe'
import { useSammlung } from './useSammlung'

afterEach(() => {
  raeumeApi()
})

interface Offener {
  ja: () => void
  nein: (fehler: Error) => void
}

// belegeEinzeln gibt das erste Fenster sofort und jedes spätere erst, wenn
// es einzeln freigegeben wird — und nennt jede Antwort nach ihrer Nummer.
function belegeEinzeln(): Offener[] {
  const offen: Offener[] = []

  belegeApi({
    Sammlungsfenster: (_s, _a, _o, ab, anzahl) => {
      const nummer = offen.length
      const antwort = {
        exemplare: Array.from(
          { length: Math.min(anzahl, bestand - ab) },
          (_, versatz): ExemplarDTO =>
            exemplar(
              ab + versatz,
              `Antwort ${String(nummer)} Karte ${String(ab + versatz)}`,
            ),
        ),
        gesamt: bestand,
        insgesamt: bestand,
        stueck: bestand,
        fehler: null,
      }

      if (ab === 0) {
        return Promise.resolve(antwort)
      }

      return new Promise((ja, nein) => {
        offen.push({
          ja: () => {
            ja(antwort)
          },
          nein,
        })
      })
    },
  })

  return offen
}

// abba holt ein Fenster in A, wechselt zu B und zurück und holt erneut.
async function abba(result: { current: ReturnType<typeof useSammlung> }) {
  act(() => {
    result.current.aufFenster(400, 420)
  })
  act(() => {
    result.current.setzeSicht({ suche: 'B' })
  })
  await waitFor(() => {
    expect(result.current.laedt).toBe(false)
  })
  act(() => {
    result.current.setzeSicht({ suche: '' })
  })
  await waitFor(() => {
    expect(result.current.laedt).toBe(false)
  })
  act(() => {
    result.current.aufFenster(400, 420)
  })
}

// nachEinerRunde lässt einer Antwort Zeit, ihre Wirkung zu tun.
const nachEinerRunde = () => new Promise((los) => setTimeout(los, 20))

describe('nach A → B → A', () => {
  it('trägt die Daten der alten Antwort nicht ein', async () => {
    const offen = belegeEinzeln()
    const { result } = renderHook(() => useSammlung())
    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })
    await abba(result)

    await act(async () => {
      offen[0]?.ja()
      await nachEinerRunde()
    })
    expect(result.current.exemplare[400]?.id).toBeLessThan(0)

    act(() => {
      offen[1]?.ja()
    })
    await waitFor(() => {
      expect(result.current.exemplare[400]?.name).toBe('Antwort 1 Karte 400')
    })
  })

  it('setzt den Fehler der alten Antwort nicht', async () => {
    const offen = belegeEinzeln()
    const { result } = renderHook(() => useSammlung())
    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })
    await abba(result)

    act(() => {
      offen[1]?.ja()
    })
    await waitFor(() => {
      expect(result.current.exemplare[400]?.name).toBe('Antwort 1 Karte 400')
    })

    await act(async () => {
      offen[0]?.nein(new Error('planmäßig: das erste A'))
      await nachEinerRunde()
    })
    expect(result.current.fehler).toBeUndefined()
  })

  it('löscht mit einem alten Erfolg nicht den neuen Fehler', async () => {
    const offen = belegeEinzeln()
    const { result } = renderHook(() => useSammlung())
    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })
    await abba(result)

    act(() => {
      offen[1]?.nein(new Error('planmäßig: das zweite A'))
    })
    await waitFor(() => {
      expect(result.current.fehler).toBeDefined()
    })

    await act(async () => {
      offen[0]?.ja()
      await nachEinerRunde()
    })
    expect(result.current.fehler).toBeDefined()
  })
})
