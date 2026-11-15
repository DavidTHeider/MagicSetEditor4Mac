// **Ein bekannter Peer an einer neuen Adresse** — NET-02.NAT-02 und
// S14.NAT-04 (#1706).
//
// # Der Befund
//
// **Die Suche fand ihn und sagte nur „steht schon in der Liste"** — und die
// Adresse, unter der die Kontaktliste ihn führte, stimmte nicht mehr. Wer
// verbinden wollte, tippte die neue von Hand ab (Handprüfung vom
// 2026-09-20). Jetzt steht beides da, und ein Knopf übernimmt sie — ohne
// am Vertrauen etwas zu ändern.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Nachbarn } from './Nachbarn'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { Bindungsfehler } from '../bindungsfehler'

import type { Nachbar } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik

afterEach(() => {
  raeumeApi()
})

function umgezogen(anders: Partial<Nachbar> = {}): Nachbar {
  return {
    kennung: 'bumgezogen',
    fingerabdruck: '1111-2222',
    name: 'Annas Mac',
    adresse: '10.0.0.8:41234',
    merkadresse: 'annas-neuer.local:41234',
    bekannt: true,
    stand: 'bestaetigt',
    gemerkteAdresse: 'annas-mac.local:41234',
    adresseGeaendert: true,
    ...anders,
  }
}

async function suche(
  nachbar: Nachbar,
  mehr: Parameters<typeof belegeApi>[0] = {},
) {
  belegeApi({
    Nachbarn: () =>
      Promise.resolve({ nachbarn: [nachbar], verworfen: 0, fehler: null }),
    ...mehr,
  })

  render(
    <Sprachraum sprache="de">
      <Nachbarn />
    </Sprachraum>,
  )

  const hand = userEvent.setup()
  await hand.click(screen.getByRole('button', { name: eigene.nachbarnSuchen }))

  return hand
}

it('übernimmt die neue Adresse eines bekannten Peers, ohne das Vertrauen anzufassen', async () => {
  const adresse = vi.fn(() =>
    Promise.resolve({ kontakte: [], zwecke: [], fehler: null }),
  )
  const bestaetigt = vi.fn()

  const hand = await suche(umgezogen(), {
    PeerkontaktAdresse: adresse,
    PeerkontaktBestaetigen: bestaetigt,
  })

  // **Beides steht da** — was die Liste führte, und wo er jetzt ist.
  expect(await screen.findByText('annas-mac.local:41234')).toBeInTheDocument()
  expect(screen.getByText(eigene.nachbarnAdresseNeu)).toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', { name: eigene.nachbarnAdresseUebernehmen }),
  )

  expect(adresse).toHaveBeenCalledWith('bumgezogen', 'annas-neuer.local:41234')
  expect(bestaetigt).not.toHaveBeenCalled()
})

it('bietet ohne gemerkte Adresse dieselbe Übernahme an', async () => {
  await suche(umgezogen({ gemerkteAdresse: '' }))

  expect(
    await screen.findByText(eigene.nachbarnAdresseKeine),
  ).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: eigene.nachbarnAdresseUebernehmen }),
  ).toBeInTheDocument()
})

it('bietet nichts an, wo die Adresse stimmt', async () => {
  await suche(umgezogen({ adresseGeaendert: false }))

  expect(
    await screen.findByText(eigene.nachbarnSchonBekannt),
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: eigene.nachbarnAdresseUebernehmen }),
  ).toBeNull()
})

// **Ein unerwarteter Fehler der Suche führt zur Diagnose** — S13.NAT-04:
// die Kennung reist mit, statt abgetippt zu werden.
it('bietet an einem unerwarteten Fehler „Diagnose öffnen" an', async () => {
  belegeApi({
    Nachbarn: () =>
      Promise.reject(
        new Bindungsfehler({
          code: 'unbekannt',
          schluessel: 'unbekanntMitVorgang',
          meldung: 'kaputt',
          werte: { vorgang: 'a1b2c3d4' },
        }),
      ),
  })

  render(
    <Sprachraum sprache="de">
      <Nachbarn />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', { name: eigene.nachbarnSuchen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent('a1b2c3d4')
  expect(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.diagnoseverweis.oeffnen,
    }),
  ).toBeInTheDocument()
})
