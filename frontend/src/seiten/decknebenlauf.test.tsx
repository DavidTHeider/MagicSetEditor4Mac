// **Die vier Regeln aus §7 auf der Deckseite** — Befund 5 des Zweitblicks
// über Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, weil `decks.test.tsx` über die Zeilengrenze aus §2
// ging. Der Schnitt folgt der Frage: dort steht, **was die Seite kann**, hier,
// **was sie unter Nebenläufigkeit tut**.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  aufschlagen,
  bauartprobe,
  deckprobe,
  formatprobe,
  platzprobe,
} from './deckprobe'

import type { DeckAntwort } from '../apidecks'

import type { Api } from '../api'
import type { DeckDTO } from '../apidecks'
import { deckhandlung } from '../test/deckhandlung'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// zeige zeichnet die Seite über den gegebenen Belegen.
function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
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
}

describe('die Deckseite unter Nebenläufigkeit', () => {
  it('verwirft die Antwort eines Decks, das niemand mehr aufgeschlagen hat', async () => {
    // **Befund 5 des Zweitblicks über Phase 4.6** (§7, Regel 2). Die
    // langsame Antwort für Deck 1 kommt **nach** der für Deck 2 — ohne
    // Standzähler setzt sie Deck 1 wieder ein, und das Fenster zeigt einen
    // Stand, den niemand aufgeschlagen hat.
    let loeseErstes: (() => void) | undefined

    const Deck = vi.fn((deckID: string) => {
      if (deckID === '1') {
        return new Promise<{ deck: DeckDTO | null; fehler: null }>(
          (aufloesen) => {
            loeseErstes = () => {
              aufloesen({
                deck: deckprobe({ id: '1', name: 'Erstes' }),
                fehler: null,
              })
            }
          },
        )
      }

      return Promise.resolve({
        deck: deckprobe({ id: '2', name: 'Zweites' }),
        fehler: null,
      })
    })

    zeige({
      Deck,
      Decks: () =>
        Promise.resolve({
          decks: [
            deckprobe({ id: '1', name: 'Erstes' }),
            deckprobe({ id: '2', name: 'Zweites' }),
          ],
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Erstes` }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Zweites` }),
    )

    // **Das zweite steht offen** — als Überschrift der Werkbank.
    expect(
      await screen.findByRole('heading', { name: 'Zweites' }),
    ).toBeInTheDocument()

    // **Und jetzt kommt die überholte Antwort.**
    loeseErstes?.()

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledTimes(2)
    })

    expect(
      screen.queryByRole('heading', { name: 'Erstes' }),
    ).not.toBeInTheDocument()
  })

  it('sperrt zwei Schreibarten desselben Decks gegeneinander', async () => {
    // **Befund 5, zweite Hälfte** (§7, Regel 4). Vorher hießen die Marken
    // `platz-legen:7` und `deck-format:7` — drei verschiedene Schreibarten
    // auf **dasselbe** Deck sperrten einander nicht.
    let loesePlatz: (() => void) | undefined

    const DeckplatzLegen = vi.fn(
      () =>
        new Promise<{ platz: null; fehler: null }>((aufloesen) => {
          loesePlatz = () => {
            aufloesen({ platz: null, fehler: null })
          }
        }),
    )

    zeige({
      DeckplatzLegen,
      Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    await userEvent.type(await screen.findByLabelText(texte.druck), 'abc')
    await userEvent.click(screen.getByRole('button', { name: texte.legen }))

    // **Solange das Legen läuft, ist die Formatwahl gesperrt** (§7, Regel 4).
    // Sie steht seit Werkstück I über den Laschen — als Eigenschaft des Decks
    // und nicht als Gebiet —, und damit ist sie neben der Kartenwahl sichtbar.
    expect(screen.getByLabelText(texte.formatWaehlen)).toBeDisabled()

    loesePlatz?.()

    await waitFor(() => {
      expect(screen.getByLabelText(texte.formatWaehlen)).not.toBeDisabled()
    })
  })

  it('verschweigt auch den Fehler eines überholten Decks', async () => {
    // **Die unauffälligere Hälfte von Regel 2.** Eine Meldung „Deck nicht
    // lesbar" zu einem Deck, das niemand mehr offen hat, sieht aus wie ein
    // Fehler des aktuellen.
    //
    // # Warum das Versprechen **abgelehnt** wird und nicht mit einem Fehler
    // **aufgelöst**
    //
    // Weil sonst der frühe Rücksprung greift: eine aufgelöste Antwort geht
    // durch `istAktuell` und kehrt zurück, bevor `auspacken` wirft — der
    // `catch` liefe gar nicht. Eine **abgelehnte** Bindung ist der Weg, den
    // Wails im Ernstfall nimmt (Zeitablauf, geschlossenes Fenster), und nur
    // sie erreicht den Zweig.
    //
    // Gemessen am 2026-09-01: mit der aufgelösten Fassung stand der
    // else-Zweig bei `[2, 0]` — zweimal ja, nie nein.
    let scheitereErstes: (() => void) | undefined

    const Deck = vi.fn((deckID: string) => {
      if (deckID === '1') {
        return new Promise<DeckAntwort>((_aufloesen, ablehnen) => {
          scheitereErstes = () => {
            ablehnen(new Error('die Brücke antwortete nicht'))
          }
        })
      }

      return Promise.resolve({
        deck: deckprobe({ id: '2', name: 'Zweites' }),
        fehler: null,
      })
    })

    zeige({
      Deck,
      Decks: () =>
        Promise.resolve({
          decks: [
            deckprobe({ id: '1', name: 'Erstes' }),
            deckprobe({ id: '2', name: 'Zweites' }),
          ],
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Erstes` }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Zweites` }),
    )

    await screen.findByRole('heading', { name: 'Zweites' })

    scheitereErstes?.()

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledTimes(2)
    })

    expect(
      screen.queryByText(kataloge.de.fehler.nichtGefunden),
    ).not.toBeInTheDocument()
  })

  it('meldet den Fehler des aufgeschlagenen Decks sehr wohl', async () => {
    // **Die andere Hälfte derselben Verzweigung.** Ein Fehler zum
    // **aktuellen** Deck gehört gezeigt — sonst wäre die Regel eine
    // Ausrede, jeden Fehler zu verschlucken.
    zeige({
      Deck: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'nichtGefunden',
            meldung: 'weg',
            feld: '',
            werte: {},
          },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    expect(
      await screen.findByText(kataloge.de.fehler.nichtGefunden),
    ).toBeInTheDocument()
  })
})

// **Ein verspäteter Schreibfehler von Deck A erreicht Deck B nicht.**
//
// # Warum das hier steht
//
// **Angestrichen am Turnierweg** — Befund 9 des Zweitblicks über Phase 4.7 —
// und hier nachgezogen: es ist dieselbe Gestalt, und behoben wird die Klasse
// (§10). `useVorgang` sperrt nur denselben Schlüssel; ein Deckwechsel
// dazwischen ist erlaubt, und ohne die Laufnummer stünde die Meldung über
// dem falschen Deck (§7, Regel 2).
describe('ein verspäteter Schreibfehler am Deckweg', () => {
  it('wird nicht dem falschen Deck zugeschrieben', async () => {
    // **Mit einem Platz**, sonst zeigt die Werkbank ihren leeren Zustand
    // statt der Plätzeliste — und der Beleg, dass B aufgeschlagen ist, wäre
    // keiner.
    const erste = deckprobe({ id: '7', name: 'Aggro' })
    const zweite = deckprobe({
      id: '8',
      name: 'Control',
      plaetze: [platzprobe()],
    })

    let lehneAb: (() => void) | undefined
    let antwort: Promise<{ fehler: null }> | undefined

    const DeckLoeschen = vi.fn((): Promise<{ fehler: null }> => {
      antwort = new Promise((_, ablehnen) => {
        lehneAb = () => {
          ablehnen(new Error('Dieses Deck ließ sich nicht löschen.'))
        }
      })

      return antwort
    })

    belegeApi({
      Decks: () => Promise.resolve({ decks: [erste, zweite], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      Deck: (id: string) =>
        Promise.resolve({ deck: id === '8' ? zweite : erste, fehler: null }),
      DeckLoeschen,
    })

    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    // **Das Löschen von A anstoßen — die Antwort hängt.** Löschen steht
    // seit UXR-11 (#1919) am offenen Deck.
    await aufschlagen()
    await deckhandlung('Aggro', kataloge.de.seiten.decks.loeschen)
    await userEvent.click(
      screen.getByRole('button', { name: kataloge.de.seiten.decks.loeschenJa }),
    )

    // **Und währenddessen B aufschlagen.**
    await userEvent.click(
      screen.getByRole('button', {
        name: `${kataloge.de.seiten.decks.waehlen}: Control`,
      }),
    )

    expect(
      await screen.findByRole('list', {
        name: kataloge.de.seiten.decks.plaetze,
      }),
    ).toBeInTheDocument()

    lehneAb?.()

    await act(async () => {
      await antwort?.catch(() => undefined)
    })

    expect(
      screen.queryByText('Dieses Deck ließ sich nicht löschen.'),
    ).toBeNull()
  })
})
