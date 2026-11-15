// **Wo ein Feldwert als Fließtext steht, steht er ohne seine Auszeichnung**
// (#952).
//
// # Die Klasse, nicht der Fall
//
// Gefunden wurde es an der **Suche**: ein Treffer im Aromatext zeigte
// `<i-flavor>As a ruthless leader …</i-flavor>` statt des Satzes. Der Fall ist
// aber nicht die Suche, sondern **jede** Stelle, die einen Feldwert außerhalb
// der Kartenvorschau als Text zeigt — und das sind drei: die Trefferliste, der
// Kartenname in jeder Liste, und die Verteilung in der Statistik.
//
// **Die Vorschau bleibt außen vor.** Dort *deutet* der Renderer die
// Auszeichnung, statt sie wegzuwerfen; das ist der Unterschied zwischen
// „zeigen, was dasteht" und „zeigen, was die Karte zeigt".
//
// **Und der Wert selbst wird nicht angefasst.** `ohneAuszeichnung` nimmt nur
// für die Anzeige weg; geschrieben wird weiterhin die rohe Form.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { Kartenliste } from './Kartenliste'
import { Statistik } from './Statistik'
import { Suche } from './Suche'

afterEach(() => {
  raeumeApi()
})

// ausgezeichnet ist ein Feldwert, wie er in einer echten `.mse-set` steht.
const ausgezeichnet = '<i-flavor>Ein Satz mit Rand.</i-flavor>'
const klar = 'Ein Satz mit Rand.'

describe('Auszeichnung in Listen', () => {
  it('zeigt einen Treffer der Suche ohne seine Marken', async () => {
    belegeApi({
      Suchfelder: () => Promise.resolve({ felder: ['flavor text'] }),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchen: () =>
        Promise.resolve({
          treffer: [
            {
              karteID: '1',
              setID: '7',
              nummer: 0,
              kartenname: `<i>Mirri</i>`,
              feld: 'flavor text',
              wert: ausgezeichnet,
            },
          ],
          gekuerzt: false,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Suche oeffnen={vi.fn()} />
      </Sprachraum>,
    )

    // **Mit Namen** (A16, #1515): seit die Recherche im selben
    // Arbeitsplatz steht, gibt es zwei Suchfelder.
    await userEvent.type(
      await screen.findByRole('searchbox', {
        name: kataloge.de.seiten.suche.eingabe,
      }),
      'Rand',
    )

    await waitFor(() => {
      expect(screen.getByText(klar)).toBeInTheDocument()
    })

    // **Auch der Kartenname**: er ist derselbe Feldwert, nur an anderer Stelle.
    expect(screen.getByRole('button', { name: 'Mirri' })).toBeInTheDocument()
  })

  it('zeigt einen Kartennamen in der Liste ohne seine Marken', () => {
    render(
      <Sprachraum sprache="de">
        <Kartenliste
          spalten={[]}
          marken={{}}
          setID="7"
          stand={0}
          karten={[{ id: '1', nummer: 1, name: '<i>Mirri</i>', felder: 1 }]}
          gewaehlt="1"
          waehlen={vi.fn()}
          angehakt={[]}
          anhaken={vi.fn()}
          mitKaestchen={false}
        />
      </Sprachraum>,
    )

    expect(screen.getByRole('button', { name: /Mirri/ })).toBeInTheDocument()
    expect(screen.queryByText(/<i>/)).not.toBeInTheDocument()
  })

  it('zeigt einen Feldwert in der Verteilung ohne seine Marken', async () => {
    belegeApi({
      Statistik: () =>
        Promise.resolve({
          statistik: {
            karten: 1,
            felder: 1,
            leere: 0,
            belegung: [{ name: 'flavor text', karten: 1 }],
          },
        }),
      Verteilung: () =>
        Promise.resolve({
          verteilung: [{ name: ausgezeichnet, karten: 1, kartenIDs: ['1'] }],
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Statistik setID="7" />
      </Sprachraum>,
    )

    // **Die Verteilung braucht ein gewähltes Feld** — ohne eines fragt die
    // Seite gar nicht erst.
    await userEvent.selectOptions(
      await screen.findByRole('combobox'),
      'flavor text',
    )

    await waitFor(() => {
      expect(screen.getByText(klar)).toBeInTheDocument()
    })

    // **Und die Belegung darüber zeigt weiterhin den Feldnamen**, unangetastet:
    // sie nennt Felder, nicht Werte.
    expect(screen.getAllByText('flavor text').length).toBeGreaterThan(0)
  })
})
