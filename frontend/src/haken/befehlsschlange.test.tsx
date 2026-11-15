// **Drei schnelle Gesten, keine verloren** — P5
// (#2373),
// Befunde C-21 und C-24: ein zweiter Befehl, während der erste unterwegs war,
// wurde verworfen (`useVorgang` als Sperre), und wer drei Länder schnell
// hintereinander tappte, sah eines getappt. Jetzt stehen die Wünsche an,
// einer ist unterwegs, und der nächste rechnet gegen die Fassung, die der
// vorige hinterlassen hat.
import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { BefehlAntwort } from '../api'
import type { PartieAntwort } from '../apipartie'
import { einBefehlsausgang, eineAntwort } from '../seiten/partieprobe'
import { einEntwurf } from '../test/partieentwurf'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { usePartie } from './usePartie'

afterEach(raeumeApi)

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// verzoegert gibt eine Fassade, deren Befehle erst auf `frei` antworten.
function verzoegert(mehr: Parameters<typeof belegeApi>[0] = {}) {
  const offen: ((antwort: BefehlAntwort) => void)[] = []
  // stand ist die Fassung, die ein Nachlesen nach dem letzten Befehl sieht.
  let stand = '1'
  const befehl = vi.fn<
    (
      id: string,
      kennung: string,
      fuer: string,
      art: string,
      nutzlast: Record<string, unknown>,
      fassung: string,
    ) => Promise<BefehlAntwort>
  >(
    () =>
      new Promise<BefehlAntwort>((fertig) => {
        offen.push(fertig)
      }),
  )
  belegeApi({
    PartieAnlegen: () => Promise.resolve(eineAntwort({ fassung: '1' })),
    Partie: () => Promise.resolve(eineAntwort({ fassung: stand })),
    PartieBefehl: befehl,
    ...mehr,
  })

  return {
    befehl,
    // frei beantwortet den ältesten wartenden Befehl.
    frei: (antwort: BefehlAntwort) => {
      if (antwort.fehler === null) {
        stand = antwort.fassung
      }
      offen.shift()?.(antwort)
    },
  }
}

async function angelegt() {
  const { result } = renderHook(() => usePartie(), { wrapper: raum })
  act(() => {
    result.current.anfangen(einEntwurf)
  })
  await waitFor(() => {
    expect(result.current.partie?.id).toBe('p1')
  })

  return result
}

describe('die Befehle einer Partie', () => {
  it('stellt drei schnelle Wünsche nacheinander, keiner geht verloren', async () => {
    const fassade = verzoegert()
    const lauf = await angelegt()

    act(() => {
      lauf.current.befehlen('objekt.tappen', { objekt: 'a' })
      lauf.current.befehlen('objekt.tappen', { objekt: 'b' })
      lauf.current.befehlen('objekt.tappen', { objekt: 'c' })
    })

    // **Einer ist unterwegs**, die anderen warten — und das ist zu sehen.
    expect(fassade.befehl).toHaveBeenCalledTimes(1)
    expect(lauf.current.wartend).toBe(3)

    await act(async () => {
      fassade.frei({ ...einBefehlsausgang(), fassung: '2' })
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(fassade.befehl).toHaveBeenCalledTimes(2)
    })
    await act(async () => {
      fassade.frei({ ...einBefehlsausgang(), fassung: '3' })
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(fassade.befehl).toHaveBeenCalledTimes(3)
    })
    await act(async () => {
      fassade.frei({ ...einBefehlsausgang(), fassung: '4' })
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(lauf.current.wartend).toBe(0)
    })

    // **Jeder gegen die Fassung, die der vorige hinterlassen hat** — nicht
    // alle drei gegen die, die beim ersten Klick zu sehen war.
    expect(fassade.befehl.mock.calls.map((ruf) => [ruf[4], ruf[5]])).toEqual([
      [{ objekt: 'a' }, '1'],
      [{ objekt: 'b' }, '2'],
      [{ objekt: 'c' }, '3'],
    ])
  })

  // **Eine Ablehnung bleibt sichtbar**, auch wenn danach der nächste Wunsch
  // hinausgeht — sonst wäre sie still verloren.
  it('lässt eine Ablehnung stehen, während die Folge weiterläuft', async () => {
    const fassade = verzoegert()
    const lauf = await angelegt()

    act(() => {
      lauf.current.befehlen('objekt.land', { objekt: 'a' })
      lauf.current.befehlen('objekt.tappen', { objekt: 'b' })
    })

    await act(async () => {
      fassade.frei({
        ...einBefehlsausgang(),
        fehler: {
          code: 'validierung',
          schluessel: 'validierung',
          meldung: 'in diesem Zug ist schon ein Land gespielt (CR 305.2)',
        },
      })
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(fassade.befehl).toHaveBeenCalledTimes(2)
    })

    expect(lauf.current.fehler).toBe(kataloge.de.fehler.validierung)
  })

  // **Ein Wechsel räumt die Schlange** (§7, Regel 1): wartende Wünsche
  // gehören zur alten Partie.
  it('verwirft wartende Wünsche, wenn die Partie geräumt wird', async () => {
    const fassade = verzoegert()
    const lauf = await angelegt()

    act(() => {
      lauf.current.befehlen('objekt.tappen', { objekt: 'a' })
      lauf.current.befehlen('objekt.tappen', { objekt: 'b' })
    })
    act(() => {
      lauf.current.raeumen()
    })
    await act(async () => {
      fassade.frei({ ...einBefehlsausgang(), fassung: '2' })
      await Promise.resolve()
    })

    await waitFor(() => {
      expect(lauf.current.wartend).toBe(0)
    })
    expect(fassade.befehl).toHaveBeenCalledTimes(1)
  })

  // **Läuft gerade ein Halt**, geht ein Wunsch nicht hinaus — das wird
  // gesagt statt verschwiegen, und die Schlange bleibt nicht hängen.
  it('meldet einen Wunsch, der neben einem Halt nicht hinausging', async () => {
    let halt: (antwort: PartieAntwort) => void = () => undefined
    const fassade = verzoegert({
      PartiePausieren: () =>
        new Promise<PartieAntwort>((fertig) => {
          halt = fertig
        }),
    })
    const lauf = await angelegt()

    act(() => {
      lauf.current.wechseln('pausieren')
    })
    act(() => {
      lauf.current.befehlen('objekt.tappen', { objekt: 'a' })
    })

    await waitFor(() => {
      expect(lauf.current.fehler).toBe(
        kataloge.de.seiten.spielen.wunschNichtGestellt,
      )
    })
    expect(fassade.befehl).not.toHaveBeenCalled()
    expect(lauf.current.wartend).toBe(0)

    await act(async () => {
      halt(eineAntwort({ fassung: '2' }))
      await Promise.resolve()
    })
  })
})
