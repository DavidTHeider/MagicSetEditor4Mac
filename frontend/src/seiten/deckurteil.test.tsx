// **Prüfung und Abgleich am Fenster** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `decks.test.tsx` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, wie ein Deck entsteht und
// was hineinkommt, hier, was man darüber erfährt.
//
// # Die eine Zusage, um die es geht
//
// **`ungeprüft` ist kein `nicht zulässig`** (Vermerk 0061) — und das Fenster
// muss es **sagen**, nicht bloß nicht rot färben.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  bauartprobe,
  deckprobe,
  formatprobe,
  platzprobe,
  urteilprobe,
  zumDeckgebiet,
} from './deckprobe'

import type { Api } from '../api'
import { kurzform } from '../sprache/kurzform'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// zeigeMitDeck schlägt ein Deck auf und gibt die Seite.
async function zeigeMitDeck(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Deck: () =>
      Promise.resolve({
        deck: deckprobe({
          plaetze: [platzprobe()],
          hauptdeck: 4,
          formatID: '5',
        }),
        fehler: null,
      }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )

  // **Urteil und Abgleich liegen im Gebiet »Prüfung«** — Werkstück I von
  // Phase 4.10
  // (#1325).
  await zumDeckgebiet('pruefung')
}

describe('die Prüfung eines Decks', () => {
  it('nennt das Urteil, die Befunde und die Wörter', async () => {
    await zeigeMitDeck({
      DeckPruefen: () =>
        Promise.resolve({
          urteil: urteilprobe({
            zulaessig: false,
            groessen: [{ feld: 'hauptdeck', meldung: 'zu wenige Karten' }],
            urteile: [
              {
                platzID: '3',
                bezeichnung: 'Blitz',
                menge: 4,
                bereich: 'hauptdeck',
                zulaessigkeit: 'banned',
                grund: '',
                quelle: 'Scryfall',
              },
            ],
          }),
          befunde: ['Scryfall schwieg'],
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.pruefen }),
    )

    expect(await screen.findByText(texte.unzulaessig)).toBeInTheDocument()
    expect(screen.getByText('zu wenige Karten')).toBeInTheDocument()
    expect(screen.getByText(texte.banned)).toBeInTheDocument()
    expect(screen.getByText('Scryfall schwieg')).toBeInTheDocument()
  })

  it('sagt, was ungeprüft heißt — und färbt es nicht rot', async () => {
    await zeigeMitDeck({
      DeckPruefen: () =>
        Promise.resolve({
          urteil: urteilprobe({
            ungeprueft: 3,
            urteile: [
              {
                platzID: '3',
                bezeichnung: '',
                menge: 3,
                bereich: 'hauptdeck',
                zulaessigkeit: 'ungeprüft',
                grund: 'eigener Entwurf',
                quelle: '',
              },
            ],
          }),
          befunde: [],
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.pruefen }),
    )

    // **Zulässig, obwohl drei Karten ungeprüft sind.**
    expect(await screen.findByText(texte.zulaessig)).toBeInTheDocument()

    expect(screen.getByText(/3 Karten sind ungeprüft/)).toBeInTheDocument()
    expect(
      screen.getByText(kurzform(texte.ungepruefterklaerung)),
    ).toBeInTheDocument()
  })

  it('lässt ein Wort stehen, das der Katalog nicht kennt', async () => {
    await zeigeMitDeck({
      DeckPruefen: () =>
        Promise.resolve({
          urteil: urteilprobe({
            urteile: [
              {
                platzID: '3',
                bezeichnung: 'Fremd',
                menge: 1,
                bereich: 'hauptdeck',
                zulaessigkeit: 'vielleicht',
                grund: '',
                quelle: 'Scryfall',
              },
            ],
          }),
          befunde: [],
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.pruefen }),
    )

    // **Unübersetzt statt leer** (§9): Scryfall darf ein sechstes Wort
    // schicken, und dann steht es da.
    expect(await screen.findByText('vielleicht')).toBeInTheDocument()
  })

  it('meldet einen Fehler der Prüfung', async () => {
    await zeigeMitDeck({
      DeckPruefen: () =>
        Promise.resolve({
          urteil: null,
          befunde: [],
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'kein Format',
            feld: 'deckID',
            werte: {},
          },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.pruefen }),
    )

    expect(
      await screen.findByText(kataloge.de.fehler.validierung),
    ).toBeInTheDocument()
  })
})

describe('der Abgleich gegen die Sammlung', () => {
  // **Ohne Klick** — AP-9 (#2366): die Leiste über dem Deck nennt, was in
  // der Sammlung fehlt, und führt hierher; hier muss es dann schon stehen.
  // Der Abgleich liest nur und geht nicht ins Netz.
  it('zeigt ohne Klick, was fehlt', async () => {
    await zeigeMitDeck({
      DeckAbgleichen: () =>
        Promise.resolve({
          zeilen: [
            {
              printing: 'abc',
              druckID: '',
              vorlaeufig: false,
              bezeichnung: 'Blitz',
              gebraucht: 4,
              vorhanden: 1,
              fehlend: 3,
              grund: '',
            },
          ],
          fehlend: 3,
          unbekannt: 0,
          fehler: null,
        }),
    })

    expect(await screen.findByText(/3 Karten fehlen/)).toBeInTheDocument()
    expect(screen.getByText('1/4')).toBeInTheDocument()
  })

  it('sagt, was fehlt — und was unbekannt ist', async () => {
    await zeigeMitDeck({
      DeckAbgleichen: () =>
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
            {
              printing: '',
              druckID: '',
              vorlaeufig: false,
              bezeichnung: '',
              gebraucht: 3,
              vorhanden: 0,
              fehlend: 0,
              grund: 'eigener Entwurf',
            },
          ],
          fehlend: 2,
          unbekannt: 3,
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.bestand }),
    )

    expect(await screen.findByText(/2 Karten fehlen/)).toBeInTheDocument()
    expect(screen.getByText('2/4')).toBeInTheDocument()
    expect(screen.getByText('eigener Entwurf')).toBeInTheDocument()
  })

  it('sagt es, wenn alles da ist', async () => {
    await zeigeMitDeck({
      DeckAbgleichen: () =>
        Promise.resolve({ zeilen: [], fehlend: 0, unbekannt: 0, fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.bestand }),
    )

    expect(
      await screen.findByText(texte.bestandVollstaendig),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler des Abgleichs', async () => {
    await zeigeMitDeck({
      DeckAbgleichen: () =>
        Promise.resolve({
          zeilen: [],
          fehlend: 0,
          unbekannt: 0,
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'kaputt',
            feld: '',
            werte: {},
          },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.bestand }),
    )

    expect(
      await screen.findByText(kataloge.de.fehler.unbekannt),
    ).toBeInTheDocument()
  })

  it('nennt Grund und Quelle je Karte', async () => {
    // **Befund 7 des Zweitblicks über Phase 4.6.** Die Abnahme des Issues
    // verlangt jede unzulässige Karte „mit dem Grund und der Quelle"; die
    // Naht trug beides, das Fenster zeigte keines.
    await zeigeMitDeck({
      DeckPruefen: () =>
        Promise.resolve({
          urteil: urteilprobe({
            zulaessig: false,
            urteile: [
              {
                platzID: '3',
                bezeichnung: 'Blitz',
                menge: 4,
                bereich: 'hauptdeck',
                zulaessigkeit: 'banned',
                grund: '',
                quelle: 'Scryfall',
              },
              {
                platzID: '4',
                bezeichnung: '',
                menge: 1,
                bereich: 'hauptdeck',
                zulaessigkeit: 'ungeprüft',
                grund: 'Diese Karte ist ein eigener Entwurf.',
                quelle: '',
              },
            ],
          }),
          befunde: [],
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.pruefen }),
    )

    expect(await screen.findByText('Scryfall')).toBeInTheDocument()
    expect(
      screen.getByText('Diese Karte ist ein eigener Entwurf.'),
    ).toBeInTheDocument()
  })
})
