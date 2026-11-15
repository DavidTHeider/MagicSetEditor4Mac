// **Eine laufende Anmeldung lässt sich abbrechen** — BK-07d, AK-4
// (#2118).
//
// Der Weg bis zum Horcher steht seit BK-01: jede Bindung nimmt den Rufkontext
// von Wails, und ein abgebrochener Ruf beendet ihn. Was fehlte, war die
// Bedienung — wer im Browser nicht weitermachte, wartete bis zur Frist.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { EinNutzerAntwort } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung, einNutzer } from '../test/anmeldebelegung'
import { Nutzerseite } from './Nutzerseite'

afterEach(() => {
  raeumeApi()
})

// abbrechbar gibt eine Antwort, die erst mit ihrem `cancel` endet — wie ein
// Ruf der Wails-Laufzeit, der auf den Browser wartet.
function abbrechbar() {
  let ablehnen: (grund: Error) => void = () => undefined
  const lauf = new Promise<EinNutzerAntwort>((_, nein) => {
    ablehnen = nein
  }) as Promise<EinNutzerAntwort> & { cancel: () => void }
  lauf.cancel = vi.fn(() => {
    ablehnen(new Error('call cancelled'))
  })

  return lauf
}

describe('der Abbruch einer Anmeldung', () => {
  it('bricht den Ruf ab, meldet keinen Fehler und gibt den Knopf frei', async () => {
    const lauf = abbrechbar()

    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() =>
        Promise.resolve({ profil: einNutzer().profile[0] ?? null }),
      ),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einNutzer()] })),
      Anmelden: vi.fn(() => lauf),
    })

    render(
      <Sprachraum sprache="de">
        <Nutzerseite />
      </Sprachraum>,
    )

    await userEvent.click(await screen.findByRole('tab', { name: 'Konto' }))
    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Anmeldung über einen Anbieter',
      }),
    )

    const abschnitt = within(
      await screen.findByRole('region', {
        name: 'Anmeldung über einen Anbieter',
      }),
    )
    await userEvent.click(
      abschnitt.getByRole('button', { name: 'Bei GitHub anmelden' }),
    )

    await userEvent.click(
      await abschnitt.findByRole('button', {
        name: 'Anmeldung bei GitHub abbrechen',
      }),
    )

    expect(lauf.cancel).toHaveBeenCalledTimes(1)
    expect(
      await abschnitt.findByRole('button', { name: 'Bei GitHub anmelden' }),
    ).toBeEnabled()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(
      abschnitt.queryByRole('button', {
        name: 'Anmeldung bei GitHub abbrechen',
      }),
    ).not.toBeInTheDocument()
  })
})
