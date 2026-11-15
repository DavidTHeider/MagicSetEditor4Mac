// **Die Zwischenspeicher sehen und leeren** — OPS-08
// (#1725):
// je Speicher Ort, Treffer, Einträge, Größe und Grenze; ein Knopf leert alle.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Zwischenspeicher } from './Zwischenspeicher'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { ZwischenspeicherAntwort } from '../apizwischenspeicher'

const texte = kataloge.de.seiten.technik.zwischenspeicher

afterEach(raeumeApi)

function stand(eintraege: number): ZwischenspeicherAntwort {
  return {
    speicher: [
      {
        name: 'scryfall.antworten',
        ort: 'platte',
        treffer: 9,
        fehlgriffe: 3,
        eintraege,
        bytes: 2048,
        grenzeEintraege: 0,
        grenzeBytes: 536870912,
      },
      {
        name: 'scryfall.sammeldaten',
        ort: 'platte',
        treffer: 0,
        fehlgriffe: 0,
        eintraege: 1,
        bytes: 1024,
        grenzeEintraege: 0,
        grenzeBytes: 0,
      },
      {
        name: 'unbekannt.speicher',
        ort: 'mond',
        treffer: 0,
        fehlgriffe: 0,
        eintraege: 0,
        bytes: 0,
        grenzeEintraege: 64,
        grenzeBytes: 0,
      },
    ],
    fehler: null,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Zwischenspeicher />
    </Sprachraum>,
  )
}

describe('Die Zwischenspeicher', () => {
  it('nennen je Speicher Ort, Trefferquote, Einträge, Größe und Grenze', async () => {
    belegeApi({ Zwischenspeicher: vi.fn(() => Promise.resolve(stand(4))) })
    zeige()

    const tafel = await screen.findByRole('table', { name: texte.ueberschrift })
    const zeile = within(tafel).getAllByRole('row')[1]

    expect(zeile).toHaveTextContent(texte.namen['scryfall.antworten'])
    expect(zeile).toHaveTextContent(texte.orte.platte)
    expect(zeile).toHaveTextContent('75')
    expect(zeile).toHaveTextContent('2,0 KB')
    expect(zeile).toHaveTextContent('512,0 MB')

    // **Unbekanntes im Wortlaut**, ohne Treffer keine Quote.
    expect(within(tafel).getAllByRole('row')[2]).toHaveTextContent(
      texte.keineGrenze,
    )

    const zweite = within(tafel).getAllByRole('row')[3]
    expect(zweite).toHaveTextContent('unbekannt.speicher')
    expect(zweite).toHaveTextContent(texte.keineQuote)
    expect(zweite).toHaveTextContent('64')
    expect(zweite).toHaveTextContent('mond')
  })

  it('leert alle und zeigt danach den neuen Stand', async () => {
    const leeren = vi.fn(() => Promise.resolve(stand(4)))
    belegeApi({
      Zwischenspeicher: vi
        .fn()
        .mockResolvedValueOnce(stand(4))
        .mockResolvedValue(stand(0)),
      ZwischenspeicherLeeren: leeren,
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.leeren }),
    )

    await waitFor(() => {
      expect(leeren).toHaveBeenCalledOnce()
    })
    expect(await screen.findByText(texte.geleert)).toBeInTheDocument()
  })

  it('zeigt einen Fehler beim Leeren', async () => {
    belegeApi({
      Zwischenspeicher: vi.fn(() => Promise.resolve(stand(1))),
      ZwischenspeicherLeeren: vi.fn(() =>
        Promise.resolve({
          speicher: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: '',
            meldung: 'Der Ordner ließ sich nicht leeren.',
          },
        }),
      ),
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.leeren }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('zeigt einen Fehler der Auskunft', async () => {
    belegeApi({
      Zwischenspeicher: vi.fn(() =>
        Promise.resolve({
          speicher: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: '',
            meldung: 'Der Ort für Zwischenspeicher fehlt.',
          },
        }),
      ),
    })
    zeige()

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.queryByRole('table')).toBeNull()
  })
})
