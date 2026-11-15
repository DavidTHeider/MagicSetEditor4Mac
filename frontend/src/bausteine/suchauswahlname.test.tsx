// **Der Name des Feldes bleibt, auch wenn es offen ist und in einem `<label>`
// steht** — UXR-F1 (#2007).
//
// Mehrere Stellen (die Einstellungen, die Suche) umschließen ihr Bedienelement
// mit dem `<label>`. Dessen Text ist der Name des Feldes — **und alles, was in
// ihm steht**, auch ein Text, den man nicht sieht. Eine Live-Region im Baum des
// Feldes machte aus „Zeitzone" bei offener Liste „Zeitzone Treffer: 4": der
// Name änderte sich mit jeder Eingabe. Sie steht deshalb am Ende von `body`.
//
// **Ohne Hervorhebung geprüft**: `dom-accessibility-api` nimmt bei einem
// Kombinationsfeld im eigenen `<label>` den Text der hervorgehobenen Zeile in
// den Namen auf, was ein Browser nicht tut — eine Eigenheit der Prüfung, die
// hier nicht gemeint ist.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { zonen } from '../test/suchauswahlprobe'
import { Suchauswahl } from './Suchauswahl'

describe('Suchauswahl — der Name im label', () => {
  it('bleibt bei offener Liste und bei einer Eingabe derselbe', async () => {
    render(
      <Sprachraum sprache="de">
        <label htmlFor="a">
          <span>Zeitzone</span>
          <Suchauswahl
            id="a"
            bezeichnung="Zeitzone"
            eintraege={zonen.slice(1)}
            wert=""
            aendern={vi.fn()}
          />
        </label>
      </Sprachraum>,
    )

    const feld = () => screen.getByRole('combobox', { name: 'Zeitzone' })

    expect(feld()).toBeInTheDocument()

    await userEvent.click(feld())

    expect(feld()).toBeInTheDocument()

    // Nichts passt: die Live-Region sagt es, der Name bleibt.
    await userEvent.keyboard('tokio')

    expect(feld()).toBeInTheDocument()
  })

  // **Die Live-Region ist trotzdem da** und sagt an, was gilt.
  it('sagt an, ohne im label zu stehen', async () => {
    render(
      <Sprachraum sprache="de">
        <label htmlFor="a">
          <span>Zeitzone</span>
          <Suchauswahl
            id="a"
            bezeichnung="Zeitzone"
            eintraege={zonen.slice(1)}
            wert=""
            aendern={vi.fn()}
          />
        </label>
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('combobox', { name: 'Zeitzone' }))

    const ansage = document.querySelector('[data-suchauswahl]')

    expect(ansage).toHaveTextContent('Treffer: 4')
    expect(screen.getByText('Zeitzone').closest('label')).not.toContainElement(
      ansage as HTMLElement,
    )
  })
})
