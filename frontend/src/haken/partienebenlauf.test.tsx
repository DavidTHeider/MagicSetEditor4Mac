// **Der Partiehaken unter Nebenläufigkeit** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Aus `partie.test.tsx` herausgelöst**, als die Datei über 400 Zeilen ging
// (§2). Der Schnitt folgt der Frage: dort, **was** der Haken tut; hier, was
// er tut, wenn zwei Dinge zugleich unterwegs sind — die Regeln 2 und 3
// aus §7.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { usePartie } from './usePartie'

import { eineAntwort, einBefehlsausgang } from '../seiten/partieprobe'

import type { BefehlAntwort, PartieAntwort } from '../apipartie'
import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
})

// raum ist der Sprachkontext, den `usePartie` für seine Meldungen braucht.
function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// einEntwurf ist der kleinste, der durchgeht.
const einEntwurf = {
  spiel: 'magic',
  regelwerk: { kennung: '', fassung: 0, strenge: '', abgewaehlt: [] },
  wirt: '',
  sitze: [
    {
      kennung: 'sitz1',
      name: 'Anna',
      deckID: '1',
      leiheID: '',
      profilID: '',
    },
    {
      kennung: 'sitz2',
      name: 'Bert',
      deckID: '2',
      leiheID: '',
      profilID: '',
    },
  ],
}

// **Eine Antwort aus einem überholten Stand trägt sich nicht mehr ein**
// (§7, Regel 2).
//
// # Wie dieser Lauf den überholten Stand herstellt
//
// **Zwei Partien, zwei Vorgangsschlüssel.** `useVorgang` ließe zwei Rufe auf
// **dieselbe** Partie nicht nebeneinander laufen (Regel 4); über zwei
// Kennungen laufen sie, und der zweite zählt den Stand weiter. Die Antwort
// des ersten gehört danach zu einer Sicht, die niemand mehr ansieht.
it('trägt eine überholte Antwort nicht mehr ein', async () => {
  let ersteFreigeben: (() => void) | undefined

  belegeApi({
    Partie: (kennung: string) =>
      kennung === 'p1'
        ? new Promise<PartieAntwort>((fertig) => {
            ersteFreigeben = () => {
              fertig(eineAntwort({ id: 'p1', fuer: 'sitz1' }))
            }
          })
        : Promise.resolve(eineAntwort({ id: 'p2', fuer: 'sitz2' })),
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.ansehen('p1', 'sitz1')
  })

  act(() => {
    result.current.ansehen('p2', 'sitz2')
  })

  await waitFor(() => {
    expect(result.current.partie?.id).toBe('p2')
  })

  await act(async () => {
    ersteFreigeben?.()
    await Promise.resolve()
  })

  await antwortAbwarten()

  // **Die späte Antwort der ersten Partie steht nicht über der zweiten.**
  expect(result.current.partie?.id).toBe('p2')
})

// **Und ein überholter Befehl lädt auch nicht mehr nach.**
//
// **Ein eigener Lauf**, weil der Befehl einen zweiten Ausgang hat: er trägt
// nicht nur nichts ein, er **hört auf** — ein Nachladen für eine Partie, die
// niemand mehr ansieht, wäre ein Ruf ins Leere.
it('lädt nach einem überholten Befehl nicht nach', async () => {
  let befehlFreigeben: (() => void) | undefined

  const gelesen = vi.fn((kennung: string) =>
    Promise.resolve(eineAntwort({ id: kennung })),
  )

  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartieBefehl: () =>
      new Promise<BefehlAntwort>((fertig) => {
        befehlFreigeben = () => {
          fertig(einBefehlsausgang())
        }
      }),
    Partie: gelesen,
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.partie).toBeDefined()
  })

  act(() => {
    result.current.befehlen('objekt.tappen', { objekt: 'sitz1/k1/0' })
  })

  await waitFor(() => {
    expect(result.current.laeuft('partie-p1')).toBe(true)
  })

  // **Der Stand zählt weiter** — eine andere Partie wird angesehen.
  act(() => {
    result.current.ansehen('p2', 'sitz1')
  })

  await waitFor(() => {
    expect(gelesen).toHaveBeenCalledTimes(1)
  })

  await act(async () => {
    befehlFreigeben?.()
    await Promise.resolve()
  })

  await antwortAbwarten()

  // **Kein zweiter Lesevorgang** — der Befehl hat aufgehört.
  expect(gelesen).toHaveBeenCalledTimes(1)
})

// **Scheitert das Nachladen, bleibt der Befehl trotzdem gewirkt**
// (§7, Regel 3).
//
// **Bestenfalls-Arbeit**: die Meldung sagt, dass die Sicht alt ist — sie
// nimmt nicht zurück, was geschehen ist.
it('meldet einen Fehler beim Nachladen, ohne etwas zurückzunehmen', async () => {
  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartieBefehl: () => Promise.resolve(einBefehlsausgang()),
    Partie: () => Promise.reject(new Error('Der Bestand ist zu.')),
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.partie).toBeDefined()
  })

  act(() => {
    result.current.befehlen('objekt.tappen', { objekt: 'sitz1/k1/0' })
  })

  await waitFor(() => {
    expect(result.current.fehler).toBe('Der Bestand ist zu.')
  })

  // **Die Ereignisse stehen, und die Partie ist noch da.**
  expect(result.current.ereignisse).toHaveLength(1)
  expect(result.current.partie?.id).toBe('p1')
})

// **Und ein gescheiterter Standwechsel meldet sich.**
it('meldet, was bei einem Standwechsel schiefging', async () => {
  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartiePausieren: () => Promise.reject(new Error('Nicht jetzt.')),
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.partie).toBeDefined()
  })

  act(() => {
    result.current.wechseln('pausieren')
  })

  await waitFor(() => {
    expect(result.current.fehler).toBe('Nicht jetzt.')
  })

  // **Und der Stand ist geblieben** — nichts halb Gewechseltes.
  expect(result.current.partie?.stand).toBe('laeuft')
})
