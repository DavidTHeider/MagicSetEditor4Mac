// **Die Leiste über den Paketen führt hin** — AP-9, Teil 2
// (#2366):
// „Installiert“ nimmt die Suche zurück, „Mit Befund“ zeigt genau die Pakete
// mit einem Abhängigkeitsbefund — und die Liste darunter folgt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AbhaengigkeitDTO, PaketDTO } from '../api'
import { Anschlussraum } from '../anschlussraum'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { belegeSeite, texte } from '../test/paketseite'
import { katalog } from '../test/paketliste'
import { Pakete } from './Pakete'

afterEach(() => {
  raeumeApi()
})

function paket(name: string): PaketDTO {
  const [vorlage] = katalog().pakete
  if (vorlage === undefined) {
    throw new Error('die Vorrichtung trägt kein Paket')
  }

  return { ...vorlage, name, anzeigename: name }
}

const befund: AbhaengigkeitDTO = {
  paket: 'probe-zwei.mse-style',
  braucht: 'probe.mse-game',
  art: 'fehlt',
  verlangt: '',
  da: '',
}

function zeige() {
  belegeSeite({
    Pakete: () =>
      Promise.resolve({
        pakete: [paket('probe-eins'), paket('probe-zwei'), paket('probe-drei')],
        befunde: [],
      }),
    Abhaengigkeiten: () => Promise.resolve({ befunde: [befund] }),
  })
  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="pakete" wechseln={vi.fn()}>
        <Pakete wechseln={vi.fn()} bestandsstand={0} />
      </Anschlussraum>
    </Sprachraum>,
  )
}

const eigene = texte.seiten.pakete
const leiste = async () =>
  within(
    await screen.findByRole('toolbar', { name: eigene.leiste.bezeichnung }),
  )

describe('die Ziele der Leiste', () => {
  it('zeigt mit Befund genau die Pakete mit Befund', async () => {
    zeige()
    expect(await screen.findByText('probe-eins')).toBeInTheDocument()

    await userEvent.click(
      await (
        await leiste()
      ).findByRole('button', { name: `${eigene.leiste.befund} 1` }),
    )

    expect(await screen.findByText('1 Paket mit Befund')).toBeInTheDocument()
    expect(screen.queryByText('probe-eins')).toBeNull()
    expect(screen.getByText('probe-zwei')).toBeInTheDocument()
  })

  it('nimmt mit „Installiert“ die Suche zurück', async () => {
    zeige()
    const suche = await screen.findByRole('searchbox', {
      name: eigene.katalogSuche,
    })
    await userEvent.type(suche, 'drei')
    expect(screen.queryByText('probe-eins')).toBeNull()

    await userEvent.click(
      (await leiste()).getByRole('button', {
        name: `${eigene.leiste.installiert} 3`,
      }),
    )

    expect(await screen.findByText('probe-eins')).toBeInTheDocument()
    expect(suche).toHaveValue('')
  })
})
