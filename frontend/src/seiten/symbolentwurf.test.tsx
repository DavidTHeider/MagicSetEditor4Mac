// **Ein Einschub aus der Symbolpalette geht durch den Entwurf** — Befund 1
// des Zweitblicks über Phase 0.7.0-I zu U02b
// (#2179).
//
// Die Palette liest, was im Feld steht — also den Entwurf —, setzt das
// Symbol ein und schrieb bis dahin **ohne Bedingung**. Stand das Feld im
// Konflikt, weil ein anderes Programm denselben Wert geändert hatte, ging
// der fremde Wert mit einem Klick verloren, ohne dass jemand „Meinen
// Entwurf speichern" gewählt hätte. Ein Einschub ist Tippen und Verlassen
// in einem (`useEntwurf.uebernehmen`): im Konflikt nur Tippen, sonst
// Schreiben mit den Werten, die der Entwurf kennt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO, KarteDTO, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const entwurf = kataloge.de.bausteine.entwurf

const vorschau: VorschauAntwort = {
  vorschau: {
    stil: '',
    gezeichnet: '',
    bild: '/vorschau/AAAA.png',
    breite: 10,
    hoehe: 14,
    dpi: 150,
    befunde: [],
    flaechen: [{ feld: 'kosten', links: 0, oben: 0, breite: 10, hoehe: 5 }],
    nennungen: [],
    vorgaben: [],
  },
}

const form: FeldbeschreibungDTO = {
  name: 'kosten',
  aufschrift: 'kosten',
  art: 'text',
  mehrzeilig: false,
  wahlen: [],
  beschreibung: '',
  vorgabe: '',
  teilfelder: [],
  reihe: '',
  stelle: 0,
  editierbar: true,
  gespeichert: true,
}

const karte = (kosten: string): KarteDTO => ({
  id: '5',
  nummer: 0,
  werte: [{ feld: 'kosten', wert: kosten }],
})

function zeige(kosten: string) {
  const FeldSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi({
    Vorschau: () => Promise.resolve(vorschau),
    Probevorschau: () => Promise.resolve(vorschau),
    Symbolpalette: () =>
      Promise.resolve({
        eintraege: [
          {
            art: 'code',
            code: 'W',
            beschriftung: 'Weißes Mana',
            rueckfrage: '',
            bild: 'data:image/png;base64,BBBB',
            kinder: [],
          },
        ],
        roh: true,
        fehler: null,
      }),
    FeldSetzen,
  })

  const seite = (wert: string) => (
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte(wert)}
        formen={[form]}
        geaendert={() => undefined}
      />
    </Sprachraum>
  )

  const { rerender } = render(seite(kosten))

  return {
    FeldSetzen,
    fremd: (wert: string) => {
      rerender(seite(wert))
    },
  }
}

describe('Ein Einschub aus der Symbolpalette', () => {
  it('schreibt mit den Werten, die der Entwurf kennt', async () => {
    const { FeldSetzen } = zeige('R')

    const feld = await screen.findByLabelText('kosten')
    await userEvent.click(feld)
    await userEvent.click(
      await screen.findByRole('button', { name: 'Weißes Mana' }),
    )

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith('1', '5', 'kosten', 'RW', [
        'R',
        'RW',
      ])
    })
  })

  it('schreibt im Konflikt nicht, sondern setzt das Symbol in den Entwurf', async () => {
    const { FeldSetzen, fremd } = zeige('R')

    const feld = await screen.findByLabelText('kosten')
    await userEvent.click(feld)
    await userEvent.type(feld, 'R')
    fremd('U')

    // **Erst der Konflikt, positiv** — sonst prüfte die Verneinung darunter
    // einen Editor, der den fremden Wert gar nicht gesehen hat.
    expect(
      await screen.findByRole('button', { name: entwurf.behalten }),
    ).toBeInTheDocument()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Weißes Mana' }),
    )

    await waitFor(() => {
      expect(screen.getByLabelText('kosten')).toHaveValue('RRW')
    })
    expect(FeldSetzen).not.toHaveBeenCalled()

    // **Und „behalten" schreibt ihn ausdrücklich** — mit dem fremden Wert
    // unter den bekannten, denn jetzt ist er gesehen.
    await userEvent.click(
      screen.getByRole('button', { name: entwurf.behalten }),
    )

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'kosten',
        'RRW',
        expect.arrayContaining(['R', 'U', 'RRW']),
      )
    })
  })
})
