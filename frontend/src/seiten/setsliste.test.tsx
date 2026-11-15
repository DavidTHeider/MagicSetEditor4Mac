// **Die Setliste in beiden Gestalten** — UX-11
// (#1683):
// suchen über Name, Spiel und Stil, ordnen, und als Zeile, wenn die gemerkte
// Ansicht es sagt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { useSetSpeichern } from '../haken/useSetSpeichern'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sets } from './Sets'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.bausteine.listenansicht
const oeffnen = kataloge.de.seiten.sets.oeffnen

function satz(id: string, name: string, karten: number, spiel: string): SetDTO {
  return {
    id,
    name,
    spiel,
    stil: 'm15',
    herkunft: `/pfad/${name}.mse-set`,
    karten,
    herkunftFehlt: false,
  }
}

function MitLauf() {
  const speichern = useSetSpeichern()

  return <Sets wechseln={vi.fn()} waehlen={vi.fn()} speichern={speichern} />
}

function zeige(ansicht: string) {
  belegeApi({
    Sets: () =>
      Promise.resolve({
        sets: [
          satz('1', 'Zeder', 12, 'magic'),
          satz('2', 'Ahorn', 3, 'yugioh'),
          satz('3', 'Birke', 7, 'magic'),
        ],
      }),
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'arbeitsplatz.ansicht.sets',
            wert: 'erben',
            wirksam: ansicht,
            werte: ['erben', 'kacheln', 'liste'],
            bereich: 'arbeitsplatz',
          },
        ],
      }),
  })

  render(
    <Sprachraum sprache="de">
      <MitLauf />
    </Sprachraum>,
  )
}

// namen liest die Sets in der Reihenfolge, in der sie dastehen.
function namen() {
  return screen
    .getAllByRole('button', { name: new RegExp(` — ${oeffnen}$`) })
    .map((knopf) => knopf.textContent)
}

describe('Die Setliste', () => {
  it('zeigt die Sets als Zeilen, wenn die gemerkte Ansicht es sagt', async () => {
    zeige('liste')

    expect(await screen.findByRole('list')).toHaveAttribute(
      'data-form',
      'liste',
    )
    expect(namen()).toEqual(['Zeder', 'Ahorn', 'Birke'])
  })

  it('sucht über das Spiel und nicht nur über den Namen', async () => {
    zeige('kacheln')
    await screen.findByRole('list')

    await userEvent.type(screen.getByLabelText(worte.suche), 'yugioh')

    expect(namen()).toEqual(['Ahorn'])
  })

  it('ordnet nach Name und nach Kartenzahl', async () => {
    zeige('kacheln')
    await screen.findByRole('list')
    const ordnung = screen.getByLabelText(worte.sortierung)

    await userEvent.selectOptions(ordnung, 'name')
    expect(namen()).toEqual(['Ahorn', 'Birke', 'Zeder'])

    await userEvent.selectOptions(ordnung, 'karten')
    expect(namen()).toEqual(['Zeder', 'Birke', 'Ahorn'])

    await userEvent.selectOptions(ordnung, 'bestand')
    expect(namen()).toEqual(['Zeder', 'Ahorn', 'Birke'])
  })

  // **Sechzig Sets auf einer Seite der Liste** — UXR-04 (#1912). Mit fünf
  // gleichrangigen Knöpfen je Karte waren es 420 Bedienelemente, R6 aus
  // `docs/UIREVIEW.md` erlaubt 200. Jetzt trägt eine Karte ihren Namen, ihr
  // Menü und ihre Marken.
  it('zeichnet mit sechzig Sets weniger als zweihundert Bedienelemente', async () => {
    belegeApi({
      Sets: () =>
        Promise.resolve({
          sets: Array.from({ length: 60 }, (_, stelle) =>
            satz(String(stelle), `Set ${String(stelle)}`, stelle, 'magic'),
          ),
        }),
    })
    render(
      <Sprachraum sprache="de">
        <MitLauf />
      </Sprachraum>,
    )

    const liste = await screen.findByRole('region', {
      name: kataloge.de.seiten.sets.titel,
    })
    await within(liste).findByRole('button', { name: `Set 59 — ${oeffnen}` })

    expect(within(liste).getAllByRole('button').length).toBeLessThan(200)
  })
})
