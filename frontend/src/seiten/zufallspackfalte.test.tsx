// **Ein großes Pack steht gefaltet da** — UXR-08
// (#1916).
//
// # Der Befund
//
// Am Probebestand zieht ein Pack eine Handvoll Karten, und die Lasche misst
// eine Rolltiefe von 1,2. `magic.mse-game` bietet zehn wählbare Packarten an,
// und die **erste** — die, die ungefragt dasteht — ist das *tournament pack*:
// 3 + 9 + 33 + 30 + 1 = **76** Karten (`packs`, nachgezählt am 2026-10-02).
// Untereinander in einer Schublade sind das gut 1.800 Punkte, R4 aus
// `docs/UIREVIEW.md` erlaubt das Zweieinhalbfache des Bildes.
//
// # Was gilt
//
// **Vorn steht ein Booster** — sechzehn Karten, so viele, wie das
// *booster pack* desselben Spiels zieht (1 + 3 + 10 + 1 + 1). Der Rest steht
// hinter einem Knopf, und die Zahl daneben sagt, dass es einen Rest gibt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { Zufallspack } from './Zufallspack'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.pack
const worte = kataloge.de.bausteine.listenansicht

// turnierpack sind 76 Karten — „Karte 1" bis „Karte 76".
const turnierpack: KarteDTO[] = Array.from({ length: 76 }, (_, stelle) => ({
  id: String(stelle + 1),
  nummer: stelle + 1,
  werte: [{ feld: 'name', wert: ['Karte', String(stelle + 1)].join(' ') }],
}))

describe('Ein großes Pack', () => {
  it('zeigt vorn einen Booster und den Rest auf Verlangen', async () => {
    belegeApi({
      Packarten: () =>
        Promise.resolve({
          arten: [{ name: 'tournament pack', waehlbar: true, karten: 300 }],
          befunde: [],
          fehler: null,
        }),
      PackZiehen: vi.fn(() =>
        Promise.resolve({ karten: turnierpack, befunde: [], fehler: null }),
      ),
    })

    const { container } = render(
      <Sprachraum sprache="de">
        <Zufallspack setID="3" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ziehen }),
    )

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')).toHaveLength(16)
    })
    expect(
      screen.getByText(einsetzen(worte.zahl, { gezeigt: '16', gesamt: '76' })),
    ).toBeInTheDocument()

    // **Gefaltet hält es WCAG ein** — Zahl und Knopf neben der Liste.
    expect(await verstoesse(container)).toEqual([])

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(worte.weitere, { anzahl: '60' }),
      }),
    )

    // **Alle 76, in ihrer Ziehreihenfolge.**
    const gezogen = screen.getAllByRole('listitem')
    expect(gezogen).toHaveLength(76)
    expect(gezogen[75]).toHaveTextContent('Karte 76')
  })
})
