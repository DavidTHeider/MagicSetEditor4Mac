// **Ein Fenster, das unterwegs ist, wird nicht noch einmal geholt** — U09
// (#2157).
//
// `aufFenster` merkte sich nur, was **angekommen** war. Wer scrollte,
// während ein Fenster noch lief, stieß denselben Ruf ein zweites Mal an.
// Nach einem Fehler muss es dagegen wieder abrufbar sein.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { bestand, exemplar } from '../test/sammlungsprobe'
import { useSammlung } from './useSammlung'

afterEach(() => {
  raeumeApi()
})

// belegeZaehlend gibt das erste Fenster sofort und jedes spätere erst auf
// `freigeben` — mit einem Zähler je Anfang. `scheitern` lässt die späten
// Rufe ablehnen.
function belegeZaehlend(scheitern = false) {
  const gezaehlt = new Map<number, number>()
  const offen: (() => void)[] = []

  belegeApi({
    Sammlungsfenster: (
      _s: string,
      _a: string,
      _o: string,
      ab: number,
      anzahl: number,
    ) => {
      gezaehlt.set(ab, (gezaehlt.get(ab) ?? 0) + 1)

      const stueck = {
        exemplare: Array.from(
          { length: Math.min(anzahl, bestand - ab) },
          (_, versatz) =>
            exemplar(ab + versatz, `Karte ${String(ab + versatz)}`),
        ),
        gesamt: bestand,
        insgesamt: bestand,
        stueck: bestand,
        fehler: null,
      }

      if (ab === 0) {
        return Promise.resolve(stueck)
      }

      return new Promise<typeof stueck>((fertig, nein) => {
        offen.push(() => {
          if (scheitern) {
            nein(new Error('planmäßig: das Fenster kam nicht'))
          } else {
            fertig(stueck)
          }
        })
      })
    },
  })

  return {
    gezaehlt,
    freigeben: () => {
      for (const los of offen.splice(0)) {
        los()
      }
    },
  }
}

describe('ein Fenster unterwegs', () => {
  it('wird für dieselbe Sicht nur einmal geholt', async () => {
    const probe = belegeZaehlend()
    const { result } = renderHook(() => useSammlung())

    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })

    act(() => {
      result.current.aufFenster(400, 420)
    })
    act(() => {
      result.current.aufFenster(400, 420)
    })

    expect(probe.gezaehlt.get(400)).toBe(1)

    act(() => {
      probe.freigeben()
    })

    await waitFor(() => {
      expect(result.current.exemplare[400]?.name).toBe('Karte 400')
    })
  })

  it('ist nach einem Fehler wieder abrufbar', async () => {
    const probe = belegeZaehlend(true)
    const { result } = renderHook(() => useSammlung())

    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })

    act(() => {
      result.current.aufFenster(400, 420)
    })
    act(() => {
      probe.freigeben()
    })

    await waitFor(() => {
      expect(result.current.fehler).toBeDefined()
    })

    act(() => {
      result.current.aufFenster(400, 420)
    })

    expect(probe.gezaehlt.get(400)).toBe(2)
  })
})

// **Ein Fenster deckt seinen Bereich, nicht seinen Anfang** — Befund 2 des
// Zweitblicks über Phase G. Gemerkt war der Anfang: wer nach `[400,420)`
// `[400,440)` brauchte, etwa nach einer Größenänderung, bekam die Zeilen
// 420–439 nie — weder solange das kurze Fenster lief noch danach.
describe('ein größeres Fenster', () => {
  it('holt den Rest, solange das kleinere unterwegs ist', async () => {
    const probe = belegeZaehlend()
    const { result } = renderHook(() => useSammlung())

    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })

    act(() => {
      result.current.aufFenster(400, 420)
    })
    act(() => {
      result.current.aufFenster(400, 440)
    })
    act(() => {
      probe.freigeben()
    })

    await waitFor(() => {
      expect(result.current.exemplare[439]?.name).toBe('Karte 439')
    })
    expect(probe.gezaehlt.get(400)).toBe(1)
  })

  it('holt den Rest, nachdem das kleinere angekommen ist', async () => {
    const probe = belegeZaehlend()
    const { result } = renderHook(() => useSammlung())

    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })

    act(() => {
      result.current.aufFenster(400, 420)
    })
    act(() => {
      probe.freigeben()
    })
    await waitFor(() => {
      expect(result.current.exemplare[419]?.name).toBe('Karte 419')
    })

    act(() => {
      result.current.aufFenster(400, 440)
    })
    act(() => {
      probe.freigeben()
    })

    await waitFor(() => {
      expect(result.current.exemplare[439]?.name).toBe('Karte 439')
    })
    expect(probe.gezaehlt.get(400)).toBe(1)
  })
})

// **Eine Sicht, die wiederkommt, ist eine neue** — derselbe Befund: nach
// A → B → A galt der Ruf aus dem ersten A als unterwegs, und das zweite A
// fragte nicht.
describe('eine wiederkehrende Sicht', () => {
  it('holt ihr Fenster neu', async () => {
    const probe = belegeZaehlend()
    const { result } = renderHook(() => useSammlung())

    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })

    act(() => {
      result.current.aufFenster(400, 420)
    })
    act(() => {
      result.current.setzeSicht({ suche: 'anders' })
    })
    await waitFor(() => {
      expect(probe.gezaehlt.get(0)).toBe(2)
    })
    act(() => {
      result.current.setzeSicht({ suche: '' })
    })
    await waitFor(() => {
      expect(probe.gezaehlt.get(0)).toBe(3)
    })
    await waitFor(() => {
      expect(result.current.gesamt).toBe(bestand)
    })

    act(() => {
      result.current.aufFenster(400, 420)
    })

    expect(probe.gezaehlt.get(400)).toBe(2)

    // **Die Antwort des ersten A gehört nicht ins zweite** — sie kommt mit
    // und wird verworfen; die Zeile füllt der neue Ruf.
    act(() => {
      probe.freigeben()
    })

    await waitFor(() => {
      expect(result.current.exemplare[400]?.name).toBe('Karte 400')
    })
  })
})
