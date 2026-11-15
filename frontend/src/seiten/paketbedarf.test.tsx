// **Was dem Set an Paketen fehlt, und wie es dazukommt** — DOM-01.AK-04
// (#1696):
// aus dem Set heraus diagnostiziert und gezielt behoben — mit dem Plan hier
// oder mit der Suche bei den Paketen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Paketbedarf } from './Paketbedarf'

afterEach(() => {
  raeumeApi()
})

const t = kataloge.de.seiten.pakete.plan

const fehlt = {
  befunde: [
    {
      paket: 'Erstes',
      braucht: 'probe-stil.mse-style',
      art: 'fehlt',
      verlangt: '',
      da: '',
    },
  ],
}

function zeige(wechseln = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="karten" wechseln={wechseln}>
        <Paketbedarf setID="1" />
      </Anschlussraum>
    </Sprachraum>,
  )

  return wechseln
}

describe('Paketbedarf', () => {
  it('nennt, was fehlt, und springt zur Suche bei den Paketen', async () => {
    belegeApi({ Paketbedarf: vi.fn(() => Promise.resolve(fehlt)) })
    const wechseln = zeige()

    expect(
      await screen.findByText(/Erstes braucht probe-stil\.mse-style/),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: t.bedarf.suchen }))
    expect(wechseln).toHaveBeenCalledWith('pakete')
  })

  it('bindet aus dem Set heraus ein — nach dem Plan — und fragt danach neu', async () => {
    const bedarf = vi
      .fn()
      .mockResolvedValueOnce(fehlt)
      .mockResolvedValue({ befunde: [] })
    const PaketEinbindenVon = vi.fn(() =>
      Promise.resolve({ paket: null, abgebrochen: false }),
    )

    belegeApi({
      Paketbedarf: bedarf,
      PaketPlanWaehlen: () =>
        Promise.resolve({
          plan: {
            vorgang: 'einbinden' as const,
            paket: 'probe-stil.mse-style',
            quelle: '/q/probe-stil.mse-style',
            von: '',
            zu: '1.0',
            geaendert: false,
            moeglich: true,
            sets: [],
            konflikte: [],
          },
          abgebrochen: false,
        }),
      PaketEinbindenVon,
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: t.einbinden }),
    )

    // **Die Quelle steht im Plan** — woher es kommt, sieht man vorher (AK-04).
    expect(
      await screen.findByText('/q/probe-stil.mse-style', { exact: false }),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: t.bestaetigen.einbinden }),
    )

    expect(PaketEinbindenVon).toHaveBeenCalledWith('/q/probe-stil.mse-style')
    // **Gewartet wird auf die Wirkung**: der zweite Ruf beweist, dass jemand
    // nachfragt, und die Antwort kommt eine Makroaufgabe später.
    await waitFor(() => {
      expect(bedarf).toHaveBeenCalledTimes(2)
    })
    await waitFor(() => {
      expect(screen.queryByText(/Erstes braucht/)).not.toBeInTheDocument()
    })
  })

  it('zeigt nichts, wenn nichts fehlt — und einen Fehler am Ort', async () => {
    const leer = vi.fn().mockResolvedValueOnce({ befunde: [] })
    belegeApi({ Paketbedarf: leer })
    const { unmount } = render(
      <Sprachraum sprache="de">
        <Paketbedarf setID="1" />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(leer).toHaveBeenCalledTimes(1)
    })
    expect(screen.queryByText(t.bedarf.ueberschrift)).not.toBeInTheDocument()
    unmount()

    belegeApi({
      Paketbedarf: () =>
        Promise.resolve({
          befunde: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Katalog schweigt.',
          },
        }),
    })
    render(
      <Sprachraum sprache="de">
        <Paketbedarf setID="1" />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Katalog schweigt.',
    )
  })
})
