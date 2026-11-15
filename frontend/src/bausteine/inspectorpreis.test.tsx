// **Ein Preis wird beim Bearbeiten gelesen wie beim Anlegen** — LOC-03.AK-03
// (#1689),
// angestrichen vom Zweitblick über Phase 6.3 (S6): der Inspector schickte den
// Text roh an Go. »1.234« Dinar wurde still zu 1,234, »19,99« Euro
// abgewiesen, und das Feld stand in kanonischer Form da.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import type { ExemplarDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlungsinspector } from './Sammlungsinspector'

const texte = kataloge.de.seiten.sammlung

function exemplar(erwerbspreis: string, erwerbswaehrung: string): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'probe-druck',
    druckID: '',
    vorlaeufig: false,
    menge: 1,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis,
    erwerbswaehrung,
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'probe-druck',
    herkunft: keineHerkunft,
  }
}

async function bearbeite(preis: string, waehrung: string) {
  belegeApi({
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
  })
  const aufAendern = vi.fn(() => Promise.resolve(true))

  render(
    <Sprachraum sprache="de">
      <Sammlungsinspector
        exemplar={exemplar(preis, waehrung)}
        inSicht
        aufAendern={aufAendern}
        aufLoeschen={vi.fn(() => Promise.resolve(true))}
        aufGeholt={vi.fn()}
        laeuft={false}
      />
    </Sprachraum>,
  )
  await userEvent.click(screen.getByRole('button', { name: texte.bearbeiten }))

  // **Der Preis steht seit UXR-10 (#1918) unter „Weitere Angaben"** —
  // eingeklappt, bis jemand ihn braucht.
  await userEvent.click(
    screen.getByRole('button', { name: texte.weitereAngaben }),
  )

  return aufAendern
}

const sichern = () =>
  userEvent.click(screen.getByRole('button', { name: texte.aenderungSichern }))

afterEach(() => {
  raeumeApi()
})

describe('Der Preis im Inspector', () => {
  it('steht in der Schreibweise des Fensters im Feld', async () => {
    await bearbeite('19.99', 'EUR')

    expect(screen.getByLabelText(texte.erwerbspreis)).toHaveValue('19,99')
  })

  it('wird gelesen wie beim Anlegen und geht kanonisch hinüber', async () => {
    const aufAendern = await bearbeite('19.99', 'EUR')
    const feld = screen.getByLabelText(texte.erwerbspreis)

    await userEvent.clear(feld)
    await userEvent.type(feld, '24,50')
    await sichern()

    await waitFor(() => {
      expect(aufAendern).toHaveBeenCalledWith(
        expect.objectContaining({ erwerbspreis: '24.50' }),
      )
    })
  })

  it('bleibt unberührt, was unberührt bleibt', async () => {
    const aufAendern = await bearbeite('1.234', 'BHD')

    expect(screen.getByLabelText(texte.erwerbspreis)).toHaveValue('1,234')
    await sichern()

    await waitFor(() => {
      expect(aufAendern).toHaveBeenCalledWith(
        expect.objectContaining({ erwerbspreis: '1.234' }),
      )
    })
  })

  // **Zwei Lesarten werden erklärt und nicht geraten** (AK-03).
  it('fragt bei zwei Lesarten am Feld, statt eine zu nehmen', async () => {
    const aufAendern = await bearbeite('', 'BHD')
    const feld = screen.getByLabelText(texte.erwerbspreis)

    await userEvent.type(feld, '1.234')
    await sichern()

    expect(await screen.findByText(/1\.234,000/)).toBeInTheDocument()
    expect(aufAendern).not.toHaveBeenCalled()
  })
})
