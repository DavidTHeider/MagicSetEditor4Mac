// **„KI einrichten" öffnet den KI-Bereich** — AI-01.AK-04 (#1742).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { KIEinrichtungsverweis } from './KIEinrichtungsverweis'
import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

function Offen() {
  const anschluss = useAnschluss()
  const offen = anschluss.offen

  return (
    <output>{offen?.art === 'einstellungsgruppe' ? offen.gruppe : ''}</output>
  )
}

it('übergibt die Gruppe ki an die Einstellungen', async () => {
  const wechseln = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="karten" wechseln={wechseln}>
        <KIEinrichtungsverweis />
        <Offen />
      </Anschlussraum>
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.kieinrichtung.oeffnen,
    }),
  )

  expect(wechseln).toHaveBeenCalledWith('einstellungen')
  expect(screen.getByRole('status')).toHaveTextContent('ki')
})
