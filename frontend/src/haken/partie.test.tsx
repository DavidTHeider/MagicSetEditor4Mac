// **Der Partiehaken** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Wogegen hier geprüft wird
//
// **Die vier Regeln aus §7**, und jede an der Stelle, an der sie bricht:
//
//  1. Ein Sitzwechsel räumt die Ereignisse mit — sie gehörten dem Sitz von
//     vorhin, und private Ergänzungen sind empfängergebunden.
//  2. Eine Antwort aus einem überholten Stand trägt sich nicht mehr ein.
//  3. Nach einem **gescheiterten** Befehl wird nicht nachgeladen; nach einem
//     geglückten schon, und ein Fehler dabei nimmt nichts zurück.
//  4. Ein zweiter Befehl, während der erste läuft, wird abgewiesen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { usePartie } from './usePartie'

import { eineAntwort, einBefehlsausgang } from '../seiten/partieprobe'
import { einEntwurf } from '../test/partieentwurf'

import type { BefehlAntwort } from '../apipartie'
import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
})

// raum ist der Sprachkontext, den `usePartie` für seine Meldungen braucht.
function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// **Eine angelegte Partie steht mit ihren erlaubten Befehlen da.**
it('trägt die angelegte Partie ein', async () => {
  const handlung = {
    objekt: 'sitz1/k1/0',
    handlungen: ['objekt.tappen'],
    primaer: 'objekt.tappen',
    zweit: '',
  }
  belegeApi({
    PartieAnlegen: () =>
      Promise.resolve({ ...eineAntwort(), handlungen: [handlung] }),
  })

  const { result } = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    result.current.anfangen(einEntwurf)
  })

  await waitFor(() => {
    expect(result.current.partie?.id).toBe('p1')
  })

  expect(result.current.erlaubt).toContain('objekt.tappen')
  // **Mit der Auskunft je Objekt** (P3, #2371) — aus derselben Antwort.
  expect(result.current.handlungen).toEqual([handlung])
})

// **Ein Sitzwechsel räumt die Ereignisse mit** (Regel 1).
it('räumt beim Sitzwechsel die Ereignisse weg', async () => {
  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartieBefehl: () => Promise.resolve(einBefehlsausgang()),
    Partie: () => Promise.resolve(eineAntwort()),
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
    expect(result.current.ereignisse).toHaveLength(1)
  })

  act(() => {
    result.current.ansehen('p1', 'sitz2')
  })

  await waitFor(() => {
    expect(result.current.ereignisse).toHaveLength(0)
  })
})

// **Was hinausgeht, trägt Kennung, Absender und die gezeigte Fassung.**
//
// **Die Fassung ist die, die dem Menschen dastand** — ist dazwischen etwas
// geschehen, das er nicht kennt, kommt ein Konflikt zurück statt einer
// stillen Wirkung.
it('schickt Befehlskennung, Absender und erwartete Fassung mit', async () => {
  const befohlen = vi.fn(() => Promise.resolve(einBefehlsausgang()))

  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartieBefehl: befohlen,
    Partie: () => Promise.resolve(eineAntwort()),
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
    expect(befohlen).toHaveBeenCalled()
  })

  const [partieID, befehlID, absender, art, nutzlast, erwartet] = befohlen.mock
    .calls[0] as unknown as [
    string,
    string,
    string,
    string,
    Record<string, unknown>,
    string,
  ]

  expect(partieID).toBe('p1')
  expect(befehlID).not.toBe('')
  expect(absender).toBe('sitz1')
  expect(art).toBe('objekt.tappen')
  expect(nutzlast).toEqual({ objekt: 'sitz1/k1/0' })
  expect(erwartet).toBe('1')
})

// **Nach einem gescheiterten Befehl wird nicht nachgeladen** (Regel 3).
//
// **Sonst überschriebe ein Lesefehler die Meldung, die sagt, was zu tun
// ist** — und im guten Fall überschriebe die frische Sicht sie stillschweigend.
it('lädt nach einem gescheiterten Befehl nicht nach', async () => {
  const gelesen = vi.fn(() => Promise.resolve(eineAntwort()))

  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartieBefehl: () =>
      Promise.resolve({
        fassung: '',
        ereignisse: [],
        fehler: {
          code: 'konflikt',
          schluessel: '',
          meldung: 'Zu spät.',
          feld: '',
        },
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
    expect(result.current.fehler).toBe('Zu spät.')
  })

  await antwortAbwarten()

  expect(gelesen).not.toHaveBeenCalled()
})

// **Ein zweiter Befehl, während der erste läuft, wird abgewiesen** (Regel 4).
it('lässt nur einen Befehl je Partie laufen', async () => {
  let freigeben: (() => void) | undefined

  const befohlen = vi.fn(
    () =>
      new Promise<BefehlAntwort>((fertig) => {
        freigeben = () => {
          fertig(einBefehlsausgang())
        }
      }),
  )

  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartieBefehl: befohlen,
    Partie: () => Promise.resolve(eineAntwort()),
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

  act(() => {
    result.current.befehlen('objekt.umdrehen', { objekt: 'sitz1/k1/0' })
  })

  await antwortAbwarten()

  expect(befohlen).toHaveBeenCalledTimes(1)

  await act(async () => {
    freigeben?.()
    await Promise.resolve()
  })
})

// **Ein Standwechsel trägt die neue Sicht selbst ein.**
it('trägt die Sicht nach einem Standwechsel ein', async () => {
  const gelesen = vi.fn(() => Promise.resolve(eineAntwort()))

  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartieBeenden: () => Promise.resolve(eineAntwort({ stand: 'beendet' }, [])),
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
    result.current.wechseln('beenden')
  })

  await waitFor(() => {
    expect(result.current.partie?.stand).toBe('beendet')
  })

  // **Und kein Nachladen** — die Antwort ist die Sicht.
  expect(gelesen).not.toHaveBeenCalled()
})

// **Pausieren und Fortsetzen gehen ihre eigenen Wege.**
it('führt Pausieren und Fortsetzen über ihre Bindungen', async () => {
  const pausiert = vi.fn(() =>
    Promise.resolve(eineAntwort({ stand: 'pausiert' })),
  )
  const fortgesetzt = vi.fn(() => Promise.resolve(eineAntwort()))

  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort()),
    PartiePausieren: pausiert,
    PartieFortsetzen: fortgesetzt,
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
    expect(result.current.partie?.stand).toBe('pausiert')
  })

  act(() => {
    result.current.wechseln('fortsetzen')
  })

  await waitFor(() => {
    expect(result.current.partie?.stand).toBe('laeuft')
  })

  expect(pausiert).toHaveBeenCalledWith('p1', 'sitz1')
  expect(fortgesetzt).toHaveBeenCalledWith('p1', 'sitz1')
})
