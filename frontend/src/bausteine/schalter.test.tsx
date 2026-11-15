// **Der Schalter einer Einstellung** — UXR-F2
// (#2008),
// herausgelöst aus dem `Bereichsschalter` (UXR-19, #1927): derselbe
// Baustein für jede Einstellung mit zwei Werten, die an oder aus sind.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Schalter } from './Schalter'

function Probe({
  an,
  angehalten = false,
  aendern,
}: {
  an: boolean
  angehalten?: boolean
  aendern: (an: boolean) => void
}) {
  return (
    <>
      <input aria-label="davor" />
      {/* eslint-disable-next-line jsx-a11y/label-has-associated-control -- das Kästchen ist der Baustein, die Regel kennt ihn nicht */}
      <label>
        <Schalter
          schluessel="ersteschritte"
          an={an}
          angehalten={angehalten}
          aendern={aendern}
        />
        Einführung beim Start
      </label>
    </>
  )
}

describe('der Schalter einer Einstellung', () => {
  it('ist ein Schalter mit dem Stand und dem Schlüssel', () => {
    render(<Probe an aendern={vi.fn()} />)

    const schalter = screen.getByRole('switch', {
      name: 'Einführung beim Start',
    })

    expect(schalter).toBeChecked()
    expect(schalter).toHaveAttribute('data-einstellung', 'ersteschritte')
  })

  it('meldet, wohin er umgelegt wird', async () => {
    const aendern = vi.fn()
    const { rerender } = render(<Probe an aendern={aendern} />)

    await userEvent.click(screen.getByRole('switch'))
    expect(aendern).toHaveBeenLastCalledWith(false)

    rerender(<Probe an={false} aendern={aendern} />)
    await userEvent.click(screen.getByRole('switch'))
    expect(aendern).toHaveBeenLastCalledWith(true)
  })

  // **Mit der Tastatur wie jeder Schalter** — die Leertaste legt um (§6).
  it('legt mit der Leertaste um', async () => {
    const aendern = vi.fn()
    render(<Probe an aendern={aendern} />)

    screen.getByRole('switch').focus()
    await userEvent.keyboard(' ')

    expect(aendern).toHaveBeenCalledExactlyOnceWith(false)
  })

  it('nimmt nichts an, solange geschrieben wird', () => {
    render(<Probe an angehalten aendern={vi.fn()} />)

    expect(screen.getByRole('switch')).toBeDisabled()
  })

  // **Der Fokus bleibt auf dem Schalter** — #1990: ein fokussiertes
  // Element, das `disabled` wird, verliert ihn an den Seitenanfang (WebKit),
  // und wer mit der Leertaste umlegt, stünde danach am Anfang der Seite.
  // jsdom tut das nicht; der Test stellt es mit `blur()` nach.
  it('behält den Fokus über das Schreiben hinweg', async () => {
    const { rerender } = render(<Probe an aendern={vi.fn()} />)
    const schalter = screen.getByRole('switch')

    schalter.focus()
    await userEvent.keyboard(' ')

    schalter.blur()
    rerender(<Probe an={false} angehalten aendern={vi.fn()} />)
    expect(document.body).toHaveFocus()

    rerender(<Probe an={false} aendern={vi.fn()} />)

    expect(schalter).toHaveFocus()
  })

  // **Ein Klick mit dem Zeiger gibt in WebKit keinen Fokus** — danach soll
  // der Schalter ihn auch nicht an sich ziehen.
  it('nimmt keinen Fokus, den er beim Umlegen nicht hatte', () => {
    const { rerender } = render(<Probe an aendern={vi.fn()} />)
    const schalter = screen.getByRole('switch')

    fireEvent.click(schalter)

    rerender(<Probe an={false} angehalten aendern={vi.fn()} />)
    rerender(<Probe an={false} aendern={vi.fn()} />)

    expect(schalter).not.toHaveFocus()
  })
})
