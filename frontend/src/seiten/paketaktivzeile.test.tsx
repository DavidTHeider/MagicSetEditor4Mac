// **Das Menü steht an der aktiven Zeile, nicht an jeder** — UXR-17
// (#1925).
//
// Gemessen am 2026-10-02 an der Hüllenprobe `uirevue`: die Paketseite zeichnete
// **671 Bedienelemente**, 22 davon im Bild — fast alle davon ein Menüknopf je
// Paketzeile. Das Prüfmaß aus `docs/UIREVIEW.md` lässt 200 zu (R6) und 20 im
// Bild (R2). Der Knopf steht jetzt an der Zeile, die gewählt ist oder die
// Tabstelle trägt; erreichbar bleibt er auf denselben Wegen — mit der Maus
// über die Wahl, mit der Tastatur über → aus der Zeile.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { anlegbar, katalog, zeige } from '../test/paketliste'

afterEach(() => {
  raeumeApi()
})

// zwei gibt einen Katalog mit zwei Paketen ohne Gruppe — die erste Zeile ist
// damit die Gruppe „Ohne Gruppe", und keine Paketzeile trägt anfangs die
// Tabstelle.
function zwei() {
  const antwort = katalog()

  antwort.pakete.push(
    ...katalog().pakete.map((paket) => ({
      ...paket,
      name: 'probe-bild',
      anzeigename: 'probe-bild',
      pfad: '/pakete/probe-bild.mse-style',
    })),
  )

  return antwort
}

const menues = () => screen.queryAllByRole('button', { name: /^Handlungen — / })

describe('Das Menü einer Paketzeile', () => {
  it('steht nicht an jeder Zeile', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(zwei()),
    })
    zeige()

    await screen.findByText('probe-bild')

    expect(menues()).toHaveLength(0)
  })

  it('steht an der gewählten Zeile', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(zwei()),
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('row', { name: 'probe-bild' }),
    )

    expect(menues().map((knopf) => knopf.getAttribute('aria-label'))).toEqual([
      'Handlungen — probe-bild',
    ])
  })

  // **Mit der Tastatur wie vorher**: ↓ in die Zeile, → auf ihren Knopf —
  // ohne erst zu wählen. Die Zeile mit der Tabstelle zeichnet ihn.
  it('steht an der Zeile mit der Tabstelle und ist mit → erreichbar', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(zwei()),
    })
    zeige()

    await screen.findByText('probe-bild')

    const zeilen = screen
      .getAllByRole('row')
      .filter((zeile) => zeile.closest('tbody') !== null)
    zeilen[0]?.focus()

    await userEvent.keyboard('{ArrowDown}')
    await userEvent.keyboard('{ArrowRight}')

    expect(document.activeElement).toBe(
      screen.getByRole('button', { name: 'Handlungen — probe-bild' }),
    )
  })

  // **Und die Gruppe schaltet an der Zeile** (Paketgruppenzeile.tsx): das
  // Gitter mit `aria-expanded` an der Zeile und dem Menü an der gewählten
  // hält WCAG 2.2 AA — `barrierefreiheit.test.tsx` sieht es nicht, denn sein
  // Probebestand hat keine Pakete.
  it('hält WCAG 2.2 AA mit Gruppe und aktiver Zeile', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(zwei()),
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('row', { name: 'probe-bild' }),
    )

    expect(
      screen.getByRole('row', { name: 'Gruppe Ohne Gruppe' }),
    ).toHaveAttribute('aria-expanded', 'true')
    expect(await verstoesse(document.body)).toEqual([])
  })
})
