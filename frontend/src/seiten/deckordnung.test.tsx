// **Die Deckliste sucht und ordnet** — UX-11
// (#1683):
// dieselbe Listenansicht wie bei den Sets, mit eigenem Suchtext und eigenen
// Ordnungen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.bausteine.listenansicht
const waehlen = kataloge.de.seiten.decks.waehlen

function zeige(ansicht = 'liste') {
  belegeApi({
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'arbeitsplatz.ansicht.decks',
            wert: 'erben',
            wirksam: ansicht,
            werte: ['erben', 'liste', 'kacheln'],
            bereich: 'arbeitsplatz',
          },
        ],
      }),
    Decks: () =>
      Promise.resolve({
        decks: [
          deckprobe({ id: '1', name: 'Zorn', hauptdeck: 60, notiz: '' }),
          deckprobe({
            id: '2',
            name: 'Asche',
            hauptdeck: 40,
            notiz: 'Turnier',
          }),
          deckprobe({ id: '3', name: 'Brandung', hauptdeck: 50, notiz: '' }),
        ],
        fehler: null,
      }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

// namen liest die Decks in der Reihenfolge, in der sie dastehen.
function namen() {
  return screen
    .getAllByRole('button', { name: new RegExp(`^${waehlen}: `) })
    .map((knopf) => knopf.getAttribute('aria-label')?.slice(waehlen.length + 2))
}

describe('Die Deckliste', () => {
  it('zeigt die Decks als Kacheln, wenn die gemerkte Ansicht es sagt', async () => {
    zeige('kacheln')

    expect(await screen.findByRole('list')).toHaveAttribute(
      'data-form',
      'kacheln',
    )
    expect(namen()).toEqual(['Zorn', 'Asche', 'Brandung'])
  })

  it('sucht auch in der Notiz', async () => {
    zeige()
    await screen.findByText('Zorn')

    await userEvent.type(screen.getByLabelText(worte.suche), 'turnier')

    expect(namen()).toEqual(['Asche'])
  })

  it('ordnet nach Name und nach der Größe des Hauptdecks', async () => {
    zeige()
    await screen.findByText('Zorn')
    const ordnung = screen.getByLabelText(worte.sortierung)

    await userEvent.selectOptions(ordnung, 'name')
    expect(namen()).toEqual(['Asche', 'Brandung', 'Zorn'])

    await userEvent.selectOptions(ordnung, 'karten')
    expect(namen()).toEqual(['Zorn', 'Brandung', 'Asche'])
  })
})
