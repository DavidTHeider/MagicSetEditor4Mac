// **Ein leerer Katalog zeigt den nächsten Schritt, und eine Suche ohne
// Treffer führt zurück** — G02
// (#2249).
//
// Bis dahin stand bei beiden ein Satz und keine Handlung: „Kein Paket
// installiert" ohne den Knopf, der eines einbindet, und „den Artfilter
// zurücksetzen" ohne den Knopf, der es tut.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { anlegbar, katalog, texte, zeige } from '../test/paketliste'

const eigene = texte.seiten.pakete

afterEach(() => {
  raeumeApi()
})

describe('Leerzustände der Paketliste', () => {
  it('bietet im leeren Katalog das Einbinden an', async () => {
    const waehlen = vi.fn(() =>
      Promise.resolve({ plan: null, abgebrochen: true }),
    )
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn(() => Promise.resolve({ pakete: [], befunde: [] })),
      PaketPlanWaehlen: waehlen,
    })
    zeige()

    const leer = (await screen.findByText(eigene.katalogLeer)).closest(
      '[data-leerzustand]',
    )

    expect(leer).not.toBeNull()
    const einbinden = within(leer as HTMLElement).getByRole('button', {
      name: eigene.plan.einbinden,
    })

    // **Und er tut, was der Knopf im Kopf tut**: er öffnet den Dialog.
    await userEvent.click(einbinden)
    expect(waehlen).toHaveBeenCalledTimes(1)
  })

  it('führt aus einer Suche ohne Treffer zurück', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn(() => Promise.resolve(katalog())),
    })
    zeige()

    const nutzer = userEvent.setup()

    await screen.findByText('probe-schlicht')
    await nutzer.type(
      screen.getByRole('searchbox', { name: eigene.katalogSuche }),
      'gibtsnicht',
    )

    const leer = (
      await screen.findByText(eigene.katalogNichtsGefunden)
    ).closest('[data-leerzustand]')

    expect(leer).not.toBeNull()

    await nutzer.click(
      within(leer as HTMLElement).getByRole('button', {
        name: texte.bausteine.listenansicht.filterZuruecksetzen,
      }),
    )

    expect(await screen.findByText('probe-schlicht')).toBeInTheDocument()
    expect(
      screen.getByRole('searchbox', { name: eigene.katalogSuche }),
    ).toHaveValue('')
  })
})
