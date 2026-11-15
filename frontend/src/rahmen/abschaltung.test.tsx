// **Ein abgeschalteter Bereich verschwindet aus Leiste und Palette** —
// OPS-10.AK-02 (#1908).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Einstellungsraum } from '../haken/einstellungsraum'
import { useEinstellungen } from '../haken/useEinstellungen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Rahmen } from './Rahmen'

const bereiche = kataloge.de.allgemein.bereiche
const navigation = kataloge.de.allgemein.navigation

afterEach(() => {
  raeumeApi()
})

function Wurzel({ children }: { children: ReactNode }) {
  return (
    <Einstellungsraum stand={useEinstellungen()}>{children}</Einstellungsraum>
  )
}

function zeige(spielen: string) {
  belegeApi({
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'schalter.spielen',
            wert: spielen,
            wirksam: spielen,
            werte: ['an', 'aus'],
            bereich: 'schalter',
          },
        ],
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Wurzel>
        <Rahmen
          aktuell="start"
          wechseln={vi.fn()}
          seitenleiste={{ offen: true, umschalten: vi.fn() }}
        >
          <p>Inhalt</p>
        </Rahmen>
      </Wurzel>
    </Sprachraum>,
  )
}

function leiste() {
  return within(
    screen.getByRole('navigation', { name: navigation.bezeichnung }),
  )
}

describe('ein abgeschalteter Bereich', () => {
  it('steht nicht mehr in der Seitenleiste', async () => {
    zeige('aus')

    await waitFor(() => {
      expect(
        leiste().queryByRole('button', { name: new RegExp(bereiche.spielen) }),
      ).toBeNull()
    })
    expect(
      leiste().getByRole('button', { name: new RegExp(bereiche.sammeln) }),
    ).toBeVisible()
  })

  it('steht eingeschaltet da', async () => {
    zeige('an')

    expect(
      await leiste().findByRole('button', {
        name: new RegExp(bereiche.spielen),
      }),
    ).toBeVisible()
  })

  it('steht nicht in der Palette', async () => {
    zeige('aus')
    await waitFor(() => {
      expect(
        leiste().queryByRole('button', { name: new RegExp(bereiche.spielen) }),
      ).toBeNull()
    })

    await userEvent.keyboard('{Meta>}k{/Meta}')
    const seiten = await screen.findByRole('list', {
      name: kataloge.de.allgemein.palette.seiten,
    })

    expect(
      within(seiten).queryByText(kataloge.de.seiten.decks.titel),
    ).toBeNull()
    expect(
      within(seiten).getByText(kataloge.de.seiten.sammlung.titel),
    ).toBeVisible()
  })
})

// **Die Reife steht am Bereich selbst** — OPS-10.AK-04: *Vorschau* und
// *auslaufend* als Marke, *stabil* ohne.
describe('die Reife eines Bereichs', () => {
  it('steht als Marke an seinem Eintrag', async () => {
    belegeApi({
      Schalter: () =>
        Promise.resolve({
          schalter: [
            {
              kennung: 'spielen',
              art: 'bereich',
              reife: 'vorschau',
              vorgabe: 'an',
              an: true,
            },
            {
              kennung: 'handeln',
              art: 'bereich',
              reife: 'auslaufend',
              vorgabe: 'an',
              an: true,
            },
            {
              kennung: 'sammeln',
              art: 'bereich',
              reife: 'stabil',
              vorgabe: 'an',
              an: true,
            },
            {
              kennung: 'ressourcen',
              art: 'bereich',
              reife: 'unbekannt',
              vorgabe: 'an',
              an: true,
            },
          ],
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Rahmen
          aktuell="start"
          wechseln={vi.fn()}
          seitenleiste={{ offen: true, umschalten: vi.fn() }}
        >
          <p>Inhalt</p>
        </Rahmen>
      </Sprachraum>,
    )

    const reife = kataloge.de.seiten.einstellungen.schalter.reife
    expect(await leiste().findByText(reife.vorschau)).toBeVisible()
    expect(leiste().getByText(reife.auslaufend)).toBeVisible()
    expect(leiste().getByText('unbekannt')).toBeVisible()
    expect(leiste().queryByText(reife.stabil)).toBeNull()
  })
})
