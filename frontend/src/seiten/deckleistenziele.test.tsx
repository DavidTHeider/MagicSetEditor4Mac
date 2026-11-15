// **Die Leiste über dem Deck führt hin** — AP-9
// (#2366):
// „Fehlt in der Sammlung“ schlägt die Lasche „Prüfung“ auf, wo dieselben
// Karten schon stehen; „Hauptdeck“ führt den Fokus zu den Plätzen.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe, platzprobe } from './deckprobe'
import { Decks } from './Decks'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// einstellungen sind die vier Schlüssel des Arbeitsplatzes — mit
// eingeklapptem Inspector, wenn `verborgen`.
function einstellungen(inspektor: string) {
  const werte = {
    'arbeitsplatz.navigator': 'sichtbar',
    'arbeitsplatz.inspektor': inspektor,
    'arbeitsplatz.navigatorbreite': '240',
    'arbeitsplatz.inspektorbreite': '320',
  }

  return () =>
    Promise.resolve({
      einstellungen: Object.entries(werte).map(([schluessel, wert]) => ({
        schluessel,
        wert,
        wirksam: wert,
        werte: [],
        bereich: 'arbeitsplatz',
      })),
    })
}

async function zeige(inspektor = 'sichtbar') {
  const EinstellungSetzen = vi.fn().mockResolvedValue({})
  const DeckAbgleichen = vi.fn(() =>
    Promise.resolve({
      zeilen: [
        {
          printing: 'abc',
          druckID: '',
          vorlaeufig: false,
          bezeichnung: 'Blitz',
          gebraucht: 4,
          vorhanden: 2,
          fehlend: 2,
          grund: '',
        },
      ],
      fehlend: 2,
      unbekannt: 0,
      fehler: null,
    }),
  )
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Deck: () =>
      Promise.resolve({
        deck: deckprobe({ plaetze: [platzprobe()], hauptdeck: 4 }),
        fehler: null,
      }),
    DeckAbgleichen,
    Einstellungen: einstellungen(inspektor),
    EinstellungSetzen,
  })
  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )
  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )

  return { DeckAbgleichen, EinstellungSetzen }
}

const leiste = async () =>
  within(await screen.findByRole('toolbar', { name: texte.leiste.bezeichnung }))

describe('die Ziele der Leiste', () => {
  it('führt vom Fehlenden in die Prüfung, wo es schon steht', async () => {
    const { DeckAbgleichen } = await zeige()

    await userEvent.click(
      await (
        await leiste()
      ).findByRole('button', { name: `${texte.leiste.fehlt} 2` }),
    )

    expect(
      await screen.findByRole('tab', {
        name: texte.gebietPruefung,
        selected: true,
      }),
    ).toBeInTheDocument()
    expect(await screen.findByText(/2 Karten fehlen/)).toBeInTheDocument()
    // **Einmal gefragt** — Leiste und Lasche lesen dieselbe Antwort.
    expect(DeckAbgleichen).toHaveBeenCalledTimes(1)
  })

  it('führt vom Hauptdeck zu den Plätzen', async () => {
    await zeige()

    await userEvent.click(
      (await leiste()).getByRole('button', { name: /^Hauptdeck 4 / }),
    )

    expect(screen.getByRole('heading', { name: texte.plaetze })).toHaveFocus()
  })

  // **Auch bei eingeklapptem Inspector** — der Klick ist der Wunsch, ihn zu
  // sehen, und wird gemerkt wie „Inspector einblenden“.
  it('klappt den Inspector auf, wenn er eingeklappt war', async () => {
    const { EinstellungSetzen } = await zeige('verborgen')
    expect(screen.queryByRole('tab', { name: texte.gebietPruefung })).toBeNull()

    await userEvent.click(
      await (
        await leiste()
      ).findByRole('button', { name: `${texte.leiste.fehlt} 2` }),
    )

    expect(await screen.findByText(/2 Karten fehlen/)).toBeInTheDocument()
    expect(EinstellungSetzen).toHaveBeenCalledWith(
      'arbeitsplatz.inspektor',
      'sichtbar',
    )
  })
})
