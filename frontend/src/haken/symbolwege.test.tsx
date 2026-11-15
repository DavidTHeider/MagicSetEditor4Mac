// **Eine überholte Antwort trägt sich nicht ein** — §7 Regel 2 an den
// Symbolwegen, Befund 2 des Zweitblicks über Phase 4.11.
//
// # Warum eigens
//
// Der Wächter selbst steht in `useZeichenschutz.test.tsx`; dass die Seite ihn
// **benutzt**, in `symbolgleichzeitig.test.tsx`. Was hier steht, ist der
// Zweig, den beide nicht betreten: eine Antwort, die **nach** dem Abräumen
// eintrifft.
//
// **Sie kommt vor**, sobald jemand während eines laufenden Rufs die Zeichnung
// wechselt — ein anderes Symbol, ein anderes Set.
import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { SymbolteilDTO } from '../antworten'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { usePunktwege } from './usePunktwege'
import { useTeilliste } from './useTeilliste'
import { useSymmetrie } from './useSymmetrie'
import { useZeichenschutz } from './useZeichenschutz'

// teile ist eine Zeichnung mit einem Teil und drei Punkten.
const teile = [
  {
    art: 'Form',
    name: 'Dreieck',
    kombination: 'overlap',
    punkte: [0, 1, 2].map(() => ({
      x: 0.5,
      y: 0.5,
      griffDavorX: 0,
      griffDavorY: 0,
      griffDanachX: 0,
      griffDanachY: 0,
      kurveDanach: false,
      sperre: 'free',
    })),
    spiegelart: '',
    kopien: 0,
    mitteX: 0,
    mitteY: 0,
    griffX: 0,
    griffY: 0,
    teile: [],
  },
] as unknown as SymbolteilDTO[]

// antwort ist eine geglückte Antwort mit einer geänderten Zeichnung.
const antwort = { teile: [], name: '', fehler: null }

describe('Eine überholte Antwort', () => {
  it('trägt sich am Punktweg nicht ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ PunktEinfuegen: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        schutz,
        wege: usePunktwege({
          teile,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    act(() => {
      void result.current.wege.punktEinfuegen(0, 0)

      // **Mitten im Ruf wechselt das Bezugsobjekt.**
      result.current.schutz.raeume()
    })

    // **Erst muss die Antwort da gewesen sein.** Vorher ist „nichts
    // eingetragen" keine Aussage, sondern eine Selbstverständlichkeit (§3).
    await antwortAbwarten()

    expect(eintragen).not.toHaveBeenCalled()

    raeumeApi()
  })

  // **Der Beleg, dass jemand zuhört**: derselbe Weg ohne Abräumen trägt ein.
  // Ohne ihn wäre die Verneinung darüber auch dann grün, wenn der Ruf gar
  // nicht ankäme (§3).
  it('trägt sich ohne Abräumen ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ PunktEinfuegen: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        wege: usePunktwege({
          teile,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    await act(async () => {
      await result.current.wege.punktEinfuegen(0, 0)
    })

    expect(eintragen).toHaveBeenCalledTimes(1)

    raeumeApi()
  })

  it('trägt sich an der Symmetrie nicht ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ SymmetrieAnlegen: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        schutz,
        wege: useSymmetrie({
          teile,
          gewaehlt: 0,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    act(() => {
      result.current.wege.anlegen()
      result.current.schutz.raeume()
    })

    // **Erst muss die Antwort da gewesen sein.** Vorher ist „nichts
    // eingetragen" keine Aussage, sondern eine Selbstverständlichkeit (§3).
    await antwortAbwarten()

    expect(eintragen).not.toHaveBeenCalled()

    raeumeApi()
  })

  it('trägt sich an der Symmetrie ohne Abräumen ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ SymmetrieAnlegen: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        wege: useSymmetrie({
          teile,
          gewaehlt: 0,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    act(() => {
      result.current.wege.anlegen()
    })

    await waitFor(() => {
      expect(eintragen).toHaveBeenCalledTimes(1)
    })

    raeumeApi()
  })

  it('trägt sich an der Teilliste nicht ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ TeileGruppieren: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        schutz,
        wege: useTeilliste({
          teile,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    act(() => {
      result.current.wege.gruppieren()
      result.current.schutz.raeume()
    })

    // **Erst muss die Antwort da gewesen sein.** Vorher ist „nichts
    // eingetragen" keine Aussage, sondern eine Selbstverständlichkeit (§3).
    await antwortAbwarten()

    expect(eintragen).not.toHaveBeenCalled()

    raeumeApi()
  })

  it('trägt sich an der Teilliste ohne Abräumen ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ TeileGruppieren: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        wege: useTeilliste({
          teile,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    act(() => {
      result.current.wege.gruppieren()
    })

    await waitFor(() => {
      expect(eintragen).toHaveBeenCalledTimes(1)
    })

    raeumeApi()
  })
})

// **Auch der Griffzug hört auf den gemeinsamen Stand** — Nachlauf über Phase
// 4.11.
//
// # Der Befund
//
// `GriffZiehen` schickt wie die anderen den **ganzen** Stand hinaus und setzt
// die Antwort ganz ein, benutzte dafür aber nur seinen **eigenen** Lauf. Der
// ordnet die Antworten innerhalb eines Zuges; er weiß nichts davon, dass
// inzwischen ein anderes Symbol offen ist. Eine verspätete Griffantwort galt
// in ihrem privaten Lauf weiter als aktuell.
describe('Ein Griffzug', () => {
  it('trägt nach einem Abräumen nichts mehr ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ GriffZiehen: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        schutz,
        wege: usePunktwege({
          teile,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    act(() => {
      result.current.wege.griffAuf(0, 0, true, 0.5, 0.5)
      result.current.schutz.raeume()
    })

    // **Erst muss die Antwort da gewesen sein.** Vorher ist „nichts
    // eingetragen" keine Aussage, sondern eine Selbstverständlichkeit (§3).
    await antwortAbwarten()

    expect(eintragen).not.toHaveBeenCalled()

    raeumeApi()
  })

  // **Der Beleg, dass jemand zuhört.**
  it('trägt ohne Abräumen ein', async () => {
    const eintragen = vi.fn()

    belegeApi({ GriffZiehen: vi.fn().mockResolvedValue(antwort) })

    const { result } = renderHook(() => {
      const schutz = useZeichenschutz()

      return {
        wege: usePunktwege({
          teile,
          eintragen,
          klagen: () => undefined,
          schutz,
        }),
      }
    })

    act(() => {
      result.current.wege.griffAuf(0, 0, true, 0.5, 0.5)
    })

    await waitFor(() => {
      expect(eintragen).toHaveBeenCalledTimes(1)
    })

    raeumeApi()
  })
})
