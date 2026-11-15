// **Der Schnellfilter folgt der Spaltenwahl** — Befund 1 des Zweitblicks über
// Phase G, zu FP-01
// (#2155).
//
// Ein freier Begriff sieht seit FP-01 die gerechneten Werte der **sichtbaren**
// Spalten. Wer eine Spalte aus- oder einblendete, während ein Filter stand,
// bekam die Liste neu — den Filter nicht: er zeigte die Treffer der alten
// Wahl, bis jemand den Suchtext änderte.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO, KartenlisteAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten

const zwei: KarteDTO[] = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Feuerball' }] },
  { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Eisball' }] },
]

// aufbau gibt die Liste mit einer gerechneten Spalte, sichtbar oder nicht.
function aufbau(sichtbar: boolean): KartenlisteAntwort {
  return {
    spalten: [
      {
        feld: 'name',
        aufschrift: 'Name',
        sichtbar: true,
        ausrichtung: 'links',
      },
      { feld: 'type', aufschrift: 'Typ', sichtbar, ausrichtung: 'links' },
    ],
    zeilen: zwei.map((karte) => ({
      karteID: karte.id,
      nummer: karte.nummer + 1,
      werte: { name: karte.werte[0]?.wert ?? '', type: 'Creature — Beast' },
      marken: [],
    })),
    sortierfeld: '',
    aufsteigend: true,
    ohneVorlage: false,
    fehler: null,
  }
}

describe('der Schnellfilter und die Spaltenwahl', () => {
  it('fragt neu, wenn eine Spalte umgeschaltet wird', async () => {
    let typSichtbar = true
    const KartenFiltern = vi.fn(() =>
      Promise.resolve({ kartenIDs: typSichtbar ? ['1', '2'] : [] }),
    )

    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      ...ausKarten(() => Promise.resolve({ karten: zwei })),
      Karten: () => Promise.resolve({ karten: zwei }),
      Kartenlistenaufbau: () => Promise.resolve(aufbau(typSichtbar)),
      ListenspaltenSetzen: () => {
        typSichtbar = false

        return Promise.resolve({ fehler: null })
      },
      KartenFiltern,
    } satisfies Partial<Api>)

    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Sprachraum>,
    )

    await userEvent.type(
      await screen.findByLabelText(texte.schnellfilter.bezeichnung),
      '"Creature — Beast"',
    )
    await waitFor(() => {
      expect(screen.getByRole('status', { name: '' })).toHaveTextContent(
        '2 von 2 Karten',
      )
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.listenspalten.ueberschrift,
      }),
    )
    await userEvent.click(
      within(
        screen.getByRole('region', { name: texte.listenspalten.ueberschrift }),
      ).getByLabelText('Typ'),
    )

    await waitFor(() => {
      expect(screen.getByRole('status', { name: '' })).toHaveTextContent(
        '0 von 2 Karten',
      )
    })
  })
})
