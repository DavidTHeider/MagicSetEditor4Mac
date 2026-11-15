// **Die Vorgabe wirkt dort, wo man sie sieht** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Eine Vorgabe ohne Wirkung wäre eine Zeile im Bestand.** Was hier geprüft
// wird, ist genau die Wirkung — nicht, dass sich etwas speichern lässt.
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { useVorgabe } from './useVorgabe'

afterEach(() => {
  raeumeApi()
})

// Zeuge zeigt, was der Haken gibt.
function Zeuge({ art }: { art: string }) {
  return <p data-testid="werte">{useVorgabe(art).join(',')}</p>
}

function zeige(teile: Record<string, unknown>) {
  belegeApi(teile)

  render(
    <Sprachraum sprache="de">
      <Zeuge art="spiel" />
    </Sprachraum>,
  )
}

describe('Die Vorgabe des gezeigten Profils', () => {
  it('gibt die Werte ihrer Art', async () => {
    zeige({
      GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: { id: '10' } })),
      Profilvorgaben: vi.fn(() =>
        Promise.resolve({
          vorgaben: [
            {
              art: 'spiel',
              mehrwertig: false,
              werte: ['magic'],
              moeglich: [],
              verwaist: [],
            },
            {
              art: 'deck',
              mehrwertig: true,
              werte: ['1'],
              moeglich: [],
              verwaist: [],
            },
          ],
        }),
      ),
    })

    await waitFor(() => {
      expect(screen.getByTestId('werte')).toHaveTextContent('magic')
    })
  })

  it('gibt nichts, wenn es kein gezeigtes Profil gibt', async () => {
    // **Der Fall, den es geben kann**: `GezeigtesProfil` gibt `null`, wenn
    // der Zeiger ins Leere sieht.
    const gefragt = vi.fn(() => Promise.resolve({ vorgaben: [] }))

    zeige({
      GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: null })),
      Profilvorgaben: gefragt,
    })

    // **Auf den Ruf gewartet und nicht auf die Wirkung** — die Aussage ist
    // eine Verneinung, und die trifft vorher ohnehin zu (§3). Der Beleg, dass
    // jemand zuhört, ist der erste Ruf.
    await waitFor(() => {
      expect(screen.getByTestId('werte')).toHaveTextContent('')
    })

    expect(gefragt).not.toHaveBeenCalled()
  })

  it('bleibt stumm, wenn das Lesen scheitert', async () => {
    // **Eine Vorgabe ist Bequemlichkeit.** Wer ein Set anlegen will, soll das
    // können, auch wenn die Vorwahl nicht zu holen war — eine Meldung darüber
    // stünde im Weg und sagte nichts, was jemand täte.
    zeige({
      GezeigtesProfil: vi.fn(() =>
        Promise.resolve({
          fehler: { code: 'unbekannt', schluessel: 'unbekannt' },
        }),
      ),
    })

    await waitFor(() => {
      expect(screen.getByTestId('werte')).toHaveTextContent('')
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('gibt nichts für eine Art, die nicht dasteht', async () => {
    zeige({
      GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: { id: '10' } })),
      Profilvorgaben: vi.fn(() => Promise.resolve({ vorgaben: [] })),
    })

    await waitFor(() => {
      expect(screen.getByTestId('werte')).toHaveTextContent('')
    })
  })
})
