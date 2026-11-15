// **Was der Partielauf tut, wenn nichts offensteht** — Werkstück G1 von
// Phase 5.5 (#1458).
//
// **Neben `partie.test.tsx` und nicht darin**: dort steht der Weg, der trägt,
// hier steht, was ohne offene Partie und bei einem Fehler geschieht. Die
// Datei stand mit beidem bei 402 Zeilen (§2).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { usePartie } from './usePartie'

import { eineAntwort } from '../seiten/partieprobe'
import { einEntwurf } from '../test/partieentwurf'

import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
})

// raum ist der Sprachkontext, den `usePartie` für seine Meldungen braucht.
//
// **Hier noch einmal und nicht geteilt**: er ist vier Zeilen, und eine
// gemeinsame Datei für vier Zeilen wäre eine Abhängigkeit für nichts.
function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// **Ohne offene Partie tut ein Befehl nichts.**
//
// **Kein stiller Ruf ins Leere**: ein Fenster, das `undefined` als
// Partiekennung schickte, bekäme einen Fehler über eine Partie, die es nie
// gab.
it('tut ohne offene Partie nichts', async () => {
  const befohlen = vi.fn()

  belegeApi({ PartieBefehl: befohlen })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.befehlen('objekt.tappen', {})
    result.current.wechseln('beenden')
  })

  await antwortAbwarten()

  expect(befohlen).not.toHaveBeenCalled()
})

// **Aufräumen legt alles weg.**
it('räumt Partie, Ereignisse und Meldung weg', async () => {
  belegeApi({ PartieAnlegen: () => Promise.resolve(eineAntwort()) })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.partie).toBeDefined()
  })

  act(() => {
    result.current.raeumen()
  })

  expect(result.current.partie).toBeUndefined()
  expect(result.current.erlaubt).toHaveLength(0)
})

// **Ein Fehler beim Ansehen steht als Meldung da.**
it('meldet, was beim Ansehen schiefging', async () => {
  belegeApi({
    Partie: () =>
      Promise.resolve({
        partie: eineAntwort().partie,
        erlaubt: [],
        handlungen: [],
        befunde: [],
        fehler: {
          code: 'nichtGefunden',
          schluessel: '',
          meldung: 'Die Partie gibt es nicht.',
          feld: '',
        },
      }),
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.ansehen('p9', 'sitz1')
  })

  await waitFor(() => {
    expect(result.current.fehler).toBe('Die Partie gibt es nicht.')
  })
})
