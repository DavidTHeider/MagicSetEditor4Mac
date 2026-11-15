// **Die Starthand wird beim Anfangen gezogen** — P6
// (#2374),
// Artboard `PartieStart09`: bis dahin zog jeder Sitz sie mit einem eigenen
// Knopf, und der Beginn kostete sechs Handlungen (S2). **Ziehen ist keine
// Entscheidung** — CR 103.5 (Fassung 2026-09-25) lässt jeden Spieler ziehen;
// ob er behält oder einen Mulligan nimmt, bleibt seine Wahl und wird nicht
// vorweggenommen.
import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { BefehlAntwort } from '../api'
import type { PartiefreigabeAntwort } from '../apipartie'
import { einBefehlsausgang, eineAntwort, einSitz } from '../seiten/partieprobe'
import { einEntwurf } from '../test/partieentwurf'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { usePartie } from './usePartie'

afterEach(raeumeApi)

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

const imBeginn = eineAntwort({
  fassung: '1',
  sitze: [
    einSitz({ kennung: 'sitz1', name: 'Ada', eroeffnung: 'ausstehend' }),
    einSitz({ kennung: 'sitz2', name: 'Bo', eroeffnung: 'ausstehend' }),
  ],
})

type Befehl = (
  partie: string,
  kennung: string,
  absender: string,
  art: string,
  nutzlast: Record<string, unknown>,
  erwartet: string,
) => Promise<BefehlAntwort>

function mitBefehl(befehl: Befehl) {
  const gelesen = vi.fn(() => Promise.resolve(eineAntwort({ fassung: '3' })))
  belegeApi({
    PartieAnlegen: () => Promise.resolve(imBeginn),
    PartieBefehl: befehl,
    Partie: gelesen,
  })

  return gelesen
}

