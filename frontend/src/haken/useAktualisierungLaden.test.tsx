// **Die Aktualisierung am Fenster** —
// #1100.
//
// **Stufe 2: holen.** Die Zusagen des Meldens stehen nebenan in
// `useAktualisierung.test.tsx`; hier stehen die des Ladens — ein Vorgang je
// Zeit, ein Fortschritt, der wirklich ankommt, und kein erfundener Anteil.
//
// **Zwei Dateien, weil die eine über 400 Zeilen ging** (§2) — und der Schnitt
// folgt der Frage, nicht der Zeilenzahl: melden ist Stufe 1, holen ist
// Stufe 2.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { FassungsAntwort, LadeAntwort } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { useAktualisierung } from './useAktualisierung'

// horcher hält die Rückrufe, die der Haken bei der Laufzeit anmeldet.
const horcher = new Map<string, (...daten: unknown[]) => void>()

beforeEach(() => {
  horcher.clear()
  localStorage.clear()
  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) => {
      horcher.set(name, rueckruf)

      return () => horcher.delete(name)
    },
  }
})

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

const neuere = {
  gefragt: true,
  neuer: true,
  fassung: '9.9.9',
  seite: 'https://beispiel.test/v9',
  holen: true,
  fehler: null,
}

describe('Der Aktualisierungshaken: holen', () => {
  it('lädt und meldet den Ort', async () => {
    const NeueFassungLaden = vi.fn().mockResolvedValue({
      neuer: true,
      fassung: '9.9.9',
      pfad: '/Users/x/Downloads/neu.dmg',
      fehler: null,
    })

    belegeApi({
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden,
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    await act(async () => {
      await result.current.laden()
    })

    expect(result.current.pfad).toBe('/Users/x/Downloads/neu.dmg')
    expect(result.current.laeuft).toBe(false)
  })

  // **Der Fortschritt kommt über ein Ereignis** — 188 MB sind kein Augenblick.
  it('zeigt den Fortschritt, während geladen wird', async () => {
    let loesen: (wert: LadeAntwort) => void = () => undefined

    const NeueFassungLaden = vi.fn().mockReturnValue(
      new Promise((ja) => {
        loesen = ja
      }),
    )

    belegeApi({
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden,
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    let fertig: Promise<void> = Promise.resolve()

    act(() => {
      fertig = result.current.laden()
    })

    expect(result.current.laeuft).toBe(true)

    act(() => {
      horcher.get('aktualisierung:fortschritt')?.(500, 1000)
    })

    expect(result.current.anteil).toBe(0.5)

    await act(async () => {
      loesen({
        neuer: true,
        fassung: '9.9.9',
        pfad: '/x/neu.dmg',
        fehler: null,
      })
      await fertig
    })

    expect(result.current.laeuft).toBe(false)
  })

  // **Ohne genannte Länge kein erfundener Anteil** (§10: die Reichweite einer
  // Messung ist die Reichweite ihrer Aussage).
  it('erfindet keinen Anteil, wenn die Länge fehlt', async () => {
    let loesen: (wert: LadeAntwort) => void = () => undefined

    belegeApi({
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden: () =>
        new Promise((ja) => {
          loesen = ja
        }),
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    let fertig: Promise<void> = Promise.resolve()

    act(() => {
      fertig = result.current.laden()
    })

    act(() => {
      horcher.get('aktualisierung:fortschritt')?.(4096, -1)
    })

    expect(result.current.anteil).toBeUndefined()
    expect(result.current.gelesen).toBe(4096)

    await act(async () => {
      loesen({
        neuer: true,
        fassung: '9.9.9',
        pfad: '/x/neu.dmg',
        fehler: null,
      })
      await fertig
    })
  })

  // **Ein Vorgang je Zeit** (§7, Regel 4). Ein zweiter Knopfdruck, während der
  // erste läuft, holte 188 MB ein zweites Mal.
  it('lädt nicht zweimal gleichzeitig', async () => {
    let loesen: (wert: LadeAntwort) => void = () => undefined

    const NeueFassungLaden = vi.fn().mockReturnValue(
      new Promise((ja) => {
        loesen = ja
      }),
    )

    belegeApi({
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden,
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    let fertig: Promise<void> = Promise.resolve()

    act(() => {
      fertig = result.current.laden()
    })

    await act(async () => {
      await result.current.laden()
    })

    expect(NeueFassungLaden).toHaveBeenCalledTimes(1)

    await act(async () => {
      loesen({
        neuer: true,
        fassung: '9.9.9',
        pfad: '/x/neu.dmg',
        fehler: null,
      })
      await fertig
    })
  })

  // **Ein zweiter Versuch fängt beim Fortschritt wieder bei null an.**
  //
  // Ohne das Zurücksetzen zeigte er bis zum ersten neuen Ereignis die alte
  // Menge — eine Zahl aus einem Lauf, den es nicht mehr gibt (§10).
  // Angestrichen vom Gesamtreview vor 0.2.0.
  it('fängt beim Wiederholen wieder bei null an', async () => {
    let loesen: (wert: LadeAntwort) => void = () => undefined

    const NeueFassungLaden = vi
      .fn()
      .mockReturnValueOnce(
        new Promise<LadeAntwort>((ja) => {
          loesen = ja
        }),
      )
      .mockReturnValue(
        new Promise<LadeAntwort>(() => {
          // bleibt offen, damit der zweite Lauf beim Fortschritt steht
        }),
      )

    belegeApi({
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden,
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    let erster: Promise<void> = Promise.resolve()

    act(() => {
      erster = result.current.laden()
    })

    act(() => {
      horcher.get('aktualisierung:fortschritt')?.(700, 1000)
    })

    expect(result.current.gelesen).toBe(700)

    await act(async () => {
      loesen({
        neuer: true,
        fassung: '9.9.9',
        pfad: '',
        fehler: {
          code: 'konflikt',
          meldung: 'die Leitung ist weg',
          schluessel: 'weg',
        },
      })
      await erster
    })

    expect(result.current.fehler).not.toBe('')

    // **Der zweite Versuch**, und vor dem ersten Ereignis steht keine Zahl.
    act(() => {
      void result.current.laden()
    })

    expect(result.current.gelesen).toBe(0)
    expect(result.current.anteil).toBeUndefined()
  })

  it('meldet einen Fehler des Ladevorgangs', async () => {
    belegeApi({
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden: () =>
        Promise.resolve({
          neuer: true,
          fassung: '9.9.9',
          pfad: '',
          fehler: {
            code: 'konflikt',
            meldung: 'die Prüfsumme stimmt nicht',
            schluessel: 'pruefsumme',
          },
        }),
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    await act(async () => {
      await result.current.laden()
    })

    expect(result.current.fehler).not.toBe('')
    expect(result.current.pfad).toBe('')
  })

  it('meldet einen abgelehnten Aufruf', async () => {
    belegeApi({
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden: () => Promise.reject(new Error('weg')),
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    await act(async () => {
      await result.current.laden()
    })

    expect(result.current.fehler).not.toBe('')
  })

  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2).
  //
  // Wer das Fenster schließt, während die Frage unterwegs ist, bekommt die
  // Antwort trotzdem — React warnte früher darüber, heute schweigt es, und
  // der Zustand landete in einem Haken, den es nicht mehr gibt.
  it('trägt nichts mehr ein, wenn niemand mehr zusieht', async () => {
    let loesen: (wert: FassungsAntwort) => void = () => undefined
    let ablehnen: (grund: unknown) => void = () => undefined

    const NachNeuerFassungSehen = vi
      .fn()
      .mockReturnValueOnce(
        new Promise((ja) => {
          loesen = ja
        }),
      )
      .mockReturnValueOnce(
        new Promise((_ja, nein) => {
          ablehnen = nein
        }),
      )

    belegeApi({ NachNeuerFassungSehen })

    const erster = renderHook(() => useAktualisierung())
    erster.unmount()

    const zweiter = renderHook(() => useAktualisierung())
    zweiter.unmount()

    await act(async () => {
      loesen(neuere)
      ablehnen(new Error('weg'))
      await Promise.resolve()
    })

    // **Der Beleg, dass die Antworten wirklich kamen**: ohne ihn prüfte dieser
    // Test eine Verneinung nach einem Ereignis, das niemand gehört hat (§3).
    expect(NachNeuerFassungSehen).toHaveBeenCalledTimes(2)
    expect(zweiter.result.current.gefragt).toBe(false)
  })

  it('übergeht einen Fehler beim Fragen still', async () => {
    belegeApi({
      NachNeuerFassungSehen: () => Promise.reject(new Error('kein Netz')),
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.gefragt).toBe(true)
    })

    expect(result.current.fassung).toBe('')
  })
})
