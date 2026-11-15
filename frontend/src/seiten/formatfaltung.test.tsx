// **Die Formate ohne Formularwand** — UXR-11
// (#1919).
//
// Jedes Format stand als offenes Formular da: Name, vier Grenzen, „Sideboard
// ohne Grenze", Scryfall-Name, Sichern und Löschen. Bei zehn Formaten — so
// viele, wie die Bauarten nahelegen — waren das siebzig Felder auf einer
// Lasche (R3 in `docs/UIREVIEW.md`). Jetzt steht ein Format als Zeile da, und
// sein Formular erst, wenn jemand es bearbeitet — eines zugleich.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe } from './deckprobe'
import { deckformate } from '../test/deckbestand'
import { kataloge } from '../sprache/kataloge'
import { zaehle } from '../test/bedienzaehlung'

const texte = kataloge.de.seiten.decks
const bearbeiten = kataloge.de.bausteine.aktionskatalog.aktionen.bearbeiten.kurz

afterEach(() => {
  raeumeApi()
})

async function zeige(): Promise<HTMLElement> {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: deckformate, fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: texte.formateKnopf }),
  )

  await screen.findByRole('button', { name: `${bearbeiten}: standard` })

  return screen.getByRole('region', { name: texte.formate })
}

describe('die Formatlasche mit zehn Formaten', () => {
  it('zeichnet nur das Anlegen als Felder, keine Formatzeile', async () => {
    const lasche = await zeige()

    // **Vier Felder sind das Anlegen**: Bauart, Name, Spiel, Scryfall-Name.
    expect(zaehle(lasche).felder).toBe(4)
    expect(
      within(lasche).queryByRole('button', {
        name: `${texte.formatSichern}: standard`,
      }),
    ).toBeNull()
    expect(
      within(lasche).getAllByRole('button', {
        name: new RegExp(`^${bearbeiten}: `),
      }),
    ).toHaveLength(10)
  })

  it('öffnet das Formular eines Formats auf Wunsch, und immer nur eines', async () => {
    const lasche = await zeige()
    const standard = within(lasche).getByRole('button', {
      name: `${bearbeiten}: standard`,
    })

    await userEvent.click(standard)

    expect(standard).toHaveAttribute('aria-expanded', 'true')
    expect(
      within(lasche).getByRole('button', {
        name: `${texte.formatSichern}: standard`,
      }),
    ).toBeInTheDocument()
    expect(
      within(lasche).getByRole('button', {
        name: `${texte.formatLoeschen}: standard`,
      }),
    ).toBeInTheDocument()

    await userEvent.click(
      within(lasche).getByRole('button', { name: `${bearbeiten}: pioneer` }),
    )

    expect(standard).toHaveAttribute('aria-expanded', 'false')
    expect(
      within(lasche).queryByRole('button', {
        name: `${texte.formatSichern}: standard`,
      }),
    ).toBeNull()

    await userEvent.click(
      within(lasche).getByRole('button', { name: `${bearbeiten}: pioneer` }),
    )

    expect(zaehle(lasche).felder).toBe(4)
  })
})
