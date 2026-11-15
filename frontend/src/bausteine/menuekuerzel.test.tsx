// **Ein Menüeintrag nennt sein Kürzel** — UXR-06
// (#1914).
//
// # Wozu
//
// Kopieren, Ausschneiden und Einfügen standen als Knöpfe in einem Kasten über
// der Karte, und „Kopieren" kündigte ⌘C an — als Tooltip und über
// `aria-keyshortcuts` (`Aktionsknopf`, `mitKuerzel`). Seit sie im Menü
// „Diese Karte" stehen, sagt es der Eintrag: rechts das Zeichen, wie in
// jedem Menü von macOS, und für ein Vorleseprogramm dieselbe Eigenschaft.
//
// **Dieselbe Zusage wie beim Knopf**: angekündigt wird nur, was der Aufrufer
// verdrahtet hat — `aktionskatalog.test.ts` hält die Stelle gegen den Lauf,
// der die Taste drückt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Menue } from './Menue'

describe('Ein Menüeintrag mit Kürzel', () => {
  it('zeigt das Zeichen der Aktion und nennt die Tasten', async () => {
    render(
      <Menue
        bezeichnung="Diese Karte"
        eintraege={[
          {
            schluessel: 'kopieren',
            name: 'Kopieren (1)',
            aktion: 'kopieren',
            mitKuerzel: true,
            aufWahl: vi.fn(),
          },
          { schluessel: 'einfuegen', name: 'Einfügen', aufWahl: vi.fn() },
        ]}
      />,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Diese Karte' }))

    const kopieren = screen.getByRole('menuitem', { name: /Kopieren \(1\)/ })
    expect(kopieren).toHaveAttribute('aria-keyshortcuts', 'Meta+C')
    expect(kopieren).toHaveTextContent('⌘C')

    // **Ohne Ansage keine Eigenschaft** — ein Eintrag behauptet kein Kürzel,
    // das niemand verdrahtet hat.
    expect(
      screen.getByRole('menuitem', { name: 'Einfügen' }),
    ).not.toHaveAttribute('aria-keyshortcuts')
  })

  // **Ein Menü am rechten Rand klappt nach links auf** (UXR-06) — seit AP-6
  // (#2363) gemessen am Knopf und geprüft in `menuelage.test.tsx`.

  // **Und das Kürzel eines Objektmenüs** — AP-7 (#2364): ⌘⌫ und ⇟ stehen im
  // Menü „Karte“, wie die Leiste sie trägt.
  it('zeigt das Kürzel eines Objekteintrags wie macOS', async () => {
    render(
      <Menue
        bezeichnung="Diese Karte"
        eintraege={[
          {
            schluessel: 'loeschen',
            name: 'Löschen …',
            kuerzel: 'befehl+backspace',
            aufWahl: vi.fn(),
          },
          {
            schluessel: 'naechste',
            name: 'Nächste Karte',
            kuerzel: 'bild ab',
            aufWahl: vi.fn(),
          },
        ]}
      />,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Diese Karte' }))

    const loeschen = screen.getByRole('menuitem', { name: /Löschen/ })
    expect(loeschen).toHaveAttribute('aria-keyshortcuts', 'Meta+Backspace')
    expect(loeschen).toHaveTextContent('⌘⌫')
    expect(
      screen.getByRole('menuitem', { name: /Nächste Karte/ }),
    ).toHaveAttribute('aria-keyshortcuts', 'PageDown')
  })
})
