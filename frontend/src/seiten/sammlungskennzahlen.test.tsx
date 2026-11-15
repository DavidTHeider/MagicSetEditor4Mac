// **Die Kennzahlen der Sammlung** — AP-9
// (#2366),
// Entwurf `Bereiche09`: Exemplare, Wert, ohne Preis — jede Zahl mit Ziel,
// **Währung und ältester Preis sichtbar, kein Trend ohne Reihe**.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { Portfolioantwort } from '../api'
import { geldAusText } from '../geld'
import type { Ladezustand } from '../haken/useLaden'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Sammlungskennzahlen } from './Sammlungskennzahlen'

const texte = kataloge.de.seiten.sammlung

function portfolio(
  teil: Partial<Ladezustand<Portfolioantwort>> = {},
): Ladezustand<Portfolioantwort> {
  return {
    daten: {
      zeilen: [
        {
          waehrung: 'EUR',
          wert: '486.00',
          exemplare: 10,
          aeltesterPreis: '2026-10-14T08:00:00Z',
          alterStunden: 72,
          quelle: 'Scryfall',
        },
        {
          waehrung: 'USD',
          wert: '12.50',
          exemplare: 2,
          aeltesterPreis: '2026-10-16T08:00:00Z',
          alterStunden: 5,
          quelle: 'Scryfall',
        },
      ],
      ohnePreis: 3,
      fehler: null,
    },
    laedt: false,
    fehler: undefined,
    nachladen: vi.fn(),
    ...teil,
  }
}

function zeige(p: Ladezustand<Portfolioantwort>, stueck = 15) {
  const ziele = { zeigeAlle: vi.fn(), zeigePortfolio: vi.fn() }
  render(
    <Sprachraum sprache="de">
      <Sammlungskennzahlen
        stueck={stueck}
        insgesamt={7}
        leiste={{ portfolio: p, zeigePortfolio: ziele.zeigePortfolio }}
        zeigeAlle={ziele.zeigeAlle}
      />
    </Sprachraum>,
  )

  return ziele
}

const leiste = () =>
  within(screen.getByRole('toolbar', { name: texte.leiste.bezeichnung }))

describe('die Leiste der Sammlung', () => {
  it('nennt Exemplare, Wert je Währung und was ohne Preis blieb', async () => {
    const ziele = zeige(portfolio())

    await userEvent.click(
      leiste().getByRole('button', {
        name: `${texte.leiste.exemplare} 15 7 Einträge`,
      }),
    )
    expect(ziele.zeigeAlle).toHaveBeenCalledTimes(1)

    const wert = `${geldAusText('486.00', 'EUR', 'de')} · ${geldAusText('12.50', 'USD', 'de')}`
    await userEvent.click(
      leiste().getByRole('button', {
        name: `${texte.leiste.wert} ${wert} ältester Preis vor 3 Tagen`,
      }),
    )
    await userEvent.click(
      leiste().getByRole('button', { name: `${texte.leiste.ohnePreis} 3` }),
    )
    expect(ziele.zeigePortfolio).toHaveBeenCalledTimes(2)
  })

  // **Kein erfundener Wert**: ohne Preis steht ein Satz, keine Null in
  // einer Währung, die niemand gewählt hat.
  it('sagt „kein Preis“, solange keiner da ist', () => {
    zeige(portfolio({ daten: { zeilen: [], ohnePreis: 15, fehler: null } }))
    expect(leiste().getByText(texte.leiste.keinPreis)).toBeInTheDocument()
  })

  it('sagt „kein Preis“ auch, wenn das Portfolio nicht zu lesen war', () => {
    zeige(portfolio({ daten: undefined, fehler: 'kaputt' }))
    expect(leiste().getByText(texte.leiste.keinPreis)).toBeInTheDocument()
    expect(leiste().getByText(texte.leiste.nichtGelesen)).toBeInTheDocument()
    expect(
      screen.getByRole('toolbar', { name: texte.leiste.bezeichnung }),
    ).not.toHaveAttribute('aria-busy')
  })

  it('steht beim Laden in Endhöhe', () => {
    zeige(portfolio({ daten: undefined, laedt: true }), 0)
    expect(
      screen.getByRole('toolbar', { name: texte.leiste.bezeichnung }),
    ).toHaveAttribute('aria-busy', 'true')
  })
})
