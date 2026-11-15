// **Was in der Sammlung gewählt wurde, steht im Deck schon da** —
// UX-03.AK-01
// (#1675).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Deckkartenwahl } from './Deckkartenwahl'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'

import type { Seitenschluessel } from '../navigation'

const texte = kataloge.de.seiten.decks

const derHinweis = einsetzen(texte.anschlussWartet, { name: 'Aschenwächter' })

// Buehne stellt den Raum so auf, wie `Anwendung` es tut — und gibt den
// Anschluss auf Knopfdruck ab, statt ihn hineinzuschreiben.
function Buehne({ legen }: { legen: (druck: string) => void }) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('sammlung')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Abgeber />
        <Deckkartenwahl
          spiel="magic"
          angehalten={false}
          legen={(_karteID, printing) => {
            legen(printing)
          }}
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
          art: 'deck',
          name: 'Aschenwächter',
          printing: 'p-4',
        })
      }}
    >
      abgeben
    </button>
  )
}

function zeige(legen = vi.fn()) {
  belegeApi({
    Sets: () => Promise.resolve({ sets: [], fehler: null }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
  })

  render(<Buehne legen={legen} />)

  return legen
}

describe('die Kartenwahl des Decks nimmt auf, was mitgereist ist', () => {
  it('trägt den Druck ein und sagt, woher er kommt', async () => {
    zeige()

    expect(screen.queryByText(derHinweis)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))

    expect(screen.getByLabelText(texte.druck)).toHaveValue('p-4')
    expect(screen.getByText(derHinweis)).toBeInTheDocument()
  })

  it('legt den mitgereisten Druck, ohne dass jemand tippt', async () => {
    const legen = zeige()

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))
    await userEvent.click(screen.getByRole('button', { name: texte.legen }))

    expect(legen).toHaveBeenCalledWith('p-4')
  })

  // **Verbraucht ist verbraucht.** Wer die Karte gelegt hat, soll den Satz
  // nicht weiter vor sich haben — sonst sagt er beim nächsten Blick etwas
  // über eine Handlung, die längst geschehen ist.
  it('räumt den Hinweis, sobald die Karte liegt', async () => {
    zeige()

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))
    await userEvent.click(screen.getByRole('button', { name: texte.legen }))

    expect(screen.queryByText(derHinweis)).not.toBeInTheDocument()
  })
})
