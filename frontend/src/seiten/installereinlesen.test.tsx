// **Einen `.mse-installer` einlesen** — FMT-02.AK-04
// (#1691).
//
// In derselben Fläche wie die Übernahme aus einer MSE2-Installation: dieselbe
// Antwort, dieselbe Anzeige, und ein Vorgang zur Zeit.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { abgebrochen } from '../test/laufauftrag'
import { Uebernahme } from './Uebernahme'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.pakete

const leer = {
  uebernommen: [],
  uebergangen: [],
  befunde: [],
  abgebrochen: false,
  fehler: null,
}

function zeige(InstallerEinlesen: () => Promise<unknown>) {
  belegeApi({
    MSE2Vorschlaege: () => Promise.resolve({ orte: [] }),
    InstallerEinlesen: InstallerEinlesen as never,
  })

  const uebernommen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Uebernahme uebernommen={uebernommen} />
    </Sprachraum>,
  )

  return uebernommen
}

describe('Einen Installer einlesen', () => {
  it('liest ein und zeigt, was übergangen wurde', async () => {
    const einlesen = vi.fn(() =>
      Promise.resolve({
        ...leer,
        uebergangen: ['probe.mse-game'],
        befunde: ['probe-kaputt.mse-game: das Paket ließ sich nicht auspacken'],
      }),
    )
    const uebernommen = zeige(einlesen)

    await userEvent.click(
      await screen.findByRole('button', { name: texte.installerWaehlen }),
    )

    expect(await screen.findByText('probe.mse-game')).toBeInTheDocument()
    expect(
      screen.getByText(/das Paket ließ sich nicht auspacken/),
    ).toBeInTheDocument()
    expect(einlesen).toHaveBeenCalledOnce()
    expect(uebernommen).toHaveBeenCalledOnce()
  })

  // **Ein abgebrochener Dialog hat nichts geschrieben** — kein Nachladen.
  it('lädt nach einem Abbruch nicht nach', async () => {
    const einlesen = vi.fn(() =>
      Promise.resolve({ ...leer, abgebrochen: true }),
    )
    const uebernommen = zeige(einlesen)

    await userEvent.click(
      await screen.findByRole('button', { name: texte.installerWaehlen }),
    )

    await waitFor(() => {
      expect(einlesen).toHaveBeenCalledOnce()
    })
    expect(uebernommen).not.toHaveBeenCalled()
    expect(screen.queryByRole('status')).toBeNull()
  })

  it('zeigt einen Fehler und lädt trotzdem nach', async () => {
    const uebernommen = zeige(() => Promise.reject(new Error('kaputt')))

    await userEvent.click(
      await screen.findByRole('button', { name: texte.installerWaehlen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(uebernommen).toHaveBeenCalledOnce()
  })
})

// **Ein angehaltenes Einlesen ist keine Fehlermeldung** (FMT-03, #1692) — die
// Pakete davor sind ganz, und nachgeladen wird trotzdem.
it('zeigt ein Anhalten nicht als Fehler und lädt nach', async () => {
  const uebernommen = zeige(() =>
    Promise.resolve({ ...leer, fehler: abgebrochen }),
  )

  await userEvent.click(
    await screen.findByRole('button', { name: texte.installerWaehlen }),
  )

  await waitFor(() => {
    expect(uebernommen).toHaveBeenCalled()
  })
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})
