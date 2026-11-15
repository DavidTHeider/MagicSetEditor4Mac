// **Was die Leiste oben verdeckt, darunter rollt der Fokus nicht** — gefunden
// am nativen Fenster in Phase 6.11 (#1672, Hüllenprobe `phase611`): mit der
// Leiste oben, „Sehr groß" und 1.024 × 700 nimmt sie drei Zeilen, und Tab
// setzte den Fokus auf „Nächste Karte" unter ihr (WCAG 2.4.11). Das Dokument
// rollt; sein `scroll-padding-top` sagt dem Rollen ins Bild, wie viel oben
// nicht zählt.
import { render } from '@testing-library/react'
import { useRef } from 'react'
import { afterEach, expect, it } from 'vitest'

import { useLeistenabstand } from './useLeistenabstand'

afterEach(() => {
  document.documentElement.style.scrollPaddingTop = ''
})

function Leiste({ waagerecht }: { waagerecht: boolean }) {
  const bereich = useRef<HTMLDivElement>(null)
  useLeistenabstand(bereich, waagerecht)

  return (
    <div
      ref={(element) => {
        bereich.current = element
        if (element !== null) {
          element.getBoundingClientRect = () => new DOMRect(0, 0, 1024, 372)
        }
      }}
    />
  )
}

it('hält oben so viel frei, wie die Leiste hoch ist', () => {
  render(<Leiste waagerecht />)

  expect(document.documentElement.style.scrollPaddingTop).toBe('372px')
})

it('hält nichts frei, wenn die Leiste an der Seite steht', () => {
  render(<Leiste waagerecht={false} />)

  expect(document.documentElement.style.scrollPaddingTop).toBe('')
})

it('gibt den Rand wieder her, wenn die Leiste an die Seite geht', () => {
  const { rerender } = render(<Leiste waagerecht />)
  rerender(<Leiste waagerecht={false} />)

  expect(document.documentElement.style.scrollPaddingTop).toBe('')
})
