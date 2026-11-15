// **Was in der Sammlung gewählt wurde, steht im Angebot schon da** —
// UX-03.AK-01
// (#1675).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Handelspositionen } from './Handelspositionen'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'

import type { Handelsposition } from '../apihandelsleitung'
import type { Seitenschluessel } from '../navigation'

const texte = kataloge.de.seiten.handeln

const derHinweis = einsetzen(texte.anschlussWartet, { name: 'Glutschmiede' })

function Buehne({
  setzePositionen,
}: {
  setzePositionen: (positionen: Handelsposition[]) => void
}) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('sammlung')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Abgeber />
        <Handelspositionen
          ich="ich-1"
          gegenueber="du-1"
          positionen={[]}
          setzePositionen={setzePositionen}
          bearbeitbar
        />
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
          art: 'handel',
          name: 'Glutschmiede',
          printing: 'p-9',
          guete: 'lightPlayed',
          sprache: 'en',
          finish: 'etched',
        })
      }}
    >
      abgeben
    </button>
  )
}

describe('die Handelszeile nimmt auf, was mitgereist ist', () => {
  it('füllt Druck, Güte, Sprache und Ausführung', async () => {
    render(<Buehne setzePositionen={vi.fn()} />)

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))

    expect(screen.getByLabelText(texte.positionPrinting)).toHaveValue('p-9')
    expect(screen.getByLabelText(texte.positionGuete)).toHaveValue(
      'lightPlayed',
    )
    expect(screen.getByLabelText(texte.positionSprache)).toHaveValue('en')
    expect(screen.getByLabelText(texte.positionFinish)).toHaveValue('etched')
    expect(screen.getByText(derHinweis)).toBeInTheDocument()
  })

  it('hängt die mitgereiste Zeile an, ohne dass jemand tippt', async () => {
    const setzePositionen = vi.fn()

    render(<Buehne setzePositionen={setzePositionen} />)

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))
    await userEvent.click(
      screen.getByRole('button', { name: texte.positionHinzufuegen }),
    )

    expect(setzePositionen).toHaveBeenCalledWith([
      expect.objectContaining({
        von: 'ich-1',
        printing: 'p-9',
        guete: 'lightPlayed',
        sprache: 'en',
        finish: 'etched',
      }),
    ])
  })

  it('räumt den Hinweis, sobald die Zeile steht', async () => {
    render(<Buehne setzePositionen={vi.fn()} />)

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))
    await userEvent.click(
      screen.getByRole('button', { name: texte.positionHinzufuegen }),
    )

    expect(screen.queryByText(derHinweis)).not.toBeInTheDocument()
  })
})

// **Ein Wunsch aus einem Deck wird zu Positionen vom Gegenüber** —
// UX-03.AK-01 (#1675), durch die echten Handelszeilen.
describe('ein Tauschwunsch aus einem Deck', () => {
  it('trägt jede fehlende Karte als Position vom Gegenüber ein', async () => {
    const setzePositionen = vi.fn()
    const vorhanden: Handelsposition = {
      von: 'ich-1',
      printing: 'p-1',
      druckId: 0,
      menge: 1,
      guete: '',
      sprache: '',
      finish: '',
      abbild: '',
      bewertung: 0,
    }

    render(
      <Sprachraum sprache="de">
        <Anschlussbuehne
          anschluss={{
            art: 'tauschwunsch',
            name: '3 Karten',
            stuecke: [{ printing: 'druck-a', name: 'Aschenwächter', menge: 3 }],
          }}
        >
          <Handelspositionen
            ich="ich-1"
            gegenueber="du-1"
            positionen={[vorhanden]}
            setzePositionen={setzePositionen}
            bearbeitbar
          />
        </Anschlussbuehne>
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))
    await userEvent.click(
      screen.getByRole('button', { name: texte.wunschEintragen }),
    )

    // **Mit der Vorgabe einer neuen Zeile** (#2495): ohne Sprache ließ sich
    // die Zeile anbieten und scheiterte erst beim Verbuchen. Zustand, Sprache
    // und Ausführung sind die des Formulars (`dieLeereZeile`).
    expect(setzePositionen).toHaveBeenCalledWith([
      vorhanden,
      {
        von: 'du-1',
        printing: 'druck-a',
        druckId: 0,
        menge: 3,
        guete: 'nearMint',
        sprache: 'de',
        finish: 'nonfoil',
        abbild: '',
        bewertung: 0,
      },
    ])
  })
})
