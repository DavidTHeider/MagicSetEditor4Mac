// **Ein fremder Setstilwechsel bei offenem Editor, durch den ganzen Editor**
// — #2314 (E2314).
//
// Ein anderer Prozess wechselt den Stil des Sets. Das Fenster erfährt es über
// den Datenstand (`stand`), lädt die Vorschau neu, und die zeichnet mit dem
// neuen Paket in einer anderen Lage. **Die rohe Eingabe ist dieser Wechsel**
// (§3) — nicht eine Ansicht, in der die Ordnung schon steht.
import { act, render, screen, waitFor } from '@testing-library/react'
import { useState } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO, KarteDTO, VorschauAntwort } from '../api'
import type { FeldflaecheDTO } from '../formen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'
import { grundansicht } from './vorschauansicht'

afterEach(() => {
  raeumeApi()
})

function form(name: string): FeldbeschreibungDTO {
  return {
    name,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    aufschrift: name,
  }
}

const formen = ['name', 'kosten', 'text'].map(form)

function flaeche(feld: string, oben: number): FeldflaecheDTO {
  return { feld, links: 0, oben, breite: 10, hoehe: 5 }
}

function gezeichnetMit(
  paket: string,
  flaechen: FeldflaecheDTO[],
): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: paket,
      bild: '/vorschau/BBBB.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen,
      nennungen: [],
      vorgaben: [],
    },
  }
}

const karte = {
  id: '2',
  nummer: 1,
  fassung: 1,
  werte: [
    { feld: 'name', wert: 'Erbt den Setstil' },
    { feld: 'kosten', wert: '2R' },
  ],
} as unknown as KarteDTO

function Seite({ stand }: { stand: number }) {
  const ansicht = useState(grundansicht)

  return (
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte}
        formen={formen}
        geaendert={() => undefined}
        ansicht={ansicht}
        stand={stand}
      />
    </Sprachraum>
  )
}

function reihenfolge(): string[] {
  const formular = screen.getByRole('textbox', { name: 'name' }).closest('form')

  return [...(formular?.querySelectorAll('[aria-label]') ?? [])]
    .map((e) => e.getAttribute('aria-label') ?? '')
    .filter((name) => formen.some((f) => f.name === name))
}

it('ordnet nach dem fremden Wechsel nach dem neuen Paket, Werte und Fokus bleiben', async () => {
  const Vorschau = vi
    .fn()
    .mockResolvedValueOnce(
      gezeichnetMit('magic-alt.mse-style', [
        flaeche('name', 0),
        flaeche('kosten', 10),
        flaeche('text', 20),
      ]),
    )
    .mockResolvedValue(
      gezeichnetMit('magic-neu.mse-style', [
        flaeche('text', 0),
        flaeche('kosten', 10),
        flaeche('name', 20),
      ]),
    )
  belegeApi({ Vorschau })

  const { rerender } = render(<Seite stand={0} />)
  await waitFor(() => {
    expect(reihenfolge()).toEqual(['name', 'kosten', 'text'])
  })

  const name = screen.getByRole('textbox', { name: 'name' })
  act(() => {
    name.focus()
  })

  // **Der fremde Wechsel**: der Datenstand steigt, die Karte bleibt dieselbe.
  rerender(<Seite stand={1} />)

  await waitFor(() => {
    expect(reihenfolge()).toEqual(['text', 'kosten', 'name'])
  })
  expect(Vorschau).toHaveBeenCalledTimes(2)
  expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
    'Erbt den Setstil',
  )
  expect(screen.getByRole('textbox', { name: 'kosten' })).toHaveValue('2R')
  expect(document.activeElement).toBe(
    screen.getByRole('textbox', { name: 'name' }),
  )
})
