// **Was ein Profil vom Menschen erbt, im Fenster** — Werkstück 9 von
// Phase 4.12
// (#1351),
// Punkt 9.
//
// # Die Zusage in einem Satz
//
// **Die Auswahlliste zeigt die Wahl, die Anwendung richtet sich nach dem
// Wirksamen.** Wer nur eines von beiden hätte, zeigte entweder eine Wahl, die
// niemand getroffen hat, oder eine Anwendung, die anders aussieht, als die
// Liste sagt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, raeumeApi } from './test/api'

afterEach(() => {
  raeumeApi()
  document.documentElement.removeAttribute('data-erscheinungsbild')
})

// erbendeAntwort gibt ein Profil, das erbt — und einen Menschen, der gewählt
// hat.
function erbendeAntwort() {
  return () =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: 'sprache',
          wert: 'erben',
          wirksam: 'en',
          // **Die Liste, wie Go sie schickt**: mit neun Werten ist die
          // Sprache eine Auswahlliste (UXR-F2, #2008).
          werte: ['erben', 'de', 'en', 'es', 'fr', 'it', 'ja', 'ru', 'zh'],
          bereich: 'anwendung',
        },
        {
          schluessel: 'erscheinungsbild',
          wert: 'erben',
          wirksam: 'dunkel',
          werte: ['erben', 'system', 'hell', 'dunkel'],
          bereich: 'anwendung',
        },
      ],
    })
}

describe('Die Übersteuerung des Profils', () => {
  it('richtet die Anwendung nach dem wirksamen Wert', async () => {
    belegeApi({ Einstellungen: erbendeAntwort() })

    render(<App />)

    // **Die Sprache ist Englisch**, obwohl das Profil nichts gewählt hat —
    // der Mensch hat.
    expect(
      await screen.findByRole('button', {
        name: kataloge.en.seiten.einstellungen.titel,
      }),
    ).toBeInTheDocument()

    // **Und das Erscheinungsbild steht am Wurzelelement** — gewartet wird auf
    // die Wirkung des Effekts, nicht auf den Knoten (§3).
    await waitFor(() => {
      expect(document.documentElement).toHaveAttribute(
        'data-erscheinungsbild',
        'dunkel',
      )
    })
  })

  // **Und die Auswahlliste steht trotzdem auf „wie beim Nutzer".**
  //
  // **Das ist die andere Hälfte**: stünde dort `en`, hätte das Profil eine
  // Wahl getroffen, die niemand getroffen hat — und der nächste Klick auf
  // eine andere Einstellung schriebe sie fest.
  it('zeigt in der Auswahlliste die Wahl und nicht die Wirkung', async () => {
    belegeApi({ Einstellungen: erbendeAntwort() })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.en.seiten.einstellungen.titel,
      }),
    )

    const wahl = await screen.findByLabelText(
      kataloge.en.seiten.einstellungen.namen.sprache,
    )

    expect(wahl).toHaveValue('erben')
    expect(wahl).toHaveDisplayValue(
      kataloge.en.seiten.einstellungen.werte['sprache.erben'],
    )
  })
})
