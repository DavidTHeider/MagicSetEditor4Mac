// **Ein Überbau, der aufbewahrt wird, greift beim Einblenden nicht erneut
// nach dem Fokus** — angestrichen vom Zweitblick über Phase 6.3 (G11):
// `Entscheidung`, `Schublade`, `Popover` und `SetNeu` merkten sich beim
// Erscheinen den Auslöser und setzten den Fokus — und `<Activity>` lässt
// diese Effekte beim Einblenden der Seite neu laufen. Dann zog der Überbau
// den Fokus aus der Leiste, und sein Rückkehrziel war die Leiste.
import { render, screen } from '@testing-library/react'
import { Activity, useRef } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { useRueckkehr } from './useRueckkehr'

afterEach(() => {
  document.body.replaceChildren()
})

function Ueberbau() {
  const knopf = useRef<HTMLButtonElement>(null)
  useRueckkehr(knopf, () => knopf.current?.focus())

  return (
    <button ref={knopf} type="button">
      Im Überbau
    </button>
  )
}

function Seite({ sichtbar, offen }: { sichtbar: boolean; offen: boolean }) {
  return (
    <>
      <button type="button">Leiste</button>
      <Activity mode={sichtbar ? 'visible' : 'hidden'}>
        <button type="button">Auslöser</button>
        {offen && <Ueberbau />}
      </Activity>
    </>
  )
}

describe('useRueckkehr', () => {
  it('fokussiert beim Erscheinen und gibt beim Verschwinden zurück', () => {
    const { rerender } = render(<Seite sichtbar offen={false} />)
    screen.getByRole('button', { name: 'Auslöser' }).focus()

    rerender(<Seite sichtbar offen />)
    expect(screen.getByRole('button', { name: 'Im Überbau' })).toHaveFocus()

    rerender(<Seite sichtbar offen={false} />)
    expect(screen.getByRole('button', { name: 'Auslöser' })).toHaveFocus()
  })

  it('greift beim Wiedereinblenden nicht nach dem Fokus', () => {
    const { rerender } = render(<Seite sichtbar offen={false} />)
    screen.getByRole('button', { name: 'Auslöser' }).focus()
    rerender(<Seite sichtbar offen />)

    rerender(<Seite sichtbar={false} offen />)
    const leiste = screen.getByRole('button', { name: 'Leiste' })
    leiste.focus()
    rerender(<Seite sichtbar offen />)

    expect(leiste).toHaveFocus()

    // **Und das Rückkehrziel bleibt der Auslöser**, nicht die Leiste.
    rerender(<Seite sichtbar offen={false} />)
    expect(screen.getByRole('button', { name: 'Auslöser' })).toHaveFocus()
  })

  // **Wird der Überbau selbst verborgen, während der Fokus in ihm steht, geht
  // er zurück** — eine Schublade des Sets steht in ihrer eigenen
  // `<Activity>`, und Zuklappen ist dort Verbergen.
  it('gibt zurück, wenn der Überbau mit dem Fokus darin verborgen wird', () => {
    function Schubladenseite({ offen }: { offen: boolean }) {
      return (
        <>
          <button type="button">Auslöser</button>
          <Activity mode={offen ? 'visible' : 'hidden'}>
            <Ueberbau />
          </Activity>
        </>
      )
    }

    const { rerender } = render(<Schubladenseite offen={false} />)
    screen.getByRole('button', { name: 'Auslöser' }).focus()
    rerender(<Schubladenseite offen />)
    expect(screen.getByRole('button', { name: 'Im Überbau' })).toHaveFocus()

    rerender(<Schubladenseite offen={false} />)

    expect(screen.getByRole('button', { name: 'Auslöser' })).toHaveFocus()
  })
})
