// **Wohin eine Karte außerhalb dieses Programms führt** — Werkstück T8 von
// Phase 5.6
// (#1473).
//
// # Was hier gemessen wird
//
// **Dass an jeder Zeile steht, wie genau sie trifft.** „Bei CardTrader
// ansehen" führt bei dem einen auf diese Karte und bei dem anderen auf eine
// Liste; wer beides gleich beschriftete, verspräche zu viel.
//
// **Dass der Öffner den Anbieter mitbekommt** — ohne ihn wüsste die Prüfung
// dahinter nicht, gegen welchen Wirt sie prüft.
//
// **Und dass ein Fehler beim Öffnen die Liste nicht wegnimmt** (§7, Regel 3).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Marktlinke } from './Marktlinke'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Marktlink } from '../apimarktlinke'

const eigene = kataloge.de.seiten.karten

// dreiWege sind ein Produkt, eine Suche und eine Startseite.
function dreiWege(): Marktlink[] {
  return [
    {
      anbieter: 'tcgplayer',
      art: 'produkt',
      herkunft: 'kennung',
      adresse: 'https://www.tcgplayer.com/product/33632',
    },
    {
      anbieter: 'cardmarket',
      art: 'suche',
      herkunft: 'name',
      adresse:
        'https://www.cardmarket.com/en/Magic/Products/Search?searchString=Bolt',
    },
    {
      anbieter: 'gatherer',
      art: 'startseite',
      herkunft: 'keine',
      adresse: 'https://gatherer.wizards.com/',
    },
  ]
}

// zeige stellt den Kasten hin.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Marktlinke printing="bolt-m10" name="Lightning Bolt" />
    </Sprachraum>,
  )

  return userEvent.setup()
}

describe('Wo es diese Karte gibt', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('nennt an jeder Zeile, wie genau sie trifft', async () => {
    belegeApi({
      Marktlinks: () => Promise.resolve({ links: dreiWege(), fehler: null }),
    })

    zeige()

    const liste = await screen.findByRole('list', { name: eigene.marktTitel })

    expect(within(liste).getByText(eigene.marktGenau)).toBeInTheDocument()
    expect(within(liste).getByText(eigene.marktSuche)).toBeInTheDocument()
    expect(within(liste).getByText(eigene.marktNurAnbieter)).toBeInTheDocument()
  })

  it('gibt dem Öffner den Anbieter und die Adresse', async () => {
    const geoeffnet = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Marktlinks: () => Promise.resolve({ links: dreiWege(), fehler: null }),
      MarktlinkOeffnen: geoeffnet,
    })

    const hand = zeige()

    await hand.click(await screen.findByRole('button', { name: 'tcgplayer' }))

    expect(geoeffnet).toHaveBeenCalledWith(
      'tcgplayer',
      'https://www.tcgplayer.com/product/33632',
    )
  })

  it('meldet einen gescheiterten Weg und behält die Liste', async () => {
    belegeApi({
      Marktlinks: () => Promise.resolve({ links: dreiWege(), fehler: null }),
      MarktlinkOeffnen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung' as const,
            schluessel: 'link',
            meldung: 'diese Adresse geht nicht hinaus',
          },
        }),
    })

    const hand = zeige()

    await hand.click(await screen.findByRole('button', { name: 'cardmarket' }))

    expect(
      await screen.findByText('diese Adresse geht nicht hinaus'),
    ).toBeInTheDocument()

    // **Und die Liste steht noch da** — ein Fehler beim Öffnen ist kein
    // Fehler beim Lesen.
    expect(
      screen.getByRole('list', { name: eigene.marktTitel }),
    ).toBeInTheDocument()
  })

  it('meldet, wenn die Wege nicht zu haben sind', async () => {
    belegeApi({
      Marktlinks: () =>
        Promise.resolve({
          links: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'links',
            meldung: 'so nicht',
          },
        }),
    })

    zeige()

    expect(await screen.findByText('so nicht')).toBeInTheDocument()
  })
})
