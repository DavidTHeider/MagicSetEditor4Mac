// **Ein Deck, mit dem gespielt werden soll, sitzt schon** — UX-03.AK-01
// (#1675).
//
// # Warum der erste Sitz
//
// **Weil das Spiel aus dem ersten Deck kommt** (`Partieanfang`:
// `wuensche[0]?.deckID` entscheidet, welches Spiel gilt). Wer „mit diesem
// Deck spielen" wählt, meint sich selbst — und der erste Sitz ist der, den
// die Fläche als eigenen führt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Spielen } from './Spielen'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { mitDecks } from './spielenprobe'
import { raeumeApi } from '../test/api'

import type { Seitenschluessel } from '../navigation'

const eigene = kataloge.de.seiten.spielen

const derHinweis = einsetzen(eigene.anschlussWartet, { name: 'Kontrolle' })

function Buehne() {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('decks')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Abgeber />
        <Spielen wechseln={() => undefined} />
      </Anschlussraum>
    </Sprachraum>
  )
}

function Abgeber() {
  const anschluss = useAnschluss()

  return (
    <button
      type="button"
      onClick={() => {
        anschluss.uebergib({
          art: 'partie',
          name: 'Kontrolle',
          deckID: '2',
        })
      }}
    >
      abgeben
    </button>
  )
}

afterEach(() => {
  raeumeApi()
})

describe('der Partieanfang nimmt auf, welches Deck mitgereist ist', () => {
  it('setzt es auf den ersten Sitz und sagt, woher es kommt', async () => {
    mitDecks()

    render(<Buehne />)

    // **Erst muss die Deckliste da sein** — sonst steht die Auswahl mit
    // ihrem Platzhalter da, und ein Wert, den sie nicht kennt, fällt auf
    // leer zurück (§3).
    const plaetze = await screen.findAllByLabelText(eigene.sitzDeck)

    // **Vorbelegt mit dem ersten eigenen Deck** (P6, #2374) — das
    // mitgereiste ersetzt es.
    expect(plaetze[0]).toHaveValue('1')

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))

    expect(screen.getAllByLabelText(eigene.sitzDeck)[0]).toHaveValue('2')
    expect(screen.getByText(derHinweis)).toBeInTheDocument()
  })

  // **Der zweite Sitz bekommt nur ein eigenes Deck** — P6 (#2374) dreht
  // hier UX-03.AK-01 um: bis dahin blieb er frei („eine Entscheidung, die
  // niemand getroffen hat“), und der Anfang kostete neun Handlungen. Jetzt
  // steht ein Vorschlag da, sichtbar und änderbar, **und nie ein fremdes
  // Deck** (Artboard `PartieStart09`).
  it('belegt den zweiten Sitz mit einem eigenen Deck vor', async () => {
    mitDecks()

    render(<Buehne />)

    await screen.findAllByLabelText(eigene.sitzDeck)
    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))

    expect(screen.getAllByLabelText(eigene.sitzDeck)[1]).toHaveValue('1')
  })

  it('zeigt ohne mitgereistes Deck keinen Hinweis', async () => {
    mitDecks()

    render(<Buehne />)

    await screen.findAllByLabelText(eigene.sitzDeck)

    expect(screen.queryByText(derHinweis)).not.toBeInTheDocument()
  })
})
