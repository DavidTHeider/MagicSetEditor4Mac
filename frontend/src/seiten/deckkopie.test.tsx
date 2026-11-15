// **Ein Deck kopieren** — DOM-03.AK-02
// (#1698).
//
// **Eine eigene Datei**, weil `decks.test.tsx` mit ihr über der Zeilengrenze
// läge (§2); die Vorrichtung ist dieselbe wie beim Umbenennen.
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
    Deck: (deckID: string) =>
      Promise.resolve({ deck: deckprobe({ id: deckID }), fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

async function oeffneKopie() {
  // **Erst aufschlagen** — seit UXR-11 (#1919) steht Duplizieren nur am
  // offenen Deck.
  await aufschlagen()
  await deckhandlung('Aggro', texte.duplizieren)
}

describe('Ein Deck kopieren', () => {
  // **Der Name ist vorgeschlagen, nicht erzwungen** — im Spiel ist er
  // einmalig, und die Kopie braucht einen eigenen. Danach geht sie auf: wer
  // kopiert, will an der Kopie weiterbauen.
  it('kopiert unter dem vorgeschlagenen Namen und schlägt die Kopie auf', async () => {
    const DeckKopieren = vi.fn(() =>
      Promise.resolve({
        deck: deckprobe({ id: '2', name: 'Aggro (Kopie)' }),
        fehler: null,
      }),
    )
    const Deck = vi.fn((deckID: string) =>
      Promise.resolve({ deck: deckprobe({ id: deckID }), fehler: null }),
    )

    zeige({ DeckKopieren, Deck })
    await oeffneKopie()

    expect(DeckKopieren).not.toHaveBeenCalled()
    expect(screen.getByLabelText(texte.kopieName)).toHaveValue('Aggro (Kopie)')

    await userEvent.click(
      screen.getByRole('button', { name: texte.kopieAnlegen }),
    )

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledWith('2')
    })
    expect(DeckKopieren).toHaveBeenCalledWith('1', 'Aggro (Kopie)')
    expect(screen.queryByLabelText(texte.kopieName)).toBeNull()
  })

  it('nimmt einen eigenen Namen, geglättet', async () => {
    const DeckKopieren = vi.fn(() =>
      Promise.resolve({ deck: deckprobe({ id: '2' }), fehler: null }),
    )

    zeige({ DeckKopieren })
    await oeffneKopie()

    const feld = screen.getByLabelText(texte.kopieName)
    await userEvent.clear(feld)
    await userEvent.type(feld, '  Aggro für Freitag ')
    await userEvent.click(
      screen.getByRole('button', { name: texte.kopieAnlegen }),
    )

    await waitFor(() => {
      expect(DeckKopieren).toHaveBeenCalledWith('1', 'Aggro für Freitag')
    })
  })

  // **Ohne Namen ruht der Knopf** — und das Feld sagt, warum.
  it('ruht ohne Namen', async () => {
    zeige({ DeckKopieren: vi.fn() })
    await oeffneKopie()

    await userEvent.clear(screen.getByLabelText(texte.kopieName))

    expect(
      screen.getByRole('button', { name: texte.kopieAnlegen }),
    ).toBeDisabled()
  })

  it('lässt das Kopieren abbrechen, ohne zu schreiben', async () => {
    const DeckKopieren = vi.fn()

    zeige({ DeckKopieren })
    await oeffneKopie()
    await userEvent.click(
      screen.getByRole('button', { name: texte.kopieAbbrechen }),
    )

    expect(screen.queryByLabelText(texte.kopieName)).toBeNull()
    expect(DeckKopieren).not.toHaveBeenCalled()
  })

  // **Escape schließt, ohne zu schreiben** — und der Fokus kehrt zum Knopf
  // zurück, der das Popover geöffnet hat (UX-08.AK-01).
  it('schließt mit Escape und gibt den Fokus dem Knopf zurück', async () => {
    const DeckKopieren = vi.fn()

    zeige({ DeckKopieren })
    await aufschlagen()

    const knopf = await deckmenue('Aggro')
    await deckhandlung('Aggro', texte.duplizieren)
    expect(screen.getByLabelText(texte.kopieName)).toHaveFocus()

    await userEvent.keyboard('{Escape}')

    expect(screen.queryByLabelText(texte.kopieName)).toBeNull()
    expect(knopf).toHaveFocus()
    expect(DeckKopieren).not.toHaveBeenCalled()
  })

  // **Ein vergebener Name steht am Namen** (§6) — nicht über der Seite, und
  // das Popover bleibt offen, damit der nächste Versuch einer ist.
  it('zeigt einen vergebenen Namen am Feld', async () => {
    const meldung = 'Diesen Namen gibt es schon.'

    zeige({
      DeckKopieren: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'validierung',
            schluessel: 'doppelterName',
            meldung,
            feld: 'name',
          },
        }),
    })
    await oeffneKopie()
    await userEvent.click(
      screen.getByRole('button', { name: texte.kopieAnlegen }),
    )

    const feld = await screen.findByLabelText(texte.kopieName)
    await waitFor(() => {
      expect(feld).toHaveAccessibleDescription(kataloge.de.fehler.doppelterName)
    })
    expect(screen.queryByRole('alert', { name: meldung })).toBeNull()

    // **Wer tippt, räumt den Befund** — er gilt dem Namen von eben.
    await userEvent.type(feld, ' 2')
    expect(feld).not.toHaveAccessibleDescription(
      kataloge.de.fehler.doppelterName,
    )
  })

  // **Ein anderer Fehler steht über der Seite**, und das Popover bleibt.
  it('meldet einen anderen Fehler über der Seite', async () => {
    zeige({
      DeckKopieren: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'nichtGefunden',
            meldung: 'Das Deck gibt es nicht mehr.',
          },
        }),
    })
    await oeffneKopie()
    await userEvent.click(
      screen.getByRole('button', { name: texte.kopieAnlegen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.getByLabelText(texte.kopieName)).toBeInTheDocument()
  })

  // **Eine verspätete Kopie schlägt nicht über eine neue Wahl auf** —
  // Befund 9 des Zweitblicks über Phase 6 von 0.9.0 (§7, Regel 2): wer
  // inzwischen ein anderes Deck geöffnet hat, behält es; die Kopie entsteht
  // trotzdem und **steht in der Liste** (Regel 3) — das ist zugleich der
  // Beleg, dass die Antwort angekommen ist.
  it('schlägt die Kopie nicht auf, wenn inzwischen ein anderes Deck offen ist', async () => {
    const { freigeben, Deck } = await verspaeteteKopie()

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Kontrolle` }),
    )
    await deckmenue('Kontrolle')

    freigeben()

    expect(await screen.findByText('Aggro (Kopie)')).toBeInTheDocument()
    expect(await deckmenue('Kontrolle')).toBeInTheDocument()
    expect(Deck).not.toHaveBeenCalledWith('2')
  })

  // **Auch nicht nach A → B → A** — Rest R3 des Nachlaufs über Phase 6 von
  // 0.9.0: wer zum Original zurückkehrt, hat es neu aufgeschlagen, und eine
  // Kopie von vorhin ist keine Antwort auf diese Wahl.
  it('schlägt die Kopie nicht auf, wenn das Original inzwischen neu geöffnet wurde', async () => {
    const { freigeben, Deck } = await verspaeteteKopie()

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Kontrolle` }),
    )
    await deckmenue('Kontrolle')
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )
    await deckmenue('Aggro')

    freigeben()

    expect(await screen.findByText('Aggro (Kopie)')).toBeInTheDocument()
    expect(await deckmenue('Aggro')).toBeInTheDocument()
    expect(Deck).not.toHaveBeenCalledWith('2')
  })
})

