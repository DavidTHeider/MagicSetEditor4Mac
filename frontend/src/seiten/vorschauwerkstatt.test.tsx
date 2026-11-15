// **Zoom, Überlagerungen und Problemliste an der gerenderten Vorschau** —
// Werkstück U5 von Phase 5.1
// (#1416).
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { vorschau, zeigeBildflaeche } from '../test/bildflaeche'

afterEach(() => {
  raeumeApi()
})

function mitBefund() {
  const eine = vorschau()

  return {
    vorschau: {
      ...eine.vorschau,
      befunde: ['Ebene 2 (name): die Schrift MPlantin ist nicht zu finden'],
    },
  }
}

describe('Die Vorschau im Arbeitsplatz', () => {
  it('rastert ab 150 % neu und lässt die Druckgröße zu, wenn die Vorlage eine Auflösung nennt', async () => {
    const Vorschau = vi.fn(() => Promise.resolve(vorschau()))
    zeigeBildflaeche({ Vorschau })

    await screen.findByRole('img', { name: 'Vorschau der Karte' })

    // **Der Maßstab ist eine Auswahl** (UXR-06, `massstab.test.tsx`) — bis
    // dahin eine Leiste aus sechs Knöpfen.
    const massstab = screen.getByRole('combobox', {
      name: 'Größe der Vorschau',
    })
    await userEvent.selectOptions(massstab, '200')

    await waitFor(() => {
      expect(Vorschau).toHaveBeenLastCalledWith(expect.any(String), 2)
    })

    await userEvent.selectOptions(massstab, 'Druckgröße')

    await waitFor(() => {
      expect(Vorschau).toHaveBeenLastCalledWith(expect.any(String), 1)
    })
    expect(massstab).toHaveDisplayValue('Druckgröße')
  })

  it('zeichnet Bildgrenzen nur über Bildfelder und die Diagnose nur über Felder mit Befund', async () => {
    zeigeBildflaeche({ Vorschau: () => Promise.resolve(mitBefund()) })

    await screen.findByRole('img', { name: 'Vorschau der Karte' })

    expect(document.querySelectorAll('[data-overlay]')).toHaveLength(0)

    await userEvent.click(screen.getByRole('checkbox', { name: 'Bildgrenzen' }))
    const bildlinien = document.querySelectorAll('[data-overlay="bildgrenzen"]')
    expect(bildlinien).toHaveLength(1)
    expect(bildlinien[0]).toHaveAttribute('title', 'kunst')

    await userEvent.click(screen.getByRole('checkbox', { name: 'Diagnose' }))
    const diagnose = document.querySelectorAll('[data-overlay="diagnose"]')
    expect(diagnose).toHaveLength(1)
    expect(diagnose[0]).toHaveAttribute('title', 'Diagnose: name')

    await userEvent.click(screen.getByRole('checkbox', { name: 'Beschnitt' }))
    expect(
      document.querySelectorAll('[data-overlay="beschnitt"]'),
    ).toHaveLength(1)
  })

  it('führt aus der Problemliste zum Feld des Befunds', async () => {
    zeigeBildflaeche({ Vorschau: () => Promise.resolve(mitBefund()) })

    // **Gewartet wird auf die Wirkung**, nicht auf die Liste: die steht schon
    // vor der Vorschau da, leer.
    await screen.findByText(/Ebene 2 \(name\)/)

    await userEvent.click(screen.getByRole('button', { name: 'Zum Feld name' }))

    expect(screen.getByLabelText('name')).toHaveFocus()
  })

  // **Das Bild in den Finder ziehen** — AP-12 (#2369), B-24: neben den
  // Überlagerungen steht ein ziehbares Bild der Karte; die Datei heißt nach
  // ihr (hier ohne Namen: „Karte“).
  it('bietet das Kartenbild zum Ziehen an, unter dem Namen der Karte', async () => {
    zeigeBildflaeche()

    await screen.findByRole('img', { name: 'Vorschau der Karte' })

    const griff = screen.getByText('Bild ziehen')
    expect(griff.parentElement?.querySelector('img')).toHaveAttribute(
      'src',
      '/vorschau/AAAA/Karte.png',
    )
  })
})
