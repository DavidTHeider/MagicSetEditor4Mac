// **Ein nachgeholtes Fenster gehört zu der Sicht, aus der es stammt** —
// Werkstück #1500.
//
// Das ist §7, Regel 2 für eine Liste, die ihre Zeilen erst holt: zwischen dem
// Ruf und der Antwort kann jemand gesucht, sortiert oder gefiltert haben, und
// eine Antwort von vorhin trüge dann Zeilen ein, die in dieser Sicht an ganz
// anderer Stelle stehen.
//
// **Zwei Tests, nicht einer.** Der zweite ist die Gegenprobe (§10): dieselbe
// späte Antwort landet, wenn die Sicht dieselbe geblieben ist. Ohne ihn sagte
// der erste nur, dass irgendetwas sie fallen lässt — auch ein Beleg, der nie
// antwortet, wäre grün.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { bestand, exemplar } from '../test/sammlungsprobe'
import { zeilenAus } from './sammlungszeilen'
import { useSammlung } from './useSammlung'

afterEach(() => {
  raeumeApi()
})

function belegeFenster(): () => void {
  let freigeben: (() => void) | undefined

  belegeApi({
    Sammlungsfenster: (
      suche: string,
      _ausfuehrung: string,
      _sortierung: string,
      ab: number,
      anzahl: number,
    ) => {
      const versatzDerSicht = suche === '' ? 0 : bestand
      const art = ab > 0 ? 'spät' : suche === '' ? 'Karte' : 'Treffer'
      const stueck = {
        exemplare: Array.from(
          { length: Math.min(anzahl, bestand - ab) },
          (_, versatz) =>
            exemplar(
              ab + versatz + versatzDerSicht,
              `${art} ${String(ab + versatz)}`,
            ),
        ),
        gesamt: bestand,
        insgesamt: bestand,
        stueck: bestand,
        fehler: null,
      }

      if (ab === 0) {
        return Promise.resolve(stueck)
      }

      return new Promise<typeof stueck>((fertig) => {
        freigeben = () => {
          fertig(stueck)
        }
      })
    },
  })

  return () => {
    freigeben?.()
  }
}

it('verwirft ein Fenster, dessen Sicht inzwischen gewechselt hat', async () => {
  const freigeben = belegeFenster()
  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.aufFenster(400, 420)
  })

  // Die Sicht wechselt, während das Fenster unterwegs ist.
  act(() => {
    result.current.setzeSicht({ suche: 'Drache' })
  })
  await waitFor(() => {
    expect(result.current.exemplare[0]?.name).toBe('Treffer 0')
  })

  await act(async () => {
    freigeben()
    await Promise.resolve()
  })

  expect(result.current.exemplare[400]?.name).toBe('')
  expect(result.current.exemplare[400]?.id).toBeLessThan(0)
})

it('trägt ein Fenster ein, dessen Sicht dieselbe geblieben ist', async () => {
  const freigeben = belegeFenster()
  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.aufFenster(400, 420)
  })

  await act(async () => {
    freigeben()
    await Promise.resolve()
  })

  await waitFor(() => {
    expect(result.current.exemplare[400]?.name).toBe('spät 400')
  })
})

// **Ein zweiter Ruf auf dasselbe Fenster geht nicht noch einmal hinaus** —
// sonst holte jedes Zeichnen dieselben Zeilen erneut.
it('holt ein Fenster nur einmal', async () => {
  const freigeben = belegeFenster()
  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.aufFenster(400, 420)
  })
  await act(async () => {
    freigeben()
    await Promise.resolve()
  })
  await waitFor(() => {
    expect(result.current.exemplare[400]?.name).toBe('spät 400')
  })

  // Der zweite Ruf findet die Zeilen vor; gäbe er sie noch einmal hinaus,
  // stünde die Antwort aus und die Zeile fiele auf den Platzhalter zurück.
  act(() => {
    result.current.aufFenster(400, 420)
  })
  expect(result.current.exemplare[400]?.name).toBe('spät 400')
})

// **Was über das Ende hinausreicht, fällt weg** — die Liste ist kürzer
// geworden, während ein Fenster unterwegs war.
it('lässt ein Fenster hinter dem Ende der Liste liegen', () => {
  const geholte = [exemplar(2, 'A'), exemplar(3, 'B'), exemplar(4, 'C')]
  const zeilen = zeilenAus([], 3, {
    schluessel: 'x',
    zeilen: new Map([[2, geholte]]),
  })

  expect(zeilen).toHaveLength(3)
  expect(zeilen[2]?.name).toBe('A')
})

// **Die Wahl überlebt eine Sicht, in der ihre Zeile fehlt** — die Abnahme von
// #1500 sagt
// es so: „Auswahl … stimmt über nicht geholte Zeilen hinweg". Der Inspector
// zeigt weiter, was jemand bearbeiten wollte, statt leer zu stehen.
it('behält die Wahl, wenn ihre Zeile nicht mehr im Fenster liegt', async () => {
  belegeFenster()
  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.waehlen(5)
  })
  expect(result.current.gehaltenes?.name).toBe('Karte 4')

  act(() => {
    result.current.setzeSicht({ suche: 'Drache' })
  })
  await waitFor(() => {
    expect(result.current.exemplare[0]?.name).toBe('Treffer 0')
  })

  expect(result.current.gehaltenes?.name).toBe('Karte 4')
})

// **Und sie gibt nichts vor, wo nichts ist.** Eine Kennung ohne Zeile — weder
// im Fenster noch gemerkt — lässt den Inspector leer; ein Rückfall auf
// irgendeine Zeile wäre schlimmer als keine.
it('gibt nichts, wenn die Kennung zu keiner Zeile gehört', async () => {
  belegeFenster()
  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.waehlen(999999)
  })

  expect(result.current.gehalten).toBe(999999)
  expect(result.current.gehaltenes).toBeUndefined()
})
