import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Meldungsband } from './Meldungsband'

function zeige(meldung: Parameters<typeof Meldungsband>[0]['meldung']) {
  const schliessen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Meldungsband meldung={meldung} schliessen={schliessen} />
    </Sprachraum>,
  )

  return schliessen
}

describe('Meldungsband', () => {
  // **Der Bereich steht auch leer da.** Ein Vorleser meldet nur, was sich in
  // einem Bereich ändert, den er schon kennt; einer, der erst mit der Meldung
  // entsteht, bleibt stumm — und genau darum geht es bei #180.
  it('hält seinen Bereich auch ohne Meldung bereit', () => {
    zeige(undefined)

    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('zeigt eine Meldung', () => {
    zeige({ text: 'Gesichert nach /tmp/x', art: 'gut' })

    expect(screen.getByRole('status')).toHaveTextContent(
      'Gesichert nach /tmp/x',
    )
  })

  it('lässt sich schließen', async () => {
    const schliessen = zeige({ text: 'Das ging schief', art: 'schlecht' })

    await userEvent.click(screen.getByRole('button', { name: 'Schließen' }))

    expect(schliessen).toHaveBeenCalledOnce()
  })
})
