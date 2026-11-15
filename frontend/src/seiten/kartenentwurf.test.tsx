// **Ein Entwurf überlebt fremde Änderungen** — U01 und U02
// (#2119).
//
// Der Editor trug den Verlaufsstand im `key`, und jede Änderung irgendwo —
// eine andere Karte, ein anderes Feld, ein anderes Programm — baute ihn neu
// auf: das Eingabefeld, in dem jemand tippte, verschwand samt Entwurf und
// Fokus. Und schrieb ein anderer **dasselbe** Feld, ging der Entwurf still
// verloren oder überschrieb beim Verlassen den fremden Wert.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Umschlag } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'

const texte = kataloge.de.bausteine.entwurf

afterEach(() => {
  raeumeApi()
})

// bestand ist, was im Bestand steht — ein anderes Programm ändert ihn.
function mitBestand(bestand: Record<string, string>, scheitert = false) {
  const gesetzt = vi.fn<
    (
      setID: string,
      karteID: string,
      feld: string,
      wert: string,
    ) => Promise<Umschlag>
  >(() =>
    Promise.resolve({
      fehler: scheitert
        ? {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Dieses Feld hat sich inzwischen geändert.',
          }
        : null,
    }),
  )

  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 2 }],
      }),
    Karte: () =>
      Promise.resolve({
        karte: {
          id: '5',
          nummer: 0,
          werte: Object.entries(bestand).map(([feld, wert]) => ({
            feld,
            wert,
          })),
        },
        fehler: null,
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
    Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    KarteVerweis: () =>
      Promise.resolve({ verweis: null, befunde: [], fehler: null }),
    FeldSetzen: gesetzt,
  })

  return gesetzt
}

function seite(stand: number) {
  return (
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID="5"
        verlaufsstand={stand}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>
  )
}

describe('ein Entwurf im Kartenformular', () => {
  it('bleibt mit Fokus stehen, wenn ein anderes Feld fremd geändert wird', async () => {
    const bestand = { name: 'Feuerball', typ: 'Zauber' }
    mitBestand(bestand)

    const { rerender } = render(seite(0))

    const name = await screen.findByRole('textbox', { name: 'name' })
    await userEvent.type(name, ' groß')

    bestand.typ = 'Hexerei'
    rerender(seite(1))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'typ' })).toHaveValue(
        'Hexerei',
      )
    })

    expect(screen.getByRole('textbox', { name: 'name' })).toBe(name)
    expect(name).toHaveValue('Feuerball groß')
    expect(document.activeElement).toBe(name)
  })

  it('sagt, wenn dasselbe Feld fremd geändert wurde, und schreibt nicht still darüber', async () => {
    const bestand = { name: 'Feuerball', typ: 'Zauber' }
    const gesetzt = mitBestand(bestand)

    const { rerender } = render(seite(0))

    const name = await screen.findByRole('textbox', { name: 'name' })
    await userEvent.type(name, ' groß')

    bestand.name = 'Eisball'
    rerender(seite(1))

    expect(
      await screen.findByText(
        einsetzen(texte.fremdGeaendert, { wert: 'Eisball' }),
      ),
    ).toBeInTheDocument()
    expect(name).toHaveValue('Feuerball groß')

    // **Verlassen schreibt nicht** — der Wert im Bestand ist ein anderer
    // als der, von dem der Entwurf ausging.
    await act(async () => {
      name.blur()
      await Promise.resolve()
    })

    expect(gesetzt).not.toHaveBeenCalled()

    // **Meins** schreibt den Entwurf ausdrücklich.
    await userEvent.click(screen.getByRole('button', { name: texte.behalten }))

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledWith(
        '1',
        '5',
        'name',
        'Feuerball groß',
        expect.any(Array),
      )
    })
  })

  it('übernimmt auf Wunsch den fremden Wert und verwirft den Entwurf', async () => {
    const bestand = { name: 'Feuerball', typ: 'Zauber' }
    const gesetzt = mitBestand(bestand)

    const { rerender } = render(seite(0))

    const name = await screen.findByRole('textbox', { name: 'name' })
    await userEvent.type(name, ' groß')

    bestand.name = 'Eisball'
    rerender(seite(1))

    await userEvent.click(
      await screen.findByRole('button', { name: texte.nehmen }),
    )

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
        'Eisball',
      )
    })
    expect(gesetzt).not.toHaveBeenCalled()
  })

  it('behält den Entwurf, wenn sein Speichern scheitert, und schreibt ihn beim nächsten Verlassen', async () => {
    const bestand = { name: 'Feuerball', typ: 'Zauber' }
    const gesetzt = mitBestand(bestand, true)

    const { rerender } = render(seite(0))

    const name = await screen.findByRole('textbox', { name: 'name' })
    await userEvent.type(name, ' groß')

    bestand.name = 'Eisball'
    rerender(seite(1))

    await userEvent.click(
      await screen.findByRole('button', { name: texte.behalten }),
    )

    // **Die Wirkung abwarten, nicht den Ruf**: die Meldung steht am Feld.
    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).toHaveAttribute(
        'aria-invalid',
        'true',
      )
    })
    expect(gesetzt).toHaveBeenCalledTimes(1)

    // **Der Text steht noch da**, und kein Konflikt mehr: der fremde Wert
    // ist gesehen, und wer weitertippt, geht von ihm aus.
    expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
      'Feuerball groß',
    )
    expect(
      screen.queryByRole('button', { name: texte.behalten }),
    ).not.toBeInTheDocument()

    await userEvent.type(screen.getByRole('textbox', { name: 'name' }), '!')
    await userEvent.tab()

    await waitFor(() => {
      expect(gesetzt).toHaveBeenLastCalledWith(
        '1',
        '5',
        'name',
        'Feuerball groß!',
        expect.any(Array),
      )
    })
  })

  it('sieht keinen Konflikt, wenn der fremde Wert der Entwurf ist', async () => {
    const bestand = { name: 'Feuerball', typ: 'Zauber' }
    const gesetzt = mitBestand(bestand)

    const { rerender } = render(seite(0))

    const name = await screen.findByRole('textbox', { name: 'name' })
    await userEvent.type(name, ' groß')

    bestand.name = 'Feuerball groß'
    bestand.typ = 'Hexerei'
    rerender(seite(1))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'typ' })).toHaveValue(
        'Hexerei',
      )
    })
    expect(
      screen.queryByRole('button', { name: texte.behalten }),
    ).not.toBeInTheDocument()

    // **Und Verlassen schreibt nichts** — es steht schon da.
    await userEvent.tab()
    expect(gesetzt).not.toHaveBeenCalled()
  })
})
