// **Der Stand der Quelle steht neben dem Alter des Abrufs** — F04
// (#2246).
//
// Abrufzeit ist nicht Inhaltsstand: MTGJSON nennt zu jedem Preis den Tag,
// an dem er beobachtet wurde, und ein heute geholter Preis kann von gestern
// sein. Der Tag kam bis ins DTO (`beobachtet`) und stand nirgends. Scryfall
// nennt keinen — dann steht da auch keiner.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Preise } from './Preise'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(raeumeApi)

function stand(teil: Record<string, unknown>) {
  return {
    printing: 'probe-druck',
    druckID: '',
    quelle: 'mtgjson:cardmarket',
    markt: 'paper',
    art: 'retail',
    zustand: '',
    sprache: '',
    lage: 'genannt',
    beobachtet: '',
    waehrung: 'EUR',
    finish: 'nonfoil',
    wert: '4.17',
    erhoben: '2026-10-10T12:00:00Z',
    alterStunden: 3,
    ...teil,
  }
}

function zeige(staende: ReturnType<typeof stand>[]) {
  belegeApi({
    Preise: () => Promise.resolve({ staende, befunde: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Preise printing="probe-druck" aufGeholt={() => undefined} />
    </Sprachraum>,
  )
}

describe('Der Stand der Quelle', () => {
  it('steht neben dem Alter, wenn die Quelle einen Tag nennt', async () => {
    zeige([stand({ beobachtet: '2026-10-08T00:00:00Z' })])

    expect(
      await screen.findByText(/Stand der Quelle 8\. Oktober 2026/),
    ).toBeInTheDocument()
  })

  it('fehlt, wenn die Quelle keinen nennt', async () => {
    zeige([stand({ quelle: 'scryfall', beobachtet: '' })])

    // **Der Preis steht da** — der Zuhörer vor der Verneinung (§3).
    expect(await screen.findByText(/4,17\s€/u)).toBeInTheDocument()
    expect(screen.queryByText(/Stand der Quelle/)).toBeNull()
  })
})
