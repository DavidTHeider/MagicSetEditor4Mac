// **Der Tisch im Fenster** — Werkstück G3 von Phase 5.5
// (#1460).
import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { einePartie } from '../seiten/partieprobe'
import { tischplan } from './tischplan'
import { useTisch } from './useTisch'

import type { Tischmalerei } from './maler'
import type { Tischanschluss } from './useTisch'
import { einMaler as einMalerOhne } from '../test/malerprobe'

const plan = tischplan(einePartie())

const einMaler = () => einMalerOhne({ breite: 800, hoehe: 600 })

// ohneLeinwand ist der Fall, in dem der Aufrufer noch keine hat.
const ohneLeinwand = (malerei: Tischanschluss['malerei']): Tischanschluss => ({
  plan,
  bild: () => undefined,
  malerei,
  leinwand: { current: null },
})

const anschluss = (malerei: Tischanschluss['malerei']): Tischanschluss => ({
  ...ohneLeinwand(malerei),
  leinwand: { current: document.createElement('canvas') },
})

// haenge hängt den Haken an eine echte Leinwand — der Haken meldet sonst
// „kein-renderer", und das ist ein anderer Fall.
const haenge = (anschlussdaten: Tischanschluss) =>
  renderHook((eigenschaften: Tischanschluss) => useTisch(eigenschaften), {
    initialProps: anschlussdaten,
  })

