// **Drei Seiten, ein Assistent — im Arbeitsplatz** — UXR-09
// (#1917).
//
// `ausgeben`, `bilder` und `druckbogen` behalten Schlüssel, Kürzel und
// Menüeintrag, schlagen aber **denselben** Assistenten auf. Hier steht, dass
// es wirklich derselbe ist: wer mit ⌥⌘0 den Druckbogen einstellt und mit
// ⌥⌘1 zu den Bildern geht, findet beim Zurückkommen seine Einstellung — und
// es steht nie mehr als eine Frage „Wohin?" im Fenster.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { aufheben, nichtsAufgehoben } from '../haken/useAufgehobeneSeiten'
import type { Setschluessel } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Setbereich } from './Setbereich'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten

// bereich zeichnet den Arbeitsplatz für eine Folge von Seiten — aufgehoben
// wird, was der Rahmen auch aufhöbe.
function bereich(gang: readonly Setschluessel[]) {
  const stand = gang.reduce(
    (bisher, seite) => aufheben(bisher, seite, '7'),
    nichtsAufgehoben,
  )

  return (
    <Sprachraum sprache="de">
      <Setbereich
        seite={gang[gang.length - 1] ?? 'karten'}
        aufgehoben={stand}
        setID="7"
        wechseln={vi.fn()}
        karteID={undefined}
        oeffneKarte={vi.fn()}
        oeffneSymbol={vi.fn()}
        bestandGeaendert={vi.fn()}
        verlaufsstand={0}
        verlaufsfehler={undefined}
      />
    </Sprachraum>
  )
}

describe('Der Ausgabeassistent im Arbeitsplatz', () => {
  it('ist für Druckbogen und Bilder derselbe', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [] }),
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    })

    const { rerender } = render(bereich(['druckbogen']))

    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: texte.druckbogen.papier }),
      'letter',
    )

    rerender(bereich(['druckbogen', 'bilder']))

    expect(
      await screen.findByRole('radio', { name: texte.bilder.titel }),
    ).toBeChecked()

    // **Auch verborgen nur einer** — drei Assistenten nebeneinander hielten
    // drei Wahlen, und welcher gerade dasteht, entschiede die Taste.
    expect(
      screen.getAllByRole('group', {
        name: texte.ausgeben.wohin,
        hidden: true,
      }),
    ).toHaveLength(1)

    // **Und darin der eingestellte Druckbogen** — die Wahl unter „Wohin?"
    // führt in denselben Assistenten zurück, nicht in einen frischen.
    await userEvent.click(
      screen.getByRole('radio', { name: texte.druckbogen.titel }),
    )

    expect(
      screen.getByRole('combobox', { name: texte.druckbogen.papier }),
    ).toHaveValue('letter')
  })
})
