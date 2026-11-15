// **Ein Deck umbenennen** — gefunden beim Aktionskatalog, UX-10
// (#1682):
// der Knopf schrieb den alten Namen und eine leere Notiz zurück.
//
// **Eine eigene Datei**, weil `decks.test.tsx` mit ihr über der Zeilengrenze
// läge (§2); die Vorrichtung ist dieselbe.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { aufschlagen, bauartprobe, deckprobe, formatprobe } from './deckprobe'

import type { Api } from '../api'
import { deckhandlung, deckmenue } from '../test/deckhandlung'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    // **Umbenennen steht am offenen Deck** (UXR-11, #1919) — jeder Lauf
    // schlägt es zuerst auf.
    Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

describe('Ein Deck umbenennen', () => {
  // **Umbenennen fragt nach dem neuen Namen — und behält die Notiz.** Bis
  // zum 2026-09-25 rief der Knopf `DeckUmbenennen` mit dem alten Namen und
  // einer leeren Notiz: umbenannt wurde nichts, und die Notiz war weg.
  it('benennt ein Deck um und behält seine Notiz', async () => {
    const DeckUmbenennen = vi.fn(() =>
      Promise.resolve({ deck: deckprobe({ name: 'Midrange' }), fehler: null }),
    )

    zeige({
      Decks: () =>
        Promise.resolve({
          decks: [deckprobe({ notiz: 'für Freitag' })],
          fehler: null,
        }),
      DeckUmbenennen,
    })

    await aufschlagen()
    await deckhandlung('Aggro', texte.umbenennen)
    expect(DeckUmbenennen).not.toHaveBeenCalled()

    const feld = screen.getByLabelText(texte.neuerName)
    expect(feld).toHaveValue('Aggro')
    await userEvent.clear(feld)
    await userEvent.type(feld, 'Midrange')
    await userEvent.click(
      screen.getByRole('button', { name: texte.nameSichern }),
    )

    await waitFor(() => {
      expect(DeckUmbenennen).toHaveBeenCalledWith(
        '1',
        '1',
        'Midrange',
        'für Freitag',
      )
    })
  })

  it('lässt das Umbenennen abbrechen, ohne zu schreiben', async () => {
    const DeckUmbenennen = vi.fn()

    zeige({ DeckUmbenennen })

    await aufschlagen()
    await deckhandlung('Aggro', texte.umbenennen)
    await userEvent.click(
      screen.getByRole('button', { name: texte.umbenennenAbbrechen }),
    )

    expect(screen.queryByLabelText(texte.neuerName)).toBeNull()
    expect(DeckUmbenennen).not.toHaveBeenCalled()
  })

  // **Escape schließt, ohne zu schreiben** — und der Fokus kehrt zum Knopf
  // zurück, der das Popover geöffnet hat (UX-08.AK-01, #1680).
  it('schließt mit Escape und gibt den Fokus dem Knopf zurück', async () => {
    const DeckUmbenennen = vi.fn()

    zeige({ DeckUmbenennen })

    await aufschlagen()

    const knopf = await deckmenue('Aggro')
    await deckhandlung('Aggro', texte.umbenennen)
    expect(screen.getByLabelText(texte.neuerName)).toHaveFocus()

    await userEvent.keyboard('{Escape}')

    expect(screen.queryByLabelText(texte.neuerName)).toBeNull()
    expect(knopf).toHaveFocus()
    expect(DeckUmbenennen).not.toHaveBeenCalled()
  })
})

// **Ein Konflikt bietet den heutigen Stand an** — UX-08.AK-02
// (#1680).
// „Bitte neu laden" ohne einen Knopf dazu schickte den Menschen suchen, wie.
describe('Ein Konflikt beim Umbenennen', () => {
  it('bietet „Neu laden" an, holt den Stand und räumt die Meldung', async () => {
    const Decks = vi.fn(() =>
      Promise.resolve({ decks: [deckprobe()], fehler: null }),
    )

    zeige({
      Decks,
      DeckUmbenennen: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Der Eintrag wurde zwischenzeitlich geändert.',
          },
        }),
    })

    await aufschlagen()
    await deckhandlung('Aggro', texte.umbenennen)
    await userEvent.clear(screen.getByLabelText(texte.neuerName))
    await userEvent.type(screen.getByLabelText(texte.neuerName), 'Midrange')
    await userEvent.click(
      screen.getByRole('button', { name: texte.nameSichern }),
    )

    const neuLaden = await screen.findByRole('button', {
      name: kataloge.de.bausteine.zustaende.neuLaden,
    })
    const vorher = Decks.mock.calls.length

    await userEvent.click(neuLaden)

    await waitFor(() => {
      expect(Decks.mock.calls.length).toBeGreaterThan(vorher)
    })
    expect(screen.queryByRole('alert')).toBeNull()
  })

  // **Und ein offenes Deck wird mit neu gelesen** — der Konflikt kann es
  // betreffen, und ein Stand von vorhin daneben wäre der nächste.
  it('liest auch das offene Deck neu', async () => {
    const Deck = vi.fn(() =>
      Promise.resolve({ deck: deckprobe(), fehler: null }),
    )

    zeige({
      Deck,
      DeckUmbenennen: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Der Eintrag wurde zwischenzeitlich geändert.',
          },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    // **Auf die Wirkung und nicht auf den Ruf** (§3): das Deck steht
    // aufgeschlagen da, wenn der Satz „kein Deck aufgeschlagen" weg ist.
    await waitFor(() => {
      expect(screen.queryByText(texte.keinDeckOffen)).toBeNull()
    })
    expect(Deck).toHaveBeenCalledTimes(1)

    await deckhandlung('Aggro', texte.umbenennen)
    await userEvent.type(screen.getByLabelText(texte.neuerName), ' neu')
    await userEvent.click(
      screen.getByRole('button', { name: texte.nameSichern }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.neuLaden,
      }),
    )

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledTimes(2)
    })
  })

  // **Ein anderer Fehler bietet es nicht an** — sein Satz sagt selbst, was zu
  // tun ist, und ein Neuladen änderte daran nichts.
  it('bietet es bei einem anderen Fehler nicht an', async () => {
    zeige({
      DeckUmbenennen: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'validierung',
            schluessel: '',
            meldung: 'Der Name ist zu lang.',
          },
        }),
    })

    await aufschlagen()
    await deckhandlung('Aggro', texte.umbenennen)
    await userEvent.type(screen.getByLabelText(texte.neuerName), ' lang')
    await userEvent.click(
      screen.getByRole('button', { name: texte.nameSichern }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Name ist zu lang.',
    )
    expect(
      screen.queryByRole('button', {
        name: kataloge.de.bausteine.zustaende.neuLaden,
      }),
    ).toBeNull()
  })
})
