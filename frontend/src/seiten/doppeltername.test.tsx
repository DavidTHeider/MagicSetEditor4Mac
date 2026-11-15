// **Ein doppelter Name steht am Namensfeld** — ENG-05.AK-03
// (#1730),
// nativer Befund F08.
//
// Bis zum 2026-09-23 kam ein zweites Deck gleichen Namens als Hinweis über der
// Seite an, und zwar mit dem falschen Satz: „Der Eintrag wurde zwischenzeitlich
// geändert. Bitte neu laden." Wer neu lud, fand dasselbe Formular. Seit der
// Bestand sagt, **welches** Feld doppelt ist, steht die Meldung dort (§6).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, formatprobe } from './deckprobe'
import { kataloge } from '../sprache/kataloge'
import { assistent } from '../test/turnierassistent'

import type { Api } from '../api'

afterEach(() => {
  raeumeApi()
})

const doppelt = {
  code: 'konflikt',
  schluessel: 'doppelterName',
  meldung: 'Diesen Namen gibt es hier schon — wähle einen anderen.',
  feld: 'name',
  werte: {},
} as const

async function deckanlage(teile: Partial<Api>) {
  const Decks_ = vi.fn(() => Promise.resolve({ decks: [], fehler: null }))

  belegeApi({
    Decks: Decks_,
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  const texte = kataloge.de.seiten.decks

  await screen.findByText(texte.leer)
  await userEvent.click(screen.getByRole('button', { name: texte.neu }))

  const anlage = screen.getByRole('region', { name: texte.neu })
  const feld = within(anlage).getByLabelText(texte.name)

  await userEvent.type(feld, 'Aggro')
  await userEvent.click(
    within(anlage).getByRole('button', { name: texte.anlegen }),
  )

  return { feld, geladen: Decks_ }
}

describe('ein doppelter Deckname', () => {
  it('steht am Namensfeld, einmal, und der Name bleibt zum Berichtigen', async () => {
    const { feld, geladen } = await deckanlage({
      DeckAnlegen: () => Promise.resolve({ deck: null, fehler: doppelt }),
    })

    await waitFor(() => {
      expect(feld).toHaveAttribute('aria-invalid', 'true')
    })

    expect(feld).toHaveAccessibleDescription(kataloge.de.fehler.doppelterName)
    expect(screen.getAllByText(kataloge.de.fehler.doppelterName)).toHaveLength(
      1,
    )
    expect(feld).toHaveValue('Aggro')

    // **Nach einem gescheiterten Schreiben wird nicht nachgeladen** (§7).
    expect(geladen).toHaveBeenCalledTimes(1)
  })

  it('geht weg, sobald der Name sich ändert', async () => {
    const { feld } = await deckanlage({
      DeckAnlegen: () => Promise.resolve({ deck: null, fehler: doppelt }),
    })

    await waitFor(() => {
      expect(feld).toHaveAttribute('aria-invalid', 'true')
    })

    await userEvent.type(feld, ' II')

    expect(feld).toHaveAttribute('aria-invalid', 'false')
    expect(
      screen.queryByText(kataloge.de.fehler.doppelterName),
    ).not.toBeInTheDocument()
  })

  it('lässt einen Fehler ohne Feld über der Seite stehen', async () => {
    const { feld } = await deckanlage({
      DeckAnlegen: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'konflikt',
            schluessel: 'konfliktgrund',
            meldung: 'Konflikt: das Profil ist voll',
            werte: { grund: 'das Profil ist voll' },
          },
        }),
    })

    expect(
      await screen.findByText('Das ging nicht: das Profil ist voll.', {
        exact: false,
      }),
    ).toBeInTheDocument()
    expect(feld).toHaveAttribute('aria-invalid', 'false')
  })
})

describe('ein doppelter Turniername', () => {
  it('steht am Namensfeld', async () => {
    belegeApi({
      Turniere: () => Promise.resolve({ turniere: [], fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      TurnierAnlegen: () => Promise.resolve({ turnier: null, fehler: doppelt }),
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    const texte = kataloge.de.seiten.turniere
    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    // **Der Assistent geht zum Namen zurück** (DOM-04.AK-01, #1699) — und
    // das Feld dort ist ein neues Element.
    await waitFor(() => {
      expect(within(bereich).getByLabelText(texte.name)).toHaveAttribute(
        'aria-invalid',
        'true',
      )
    })
    const feld = within(bereich).getByLabelText(texte.name)

    // **Befund und Hinweis zugleich** — `Feld` nennt beide, den Befund zuerst.
    expect(feld).toHaveAccessibleDescription(
      expect.stringContaining(kataloge.de.fehler.doppelterName),
    )

    await userEvent.type(feld, ' II')

    expect(feld).toHaveAttribute('aria-invalid', 'false')
  })
})
