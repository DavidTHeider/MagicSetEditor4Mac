// **Ein Feld: Beschriftung, Bedienelement, Hinweis, Befund** — Werkstück U2
// von Phase 5.1 (#1413).
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Feld } from './Feld'

describe('Feld', () => {
  it('beschriftet sein Element über label und htmlFor', () => {
    render(
      <Feld beschriftung="Name">
        {(verbindung) => <input {...verbindung} type="text" />}
      </Feld>,
    )

    const feld = screen.getByRole('textbox', { name: 'Name' })
    expect(feld).not.toHaveAttribute('aria-describedby')
    expect(feld).toHaveAttribute('aria-invalid', 'false')
  })

  it('beschreibt das Element mit seinem Hinweis', () => {
    render(
      <Feld
        beschriftung="Spiel"
        hinweis="Der Schlüssel des Spielpakets, etwa magic."
      >
        {(verbindung) => <input {...verbindung} type="text" />}
      </Feld>,
    )

    expect(
      screen.getByRole('textbox', { name: 'Spiel' }),
    ).toHaveAccessibleDescription('Der Schlüssel des Spielpakets, etwa magic.')
  })

  // **Der Befund ist ein Alarm am Feld** (§6), und das Element sagt, dass es
  // ungültig ist.
  it('meldet den Befund am Feld und markiert das Element als ungültig', () => {
    render(
      <Feld
        beschriftung="Name"
        hinweis="Eindeutig je Nutzer"
        befund="Name braucht einen Wert."
      >
        {(verbindung) => <input {...verbindung} type="text" />}
      </Feld>,
    )

    const feld = screen.getByRole('textbox', { name: 'Name' })
    expect(feld).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Name braucht einen Wert.',
    )

    // **Befund zuerst, Hinweis danach** — beides, nicht eines.
    expect(feld).toHaveAccessibleDescription(
      'Name braucht einen Wert. Eindeutig je Nutzer',
    )
  })

  // **Das Ziel der Beschreibung trägt keine Rolle, die Hülle darum den
  // Alarm** (#1989): zeigt `aria-describedby` auf ein Element mit
  // `role="alert"`, liest VoiceOver in WKWebView die Beschreibung nicht vor —
  // gemessen am 2026-10-01 in der Abnahme-VM.
  it('beschreibt mit einem Ziel ohne Rolle, das in einem Alarm steht', () => {
    render(
      <Feld beschriftung="Name" befund="Name braucht einen Wert.">
        {(verbindung) => <input {...verbindung} type="text" />}
      </Feld>,
    )

    const kennung = screen
      .getByRole('textbox', { name: 'Name' })
      .getAttribute('aria-describedby')
    const ziel = document.getElementById(kennung ?? '')

    expect(ziel).toHaveTextContent('Name braucht einen Wert.')
    expect(ziel).not.toHaveAttribute('role')
    expect(ziel?.parentElement).toHaveAttribute('role', 'alert')
  })

  it('verbindet auch ein Auswahlfeld', () => {
    render(
      <Feld beschriftung="Stil">
        {(verbindung) => (
          <select {...verbindung}>
            <option value="m15">M15</option>
          </select>
        )}
      </Feld>,
    )

    expect(screen.getByRole('combobox', { name: 'Stil' })).toBeInTheDocument()
  })
})
