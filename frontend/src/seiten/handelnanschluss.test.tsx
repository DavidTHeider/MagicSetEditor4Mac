// **Eine mitgereiste Karte wartet sichtbar** — UX-03.AK-01
// (#1675).
//
// # Was hier gemessen wird
//
// **Dass die Fläche sagt, worauf sie wartet.** Ein Angebot braucht ein
// Gegenüber; solange keines gewählt ist, gibt es keine Zeile, in die der
// Druck fiele. Ohne diesen Satz stünde die Karte im Raum und niemand sähe,
// dass sie auf etwas wartet (§6: leere Zustände erklären den nächsten
// Schritt).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Seitenschluessel } from '../navigation'

const eigene = kataloge.de.seiten.handeln

const derHinweis = einsetzen(eigene.anschlussOhnePartner, {
  name: 'Glutschmiede',
})

function Buehne() {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('sammlung')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Abgeber />
        <Handeln />
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
          guete: 'nearMint',
          sprache: 'de',
          finish: 'nonfoil',
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

describe('Handeln und ein mitgereister Druck', () => {
  it('sagt, dass die Karte auf ein Gegenüber wartet', async () => {
    belegeApi({ Peerkontakte: () => Promise.resolve(eineListe([])) })

    render(<Buehne />)

    expect(screen.queryByText(derHinweis)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))

    expect(await screen.findByText(derHinweis)).toBeInTheDocument()
  })

  // **Die Gegenprobe**: ohne mitgereiste Karte steht der gewöhnliche Satz da
  // und nicht dieser. Sonst sagte der Lauf oben nur, dass irgendein Text
  // erscheint.
  it('lässt ohne mitgereiste Karte den gewöhnlichen Satz stehen', async () => {
    belegeApi({ Peerkontakte: () => Promise.resolve(eineListe([])) })

    render(<Buehne />)

    expect(await screen.findByText(eigene.partnerKeiner)).toBeInTheDocument()
    expect(screen.queryByText(derHinweis)).not.toBeInTheDocument()
  })
})
