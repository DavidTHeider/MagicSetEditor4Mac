// Die Abhängigkeiten der installierten Pakete (#57).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { type Api } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { AbhaengigkeitenGelesen } from '../test/abhaengigkeiten'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function zeige(teile: Partial<Api>) {
  belegeApi({
    Abhaengigkeiten: () => Promise.resolve({ befunde: [] }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <AbhaengigkeitenGelesen />
    </Sprachraum>,
  )
}

describe('Abhängigkeiten', () => {
  it('sagt, dass alles erfüllt ist', async () => {
    zeige({})

    expect(
      await screen.findByText(texte.seiten.pakete.abhaengigAllesGut),
    ).toBeInTheDocument()
  })

  // **Der Satz entsteht hier, aus Angaben von dort** (§9). Jede der vier Arten
  // bekommt ihren eigenen — ohne diesen Fall stünde für drei von ihnen der
  // Satz der vierten da, und niemand sähe es.
  it('formuliert jede der vier Arten in ihrer Sprache', async () => {
    zeige({
      Abhaengigkeiten: () =>
        Promise.resolve({
          befunde: [
            {
              paket: 'magic-m15.mse-style',
              braucht: 'magic.mse-game',
              art: 'fehlt',
              verlangt: '2009-07-23',
              da: '',
            },
            {
              paket: 'a.mse-style',
              braucht: 'b.mse-game',
              art: 'zualt',
              verlangt: '2024-01-01',
              da: '2009-07-23',
            },
            {
              paket: 'eins.mse-style',
              braucht: 'zwei.mse-style',
              art: 'zyklus',
              verlangt: '',
              da: '',
            },
            {
              paket: 'c.mse-style',
              braucht: 'd.mse-game',
              art: 'unlesbar',
              verlangt: 'bald',
              da: '2009-07-23',
            },
          ],
        }),
    })

    expect(await screen.findByText(/nicht installiert/)).toBeInTheDocument()
    expect(screen.getByText(/zeichnet die Vorlage still/)).toBeInTheDocument()
    expect(screen.getByText(/im Kreis/)).toBeInTheDocument()
    expect(screen.getByText(/nicht vergleichen/)).toBeInTheDocument()
  })

  it('meldet einen Fehler am Ort', async () => {
    zeige({
      Abhaengigkeiten: vi.fn(() =>
        Promise.resolve({
          befunde: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Katalog schweigt.',
          },
        }),
      ),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Katalog schweigt.',
    )
  })
})
