import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Kontextmenue } from './Kontextmenue'

// **Ein Menü, dessen Lage gleitet, rückt in einem Schritt ein** — unter
// wenig Bewegung bekommt jedes Element eine Übergangsdauer von 0,01 ms, und
// weil `transition-property` von Haus aus `all` ist, gleiten auch `top` und
// `left`: eine Messung im selben Bild liefert die alte Lage. Das Einrücken
// rechnete von der gemessenen Lage aus weiter und gab nach fünfzig Schritten
// React #185 — am nativen Fenster gefunden von `ap4zustaende` (Rechtsklick
// auf das Kartenbild, wenig Bewegung).
describe('die Lage, wenn die Messung hinterherhinkt', () => {
  it('rückt aus Größe und Fenster ein, nicht aus der gezeichneten Stelle', () => {
    const hoehe = window.innerHeight
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(
      function (this: HTMLElement) {
        // **Das Menü bleibt für die Messung, wo es zuerst stand.**
        return this.getAttribute('role') === 'menu'
          ? new DOMRect(400, hoehe - 100, 200, 300)
          : new DOMRect(0, 0, 800, hoehe)
      },
    )
    render(
      <Kontextmenue
        bezeichnung="Karte"
        eintraege={[{ schluessel: 'a', name: 'Öffnen', aufWahl: vi.fn() }]}
      >
        <p>Bild</p>
      </Kontextmenue>,
    )

    fireEvent.contextMenu(screen.getByText('Bild'), {
      clientX: 400,
      clientY: hoehe - 100,
    })

    expect(screen.getByRole('menu').style.top).toBe(
      `${String(hoehe - 8 - 300)}px`,
    )
    vi.restoreAllMocks()
  })
})
