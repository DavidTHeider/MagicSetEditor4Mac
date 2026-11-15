// **Die Zeile unter den Kacheln** — AP-8, Teil 2
// (#2365):
// wie alt die Prüfung ist, so grob, wie sie altert; und ein gescheiterter
// Anstoß sagt sich, ohne nachzuladen (§7, Regel 3).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import type { Problemlage } from './problemlage'
import { Problemzeile } from './Problemzeile'

const eigene = kataloge.de.seiten.start.ueberblick

afterEach(() => {
  raeumeApi()
})

const lage = (teil: Partial<Problemlage>): Problemlage => ({
  gesamt: 1,
  geprueft: 1,
  vollstaendig: true,
  mitProblemen: [],
  karten: 0,
  kartenGeprueft: 7,
  kartenGesamt: 7,
  alter: 0,
  ...teil,
})

function zeige(gegeben: Problemlage | undefined, nachladen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Problemzeile lage={gegeben} geladen nachladen={nachladen} />
    </Sprachraum>,
  )

  return nachladen
}

describe('die Zeile unter den Kacheln', () => {
  it.each([
    [30, 'gerade eben'],
    [7200, 'vor 2 h'],
    [3 * 86400, 'vor 3 Tagen'],
  ])('nennt nach %i s „%s“', (alter, wann) => {
    zeige(lage({ alter }))

    expect(
      screen.getByText(`Probleme geprüft ${wann} · 7 von 7 Karten`),
    ).toBeInTheDocument()
  })

  it('steht ohne Sets leer da, in ihrer Höhe', () => {
    const { container } = render(
      <Sprachraum sprache="de">
        <Problemzeile lage={lage({ gesamt: 0 })} geladen nachladen={vi.fn()} />
      </Sprachraum>,
    )

    const zeile = container.querySelector('.min-h-10')
    expect(zeile).not.toBeNull()
    expect(zeile?.textContent).toBe('')
  })

  it('sagt einen gescheiterten Anstoß und lädt nicht nach', async () => {
    belegeApi({
      SetproblemePruefen: () =>
        Promise.resolve({
          staende: [],
          fehler: {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Eine neuere Prüfung desselben Sets hat diese abgelöst.',
            feld: '',
            werte: {},
          },
        }),
    })
    const nachladen = zeige(
      lage({ geprueft: 0, vollstaendig: false, kartenGeprueft: 0 }),
    )

    await userEvent.click(screen.getByRole('button', { name: eigene.pruefen }))

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(nachladen).not.toHaveBeenCalled()
  })
})
