// **Was die Befehlspalette übergibt, nimmt die Zielfläche auf** — UX-02.AK-02
// (#1674).
//
// # Was hier gemessen wird
//
// **Der zweite Teil der Naht.** `palettenziele.test.ts` hält, was die Palette
// übergibt; hier steht, dass jemand es annimmt — sonst führte „Karten
// tauschen mit Turfer" auf eine Seite, auf der niemand Turfer aufschlägt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './seiten/Decks'
import { Handeln } from './seiten/Handeln'
import { useKartenwunsch } from './haken/useKartenwunsch'
import { Sprachraum } from './sprache/kontext'
import { kataloge } from './sprache/kataloge'
import { belegeApi, raeumeApi } from './test/api'
import { Anschlussbuehne, uebergabeknopf } from './test/anschlussbuehne'
import { deckprobe, formatprobe, bauartprobe } from './seiten/deckprobe'
import { eineListe, einKontakt } from './seiten/peerkontaktprobe'
import { zeigeEinstellungen } from './test/einstellungsseite'

import type { Anschluss } from './anschluss'
import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
})

async function uebergeben() {
  await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))
}

describe('Handeln nimmt ein Gegenüber auf', () => {
  it('schlägt „Karten tauschen mit …" beim Gegenüber auf', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([
            einKontakt({
              kennung: 'peer-1',
              eigenerName: 'Turfer',
              stand: 'bestaetigt',
            }),
          ]),
        ),
    })

    render(
      <Anschlussbuehne
        anschluss={{ art: 'gegenueber', name: 'Turfer', kennung: 'peer-1' }}
      >
        <Sprachraum sprache="de">
          <Handeln />
        </Sprachraum>
      </Anschlussbuehne>,
    )

    const eigene = kataloge.de.seiten.handeln
    expect(await screen.findByText(eigene.partnerKeiner)).toBeInTheDocument()

    await uebergeben()

    // **Aufgeschlagen ist, was rechts steht** — seit UXR-15 (#1923) steht die
    // Wahl des Gegenübers in der Schublade hinter „Neuer Tausch …", und der
    // Raum zeigt den neuen Handel.
    expect(await screen.findByText(eigene.assistent.neu)).toBeInTheDocument()
    expect(screen.queryByText(eigene.partnerKeiner)).toBeNull()
  })
})

describe('Decks nehmen ein Deck zum Aufschlagen auf', () => {
  it('schlägt das gefundene Deck auf', async () => {
    const Deck = vi.fn(() =>
      Promise.resolve({ deck: deckprobe({ id: '3' }), fehler: null }),
    )
    belegeApi({
      Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      Deck,
      DeckBenutzt: () => Promise.resolve({ fehler: null }),
    })

    render(
      <Anschlussbuehne
        anschluss={{ art: 'deckansicht', name: 'Aggro', deckID: '3' }}
      >
        <Sprachraum sprache="de">
          <Decks wechseln={vi.fn()} />
        </Sprachraum>
      </Anschlussbuehne>,
    )

    expect(Deck).not.toHaveBeenCalled()

    await uebergeben()

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledWith('3')
    })
  })
})

// **Ein veralteter Verweis führt nicht ins Stille** — UX-03.AK-04
// (#1675):
// wer ein Deck in der Palette findet, das inzwischen gelöscht ist, liest das,
// statt eine leere Fläche zu sehen.
describe('ein Deck, das es nicht mehr gibt', () => {
  it('sagt es auf der Deckseite', async () => {
    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      Deck: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'nichtGefunden',
            meldung: 'nicht gefunden',
            feld: '',
            schluessel: 'nichtGefunden',
          },
        }),
    })

    render(
      <Anschlussbuehne
        anschluss={{ art: 'deckansicht', name: 'Aggro', deckID: '9' }}
      >
        <Sprachraum sprache="de">
          <Decks wechseln={vi.fn()} />
        </Sprachraum>
      </Anschlussbuehne>,
    )

    await uebergeben()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.nichtGefunden,
    )
  })
})

