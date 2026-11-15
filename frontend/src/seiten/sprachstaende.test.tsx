// **Je Sprache steht da, wie weit sie ist, woher sie kommt und wer sie geprüft
// hat** — H04 (#2258).
//
// Bis dahin sagte ein Satz unter der Sprachwahl alles für alle: „alle anderen
// Sprachen sind maschinell übersetzt". Welche wann, und ob ein Mensch eine
// gelesen hat, stand nur in `docs/SPRACHINVENTAR.md`.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import {
  kataloge,
  sprachen,
  sprachherkunft,
  sprachstaende,
} from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Sprachstaende } from './Sprachstaende'

import type { Sprache } from '../sprache/kataloge'

async function zeige(sprache: Sprache) {
  render(
    <Sprachraum sprache={sprache}>
      <Sprachstaende />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', {
      name: kataloge[sprache].seiten.sprachstaende.titel,
    }),
  )
}

describe('Wie weit jede Sprache ist', () => {
  it.each(sprachen)(
    'nennt in %s jede Sprache mit Stand, Herkunft und Prüfung',
    async (fenster) => {
      await zeige(fenster)

      const texte = kataloge[fenster].seiten.sprachstaende
      const namen: Record<string, string> =
        kataloge[fenster].seiten.einstellungen.werte

      for (const sprache of sprachen) {
        const zeile = screen.getByRole('row', {
          name: new RegExp(namen[`sprache.${sprache}`] ?? ''),
        })
        const herkunft = sprachherkunft[sprache]

        expect(
          within(zeile).getByText(texte.stand[sprachstaende[sprache]]),
        ).toBeInTheDocument()

        // **Keine Redaktion, die nicht stattfand** — und jedes Stück um den
        // Tag herum steht da, gleich wo die Sprache ihn hinsetzt.
        const redaktion =
          herkunft.geprueft === '' ? texte.ungeprueft : texte.geprueft

        for (const satz of [texte.herkunft[herkunft.herkunft], redaktion]) {
          for (const stueck of satz
            .split('{datum}')
            .map((teil) => teil.trim())
            .filter(Boolean)) {
            expect(zeile).toHaveTextContent(stueck)
          }
        }
      }
    },
  )

  it('sagt den Stand in der Sprache des Fensters und nicht auf Deutsch', () => {
    for (const fenster of sprachen.filter((s) => s !== 'de')) {
      expect(
        kataloge[fenster].seiten.sprachstaende.stand.maschinell,
        fenster,
      ).not.toBe(kataloge.de.seiten.sprachstaende.stand.maschinell)
    }
  })
})