describe('useTisch', () => {
  it('meldet, dass es keine Leinwand gibt', async () => {
    const lauf = haenge(
      ohneLeinwand(() => Promise.resolve({ art: 'ohne', grund: 'kein-webgl' })),
    )

    await waitFor(() => {
      expect(lauf.result.current.hindernis).toBe('kein-renderer')
    })
  })

  it('nennt das Hindernis, das die Naht meldet', async () => {
    const lauf = haenge(
      anschluss(() => Promise.resolve({ art: 'ohne', grund: 'kein-webgl' })),
    )

    await waitFor(() => {
      expect(lauf.result.current.hindernis).toBe('kein-webgl')
    })
  })

  it('passt den Tisch einmal ein und zeichnet ihn dann', async () => {
    const maler = einMaler()
    const lauf = haenge(
      anschluss(() => Promise.resolve({ art: 'da', maler } as Tischmalerei)),
    )

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })
    expect(lauf.result.current.kamera.massstab).toBeGreaterThan(0)
    expect(lauf.result.current.hindernis).toBeUndefined()
  })

  it('zeichnet neu, wenn die Kamera sich bewegt', async () => {
    const maler = einMaler()
    const lauf = haenge(
      anschluss(() => Promise.resolve({ art: 'da', maler } as Tischmalerei)),
    )
    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    const vorher = maler.zeichne.mock.calls.length
    act(() => {
      lauf.result.current.schwenke(30, 0)
    })
    act(() => {
      lauf.result.current.zoome(2, { x: 100, y: 100 })
    })
    act(() => {
      lauf.result.current.drehe(90)
    })
    act(() => {
      lauf.result.current.passeEin()
    })

    expect(maler.zeichne.mock.calls.length).toBeGreaterThan(vorher)
    expect(lauf.result.current.kamera.drehung).toBe(90)
  })

  it('meldet den Verlust des Grafikkontexts', async () => {
    const maler = einMaler()
    const lauf = haenge(
      anschluss(() => Promise.resolve({ art: 'da', maler } as Tischmalerei)),
    )
    await waitFor(() => {
      expect(maler.beiVerlust).toHaveBeenCalled()
    })

    act(() => {
      maler.verlieren()
    })

    expect(lauf.result.current.hindernis).toBe('verloren')
  })

  // **Gefunden in der Sichtprüfung am 2026-09-16**: React hängt Effekte im
  // Entwicklungslauf doppelt ein, `destroy()` feuert `webglcontextlost`, und
  // der Horcher des alten Malers meldete „verloren" auf den neuen Zustand.
  it('lässt einen Verlust nach dem Aushängen unbeachtet', async () => {
    const maler = einMaler()
    const lauf = haenge(anschluss(() => Promise.resolve({ art: 'da', maler })))
    await waitFor(() => {
      expect(maler.beiVerlust).toHaveBeenCalled()
    })

    lauf.unmount()
    act(() => {
      maler.verlieren()
    })

    expect(lauf.result.current.hindernis).toBeUndefined()
  })

  // **Die Auswahl gehört diesem Fenster und geht trotzdem an den Maler** —
  // sie ändert am Spiel nichts, muss aber zu sehen sein.
  it('reicht die gewählte Karte an den Maler weiter', async () => {
    const maler = einMaler()
    const lauf = haenge({
      ...anschluss(() => Promise.resolve({ art: 'da', maler })),
      gewaehlt: 'k7',
    })

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    expect(maler.zeichne).toHaveBeenLastCalledWith(
      expect.objectContaining({ gewaehlt: 'k7' }),
    )
    lauf.unmount()
  })

  // **Und die Stufe der Beschriftung** — `tisch.beschriftung`, seit dem
  // 2026-09-24. Sie ändert am Plan nichts, nur daran, was der Maler
  // darauf schreibt.
  it('reicht die Stufe der Beschriftung an den Maler weiter', async () => {
    const maler = einMaler()
    const lauf = haenge({
      ...anschluss(() => Promise.resolve({ art: 'da', maler })),
      beschriftungsstufe: 'gross',
    })

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    expect(maler.zeichne).toHaveBeenLastCalledWith(
      expect.objectContaining({ beschriftungsstufe: 'gross' }),
    )
    lauf.unmount()
  })

  // **Der Abzug greift auf den laufenden Maler** — PLAY-05.AK-03
  // (#1713).
  it('macht einen Abzug über den laufenden Maler', async () => {
    const maler = einMaler()
    const lauf = haenge(anschluss(() => Promise.resolve({ art: 'da', maler })))

    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    const ergebnis = await lauf.result.current.abzug({ art: 'tisch' }, 150)

    expect(ergebnis).toEqual(
      expect.objectContaining({ bild: 'data:image/png;base64,AAA' }),
    )

    expect(maler.abzug).toHaveBeenCalled()
    lauf.unmount()
  })

  // **Und wer klickt, bevor der Maler da ist, hört, dass er kommt.** Ein
  // Maler entsteht asynchron; zwischen dem ersten Bild und ihm liegt eine
  // Spanne, in der `maler.current` noch nichts ist.
  //
  // **Hier stand „denselben Grund" wie ohne Grafikkarte**, und das war
  // falsch, sobald die Spanne häufig wurde: seit dem Nachlauf über Phase 6.9
  // wird der Tisch bei jedem Sichtwechsel neu aufgebaut, und wer gleich
  // danach ein Tischbild wollte, las, es gebe keine Grafikkarte.
  it('meldet einen Abzug vor dem ersten Maler als Aufbau', async () => {
    const lauf = haenge(
      anschluss(() => Promise.resolve({ art: 'da', maler: einMaler() })),
    )

    expect(await lauf.result.current.abzug({ art: 'tisch' }, 150)).toEqual({
      fehler: 'aufbau',
    })

    lauf.unmount()
  })

  // **Und ohne Maler gibt es keinen Abzug, sondern einen Grund** — die
  // Zonenliste daneben steht weiter, und wer das Bild will, soll hören,
  // warum es keines gibt (§8).
  it('meldet ohne Maler, dass es kein Bild gibt', async () => {
    const lauf = haenge(
      ohneLeinwand(() => Promise.resolve({ art: 'ohne', grund: 'kein-webgl' })),
    )

    await waitFor(() => {
      expect(lauf.result.current.hindernis).toBe('kein-renderer')
    })

    expect(await lauf.result.current.abzug({ art: 'tisch' }, 150)).toEqual({
      fehler: 'keinMaler',
    })
  })

  it('räumt den Maler beim Aushängen', async () => {
    const maler = einMaler()
    const lauf = haenge(
      anschluss(() => Promise.resolve({ art: 'da', maler } as Tischmalerei)),
    )
    await waitFor(() => {
      expect(maler.zeichne).toHaveBeenCalled()
    })

    lauf.unmount()

    expect(maler.raeume).toHaveBeenCalled()
  })

  it('bedient die Kamera auch ohne Maler, ohne darüber zu stolpern', async () => {
    const lauf = haenge(
      ohneLeinwand(() => Promise.resolve({ art: 'ohne', grund: 'kein-webgl' })),
    )
    await waitFor(() => {
      expect(lauf.result.current.hindernis).toBe('kein-renderer')
    })

    act(() => {
      lauf.result.current.zoome(2, { x: 0, y: 0 })
    })
    act(() => {
      lauf.result.current.passeEin()
    })

    expect(lauf.result.current.masse).toEqual({ breite: 0, hoehe: 0 })
  })

  // **Regel 2 aus §7**: eine Antwort gehört zu dem Stand, aus dem sie stammt.
  it('räumt einen Maler, der nach dem Aushängen fertig wird', async () => {
    const maler = einMaler()
    let liefere: ((eines: Tischmalerei) => void) | undefined
    const lauf = haenge(
      anschluss(
        () =>
          new Promise<Tischmalerei>((erfuellen) => {
            liefere = erfuellen
          }),
      ),
    )

    lauf.unmount()
    await act(async () => {
      liefere?.({ art: 'da', maler })
      await Promise.resolve()
    })

    expect(maler.raeume).toHaveBeenCalled()
    expect(maler.zeichne).not.toHaveBeenCalled()
  })

  it('verwirft auch ein Hindernis, das nach dem Aushängen eintrifft', async () => {
    let liefere: ((eines: Tischmalerei) => void) | undefined
    const lauf = haenge(
      anschluss(
        () =>
          new Promise<Tischmalerei>((erfuellen) => {
            liefere = erfuellen
          }),
      ),
    )

    lauf.unmount()

    // **Auf die Antwort warten, nicht auf den Ruf** (§3): ohne das `await`
    // stünde die Zusicherung vor dem `.then`, und sie träfe zu, weil noch
    // nichts geschehen ist.
    await act(async () => {
      liefere?.({ art: 'ohne', grund: 'kein-webgl' })
      await Promise.resolve()
    })

    expect(lauf.result.current.hindernis).toBeUndefined()
  })
})
