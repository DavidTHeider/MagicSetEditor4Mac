// **Ein nachgeholtes Fenster, das scheitert, gehört trotzdem zu seiner Sicht**
// — Werkstück #1500
// und Befund 6 des Nachlaufs zum Zweitblick über Phase 5.2.
//
// Aus `sammlungsfenster.test.tsx` herausgelöst, als es über 400 Zeilen ging
// (§2). Der Schnitt folgt der Frage: dort steht, **was eine Antwort einträgt**,
// hier, **was eine Meldung tut** — und wem sie gehört.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import type { ExemplarDTO } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { bestand, exemplar } from '../test/sammlungsprobe'
import { useSammlung } from './useSammlung'

afterEach(() => {
  raeumeApi()
})

// **Ein gescheitertes Fenster sagt es — und trägt nichts ein.**
//
// Befund 9 des Zweitblicks über Phase 5.2: der Nachschlag griff auf
// `antwort.exemplare` zu, ohne den Umschlag zu öffnen. Ein Fehler aus Go wurde
// damit zu `undefined` im Speicher, die Zeilen blieben leer, und niemand
// erfuhr etwas.
it('meldet einen Fehler beim Nachholen und trägt nichts ein', async () => {
  belegeApi({
    Sammlungsfenster: (
      _suche: string,
      _ausfuehrung: string,
      _sortierung: string,
      ab: number,
      anzahl: number,
    ) => {
      if (ab > 0) {
        return Promise.resolve({
          exemplare: [],
          gesamt: 0,
          insgesamt: 0,
          stueck: 0,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'intern',
            meldung: 'der Bestand schweigt',
            feld: '',
            werte: {},
          },
        })
      }

      return Promise.resolve({
        exemplare: Array.from(
          { length: Math.min(anzahl, bestand) },
          (_, versatz) => exemplar(versatz, `Karte ${String(versatz)}`),
        ),
        gesamt: bestand,
        insgesamt: bestand,
        stueck: bestand,
        fehler: null,
      })
    },
  })

  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.aufFenster(400, 420)
  })

  await waitFor(() => {
    expect(result.current.fehler).toBe('der Bestand schweigt')
  })

  // **Der Platzhalter bleibt stehen** — ein leeres Fenster im Speicher wäre
  // die Behauptung, dort stünde nichts.
  expect(result.current.exemplare[400]?.id).toBeLessThan(0)
})

// **Ein Fehler der alten Sicht landet nicht in der neuen.**
//
// Befund 6 des Nachlaufs zum Zweitblick über Phase 5.2: der Erfolgsweg prüfte
// den Sichtschlüssel, der Fehlerweg nicht. Ein abgelehnter Ruf von vorhin
// hinterließ seine Meldung in einer Sicht, die ihn nie gestellt hat — §7,
// Regel 2 gilt für Erfolg **und** Fehler.
it('lässt einen Fehler der alten Sicht nicht in die neue', async () => {
  let ablehnen: (() => void) | undefined

  belegeApi({
    Sammlungsfenster: (
      suche: string,
      _ausfuehrung: string,
      _sortierung: string,
      ab: number,
      anzahl: number,
    ) => {
      if (ab === 0) {
        const art = suche === '' ? 'Karte' : 'Treffer'

        return Promise.resolve({
          exemplare: Array.from(
            { length: Math.min(anzahl, bestand) },
            (_, versatz) => exemplar(versatz, `${art} ${String(versatz)}`),
          ),
          gesamt: bestand,
          insgesamt: bestand,
          stueck: bestand,
          fehler: null,
        })
      }

      return new Promise((_fertig, gescheitert) => {
        ablehnen = () => {
          gescheitert(new Error('der Ruf von vorhin'))
        }
      })
    },
  })

  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.aufFenster(400, 420)
  })

  await waitFor(() => {
    expect(ablehnen).toBeDefined()
  })

  // **Die Sicht wechselt, während der Ruf unterwegs ist.**
  act(() => {
    result.current.setzeSicht({ suche: 'Drache' })
  })

  await waitFor(() => {
    expect(result.current.exemplare[0]?.name).toBe('Treffer 0')
  })

  await act(async () => {
    ablehnen?.()
    await Promise.resolve()
  })

  expect(result.current.fehler).toBeUndefined()
})

// **Und ein später Erfolg der alten Sicht löscht keinen aktuellen Fehler.**
//
// Die andere Hälfte von Befund 6: der Erfolgsweg **löscht** eine Meldung, und
// ohne Prüfung des Sichtschlüssels nähme ein Ruf von vorhin einen Fehler weg,
// der zur Sicht von jetzt gehört. Ein Wächter, der die halbe Zusage hält,
// macht die andere unsichtbar (§10).
it('lässt einen späten Erfolg der alten Sicht keinen Fehler löschen', async () => {
  let antworten: (() => void) | undefined
  let ablehnen: (() => void) | undefined

  belegeApi({
    Sammlungsfenster: (
      suche: string,
      _ausfuehrung: string,
      _sortierung: string,
      ab: number,
      anzahl: number,
    ) => {
      if (ab === 0) {
        const art = suche === '' ? 'Karte' : 'Treffer'

        return Promise.resolve({
          exemplare: Array.from(
            { length: Math.min(anzahl, bestand) },
            (_, versatz) => exemplar(versatz, `${art} ${String(versatz)}`),
          ),
          gesamt: bestand,
          insgesamt: bestand,
          stueck: bestand,
          fehler: null,
        })
      }

      // **Die alte Sicht antwortet spät und gut, die neue sofort und
      // schlecht** — genau die Reihenfolge, in der das Löschen schadet.
      if (suche === '') {
        return new Promise<{
          stueck: number
          exemplare: ExemplarDTO[]
          gesamt: number
          insgesamt: number
          fehler: null
        }>((fertig) => {
          antworten = () => {
            fertig({
              exemplare: [exemplar(400, 'Karte 400')],
              gesamt: bestand,
              insgesamt: bestand,
              stueck: bestand,
              fehler: null,
            })
          }
        })
      }

      return new Promise((_fertig, gescheitert) => {
        ablehnen = () => {
          gescheitert(new Error('der Ruf von jetzt'))
        }
      })
    },
  })

  const { result } = renderHook(() => useSammlung())

  await waitFor(() => {
    expect(result.current.gesamt).toBe(bestand)
  })

  act(() => {
    result.current.aufFenster(400, 420)
  })

  await waitFor(() => {
    expect(antworten).toBeDefined()
  })

  act(() => {
    result.current.setzeSicht({ suche: 'Drache' })
  })

  await waitFor(() => {
    expect(result.current.exemplare[0]?.name).toBe('Treffer 0')
  })

  // **Die neue Sicht scheitert.**
  act(() => {
    result.current.aufFenster(400, 420)
  })

  await waitFor(() => {
    expect(ablehnen).toBeDefined()
  })

  await act(async () => {
    ablehnen?.()
    await Promise.resolve()
  })

  await waitFor(() => {
    expect(result.current.fehler).toBeDefined()
  })

  const meldung = result.current.fehler

  // **Jetzt kommt die alte Sicht zurück — und lässt die Meldung stehen.**
  await act(async () => {
    antworten?.()
    await Promise.resolve()
  })

  expect(result.current.fehler).toBe(meldung)
})