describe('der Anfang', () => {
  it('zieht für jeden Sitz die Starthand und liest dann nach', async () => {
    let fassung = 1
    const befehl = vi.fn<Befehl>(() => {
      fassung += 1

      return Promise.resolve({
        ...einBefehlsausgang(),
        fassung: String(fassung),
      })
    })
    const gelesen = mitBefehl(befehl)
    const { result } = renderHook(() => usePartie(), { wrapper: raum })

    act(() => {
      result.current.anfangen(einEntwurf)
    })

    await waitFor(() => {
      expect(gelesen).toHaveBeenCalled()
    })
    // **Absender, Art und Fassung** — jeder Sitz zieht selbst, gegen die
    // Fassung, die der vorige hinterließ.
    expect(befehl.mock.calls.map((ruf) => [ruf[2], ruf[3], ruf[5]])).toEqual([
      ['sitz1', 'spieler.starthand', '1'],
      ['sitz2', 'spieler.starthand', '2'],
    ])
    expect(result.current.partie?.fassung).toBe('3')
  })

  // **Angelegt ist angelegt** (§7, Regel 3): scheitert das Ziehen, steht die
  // Partie trotzdem da — mit der Meldung, und ohne den zweiten Zug.
  it('behält die Partie, wenn das Ziehen scheitert', async () => {
    const befehl = vi.fn<Befehl>(() =>
      Promise.resolve({
        ...einBefehlsausgang(),
        fehler: {
          code: 'validierung',
          schluessel: 'validierung',
          meldung: 'nein',
        },
      }),
    )
    mitBefehl(befehl)
    const { result } = renderHook(() => usePartie(), { wrapper: raum })

    act(() => {
      result.current.anfangen(einEntwurf)
    })

    await waitFor(() => {
      expect(result.current.fehler).toBeDefined()
    })
    expect(befehl).toHaveBeenCalledTimes(1)
    expect(result.current.partie?.id).toBe(imBeginn.partie.id)
  })

  // **Ein eingeladener Sitz wartet auf sein Deck** (P9b, #2377) — ziehen
  // kann er erst, wenn sein Gast sich gesetzt hat, und das tut der selbst.
  it('zieht für einen wartenden Sitz keine Starthand', async () => {
    const befehl = vi.fn<Befehl>(() =>
      Promise.resolve({ ...einBefehlsausgang(), fassung: '2' }),
    )
    const gelesen = vi.fn(() => Promise.resolve(eineAntwort({ fassung: '2' })))
    belegeApi({
      PartieAnlegen: () =>
        Promise.resolve(
          eineAntwort({
            fassung: '1',
            sitze: [
              einSitz({
                kennung: 'sitz1',
                name: 'Ada',
                eroeffnung: 'ausstehend',
              }),
              einSitz({
                kennung: 'sitz2',
                name: 'Bo',
                eroeffnung: 'ausstehend',
                wartet: true,
              }),
            ],
          }),
        ),
      PartieBefehl: befehl,
      Partie: gelesen,
      PartieFreigeben: () =>
        Promise.resolve({
          laeuft: true,
          partie: 'p1',
          wirt: 'sitz1',
          vergeben: [],
          fehler: null,
        }),
    })
    const { result } = renderHook(() => usePartie(), { wrapper: raum })

    act(() => {
      result.current.anfangen(einEntwurf)
    })

    await waitFor(() => {
      expect(gelesen).toHaveBeenCalled()
    })
    expect(befehl.mock.calls.map((ruf) => ruf[2])).toEqual(['sitz1'])
  })

  // **Wer einlädt, will den Tisch offen haben** (P9c, #2377): der wartende
  // Sitz wird freigegeben, ohne dass jemand „Tisch öffnen“ drückt.
  it('gibt den Tisch für einen wartenden Sitz frei', async () => {
    const freigegeben = vi.fn(() =>
      Promise.resolve({
        laeuft: true,
        partie: 'p1',
        wirt: 'sitz1',
        vergeben: [],
        fehler: null,
      }),
    )
    belegeApi({
      PartieAnlegen: () =>
        Promise.resolve(
          eineAntwort({
            fassung: '1',
            sitze: [
              einSitz({
                kennung: 'sitz1',
                name: 'Ada',
                eroeffnung: 'ausstehend',
              }),
              einSitz({ kennung: 'sitz2', name: 'Bo', wartet: true }),
            ],
          }),
        ),
      PartieBefehl: () =>
        Promise.resolve({ ...einBefehlsausgang(), fassung: '2' }),
      Partie: () => Promise.resolve(eineAntwort({ fassung: '2' })),
      PartieFreigeben: freigegeben,
    })
    const { result } = renderHook(() => usePartie(), { wrapper: raum })

    act(() => {
      result.current.anfangen(einEntwurf)
    })

    await waitFor(() => {
      expect(freigegeben).toHaveBeenCalledOnce()
    })
    const [partie, fuer, gaeste] = freigegeben.mock.calls[0] as unknown as [
      string,
      string,
      string[],
    ]
    expect(gaeste).toEqual(['sitz2'])
    expect(partie).not.toBe('')
    expect(fuer).not.toBe('')
  })

  // **Nach „Partie schließen“ geht nichts mehr weiter** — Nachlauf zum
  // Zweitblick über Phase 7, R7. Die Starthand ist noch unterwegs, als der
  // Tisch geräumt wird; danach gab das Anfangen den Tisch trotzdem frei, und
  // das Nachlesen brachte ihn zurück.
  it('gibt nach dem Schließen nicht frei und liest nicht nach', async () => {
    let gezogen: (antwort: BefehlAntwort) => void = () => undefined
    const befehl = vi.fn<Befehl>(
      () =>
        new Promise((fertig) => {
          gezogen = fertig
        }),
    )
    const freigegeben = vi.fn()
    const gelesen = vi.fn(() => Promise.resolve(eineAntwort({ fassung: '3' })))
    belegeApi({
      PartieAnlegen: () =>
        Promise.resolve(
          eineAntwort({
            fassung: '1',
            sitze: [
              einSitz({
                kennung: 'sitz1',
                name: 'Ada',
                eroeffnung: 'ausstehend',
              }),
              einSitz({ kennung: 'sitz2', name: 'Bo', wartet: true }),
            ],
          }),
        ),
      PartieBefehl: befehl,
      Partie: gelesen,
      PartieFreigeben: freigegeben,
    })
    const { result } = renderHook(() => usePartie(), { wrapper: raum })

    act(() => {
      result.current.anfangen(einEntwurf)
    })
    await waitFor(() => {
      expect(befehl).toHaveBeenCalledOnce()
    })

    act(() => {
      result.current.raeumen()
    })
    await act(async () => {
      gezogen({ ...einBefehlsausgang(), fassung: '2' })
      await Promise.resolve()
    })
    await antwortAbwarten()

    expect(freigegeben).not.toHaveBeenCalled()
    expect(gelesen).not.toHaveBeenCalled()
    expect(result.current.partie).toBeUndefined()
  })

  // **Und wer während der Freigabe schließt**, bekommt den Tisch auch nicht
  // durch das Nachlesen zurück.
  it('liest nach einem Schließen während der Freigabe nicht nach', async () => {
    let freigabeFertig: (antwort: PartiefreigabeAntwort) => void = () =>
      undefined
    const gelesen = vi.fn(() => Promise.resolve(eineAntwort({ fassung: '3' })))
    belegeApi({
      PartieAnlegen: () =>
        Promise.resolve(
          eineAntwort({
            fassung: '1',
            sitze: [
              einSitz({
                kennung: 'sitz1',
                name: 'Ada',
                eroeffnung: 'ausstehend',
              }),
              einSitz({ kennung: 'sitz2', name: 'Bo', wartet: true }),
            ],
          }),
        ),
      PartieBefehl: () =>
        Promise.resolve({ ...einBefehlsausgang(), fassung: '2' }),
      Partie: gelesen,
      PartieFreigeben: () =>
        new Promise((fertig) => {
          freigabeFertig = fertig
        }),
    })
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
    await act(async () => {
      freigabeFertig({
        laeuft: true,
        partie: 'p1',
        wirt: 'sitz1',
        vergeben: [],
        fehler: null,
      })
      await Promise.resolve()
    })
    await antwortAbwarten()

    expect(gelesen).not.toHaveBeenCalled()
    expect(result.current.partie).toBeUndefined()
  })
})
