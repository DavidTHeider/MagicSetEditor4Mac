// **Brotkrumen und Arbeitsplatz lesen das offene Set einmal** — U13
// (#2158).
//
// Das Fenster in klein: es liest wie `Anwendung`, nennt den Namen über dem
// Kopf und zeichnet darunter den echten Arbeitsplatz. Gezählt wird der Ruf
// `Sets` — je Set und Stand **einer**, nicht einer je Verbraucher.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO, Umschlag } from '../api'
import { Setarbeitsplatz } from '../rahmen/Setarbeitsplatz'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { OffenesSet, useOffenesSet } from './offenesset'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

const set = (id: string, name: string): SetDTO => ({
  id,
  name,
  spiel: 'magic',
  stil: 'magic-m15',
  karten: 3,
  herkunft: '',
  herkunftFehlt: false,
})

function Fenster({ setID, stand }: { setID: string; stand: number }) {
  const offen = useOffenesSet(setID, stand)

  return (
    <Sprachraum sprache="de">
      <p data-testid="weg">{offen?.name ?? '—'}</p>
      <OffenesSet wert={offen}>
        <Setarbeitsplatz
          setID={setID}
          aktuell="karten"
          wechseln={vi.fn()}
          verlaufsstand={0}
          oeffneKarte={vi.fn()}
        >
          {() => <p>Inhalt</p>}
        </Setarbeitsplatz>
      </OffenesSet>
    </Sprachraum>
  )
}

function mitSets(...sets: SetDTO[]) {
  const gefragt = vi.fn<() => Promise<{ sets: SetDTO[] } & Umschlag>>(() =>
    Promise.resolve({ sets, fehler: null }),
  )
  belegeApi({ Sets: gefragt })

  return gefragt
}

describe('das offene Set', () => {
  it('wird für Weg und Arbeitsplatz einmal gelesen', async () => {
    const gefragt = mitSets(set('7', 'Vier Winde'))

    render(<Fenster setID="7" stand={0} />)

    expect(
      await screen.findByRole('heading', { name: 'Vier Winde' }),
    ).toBeInTheDocument()
    expect(screen.getByTestId('weg')).toHaveTextContent('Vier Winde')
    expect(gefragt).toHaveBeenCalledTimes(1)
  })

  it('liest bei einem neuen Stand einmal nach und behält den Namen dabei', async () => {
    const gefragt = mitSets(set('7', 'Vier Winde'))

    const { rerender } = render(<Fenster setID="7" stand={0} />)
    await screen.findByRole('heading', { name: 'Vier Winde' })

    rerender(<Fenster setID="7" stand={1} />)
    expect(
      screen.getByRole('heading', { name: 'Vier Winde' }),
    ).toBeInTheDocument()

    await effekteNachholen()
    expect(gefragt).toHaveBeenCalledTimes(2)
  })

  it('räumt bei einem anderen Set und gibt kein fremdes heraus', async () => {
    mitSets(set('7', 'Vier Winde'), set('8', 'Ebbe'))

    const { rerender } = render(<Fenster setID="7" stand={0} />)
    await screen.findByRole('heading', { name: 'Vier Winde' })

    rerender(<Fenster setID="8" stand={0} />)
    expect(screen.queryByText('Vier Winde')).not.toBeInTheDocument()
    expect(
      await screen.findByRole('heading', { name: 'Ebbe' }),
    ).toBeInTheDocument()
  })

  it('zeigt bei einem Fehler in beiden keinen Namen', async () => {
    const gefragt = vi.fn(() => Promise.reject(new Error('planmäßig: weg')))
    belegeApi({ Sets: gefragt })

    render(<Fenster setID="7" stand={0} />)
    await effekteNachholen()

    expect(gefragt).toHaveBeenCalledTimes(1)
    expect(screen.getByTestId('weg')).toHaveTextContent('—')
    expect(
      screen.getByRole('heading', {
        name: texte.allgemein.arbeitsplatz.ohneNamen,
      }),
    ).toBeInTheDocument()
  })
})

describe('ein Arbeitsplatz für ein anderes Set', () => {
  it('bekommt das gelesene nicht', () => {
    render(
      <Sprachraum sprache="de">
        <OffenesSet wert={set('7', 'Vier Winde')}>
          <Setarbeitsplatz
            setID="8"
            aktuell="karten"
            wechseln={vi.fn()}
            verlaufsstand={0}
            oeffneKarte={vi.fn()}
          >
            {() => <p>Inhalt</p>}
          </Setarbeitsplatz>
        </OffenesSet>
      </Sprachraum>,
    )

    expect(
      screen.getByRole('heading', {
        name: texte.allgemein.arbeitsplatz.ohneNamen,
      }),
    ).toBeInTheDocument()
    expect(screen.queryByText('Vier Winde')).not.toBeInTheDocument()
  })
})
