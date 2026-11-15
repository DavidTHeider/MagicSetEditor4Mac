// **Eine Partie anfangen, nach dem Artboard** — PLAY-01, PLAY-02 und
// S11.NAT-01 bis 04 (UX-08,
// #1680),
// Befund F13.
//
// # Wogegen hier geprüft wird
//
// **Gegen die Regelschau zwischen Sitzen und Startknopf.** Mit zwei
// gewählten Decks war die Seite bis zum 2026-09-24 **4.890 Punkte** hoch —
// gemessen in WebKit —, und der Knopf „Partie anfangen" stand darunter. Das
// Artboard *Eine Partie anfangen* setzt Sitze und Regeln in zwei Tafeln, die
// langen Aufzählungen eingeklappt darunter, und den Anfang als Hauptknopf.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { element } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { mitDecks, zeige } from './spielenprobe'
import { raeumeApi } from '../test/api'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// deckWaehlen stellt den ersten Sitz auf das erste Deck — erst dann gibt es
// ein Spiel und eine Auskunft über die Regeln.
async function deckWaehlen(hand: ReturnType<typeof zeige>) {
  const [erstes] = await screen.findAllByLabelText(eigene.sitzDeck)

  await hand.selectOptions(element(erstes), '1')
}

it('stellt die Sitze in eine Tafel mit ihrer Zahl', async () => {
  mitDecks()

  const hand = zeige()

  const sitze = await screen.findByRole('region', { name: eigene.sitzeTitel })

  expect(
    within(sitze).getByText(einsetzen(eigene.sitzeZahl, { zahl: '2' })),
  ).toBeInTheDocument()

  await hand.click(
    within(sitze).getByRole('button', { name: eigene.sitzHinzu }),
  )

  expect(
    within(sitze).getByText(einsetzen(eigene.sitzeZahl, { zahl: '3' })),
  ).toBeInTheDocument()
})

// **Die Größe des Decks steht beim Sitz** — „60 + 15" im Artboard, für ein
// Vorleseprogramm in Worten.
it('nennt je Sitz die Größe des gewählten Decks', async () => {
  mitDecks()

  const hand = zeige()

  await deckWaehlen(hand)

  // **An beiden Sitzen** — der zweite ist vorbelegt (P6, #2374).
  expect(
    screen.getAllByText(
      einsetzen(eigene.deckGroesse, { haupt: '60', seite: '0' }),
    ),
  ).toHaveLength(2)
})

it('stellt die Regeln in eine eigene Tafel und ihre Einzelheiten dahinter', async () => {
  mitDecks()

  const hand = zeige()

  await deckWaehlen(hand)

  const regeln = await screen.findByRole('region', { name: eigene.regelnTafel })

  // **Die langen Aufzählungen sind zu** — und aufzuklappen, samt Zustand.
  expect(
    screen.queryByRole('heading', { name: eigene.regelnTitel }),
  ).not.toBeInTheDocument()

  const kasten = within(regeln).getByRole('button', {
    name: eigene.regelgrenzenTitel,
  })

  expect(kasten).toHaveAttribute('aria-expanded', 'false')

  await hand.click(kasten)

  expect(
    screen.getByRole('heading', { name: eigene.regelnTitel }),
  ).toBeInTheDocument()
})

// **Die Marke der Regeltafel folgt der Wahl** — sie sagt auf einen Blick,
// wie viel das Programm übernimmt.
it('nennt den gewählten Modus neben der Regeltafel', async () => {
  mitDecks()

  const hand = zeige()

  await deckWaehlen(hand)

  // kopf ist die Zeile mit der Überschrift der Regeltafel — **jedes Mal neu
  // gesucht**: ein Wechsel des Modus fragt die Freigabe neu, und die Tafel
  // wird neu gezeichnet.
  const kopf = async () =>
    element(
      (await screen.findByRole('heading', { name: eigene.regelnTafel }))
        .parentElement,
    )

  expect(
    await within(await kopf()).findByText(eigene.modusAnfaenger),
  ).toBeInTheDocument()

  await hand.click(screen.getByLabelText(eigene.modusFrei))

  expect(
    await within(await kopf()).findByText(eigene.modusFrei),
  ).toBeInTheDocument()
})

// **Der Anfang ist die Haupthandlung dieser Seite** und steht nach allem,
// was man vorher wissen muss — auch nach den Regeln.
it('stellt den Anfang als Hauptknopf hinter die Regeln', async () => {
  mitDecks()

  const hand = zeige()

  await deckWaehlen(hand)

  const regeln = await screen.findByRole('region', { name: eigene.regelnTafel })
  const knopf = screen.getByRole('button', { name: eigene.anfangen })

  expect(knopf).toHaveClass('bg-marke')
  expect(
    regeln.compareDocumentPosition(knopf) & Node.DOCUMENT_POSITION_FOLLOWING,
  ).toBeTruthy()
})
