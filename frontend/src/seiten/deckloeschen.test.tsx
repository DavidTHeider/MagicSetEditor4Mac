// **Was das Löschen eines aufgeschlagenen Decks abräumt** — Befund 4 des
// Nachlaufs über Phase 4.6
// (#1293).
//
// # Der Befund
//
// §7, Regel 1 sagt: „Ein Wechsel des Bezugsobjekts räumt den offenen
// Kontext." Ein **Wechsel** war abgedeckt — das Löschen nicht. `loeschen`
// räumte die Rückfrage und lud die Übersicht nach; `offen` blieb stehen. Das
// Fenster zeigte und bediente danach einen Datensatz, den der Bestand nicht
// mehr kennt.
//
// **Und die zweite Hälfte**: eine vor dem Löschen begonnene Leseantwort trug
// weiterhin die aktuelle Laufnummer und konnte das gelöschte Deck **erneut**
// aufschlagen.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { deckprobe, platzprobe } from './deckprobe'
import { deckhandlung, platzeintrag } from '../test/deckhandlung'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

describe('ein gelöschtes Deck', () => {
  it('bleibt nicht aufgeschlagen', async () => {
    // **Mit einem Platz**, denn ohne einen zeigt die Werkbank ihren leeren
    // Zustand statt der Plätzeliste — und dann prüfte der Test nichts.
    const deck = deckprobe({
      id: '7',
      name: 'Aggro',
      plaetze: [platzprobe()],
    })

    let gelöscht = false

    belegeApi({
      Decks: () =>
        Promise.resolve({ decks: gelöscht ? [] : [deck], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      Deck: () => Promise.resolve({ deck, fehler: null }),
      DeckLoeschen: () => {
        gelöscht = true

        return Promise.resolve({ deck: null, fehler: null })
      },
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    // **Erst aufschlagen** — sonst prüfte der Test das Löschen an einem
    // Fenster, in dem ohnehin nichts stand. Die Plätzeliste trägt als
    // einziges Stück der Werkbank einen eigenen Namen.
    expect(
      await screen.findByRole('list', { name: texte.plaetze }),
    ).toBeTruthy()

    await deckhandlung('Aggro', texte.loeschen)
    await userEvent.click(
      await screen.findByRole('button', { name: texte.loeschenJa }),
    )

    await waitFor(() => {
      expect(screen.queryByRole('list', { name: texte.plaetze })).toBeNull()
    })
  })

  it('wird von einer langsamen Antwort nicht wieder aufgeschlagen', async () => {
    // **Mit einem Platz** — siehe oben.
    const deck = deckprobe({
      id: '7',
      name: 'Aggro',
      plaetze: [platzprobe({ bezeichnung: 'Blitz' })],
    })

    let loeseAuf: (() => void) | undefined
    let antwort: Promise<{ deck: typeof deck; fehler: null }> | undefined

    // **Der erste Ruf antwortet sofort, der zweite hängt.**
    //
    // Ohne den ersten wäre die Zusicherung am Ende schon **vor** der Antwort
    // wahr — die Werkbank stand ja nie —, und der Test wäre grün, ohne die
    // Wirkung geprüft zu haben (§3).
    const Deck = vi.fn((): Promise<{ deck: typeof deck; fehler: null }> => {
      if (Deck.mock.calls.length === 1) {
        return Promise.resolve({ deck, fehler: null })
      }

      antwort = new Promise((fertig) => {
        loeseAuf = () => {
          fertig({ deck, fehler: null })
        }
      })

      return antwort
    })

    belegeApi({
      Decks: () => Promise.resolve({ decks: [deck], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      Deck,
      DeckLoeschen: () => Promise.resolve({ deck: null, fehler: null }),
      DeckplatzLegen: () =>
        Promise.resolve({ platz: platzprobe({ menge: 5 }), fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    const oeffnen = await screen.findByRole('button', {
      name: `${texte.waehlen}: Aggro`,
    })

    await userEvent.click(oeffnen)

    // **Sie steht** — das ist der Beleg, dass die Verneinung am Ende etwas
    // sagt.
    expect(
      await screen.findByRole('list', { name: texte.plaetze }),
    ).toBeTruthy()

    // **Jetzt ein zweites Lesen, das hängt** — und mittendrin gelöscht.
    // Nach jedem Legen liest der Haken dasselbe Deck neu; ein zweites
    // „Öffnen“ gibt es am offenen Deck seit AP-6 (#2363) nicht mehr.
    await userEvent.click(await platzeintrag('Blitz', texte.mehrVon))

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledTimes(2)
    })

    await deckhandlung('Aggro', texte.loeschen)
    await userEvent.click(
      await screen.findByRole('button', { name: texte.loeschenJa }),
    )

    // **Und jetzt kommt die Antwort von vorhin.**
    //
    // **Gewartet wird auf sie und nicht auf die Zusicherung.** `waehle` hat
    // den offenen Kontext beim zweiten Klick ohnehin geleert; ein `waitFor`
    // auf „die Liste ist weg" träfe deshalb **sofort** zu — schon bevor die
    // Antwort da ist —, und der Test wäre grün, ohne die Wirkung geprüft zu
    // haben (§3). Erst nach dem Abschluss darf zugesichert werden.
    loeseAuf?.()

    await act(async () => {
      await antwort
    })

    expect(screen.queryByRole('list', { name: texte.plaetze })).toBeNull()
  })
})
