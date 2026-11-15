// **Ein Feld, das ein Spiel nennt, schlägt die installierten vor** —
// UX-06.AK-01 (#1678).
//
// Format und Turnier nennen ihr Spiel. Bis zum 2026-09-25 stand dort ein
// nacktes Textfeld; jetzt schlägt es vor, was installiert ist, und nimmt
// weiter jedes Spiel an — ein eingelesenes Turnier kommt von einem anderen
// Rechner und darf eines nennen, das hier fehlt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { assistent } from '../test/turnierassistent'
import { bauartprobe } from './deckprobe'

afterEach(() => {
  raeumeApi()
})

function paket(art: string, name: string) {
  return {
    art,
    name,
    anzeigename: name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 0,
    artrang: 0,
    herkunftsart: 'mitgeliefert' as const,
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
  }
}

const pakete = () =>
  Promise.resolve({
    pakete: [
      paket('mse-game', 'magic'),
      paket('mse-game', 'probe'),
      paket('mse-style', 'magic-m15'),
    ],
    befunde: [],
    fehler: null,
  })

// vorgeschlagen liest die Vorschläge unter einem Feld — die Einträge der
// `datalist`, auf die sein `list` zeigt.
async function vorgeschlagen(feld: HTMLElement) {
  const liste = feld.getAttribute('list')

  if (liste === null) {
    throw new Error('das Feld hat keine Vorschläge')
  }

  const eintraege = () =>
    [...(document.getElementById(liste)?.querySelectorAll('option') ?? [])].map(
      (eintrag) => eintrag.getAttribute('value'),
    )

  // **Gewartet wird auf die Pakete**, nicht auf das Feld: das steht schon da,
  // bevor sie ankommen.
  await waitFor(() => {
    expect(eintraege().length).toBeGreaterThan(0)
  })

  return eintraege()
}

describe('ein Spiel nennen', () => {
  it('im Format: die installierten Spiele stehen als Vorschlag da', async () => {
    const texte = kataloge.de.seiten.decks

    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      Pakete: pakete,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.formateKnopf }),
    )
    const bereich = await screen.findByRole('region', { name: texte.formate })
    const spiel = within(bereich).getByLabelText(texte.spiel)

    expect(await vorgeschlagen(spiel)).toEqual(['magic', 'probe'])

    // **Und ein anderes bleibt möglich.**
    await userEvent.clear(spiel)
    await userEvent.type(spiel, 'fremd')

    expect(spiel).toHaveValue('fremd')
  })

  it('im Turnier: dieselben Vorschläge', async () => {
    const texte = kataloge.de.seiten.turniere

    belegeApi({
      Turniere: () => Promise.resolve({ turniere: [], fehler: null }),
      Pakete: pakete,
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    // **Der Assistent steht in einer Schublade** — seit UXR-12 (#1920).
    await assistent()
    const spiel = await screen.findByLabelText(texte.spiel)

    expect(await vorgeschlagen(spiel)).toEqual(['magic', 'probe'])
  })
})
