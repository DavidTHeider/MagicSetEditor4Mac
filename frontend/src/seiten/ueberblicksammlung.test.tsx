// **Die Kachel „Sammlung“ sagt, was die Sammlung wert ist** — AP-8, Teil 2b
// (#2365),
// Entwurf `Start09`: „486 € · ältester Preis vor 3 Tagen“. Gelesen aus dem
// Portfolio, **für sich**: die Zahlen warten nicht darauf (PERF02).
import { render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Portfolioantwort } from '../api'
import { geldAusText } from '../geld'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Start } from './Start'

const texte = kataloge.de.seiten.start

afterEach(() => {
  raeumeApi()
})

function zeige(portfolio: () => Promise<Portfolioantwort>) {
  belegeApi({
    Sets: () => Promise.resolve({ sets: [], fehler: null }),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: [],
        gesamt: 0,
        insgesamt: 4,
        stueck: 4,
        fehler: null,
      }),
    Setprobleme: () => Promise.resolve({ staende: [], fehler: null }),
    Portfolio: portfolio,
  })
  render(
    <Sprachraum sprache="de">
      <Start wechseln={vi.fn()} waehleSet={vi.fn()} bestandsstand={0} />
    </Sprachraum>,
  )
}

const ueberblick = () =>
  within(screen.getByRole('list', { name: texte.ueberblick.titel }))

describe('die Kachel der Sammlung', () => {
  it('nennt Wert, ältesten Preis und was ohne Preis blieb', async () => {
    zeige(() =>
      Promise.resolve({
        zeilen: [
          {
            waehrung: 'EUR',
            wert: '486.00',
            exemplare: 3,
            aeltesterPreis: '2026-10-14T08:00:00Z',
            alterStunden: 72,
            quelle: 'Scryfall',
          },
        ],
        ohnePreis: 1,
        fehler: null,
      }),
    )

    const wert = geldAusText('486.00', 'EUR', 'de')
    expect(
      await ueberblick().findByRole('button', {
        name: `4 Exemplare ${wert} · ältester Preis vor 3 Tagen · 1 ohne Preis`,
      }),
    ).toBeInTheDocument()
  })

  it('wartet mit den Zahlen nicht auf das Portfolio', async () => {
    zeige(() => new Promise(() => undefined))

    expect(
      await ueberblick().findByRole('button', { name: /^4 Exemplare$/ }),
    ).toBeInTheDocument()
  })
})
