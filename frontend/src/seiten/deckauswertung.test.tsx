// **Was ein Deck über sich sagt: Namen, Kurve, Farben** — Werkstück I von
// Phase 4.10
// (#1325),
// samt #1376.
//
// # Der Befund
//
// Die Platzliste zeigte die **Kennung** der Karte — eine Zahl oder eine UUID —,
// und sonst nichts. Ein Deck sagte damit nicht, was in ihm steckt: kein Name,
// keine Kurve, keine Farbverteilung.
//
// # Woher die Angaben kommen
//
// **Aus derselben Antwort.** Der Dienst holt Name und Manakosten beim Lesen —
// für eigene Karten aus dem Bestand, für Drucke von Scryfall in **einem** Ruf.
// Ein Fenster, das sie selbst holte, bräuchte einen Ruf je Karte und ließe MCP
// und die Kommandozeile mit UUIDs zurück (§16).
import { render, screen, waitFor, within } from '@testing-library/react'
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
  zumDeckgebiet,
} from './deckprobe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

function setzeAuf(plaetze = volleplaetze(), teile: Partial<Api> = {}) {
  const deck = deckprobe({ hauptdeck: 9, plaetze })

  belegeApi({
    Decks: () => Promise.resolve({ decks: [deck], fehler: null }),
    Deck: () => Promise.resolve({ deck, fehler: null }),
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

// volleplaetze sind drei Plätze mit Namen und Kosten.
function volleplaetze() {
  return [
    platzprobe({
      id: '1',
      karteID: '11',
      printing: '',
      menge: 4,
      bezeichnung: 'Lightning Bolt',
      manakosten: 'R',
    }),
    platzprobe({
      id: '2',
      karteID: '12',
      printing: '',
      menge: 3,
      bezeichnung: 'Counterspell',
      manakosten: 'UU',
    }),
    platzprobe({
      id: '3',
      karteID: '13',
      printing: '',
      menge: 2,
      bezeichnung: 'Serra Angel',
      manakosten: '3WW',
    }),
  ]
}

async function aufschlagen() {
  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )
}

describe('die Auswertung eines Decks', () => {
  it('nennt die Karten bei ihrem Namen statt bei ihrer Kennung', async () => {
    setzeAuf()
    await aufschlagen()

    const plaetze = await screen.findByRole('list', { name: texte.plaetze })

    expect(within(plaetze).getByText('Lightning Bolt')).toBeInTheDocument()
    expect(within(plaetze).getByText('Counterspell')).toBeInTheDocument()
    expect(within(plaetze).queryByText('11')).not.toBeInTheDocument()
  })

  // **Ohne Namen steht die Kennung** — sie ist die einzige Angabe, die dann
  // stimmt. Das kommt vor: der Name eines Drucks kommt von Scryfall, und ohne
  // Netz bleibt er aus (§16, Regel 5).
  it('zeigt die Kennung, wo kein Name zu haben war', async () => {
    setzeAuf([
      platzprobe({
        id: '1',
        karteID: '',
        printing: 'c58096e5',
        menge: 4,
        bezeichnung: '',
        manakosten: '',
      }),
    ])

    await aufschlagen()

    const plaetze = await screen.findByRole('list', { name: texte.plaetze })

    expect(within(plaetze).getByText('c58096e5')).toBeInTheDocument()
  })

  // **Die Kurve zählt Karten, nicht Plätze.** Vier Blitze bei Kosten 1 sind
  // vier, nicht einer — das ist der ganze Sinn einer Manakurve.
  it('zählt die Manakurve über die Mengen', async () => {
    setzeAuf()
    await aufschlagen()
    await zumDeckgebiet('auswertung')

    const kurve = await screen.findByRole('list', { name: texte.kurve })

    // R = 1, UU = 2, 3WW = 5.
    expect(
      within(kurve).getByRole('listitem', { name: '1' }),
    ).toHaveTextContent('4')
    expect(
      within(kurve).getByRole('listitem', { name: '2' }),
    ).toHaveTextContent('3')
    expect(
      within(kurve).getByRole('listitem', { name: '5' }),
    ).toHaveTextContent('2')
  })

  it('zählt die Farben über die Mengen', async () => {
    setzeAuf()
    await aufschlagen()
    await zumDeckgebiet('auswertung')

    const farben = await screen.findByRole('list', { name: texte.farben })

    expect(
      within(farben).getByRole('listitem', { name: texte.farbeRot }),
    ).toHaveTextContent('4')
    expect(
      within(farben).getByRole('listitem', { name: texte.farbeBlau }),
    ).toHaveTextContent('3')
    expect(
      within(farben).getByRole('listitem', { name: texte.farbeWeiss }),
    ).toHaveTextContent('2')
  })

  // **Ein Spiel ohne Manakosten bekommt keine Kurve.** Gefunden am 2026-09-03
  // in der Sichtprüfung: `Vanguard.mse-set` führt ein `casting cost`-Feld und
  // lässt es bei jeder Karte leer, und die Kurve zeigte einen einzigen Balken
  // bei null über zwölf Karten. Rechnerisch richtig, als Aussage wertlos.
  it('lässt die Kurve weg, wo keine Karte Manakosten kennt', async () => {
    setzeAuf([
      platzprobe({
        id: '1',
        karteID: '11',
        printing: '',
        menge: 4,
        bezeichnung: 'Akroma Avatar',
        manakosten: '',
      }),
    ])

    await aufschlagen()

    await waitFor(() => {
      expect(screen.getByText('Akroma Avatar')).toBeInTheDocument()
    })

    await zumDeckgebiet('auswertung')

    expect(screen.queryByRole('list', { name: texte.kurve })).toBeNull()
  })

  // **Länder stehen nicht in der Kurve** — Werkstück U10, angestrichen von
  // der Sichtprüfung (2026-09-09): bis dahin fielen sie **zufällig** heraus,
  // weil ihre Kosten leer sind. Seit der Manawert nachgelesen wird, muss die
  // Absicht dastehen — sonst wäre die Spalte null bei zwanzig Wäldern der
  // höchste Balken und sagte nichts.
  it('lässt Länder aus der Kurve, zeigt aber den Rest', async () => {
    setzeAuf([
      platzprobe({
        id: '1',
        printing: 'aaaa1111-0000-4000-8000-00000000000a',
        menge: 12,
        bezeichnung: 'Forest',
        typzeile: 'Basic Land — Forest',
        manakosten: '',
      }),
      platzprobe({
        id: '2',
        printing: 'bbbb2222-0000-4000-8000-00000000000b',
        menge: 4,
        bezeichnung: 'Lightning Bolt',
        typzeile: 'Instant',
        manakosten: '{R}',
      }),
    ])

    await aufschlagen()
    await zumDeckgebiet('auswertung')

    const kurve = await screen.findByRole('list', { name: texte.kurve })
    const zeilen = within(kurve).getAllByRole('listitem')

    // **Die Zwölf steht nirgends.** Der erste Anlauf prüfte die Zahl der
    // Zeilen und die Werte 0 und 4 — und blieb grün, als der Wächter
    // entschärft wurde: die Wälder standen dann in der Spalte null, und die
    // Zeile trug immer noch eine Null (§10, die Gegenprobe).
    expect(zeilen).toHaveLength(2)
    expect(within(kurve).queryByText('12')).toBeNull()

    // **Zweimal die Null** — die Spalte und ihre Zahl. Stünden die Wälder
    // darin, wäre es einmal die Null und einmal die Zwölf.
    expect(within(kurve).getAllByText('0')).toHaveLength(2)
    expect(within(kurve).getByText('4')).toBeInTheDocument()
  })
})
