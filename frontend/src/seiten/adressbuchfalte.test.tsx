// **Das Adressbuch mit echten Mengen** — UXR-16
// (#1924).
//
// # Der Befund
//
// Apple Kontakte gibt „alle Kontakte" als **eine** vCard aus. Mit 150
// übernommenen Kontakten und einer Datei von 200 Karten
// (`test/gegenueberbestand.ts`) gezählt am 2026-10-02: **154
// Bedienelemente** in der Liste und in der Vorschau **600 Kästchen** — je
// Karte eines und je Weg eines, auch für die 150, die unverändert schon
// dastehen und nicht wählbar waren (R2, R3, R6 in `docs/UIREVIEW.md`).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Adressbuch } from './Adressbuch'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { zaehle } from '../test/bedienzaehlung'
import { adressbuch, vcard } from '../test/gegenueberbestand'

import type { KontaktDTO } from '../apikontakte'

const texte = kataloge.de.seiten.adressbuch
const listenworte = kataloge.de.bausteine.listenansicht

afterEach(() => {
  raeumeApi()
})

async function zeige(uebernehmen = vi.fn()) {
  belegeApi({
    Kontakte: () => Promise.resolve({ kontakte: adressbuch(), fehler: null }),
    KontakteEinlesen: () =>
      Promise.resolve({
        karten: vcard(),
        befunde: [],
        abgebrochen: false,
        fehler: null,
      }),
    KontakteUebernehmen: (auswahl: KontaktDTO[]) => {
      uebernehmen(auswahl)

      return Promise.resolve({ angelegt: 50, ersetzt: 0, fehler: null })
    },
  })

  render(
    <Sprachraum sprache="de">
      <Adressbuch />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  await hand.click(screen.getByRole('button', { name: texte.titel }))
  await screen.findByText('Kontakt 001')

  return hand
}

async function vorschau(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(screen.getByRole('button', { name: texte.waehlen }))

  return screen.findByRole('region', { name: texte.vorschauTitel })
}

describe('die Kontaktliste', () => {
  it('zeigt die ersten zwanzig und den Rest stückweise', async () => {
    const hand = await zeige()

    expect(screen.getByText('Kontakt 020')).toBeInTheDocument()
    expect(screen.queryByText('Kontakt 021')).toBeNull()
    expect(screen.getAllByRole('button', { name: /entfernen$/ })).toHaveLength(
      20,
    )

    await hand.click(
      screen.getByRole('button', {
        name: einsetzen(listenworte.weitere, { anzahl: '20' }),
      }),
    )

    expect(screen.getByText('Kontakt 040')).toBeInTheDocument()
  })
})

describe('die Vorschau einer vCard mit 200 Karten', () => {
  it('zeichnet nur wählbare Karten, zehn vorn, und keine Wege als Kästchen', async () => {
    const hand = await zeige()
    const bereich = await vorschau(hand)

    expect(zaehle(bereich).felder).toBe(10)
    expect(within(bereich).getByText('Kontakt 151')).toBeInTheDocument()
    expect(within(bereich).queryByText('Kontakt 001')).toBeNull()

    // **Was unverändert ist, steht als eine Zeile da** — und auf Wunsch
    // mit Namen.
    const unveraendert = within(bereich).getByRole('button', {
      name: einsetzen(texte.unveraendert, { anzahl: '150' }),
    })

    await hand.click(unveraendert)

    expect(within(bereich).getByText('Kontakt 001')).toBeInTheDocument()

    // **Übernommen wird alles Gewählte**, auch was nicht gezeichnet ist.
    expect(
      within(bereich).getByRole('button', {
        name: einsetzen(texte.uebernehmen, { anzahl: '50' }),
      }),
    ).toBeEnabled()
  })

  it('wählt die Wege einer Karte auf Wunsch einzeln', async () => {
    const uebernehmen = vi.fn()
    const hand = await zeige(uebernehmen)
    const bereich = await vorschau(hand)
    const telefon = einsetzen(texte.wegZeile, {
      art: texte.weg.telefon,
      wert: '+49 170 150',
    })

    // **Gelesen wird der Weg immer**, gewählt erst auf Wunsch.
    expect(within(bereich).getByText(telefon)).toBeInTheDocument()
    expect(
      within(bereich).queryByRole('checkbox', { name: telefon }),
    ).toBeNull()

    const wege = within(bereich).getByRole('button', {
      name: `${texte.wegeWaehlen}: Kontakt 151`,
    })

    await hand.click(wege)
    await hand.click(within(bereich).getByRole('checkbox', { name: telefon }))
    await hand.click(wege)

    // **Abgewählt bleibt sichtbar abgewählt**, auch mit zugeklappter Wahl.
    expect(within(bereich).getByText(telefon)).toHaveClass('line-through')

    await hand.click(
      within(bereich).getByRole('button', {
        name: einsetzen(texte.uebernehmen, { anzahl: '50' }),
      }),
    )

    await waitFor(() => {
      expect(uebernehmen).toHaveBeenCalled()
    })

    const [auswahl] = uebernehmen.mock.calls[0] as [KontaktDTO[]]

    expect(auswahl).toHaveLength(50)
    expect(auswahl[0]?.wege).toEqual([
      { art: 'email', wert: 'k150@example.org' },
    ])
    expect(auswahl[1]?.wege).toHaveLength(2)
  })
})
