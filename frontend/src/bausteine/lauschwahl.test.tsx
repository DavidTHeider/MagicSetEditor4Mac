// **Wo die Dienste lauschen** — G08
// (#2272).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import type { NetzschnittstelleDTO } from '../antworten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Lauschwahl } from './Lauschwahl'

const eigene = kataloge.de.seiten.technik.schnittstellen.lauschwahl

const rechner: NetzschnittstelleDTO[] = [
  {
    name: 'lo0',
    art: 'schleife',
    aktiv: true,
    adressen: ['127.0.0.1', '::1'],
    befund: '',
  },
  {
    name: 'en0',
    art: 'netz',
    aktiv: true,
    adressen: ['192.168.178.74'],
    befund: '',
  },
  { name: 'en5', art: '', aktiv: false, adressen: [], befund: '' },
]

function mitWahl(wert: string, gesetzt: string[] = []) {
  belegeApi({
    Netzschnittstellen: () =>
      Promise.resolve({ schnittstellen: rechner, fehler: null }),
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'netz.lauschen',
            wert,
            wirksam: wert,
            werte: [],
            bereich: 'netz',
            ebene: 'geraet',
            herkunft: '',
            wirkung: 'neustart',
            vorgabe: '',
          },
        ],
        fehler: null,
      }),
    EinstellungSetzen: (schluessel: string, neu: string) => {
      gesetzt.push(`${schluessel}=${neu}`)

      return Promise.resolve({ fehler: null })
    },
  })

  return render(
    <Sprachraum sprache="de">
      <Lauschwahl />
    </Sprachraum>,
  )
}

afterEach(raeumeApi)

describe('Die Lauschwahl', () => {
  it('steht ohne Wahl auf diesem Mac und speichert das als leer', async () => {
    const gesetzt: string[] = []
    mitWahl('', gesetzt)

    expect(
      await screen.findByRole('radio', { name: eigene.schleife }),
    ).toBeChecked()
    expect(screen.queryByText(eigene.aussen)).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.speichern }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      eigene.gespeichert,
    )
    expect(gesetzt).toEqual(['netz.lauschen='])
  })

  it('liest eine Auswahl, wählt dazu und sagt, dass sie den Mac verlässt', async () => {
    const gesetzt: string[] = []
    mitWahl('127.0.0.1,::1', gesetzt)

    expect(
      await screen.findByRole('radio', { name: eigene.auswahl }),
    ).toBeChecked()
    expect(screen.getByRole('checkbox', { name: '::1' })).toBeChecked()
    expect(screen.getByText(eigene.inaktiv)).toBeInTheDocument()
    expect(screen.getByText(eigene.ohneAdresse)).toBeInTheDocument()
    expect(screen.getByText(eigene.arten.netz)).toBeInTheDocument()
    expect(screen.queryByText(eigene.aussen)).toBeNull()

    await userEvent.click(
      screen.getByRole('checkbox', { name: '192.168.178.74' }),
    )
    expect(screen.getByText(eigene.aussen)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('checkbox', { name: '::1' }))
    await userEvent.click(
      screen.getByRole('button', { name: eigene.speichern }),
    )

    expect(await screen.findByRole('status')).toBeInTheDocument()
    expect(gesetzt).toEqual(['netz.lauschen=127.0.0.1,192.168.178.74'])
  })

  it('nimmt alle nur ausdrücklich und weist eine leere Auswahl ab', async () => {
    const gesetzt: string[] = []
    mitWahl('127.0.0.1', gesetzt)

    await userEvent.click(
      await screen.findByRole('radio', { name: eigene.alle }),
    )
    expect(screen.getByText(eigene.alleSatz)).toBeInTheDocument()
    expect(screen.getByText(eigene.aussen)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.speichern }),
    )
    await screen.findByRole('status')
    expect(gesetzt).toEqual(['netz.lauschen=0.0.0.0'])

    await userEvent.click(screen.getByRole('radio', { name: eigene.auswahl }))
    await userEvent.click(
      screen.getByRole('button', { name: eigene.speichern }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      eigene.keineGewaehlt,
    )
    expect(gesetzt).toHaveLength(1)
  })

  it('liest die Wahl »alle« und meldet, wenn das Speichern scheitert', async () => {
    mitWahl('0.0.0.0')
    belegeApi({
      EinstellungSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'roh',
            werte: {},
          },
        }),
    })

    expect(
      await screen.findByRole('radio', { name: eigene.alle }),
    ).toBeChecked()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.speichern }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.validierung,
    )
  })

  it('meldet, wenn die Schnittstellen nicht zu lesen sind', async () => {
    belegeApi({
      Netzschnittstellen: () =>
        Promise.resolve({
          schnittstellen: [],
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'roh',
            werte: {},
          },
        }),
      Einstellungen: () => Promise.resolve({ einstellungen: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Lauschwahl />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.validierung,
    )
  })
})
