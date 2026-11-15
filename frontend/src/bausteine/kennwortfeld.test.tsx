// **Das Kennwortfeld für sich** — Werkstück 6 von Phase 4.12
// (#1351).
//
// **Es wird hier für sich geprüft und nicht nur über die Seite**: die Seite
// gibt ihm immer einen Hinweis, und was das Feld **ohne** einen tut, sähe dort
// niemand.
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Kennwortfeld } from './Kennwortfeld'
import { Sprachraum } from '../sprache/kontext'

function zeige(teile: Partial<Parameters<typeof Kennwortfeld>[0]> = {}) {
  render(
    <Sprachraum sprache="de">
      <Kennwortfeld
        beschriftung="Kennwort"
        wert=""
        geaendert={vi.fn()}
        {...teile}
      />
    </Sprachraum>,
  )
}

describe('das Kennwortfeld', () => {
  // **Ohne Hinweis und ohne Befund beschreibt es nichts** — ein
  // `aria-describedby`, das auf einen Knoten zeigt, den es nicht gibt, ist
  // schlimmer als keines: eine Vorlesehilfe liest dann gar nichts vor.
  it('verweist auf nichts, solange es nichts zu sagen gibt', () => {
    zeige()

    expect(screen.getByLabelText('Kennwort')).not.toHaveAttribute(
      'aria-describedby',
    )
  })

  // **Der Hinweis steht da, bevor ein Fehler kommt** (§6).
  it('beschreibt sich über seinen Hinweis', () => {
    zeige({ hinweis: 'Mindestens zwölf Zeichen' })

    expect(screen.getByLabelText('Kennwort')).toHaveAccessibleDescription(
      'Mindestens zwölf Zeichen',
    )
  })

  // **Und ein Befund verdrängt ihn** — zwei Sätze übereinander, von denen
  // einer nicht mehr gilt, sind schlimmer als einer.
  it('ersetzt den Hinweis durch den Befund', () => {
    zeige({ hinweis: 'Mindestens zwölf Zeichen', befund: 'Zu kurz.' })

    const feld = screen.getByLabelText('Kennwort')

    expect(feld).toHaveAccessibleDescription('Zu kurz.')
    expect(feld).toHaveAttribute('aria-invalid', 'true')
    expect(
      screen.queryByText('Mindestens zwölf Zeichen'),
    ).not.toBeInTheDocument()
  })

  // **Angehalten heißt angehalten** — während ein Vorgang läuft, tippt
  // niemand weiter (§7, Regel 4).
  it('lässt sich anhalten', () => {
    zeige({ angehalten: true })

    expect(screen.getByLabelText('Kennwort')).toBeDisabled()
  })

  // **Der Browser soll hier nicht ausfüllen.** Dasselbe Feld dient dem
  // Anlegen und dem Anmelden; ein automatisch eingesetztes Kennwort legte
  // sonst ein Konto mit einem an, das niemand dafür gewählt hat.
  it('lässt den Browser nicht ausfüllen', () => {
    zeige()

    expect(screen.getByLabelText('Kennwort')).toHaveAttribute(
      'autocomplete',
      'new-password',
    )
  })
})
