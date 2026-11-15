// **Ein Leser schneidet nichts aus und fügt nichts ein, auch nicht über die
// Tasten** — #2288
// (#2288).
//
// Die Einträge der Handlungszeile sind gesperrt; ⌘X und ⌘V in der
// Kartenspalte gingen denselben Weg ohne Sperre. ⌘C bleibt — Kopieren liest
// nur.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { waehleMehrere } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const zwei: KarteDTO[] = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Feuerball' }] },
  { id: '2', nummer: 0, werte: [{ feld: 'name', wert: 'Mirri' }] },
]

describe('Die Zwischenablage für einen Leser', () => {
  it('kopiert, schneidet aber nicht aus und fügt nicht ein', async () => {
    const ohneFehler = { fehler: null }
    const kopieren = vi.fn(() => Promise.resolve(ohneFehler))
    const ausschneiden = vi.fn(() => Promise.resolve(ohneFehler))
    const einfuegen = vi.fn(() =>
      Promise.resolve({ ...ohneFehler, karten: [], befunde: [] }),
    )

    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      ...ausKarten(() => Promise.resolve({ karten: zwei })),
      Karten: () => Promise.resolve({ karten: zwei }),
      AblageKopieren: kopieren,
      AblageAusschneiden: ausschneiden,
      AblageEinfuegen: einfuegen,
    })

    render(
      <Sprachraum sprache="de">
        <Schreibrechtraum nurLesend>
          <Karten
            karteGewaehlt={vi.fn()}
            setID="1"
            karteID={undefined}
            verlaufsstand={0}
            verlaufsfehler={undefined}
            bestandGeaendert={vi.fn()}
          />
        </Schreibrechtraum>
      </Sprachraum>,
    )
    await waehleMehrere()

    const [haken] = await screen.findAllByRole('checkbox', {
      name: new RegExp(kataloge.de.seiten.karten.massen.auswaehlen),
    })

    if (haken === undefined) {
      throw new Error('kein Ankreuzfeld in der Liste')
    }

    await userEvent.click(haken)
    fireEvent.keyDown(haken, { key: 'x', metaKey: true })
    fireEvent.keyDown(haken, { key: 'v', metaKey: true })

    // **Erst der Beleg, dass jemand zuhört** (§3): ⌘C geht durch.
    fireEvent.keyDown(haken, { key: 'c', metaKey: true })
    await waitFor(() => {
      expect(kopieren).toHaveBeenCalledTimes(1)
    })

    expect(ausschneiden).not.toHaveBeenCalled()
    expect(einfuegen).not.toHaveBeenCalled()
  })
})
