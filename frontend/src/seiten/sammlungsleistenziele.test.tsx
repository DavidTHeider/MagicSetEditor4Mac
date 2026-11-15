// **Die Leiste über der Sammlung führt hin** — AP-9
// (#2366):
// „Exemplare“ nimmt Suche, Ausführung und Marken zurück, „Wert“ führt zur
// Übersicht im Navigator — auch, wenn er eingeklappt war.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import {
  belegeSammlung,
  bestand,
  texte,
  zeigeSammlung,
} from '../test/sammlungsseite'

afterEach(() => {
  raeumeApi()
})

function einstellungen(navigator: string) {
  const werte = {
    'arbeitsplatz.navigator': navigator,
    'arbeitsplatz.inspektor': 'sichtbar',
    'arbeitsplatz.navigatorbreite': '240',
    'arbeitsplatz.inspektorbreite': '320',
  }

  return () =>
    Promise.resolve({
      einstellungen: Object.entries(werte).map(([schluessel, wert]) => ({
        schluessel,
        wert,
        wirksam: wert,
        werte: [],
        bereich: 'arbeitsplatz',
      })),
    })
}

function zeige(navigator = 'sichtbar') {
  const EinstellungSetzen = vi.fn().mockResolvedValue({})
  belegeSammlung(bestand(3), {
    Einstellungen: einstellungen(navigator),
    EinstellungSetzen,
  })
  zeigeSammlung()

  return { EinstellungSetzen }
}

const leiste = async () =>
  within(await screen.findByRole('toolbar', { name: texte.leiste.bezeichnung }))

describe('die Ziele der Leiste', () => {
  it('führt vom Wert zur Übersicht', async () => {
    zeige()

    await userEvent.click(
      await (
        await leiste()
      ).findByRole('button', { name: new RegExp(`^${texte.leiste.wert} `) }),
    )

    expect(
      screen.getByRole('heading', { name: texte.uebersichtTitel }),
    ).toHaveFocus()
  })

  it('nimmt mit den Exemplaren die Suche zurück', async () => {
    zeige()
    const suche = await screen.findByRole('searchbox', { name: texte.suche })
    await userEvent.type(suche, 'Wald')
    expect(suche).toHaveValue('Wald')

    await userEvent.click(
      await (
        await leiste()
      ).findByRole('button', {
        name: new RegExp(`^${texte.leiste.exemplare} `),
      }),
    )

    expect(suche).toHaveValue('')
  })

  it('klappt den Navigator auf, wenn er eingeklappt war', async () => {
    const { EinstellungSetzen } = zeige('verborgen')
    // **Erst, wenn die Einstellungen gelten**: davor steht der Navigator
    // nach der Vorgabe offen.
    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { name: texte.uebersichtTitel }),
      ).toBeNull()
    })

    await userEvent.click(
      await (
        await leiste()
      ).findByRole('button', { name: new RegExp(`^${texte.leiste.wert} `) }),
    )

    expect(
      await screen.findByRole('heading', { name: texte.uebersichtTitel }),
    ).toHaveFocus()
    expect(EinstellungSetzen).toHaveBeenCalledWith(
      'arbeitsplatz.navigator',
      'sichtbar',
    )
  })
})
