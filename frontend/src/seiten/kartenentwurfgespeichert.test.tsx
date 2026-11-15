// **Der Editor bestätigt den Entwurf mit dem Wert, den das Schreiben
// gespeichert hat** — #2535 (Nachlauf über das Gesamtreview über 0.9.0,
// Befund 1).
//
// Bis dahin galt ein geglücktes Schreiben als erledigt, sobald der Bestand
// sich änderte. Schrieb ein anderes Programm zwischen Speichern und
// Nachladen, nahm der Editor dessen Wert als den eigenen, und niemand sah den
// Konflikt. Die Vorrichtung ist die aus `kartenentwurfgerechnet.test.tsx`;
// ihr `FeldSetzen` sagt hier, was es gespeichert hat.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'

const texte = kataloge.de.bausteine.entwurf

afterEach(() => {
  raeumeApi()
})

// mitBestand rechnet wie ein Wertskript und lässt danach `fremd` ins Feld
// schreiben, wenn es gesetzt ist — ein anderes Programm vor dem Nachladen.
function mitBestand(bestand: Record<string, string>, fremd?: string) {
  const gesetzt = vi.fn(
    (_set: string, _karte: string, feld: string, wert: string) => {
      const gespeichert = `[${wert}]`
      bestand[feld] = fremd ?? gespeichert

      return Promise.resolve({ fehler: null, gespeichert })
    },
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

async function tippeUndVerlasse(text: string) {
  const name = await screen.findByRole('textbox', { name: 'name' })
  await userEvent.type(name, text)
  await act(async () => {
    name.blur()
    await Promise.resolve()
  })
}

describe('ein Entwurf, dessen Schreiben sagt, was es gespeichert hat', () => {
  it('ist erledigt, wenn der Bestand den gespeicherten Wert zeigt', async () => {
    const gesetzt = mitBestand({ name: 'Feuerball', typ: 'Zauber' })
    const { rerender } = render(seite(0))

    await tippeUndVerlasse(' groß')
    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledOnce()
    })
    rerender(seite(1))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
        '[Feuerball groß]',
      )
    })
  })

  it('zeigt einen fremden Wert danach als Konflikt', async () => {
    const gesetzt = mitBestand({ name: 'Feuerball', typ: 'Zauber' }, 'Eisball')
    const { rerender } = render(seite(0))

    await tippeUndVerlasse(' groß')
    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledOnce()
    })
    rerender(seite(1))

    expect(
      await screen.findByText(
        einsetzen(texte.fremdGeaendert, { wert: 'Eisball' }),
      ),
    ).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
      'Feuerball groß',
    )
  })
})