// verspaeteteKopie stößt eine Kopie an, deren Antwort `freigeben` erst
// schickt; das Popover ist danach zu. Die Liste nennt die Kopie, sobald sie
// geschrieben ist.
async function verspaeteteKopie() {
  let geschrieben = false
  let freigeben: () => void = () => undefined
  const DeckKopieren = vi.fn(
    () =>
      new Promise<{ deck: ReturnType<typeof deckprobe>; fehler: null }>(
        (fertig) => {
          freigeben = () => {
            geschrieben = true
            fertig({
              deck: deckprobe({ id: '2', name: 'Aggro (Kopie)' }),
              fehler: null,
            })
          }
        },
      ),
  )
  const Deck = vi.fn((deckID: string) =>
    Promise.resolve({
      deck: deckprobe({
        id: deckID,
        name: { '2': 'Aggro (Kopie)', '3': 'Kontrolle' }[deckID] ?? 'Aggro',
      }),
      fehler: null,
    }),
  )

  zeige({
    DeckKopieren,
    Deck,
    Decks: () =>
      Promise.resolve({
        decks: [
          deckprobe(),
          deckprobe({ id: '3', name: 'Kontrolle' }),
          ...(geschrieben
            ? [deckprobe({ id: '2', name: 'Aggro (Kopie)' })]
            : []),
        ],
        fehler: null,
      }),
  })
  await oeffneKopie()
  await userEvent.click(
    screen.getByRole('button', { name: texte.kopieAnlegen }),
  )
  await userEvent.keyboard('{Escape}')
  await waitFor(() => {
    expect(DeckKopieren).toHaveBeenCalled()
  })

  return {
    freigeben: () => {
      freigeben()
    },
    Deck,
  }
}
