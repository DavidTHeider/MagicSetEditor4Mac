// **Der Markenfilter faltet ab vielen Marken** — #2017, Folge aus UXR-10
// (#1918).
//
// Je Marke ein Knopf: mit acht Marken acht, und es wächst mit dem Bestand
// über R2 aus `docs/UIREVIEW.md` (höchstens zwanzig Bedienelemente im Bild).
// Ab mehr als acht stehen die ersten acht und jede gewählte da, der Rest
// hinter „Weitere Marken (N)".
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { Markenfilter, type Markenwahl } from './Markenfilter'

const eigene = kataloge.de.bausteine.objektmarken

function markenknoepfe() {
  // **Die Marken sind Umschalter** — sie tragen `aria-pressed`, der
  // Aufklapper `aria-expanded`.
  return within(screen.getByRole('region', { name: eigene.filterUeberschrift }))
    .queryAllByRole('button')
    .filter((knopf) => knopf.hasAttribute('aria-pressed'))
}

function zeige(anzahl: number, filter: Markenwahl) {
  const je: Record<string, string[]> = {}
  for (let i = 0; i < anzahl; i++) {
    je[String(i)] = [`marke${String(i).padStart(2, '0')}`]
  }

  render(
    <Sprachraum sprache="de">
      <Markenfilter
        art="exemplar"
        je={je}
        filter={filter}
        setzeFilter={vi.fn()}
        gespeichert="[]"
        setzeGespeichert={vi.fn()}
        getroffen={undefined}
      />
    </Sprachraum>,
  )
}

describe('Der Markenfilter mit vielen Marken', () => {
  it('zeigt bis acht Marken alle und keinen Aufklapper', () => {
    zeige(8, { marken: [], verknuepfung: 'oder' })

    expect(markenknoepfe()).toHaveLength(8)
    expect(screen.queryByRole('button', { expanded: false })).toBeNull()
  })

  it('zeigt ab neun die ersten acht und den Rest auf Wunsch', async () => {
    zeige(20, { marken: [], verknuepfung: 'oder' })

    expect(markenknoepfe()).toHaveLength(8)

    const weiter = screen.getByRole('button', {
      name: einsetzen(eigene.weitereMarken, { zahl: '12' }),
    })
    expect(weiter).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(weiter)

    expect(markenknoepfe()).toHaveLength(20)
    expect(weiter).toHaveAttribute('aria-expanded', 'true')
  })

  it('zeigt eine gewählte Marke auch hinter der Falte', () => {
    zeige(20, { marken: ['marke15'], verknuepfung: 'oder' })

    const namen = markenknoepfe().map((b) => b.textContent)
    expect(namen).toHaveLength(9)
    expect(namen).toContain('marke15')
  })
})
