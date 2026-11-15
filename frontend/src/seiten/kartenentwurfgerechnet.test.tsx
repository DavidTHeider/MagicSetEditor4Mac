// **Ein geglückter Entwurf ist erledigt, auch wenn der Bestand rechnet** —
// Gesamtreview über 0.9.0, Befund 1
// (#2382).
//
// Seit #2396 speichert der Bestand, was das Wertskript rechnet. Der Editor
// verwarf, ob das Schreiben glückte, und der Entwurf wartete auf seinen
// wörtlichen Wert — nach dem Speichern stand „von anderer Seite geändert“ da.
// Die Vorrichtung ist die aus `kartenentwurf.test.tsx`, ihr `FeldSetzen`
// rechnet hier wie ein Wertskript.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Umschlag } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'

const texte = kataloge.de.bausteine.entwurf

afterEach(() => {
  raeumeApi()
})

// bestand ist, was im Bestand steht — ein anderes Programm ändert ihn.
function mitBestand(
  bestand: Record<string, string>,
  scheitert = false,
  rechnet = false,
) {
  const gesetzt = vi.fn<
    (
      setID: string,
      karteID: string,
      feld: string,
      wert: string,
    ) => Promise<Umschlag>
  >((_set, _karte, feld, wert) => {
    if (rechnet) {
      bestand[feld] = `[${wert}]`
    }

    return Promise.resolve({
      fehler: scheitert
        ? {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Dieses Feld hat sich inzwischen geändert.',
          }
        : null,
    })
  })

  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 2 }],
      }),
    Karte: () =>
      Promise.resolve({
        karte: {
          id: '5',
          nummer: 0,
          werte: Object.entries(bestand).map(([feld, wert]) => ({
            feld,
            wert,
          })),
        },
        fehler: null,
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
    Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    KarteVerweis: () =>
      Promise.resolve({ verweis: null, befunde: [], fehler: null }),
    FeldSetzen: gesetzt,
  })

  return gesetzt
}

function seite(stand: number) {
  return (
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID="5"
        verlaufsstand={stand}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>
  )
}

describe('ein Entwurf über einem Wertskript', () => {
  it('ist nach geglücktem Schreiben erledigt und zeigt den gerechneten Wert', async () => {
    const bestand: Record<string, string> = { name: 'Feuerball', typ: 'Zauber' }
    const gesetzt = mitBestand(bestand, false, true)

    const { rerender } = render(seite(0))

    const name = await screen.findByRole('textbox', { name: 'name' })
    await userEvent.type(name, ' groß')
    await act(async () => {
      name.blur()
      await Promise.resolve()
    })

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledOnce()
    })
    rerender(seite(1))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
        '[Feuerball groß]',
      )
    })
    expect(
      screen.queryByText(
        einsetzen(texte.fremdGeaendert, { wert: '[Feuerball groß]' }),
      ),
    ).toBeNull()
  })
})
