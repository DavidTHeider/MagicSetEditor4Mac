// **Die Spaltenwahl bei einem großen Spielpaket** — UXR-05
// (#1913).
//
// # Was hier gemessen wird
//
// **Mit der Menge, die es wirklich gibt.** `magic.mse-game` führt 249
// Kartenfelder, 243 davon mit einer Spalte und elf sichtbar (gezählt am
// 2026-10-02 in `card_fields` des Full-Magic-Pack). Bis UXR-05 zeichnete der
// aufgeklappte Kasten jede: 243 Kästchen und 486 Pfeilknöpfe — R6 aus
// `docs/UIREVIEW.md` lässt 200 Bedienelemente zu. Und die Ordnung bot alle
// 243 in einer Auswahl an (R7: ab 15 Werten nur mit Suche).
//
// **Was bleibt**: die gezeigten Spalten stehen offen da, jede ausgeblendete
// ist über das Suchfeld erreichbar, und sortiert wird nach dem, was die
// Liste zeigt — wie in MSE2, dessen Ordnung auf eine **sichtbare** Spalte
// zeigt (`CardListBase::rebuild`, `src/gui/control/card_list.cpp`, Zeilen
// 231–243, Fassung f9d9356 vom 2021-07-06).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KartenlisteAntwort, ListenspalteDTO } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.listenspalten

// magicspalten sind 243 Spalten, die ersten elf sichtbar — die Gestalt von
// `magic.mse-game`. Die Namen sind erfunden; gezählt wird, nicht gelesen.
const magicspalten: ListenspalteDTO[] = Array.from(
  { length: 243 },
  (_, stelle) => ({
    feld: `feld ${String(stelle + 1).padStart(3, '0')}`,
    aufschrift: `Spalte ${String(stelle + 1).padStart(3, '0')}`,
    sichtbar: stelle < 11,
    ausrichtung: 'links',
  }),
)

function aufbau(
  abweichung: Partial<KartenlisteAntwort> = {},
): KartenlisteAntwort {
  return {
    spalten: magicspalten,
    zeilen: [{ karteID: '1', nummer: 1, werte: {}, marken: [] }],
    sortierfeld: '',
    aufsteigend: true,
    ohneVorlage: false,
    fehler: null,
    ...abweichung,
  }
}

function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    ...ausKarten(() =>
      Promise.resolve({
        karten: [
          { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Erste' }] },
        ],
      }),
    ),
    Kartenlistenaufbau: () => Promise.resolve(aufbau()),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={vi.fn()}
      />
    </Sprachraum>,
  )
}

async function aufklappen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.ueberschrift }),
  )

  const kasten = within(
    screen.getByRole('region', { name: texte.ueberschrift }),
  )
  await kasten.findByLabelText('Spalte 001')

  return kasten
}

describe('Die Spaltenwahl bei 243 Spalten', () => {
  it('zeigt die elf gezeigten offen und faltet die übrigen', async () => {
    zeige()

    const kasten = await aufklappen()

    expect(kasten.getAllByRole('checkbox')).toHaveLength(11)
    expect(kasten.getAllByRole('button').length).toBeLessThan(40)
    expect(
      kasten.getByText(einsetzen(texte.ausgeblendet, { anzahl: '232' })),
    ).toBeInTheDocument()
  })

  it('findet eine ausgeblendete Spalte über das Suchfeld und blendet sie ein', async () => {
    const setzen = vi.fn(() => Promise.resolve({ fehler: null }))
    zeige({ ListenspaltenSetzen: setzen })

    const kasten = await aufklappen()
    await userEvent.type(kasten.getByLabelText(texte.filtern), 'Spalte 200')
    await userEvent.click(kasten.getByLabelText('Spalte 200'))

    await waitFor(() => {
      expect(setzen).toHaveBeenCalled()
    })
    const [, gesendet] = setzen.mock.calls[0] as unknown as [
      string,
      ListenspalteDTO[],
    ]
    expect(
      gesendet.find((spalte) => spalte.feld === 'feld 200')?.sichtbar,
    ).toBe(true)
    expect(gesendet).toHaveLength(243)
  })

  it('zeichnet bei einer ungenauen Suche höchstens zwölf und sagt, dass es mehr gibt', async () => {
    zeige()

    const kasten = await aufklappen()
    await userEvent.type(kasten.getByLabelText(texte.filtern), 'Spalte')

    expect(kasten.getAllByRole('checkbox')).toHaveLength(12)
    expect(kasten.getByText(texte.genauer)).toBeInTheDocument()
  })

  it('bietet die Ordnung nach den gezeigten Spalten an — und hält eine gewählte ausgeblendete', async () => {
    zeige({
      Kartenlistenaufbau: () =>
        Promise.resolve(aufbau({ sortierfeld: 'feld 100' })),
    })

    const kasten = await aufklappen()
    const ordnung = kasten.getByLabelText(texte.sortieren)
    const angeboten = within(ordnung)
      .getAllByRole('option')
      .map((eine) => eine.textContent)

    expect(angeboten).toEqual([
      texte.garNicht,
      ...magicspalten.slice(0, 11).map((spalte) => spalte.aufschrift),
      'Spalte 100',
    ])
    expect(ordnung).toHaveValue('feld 100')
  })
})