describe('Die Einstellungen nehmen eine Gruppe auf', () => {
  it('öffnet „Daten und Sicherung" für „Sicherung erstellen"', async () => {
    const gruppen = kataloge.de.seiten.einstellungen.gruppen
    const Huelle = ({ children }: { children: ReactNode }) => (
      <Anschlussbuehne
        anschluss={{
          art: 'einstellungsgruppe',
          name: gruppen.daten,
          gruppe: 'daten',
        }}
      >
        {children}
      </Anschlussbuehne>
    )

    zeigeEinstellungen({}, Huelle)

    expect(
      screen.getByRole('heading', { name: gruppen.allgemein }),
    ).toBeInTheDocument()

    await uebergeben()

    expect(
      await screen.findByRole('heading', { name: gruppen.daten }),
    ).toBeInTheDocument()
  })

  // **Mit Abschnitt schlägt sie ihn auch auf und gibt ihm den Fokus** —
  // AP-10 (#2367): „Schnittstellen einrichten“ endete sonst vor einem
  // zugeklappten Kasten, den man erst suchen müsste.
  it('schlägt „Werkzeuge ohne Fenster“ unter Verbindungen auf', async () => {
    const gruppen = kataloge.de.seiten.einstellungen.gruppen
    const titel = kataloge.de.seiten.technik.ueberschrift
    belegeApi({ Technikangaben: () => new Promise(() => undefined) })

    const Huelle = ({ children }: { children: ReactNode }) => (
      <Anschlussbuehne
        anschluss={{
          art: 'einstellungsgruppe',
          name: gruppen.verbindungen,
          gruppe: 'verbindungen',
          abschnitt: 'werkzeuge',
        }}
      >
        {children}
      </Anschlussbuehne>
    )

    zeigeEinstellungen({}, Huelle)

    await uebergeben()

    expect(
      await screen.findByRole('heading', { name: gruppen.verbindungen }),
    ).toBeInTheDocument()
    const kasten = await screen.findByRole('button', { name: titel })
    await waitFor(() => {
      expect(kasten).toHaveAttribute('aria-expanded', 'true')
    })
    await waitFor(() => {
      expect(kasten).toHaveFocus()
    })
  })
})

describe('Die Kartenseite nimmt „Karte erstellen" auf', () => {
  function Wunsch({
    setID,
    anlegen,
  }: {
    setID: string
    anlegen: () => Promise<void>
  }) {
    useKartenwunsch(setID, anlegen)

    return null
  }

  const karte: Anschluss = { art: 'karte', name: 'Werkstatt', setID: '7' }

  it('legt genau eine Karte an, im Set, für das der Wunsch galt', async () => {
    const anlegen = vi.fn(() => Promise.resolve())

    const { rerender } = render(
      <Anschlussbuehne anschluss={karte}>
        <Wunsch setID="7" anlegen={anlegen} />
      </Anschlussbuehne>,
    )

    expect(anlegen).not.toHaveBeenCalled()

    await uebergeben()

    expect(anlegen).toHaveBeenCalledTimes(1)

    // **Derselbe Wunsch bleibt im Raum stehen** — ein weiteres Zeichnen
    // ergibt keine zweite Karte.
    rerender(
      <Anschlussbuehne anschluss={karte}>
        <Wunsch setID="7" anlegen={anlegen} />
      </Anschlussbuehne>,
    )
    expect(anlegen).toHaveBeenCalledTimes(1)
  })

  // **Und keine zweite bei der Rückkehr** — die Kartenseite entsteht nach
  // einem Setwechsel unter `key={setID}` neu, und ihre Referenz ist dann
  // leer. Angestrichen vom Zweitblick über Phase 6.3 (S7): A → B → A legte
  // bei jeder Rückkehr eine weitere Karte an.
  it('legt bei der Rückkehr ins Set keine zweite Karte an', async () => {
    const anlegen = vi.fn(() => Promise.resolve())
    const zeichne = (setID: string, stand: string) => (
      <Anschlussbuehne anschluss={karte}>
        <Wunsch key={stand} setID={setID} anlegen={anlegen} />
      </Anschlussbuehne>
    )

    const { rerender } = render(zeichne('7', 'a'))
    await uebergeben()
    expect(anlegen).toHaveBeenCalledTimes(1)

    rerender(zeichne('8', 'b'))
    rerender(zeichne('7', 'c'))

    expect(anlegen).toHaveBeenCalledTimes(1)
  })

  // **Ist inzwischen ein anderes Set offen, tut die Seite nichts** (§7,
  // Regel 2) — eine Karte im falschen Set wäre schlimmer als keine.
  it('legt in einem anderen Set nichts an', async () => {
    const anlegen = vi.fn(() => Promise.resolve())

    render(
      <Anschlussbuehne anschluss={karte}>
        <Wunsch setID="8" anlegen={anlegen} />
      </Anschlussbuehne>,
    )

    await uebergeben()

    expect(anlegen).not.toHaveBeenCalled()
  })
})
