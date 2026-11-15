// **Die Liste der Gegenüber faltet, die offene Zeile bleibt im Bild** —
// #2035.
//
// Gemessen am nativen Fenster (`uxr16`): mit acht Gegenübern stand *Einstellungen
// › Verbindungen* bei 21 Bedienelementen im Bild (R2) und rollte 3,2 Höhen
// tief (R4). Die Liste wächst mit jedem Gegenüber. Sie zeigt jetzt vorn fünf
// und den Rest auf Wunsch — und **die offene Zeile immer**, auch wenn sie
// hinten steht: ein eben aufgenommener Fund schlägt mit seinem Vergleich auf
// und darf nicht hinter der Falte verschwinden.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Peerkontakte } from './Peerkontakte'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import {
  achtGegenueber,
  fund,
  gegenueber,
  zwecke,
} from '../test/gegenueberbestand'

import type { Peerkontakt } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik
const worte = kataloge.de.bausteine.listenansicht

afterEach(() => {
  raeumeApi()
})

function liste(kontakte: Peerkontakt[]) {
  return Promise.resolve({ kontakte, zwecke, fehler: null })
}

// zeilen nennt die gezeigten Gegenüber beim Namen — über den Knopf der Zeile.
function zeilen() {
  const vorsatz = new RegExp(
    `^(${eigene.gegenueber.vergleichen}|${eigene.gegenueber.verwalten}): `,
  )

  return screen
    .getAllByRole('button', { name: vorsatz })
    .map((knopf) => knopf.getAttribute('aria-label')?.replace(vorsatz, ''))
}

describe('die Liste der Gegenüber', () => {
  it('zeigt vorn fünf und den Rest auf Wunsch', async () => {
    belegeApi({ Peerkontakte: () => liste(achtGegenueber()) })

    render(
      <Sprachraum sprache="de">
        <Peerkontakte eingebettet />
      </Sprachraum>,
    )

    await screen.findByText('Anna')
    expect(zeilen()).toEqual(['Anna', 'Bert', 'Cleo', 'Dora', 'Emil'])

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(worte.weitere, { anzahl: '3' }),
      }),
    )

    expect(zeilen()).toHaveLength(8)
  })

  it('hält einen eben aufgenommenen Fund im Bild, auch hinter der Falte', async () => {
    let stand = achtGegenueber()

    belegeApi({
      Peerkontakte: () => liste(stand),
      Nachbarn: () =>
        Promise.resolve({
          nachbarn: [fund('Ida')],
          verworfen: 0,
          fehler: null,
        }),
      PeerkontaktAufnehmen: (kennung: string, name: string) => {
        stand = [
          ...stand,
          gegenueber('Ida', {
            kennung,
            eigenerName: '',
            selbstgewaehlterName: name,
            vergleichscode: '12345 67890 12345 67890 12345 67890',
          }),
        ]

        return liste(stand)
      },
      PeerkontaktAdresse: () => liste(stand),
    })

    render(
      <Sprachraum sprache="de">
        <Peerkontakte eingebettet />
      </Sprachraum>,
    )

    const hand = userEvent.setup()
    await hand.click(
      await screen.findByRole('button', { name: eigene.nachbarnSuchen }),
    )
    await hand.click(
      await screen.findByRole('button', { name: eigene.nachbarnAufnehmen }),
    )

    // **Der neunte steht hinter der Falte — und trotzdem da, offen.**
    expect(
      await screen.findByRole('button', {
        name: `${eigene.gegenueber.vergleichen}: Idas Mac`,
      }),
    ).toHaveAttribute('aria-expanded', 'true')
    expect(
      screen.getByText('12345 67890 12345 67890 12345 67890'),
    ).toBeVisible()
  })
})
