// **Notiz und noch nicht gesetzte Marke überleben fremde Änderungen** — U04
// und U05 (#2121).
//
// Dieselbe Klasse wie U01: `Kartenspalte` trug den Verlaufsstand im `key`
// von Notizkasten und Markenkasten, und jede Änderung irgendwo baute beide
// neu auf — der Text, den jemand tippte, war weg.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Umschlag } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'

const entwurf = kataloge.de.bausteine.entwurf
const eigene = kataloge.de.seiten.karten

afterEach(() => {
  raeumeApi()
})

interface Lage {
  werte: Record<string, string>
  marken: string[]
}

// lage ist, was im Bestand steht — ein anderes Programm ändert sie.
function mitLage(lage: Lage) {
  const gesetzt = vi.fn<
    (
      setID: string,
      karteID: string,
      feld: string,
      wert: string,
    ) => Promise<Umschlag>
  >(() => Promise.resolve({ fehler: null }))
  const markiert = vi.fn<
    (setID: string, karteID: string, marke: string) => Promise<Umschlag>
  >(() => Promise.resolve({ fehler: null }))
  const werte = () =>
    Object.entries(lage.werte).map(([feld, wert]) => ({ feld, wert }))

  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 2 }],
      }),
    Karte: () =>
      Promise.resolve({ karte: { id: '5', nummer: 0, werte: werte() } }),
    Kartenlistenaufbau: () =>
      Promise.resolve({
        spalten: [
          {
            feld: 'name',
            aufschrift: '',
            sichtbar: true,
            ausrichtung: 'links',
          },
        ],
        zeilen: [
          {
            karteID: '5',
            nummer: 1,
            werte: { name: 'Feuerball' },
            marken: [...lage.marken],
          },
        ],
        sortierfeld: '',
        aufsteigend: true,
        ohneVorlage: false,
        fehler: null,
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
    Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    KarteVerweis: () =>
      Promise.resolve({ verweis: null, befunde: [], fehler: null }),
    FeldSetzen: gesetzt,
    MarkeSetzen: markiert,
  })

  return { gesetzt, markiert }
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

describe('die Notiz einer Karte', () => {
  it('bleibt mit Fokus stehen, wenn anderswo etwas geändert wird', async () => {
    const lage = { werte: { name: 'Feuerball', notes: 'alt' }, marken: [] }
    mitLage(lage)

    const { rerender } = render(seite(0))

    const notiz = await screen.findByLabelText(eigene.notizen.ueberschrift)
    await waitFor(() => {
      expect(notiz).toHaveValue('alt')
    })
    await userEvent.type(notiz, ' und neu')

    lage.werte.name = 'Eisball'
    rerender(seite(1))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
        'Eisball',
      )
    })
    expect(screen.getByLabelText(eigene.notizen.ueberschrift)).toBe(notiz)
    expect(notiz).toHaveValue('alt und neu')
    expect(document.activeElement).toBe(notiz)
  })

  it('sagt, wenn die Notiz fremd geändert wurde, und schreibt nicht still darüber', async () => {
    const lage = { werte: { name: 'Feuerball', notes: 'alt' }, marken: [] }
    const { gesetzt } = mitLage(lage)

    const { rerender } = render(seite(0))

    const notiz = await screen.findByLabelText(eigene.notizen.ueberschrift)
    await waitFor(() => {
      expect(notiz).toHaveValue('alt')
    })
    await userEvent.type(notiz, ' und neu')

    lage.werte.notes = 'fremd'
    rerender(seite(1))

    expect(
      await screen.findByText(
        einsetzen(entwurf.fremdGeaendert, { wert: 'fremd' }),
      ),
    ).toBeInTheDocument()

    await act(async () => {
      notiz.blur()
      await Promise.resolve()
    })
    expect(gesetzt).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', { name: entwurf.behalten }),
    )

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledWith(
        '1',
        '5',
        'notes',
        'alt und neu',
        expect.any(Array),
      )
    })
  })
})

describe('eine noch nicht gesetzte Marke', () => {
  it('bleibt mit Fokus stehen, wenn anderswo etwas geändert wird', async () => {
    const lage = { werte: { name: 'Feuerball' }, marken: [] }
    mitLage(lage)

    const { rerender } = render(seite(0))

    const feld = await screen.findByLabelText(eigene.marken.setzen)
    await userEvent.type(feld, 'prüf')

    lage.werte.name = 'Eisball'
    rerender(seite(1))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
        'Eisball',
      )
    })
    expect(screen.getByLabelText(eigene.marken.setzen)).toBe(feld)
    expect(feld).toHaveValue('prüf')
    expect(document.activeElement).toBe(feld)
  })

  it('sagt, wenn ein anderer dieselbe Marke schon gesetzt hat, und setzt sie nicht noch einmal', async () => {
    const lage: Lage = { werte: { name: 'Feuerball' }, marken: [] }
    const { markiert } = mitLage(lage)

    const { rerender } = render(seite(0))

    const feld = await screen.findByLabelText(eigene.marken.setzen)
    await userEvent.type(feld, 'prüfen')

    lage.marken = ['prüfen']
    rerender(seite(1))

    expect(
      await screen.findByText(
        einsetzen(eigene.marken.schonGesetzt, { marke: 'prüfen' }),
      ),
    ).toBeInTheDocument()
    expect(feld).toHaveValue('prüfen')
    expect(
      screen.getByRole('button', { name: eigene.marken.setzen }),
    ).toBeDisabled()
    expect(markiert).not.toHaveBeenCalled()
  })
})
