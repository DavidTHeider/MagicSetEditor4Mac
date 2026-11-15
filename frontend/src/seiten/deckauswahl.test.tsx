// **Die Decks mit fehlenden Karten, gefiltert** — AP-8, Teil 2b
// (#2365):
// die Kachel „Decks“ öffnet die Deckliste mit genau ihren Decks. Der Filter
// nennt sich und lässt sich lösen — wie in Set- und Paketliste.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'
import { Decks } from './Decks'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.decks

function zeige() {
  belegeApi({
    Decks: () =>
      Promise.resolve({
        decks: [
          deckprobe({ id: '1', name: 'Waldschritt' }),
          deckprobe({ id: '2', name: 'Glutpfad' }),
          deckprobe({ id: '3', name: 'Moorwacht' }),
        ],
        fehler: null,
      }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
  })
  render(
    <Sprachraum sprache="de">
      <Anschlussbuehne
        anschluss={{
          art: 'deckauswahl',
          name: '2 Decks mit fehlenden Karten',
          deckIDs: ['1', '3'],
        }}
      >
        <Decks wechseln={vi.fn()} />
      </Anschlussbuehne>
    </Sprachraum>,
  )
}

const liste = () => within(screen.getByRole('region', { name: texte.titel }))

describe('die Decks mit fehlenden Karten', () => {
  it('stehen gefiltert da, und der Filter lässt sich lösen', async () => {
    zeige()
    expect(await screen.findByText('Glutpfad')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

    expect(
      await screen.findByText('2 Decks mit fehlenden Karten'),
    ).toBeInTheDocument()
    expect(liste().queryByText('Glutpfad')).toBeNull()
    expect(liste().getByText('Waldschritt')).toBeInTheDocument()
    expect(liste().getByText('Moorwacht')).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.auswahlLoesen }),
    )

    expect(await liste().findByText('Glutpfad')).toBeInTheDocument()
    expect(screen.queryByText('2 Decks mit fehlenden Karten')).toBeNull()
  })

  // **Neu gezeichnet kommt die Auswahl nicht wieder** — bis AP-6 (#2363)
  // zeichnete ein Wechsel der Laschen „Decks“ und „Formate“ die Liste neu;
  // seit die Formate in einer Schublade aufgehen, tut es das Aus- und
  // Einblenden des Navigators.
  it('kommen nach dem Neuzeichnen der Liste nicht wieder', async () => {
    zeige()
    expect(await screen.findByText('Glutpfad')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))
    expect(
      await screen.findByText('2 Decks mit fehlenden Karten'),
    ).toBeInTheDocument()

    const regionen = kataloge.de.bausteine.arbeitsplatz
    await userEvent.click(
      screen.getByRole('button', { name: regionen.navigatorZu }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: regionen.navigatorAuf }),
    )

    // **Der Beleg, dass die Liste wieder steht** — erst dann sagt das Fehlen
    // des Satzes etwas.
    expect(await liste().findByText('Glutpfad')).toBeInTheDocument()
    expect(screen.queryByText('2 Decks mit fehlenden Karten')).toBeNull()
  })
})
