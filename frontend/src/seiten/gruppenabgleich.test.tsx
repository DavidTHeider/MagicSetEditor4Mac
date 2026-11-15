// **Eine Wahl gilt für das Bild, aus dem sie stammt** — UXR-F6
// (#2013),
// AGENTS.md §7, Regel 2.
//
// **Kommt die Statistik zurück, rechnet sie neu.** Die Schublade bleibt
// stehen (`<Activity>`), verborgen und wieder sichtbar, und ihre Zahlen werden
// neu geholt. In MSE2 verwirft `StatsPanel::onChange` die Wahl, sobald die
// Daten neu gesetzt sind (`GraphControl::setData` leert `current_item`,
// `src/gui/control/graph.cpp` Zeile 1084, Fassung f9d9356 vom 2021-07-06);
// hier **folgt die Wahl ihrem Balken**, solange es ihn gibt
// (`useGruppenabgleich`) — angepasst, denn wer den Balken gewählt hat, meinte
// ihn und nicht den Stand von vorhin.
//
// Die Fälle teilen sich die Vorrichtung mit `gruppenwahl.test.tsx`
// (`test/gruppenproben.tsx`).
import { render, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { WertzahlDTO } from '../api'
import { raeumeApi } from '../test/api'
import {
  auszaehlung,
  balken,
  belege,
  farben,
  filter,
  seite,
  seltenheiten,
  waehleBild,
} from '../test/gruppenproben'

afterEach(() => {
  raeumeApi()
})

// nacheinander antwortet mit je einem Stand der Verteilung pro Ruf.
function nacheinander(...staende: WertzahlDTO[][]) {
  const verteilung = vi.fn()

  for (const stand of staende) {
    verteilung.mockResolvedValueOnce({ verteilung: stand })
  }

  belege({ Verteilung: verteilung })

  return verteilung
}

// zuruecksehen verbirgt die Schublade und zeigt sie wieder — und wartet, bis
// die Zahlen zum `rufe`-ten Mal geholt sind.
async function zuruecksehen(
  rerender: (neu: ReactNode) => void,
  weg: ReturnType<typeof vi.fn>,
  rufe: number,
) {
  rerender(seite(vi.fn(), false))
  rerender(seite(vi.fn(), true))

  await waitFor(() => {
    expect(weg).toHaveBeenCalledTimes(rufe)
  })
}

describe('Die Wahl nach dem Neuladen', () => {
  it('folgt dem Balken, wenn er andere Karten trägt', async () => {
    const verteilung = nacheinander(seltenheiten, [
      { name: 'rare', karten: 3, kartenIDs: ['3', '7', '9'] },
    ])
    const { rerender } = render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))
    expect(filter()).toHaveTextContent('rare | 3,7')

    await zuruecksehen(rerender, verteilung, 2)

    await waitFor(() => {
      expect(filter()).toHaveTextContent('rare | 3,7,9')
    })
  })

  it('bleibt, wenn der Balken dieselben Karten trägt', async () => {
    const verteilung = nacheinander(seltenheiten, seltenheiten)
    const { rerender } = render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))

    await zuruecksehen(rerender, verteilung, 2)

    expect(filter()).toHaveTextContent('rare | 3,7')
    expect(await balken('rare')).toHaveAttribute('aria-pressed', 'true')
  })

  it('endet, wenn der Balken verschwunden ist', async () => {
    const verteilung = nacheinander(seltenheiten, [
      { name: 'common', karten: 3, kartenIDs: ['1', '2', '5'] },
    ])
    const { rerender } = render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))

    await zuruecksehen(rerender, verteilung, 2)

    await waitFor(() => {
      expect(filter()).toHaveTextContent('keiner')
    })
  })

  // **Ein Balken ohne Karten ist keine Gruppe mehr** — die Liste zeigte
  // nichts, und die Wahl endet lieber.
  it('endet, wenn der Balken leer geworden ist', async () => {
    const verteilung = nacheinander(seltenheiten, [
      { name: 'rare', karten: 0, kartenIDs: [] },
    ])
    const { rerender } = render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))

    await zuruecksehen(rerender, verteilung, 2)

    await waitFor(() => {
      expect(filter()).toHaveTextContent('keiner')
    })
  })

  // **Ein gescheitertes Neuladen sagt nichts über den Balken.** Die Wahl
  // bleibt, und das Bild zeigt den Fehler.
  it('bleibt, wenn das Neuladen scheitert', async () => {
    const verteilung = vi
      .fn()
      .mockResolvedValueOnce({ verteilung: seltenheiten })
      .mockRejectedValueOnce(new Error('nicht erreichbar'))

    belege({ Verteilung: verteilung })
    const { rerender } = render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))

    await zuruecksehen(rerender, verteilung, 2)

    expect(filter()).toHaveTextContent('rare | 3,7')
  })

  it('folgt in der Auszählung ebenso', async () => {
    const Auszaehlen = vi
      .fn()
      .mockResolvedValueOnce(auszaehlung(farben))
      .mockResolvedValueOnce(
        auszaehlung([
          { name: 'White', karten: 3, farbe: '', kartenIDs: ['1', '3', '4'] },
        ]),
      )

    belege({ Auszaehlen })
    const { rerender } = render(seite())

    await waehleBild('dimension:color category')
    await userEvent.click(await balken('White'))
    expect(filter()).toHaveTextContent('White | 1,3')

    await zuruecksehen(rerender, Auszaehlen, 2)

    await waitFor(() => {
      expect(filter()).toHaveTextContent('White | 1,3,4')
    })
  })
})
