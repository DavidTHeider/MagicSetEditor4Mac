// **Die Vorschau folgt dem Tippen** (#288).
//
// AGENTS.md §6: „Eine Änderung an einem Feld ist in der Vorschau zu sehen, ohne
// dass jemand etwas bestätigt." Bis hierher schrieb der Editor beim
// **Verlassen** des Feldes — und Verlassen ist eine Bestätigung.
//
// **Geschrieben wird trotzdem erst beim Verlassen.** Ein Schreiben legt einen
// Schritt im Verlauf an; wer einen Kartennamen tippt, bekäme zwanzig davon, und
// ⌘Z nähme Buchstaben zurück statt Änderungen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const karte: KarteDTO = {
  id: '5',
  nummer: 0,
  werte: [{ feld: 'name', wert: 'Feuerball' }],
}

function bild(name: string) {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: `/vorschau/${name}.png`,
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [],
      nennungen: [],
      vorgaben: [],
    },
  }
}

describe('Vorschau beim Tippen', () => {
  it('zeichnet mit dem getippten Wert, ohne ihn zu schreiben', async () => {
    const Probevorschau = vi.fn(() => Promise.resolve(bild('neu')))
    const FeldSetzen = vi.fn(() => Promise.resolve({}))

    belegeApi({
      Vorschau: () => Promise.resolve(bild('alt')),
      Probevorschau,
      FeldSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await userEvent.type(await screen.findByLabelText('name'), 'X')

    await waitFor(() => {
      expect(Probevorschau).toHaveBeenCalledWith('5', 'name', 'FeuerballX', 1)
    })

    // **Noch ist nichts geschrieben.**
    expect(FeldSetzen).not.toHaveBeenCalled()
  })

  // **Beim Verlassen wird geschrieben** — und danach zeigt die gewöhnliche
  // Vorschau, was im Bestand steht.
  it('schreibt beim Verlassen des Feldes', async () => {
    const FeldSetzen = vi.fn(() => Promise.resolve({}))

    belegeApi({
      Vorschau: () => Promise.resolve(bild('alt')),
      Probevorschau: () => Promise.resolve(bild('neu')),
      FeldSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const feld = await screen.findByLabelText('name')
    await userEvent.type(feld, 'X')
    await userEvent.tab()

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'name',
        'FeuerballX',
        expect.any(Array),
      )
    })
  })

  // **Ein unveränderter Wert löst keine Probevorschau aus.** Wer ein Feld
  // anklickt und wieder verlässt, soll nicht rendern lassen.
  it('zeichnet nicht ohne Änderung', async () => {
    const Probevorschau = vi.fn(() => Promise.resolve(bild('neu')))

    belegeApi({
      Vorschau: () => Promise.resolve(bild('alt')),
      Probevorschau,
      FeldSetzen: () => Promise.resolve({}),
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const feld = await screen.findByLabelText('name')
    await userEvent.click(feld)
    await userEvent.tab()

    expect(Probevorschau).not.toHaveBeenCalled()
  })

  // **Eine Probe ohne Bild lässt das vorige stehen.** Zwischen zwei
  // Tastendrücken hat sie noch keine Antwort; wer dann umschaltete, ließe die
  // Karte bei jedem Buchstaben verschwinden.
  it('lässt das vorige Bild stehen, wenn die Probe keines gibt', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(bild('alt')),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
      FeldSetzen: () => Promise.resolve({}),
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const gezeigt = await screen.findByRole('img')
    expect(gezeigt).toHaveAttribute('src', '/vorschau/alt.png')

    await userEvent.type(await screen.findByLabelText('name'), 'X')

    await waitFor(() => {
      expect(screen.getByRole('img')).toHaveAttribute(
        'src',
        '/vorschau/alt.png',
      )
    })
  })
})
