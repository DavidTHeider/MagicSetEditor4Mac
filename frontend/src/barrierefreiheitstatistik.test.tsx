// **Barrierefreiheit der Statistik mit Zahlen** — ausgelagert aus
// `barrierefreiheit.test.tsx`, als die Datei über die Zeilengrenze wuchs (§2).
// Ohne Set zeigt die Statistik nur einen Hinweis; die Balken, die
// Beschreibungsliste und die Feldwahl sind das, was hier geprüft gehört.
import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Rahmen } from './rahmen/Rahmen'
import { Statistik } from './seiten/Statistik'
import { Sprachraum } from './sprache/kontext'
import { verstoesse } from './test/axe'
import { belegeApi, raeumeApi } from './test/api'

it('hält die Statistik mit Zahlen ein', async () => {
  belegeApi({
    Statistik: () =>
      Promise.resolve({
        statistik: {
          karten: 12,
          felder: 2,
          leere: 1,
          belegung: [
            { name: 'name', karten: 12 },
            { name: 'typ', karten: 6 },
          ],
        },
      }),
    Verteilung: () => Promise.resolve({ verteilung: [] }),
  })

  const { container } = render(
    <Sprachraum sprache="de">
      <Rahmen aktuell="statistik" wechseln={vi.fn()}>
        <Statistik setID="7" />
      </Rahmen>
    </Sprachraum>,
  )

  await screen.findByText('12')

  expect(await verstoesse(container)).toEqual([])

  raeumeApi()
})
