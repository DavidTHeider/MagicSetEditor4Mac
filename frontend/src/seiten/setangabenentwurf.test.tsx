// **Ein Entwurf in den Setangaben überlebt fremde Änderungen** — U03
// (#2120).
//
// Dieselbe Klasse wie U01 im Karteneditor: die Lasche trug den
// Verlaufsstand im `key`, und jede Änderung irgendwo baute das Formular neu
// auf. Geprüft über `Lascheninhalt`, denn **dort** stand der Schlüssel.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO, Umschlag } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Lascheninhalt } from './Lascheninhalt'

const texte = kataloge.de.bausteine.entwurf

afterEach(() => {
  raeumeApi()
})

function form(name: string): FeldbeschreibungDTO {
  return {
    name,
    aufschrift: name,
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
}

// bestaende ist, was je Set im Bestand steht — ein anderes Programm ändert es.
function mitBestaenden(bestaende: Record<string, Record<string, string>>) {
  const gesetzt = vi.fn<
    (setID: string, feld: string, wert: string) => Promise<Umschlag>
  >(() => Promise.resolve({ fehler: null }))

  belegeApi({
    Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
    Setangaben: (setID: string) =>
      Promise.resolve({
        felder: [form('title'), form('set code')],
        werte: Object.entries(bestaende[setID] ?? {}).map(([feld, wert]) => ({
          feld,
          wert,
        })),
        kopf: {
          mseFassung: '',
          spielfassung: '',
          stilfassung: '',
          versionskontrolle: '',
          apprenticecode: '',
        },
        fehler: null,
      }),
    SetwertSetzen: gesetzt,
  })

  return gesetzt
}

function lasche(setID: string, stand: number) {
  return (
    <Sprachraum sprache="de">
      <Lascheninhalt
        seite="setangaben"
        wechseln={vi.fn()}
        setID={setID}
        karteID={undefined}
        oeffneKarte={vi.fn()}
        oeffneSymbol={vi.fn()}
        bestandGeaendert={vi.fn()}
        verlaufsstand={stand}
        verlaufsfehler={undefined}
      />
    </Sprachraum>
  )
}

describe('ein Entwurf in den Setangaben', () => {
  it('bleibt mit Fokus stehen, wenn ein anderes Feld fremd geändert wird', async () => {
    const bestaende = { '1': { title: 'Saat', 'set code': 'PRB' } }
    mitBestaenden(bestaende)

    const { rerender } = render(lasche('1', 0))

    const titel = await screen.findByRole('textbox', { name: 'title' })
    await waitFor(() => {
      expect(titel).toHaveValue('Saat')
    })
    await userEvent.type(titel, ' zwei')

    bestaende['1']['set code'] = 'PRC'
    rerender(lasche('1', 1))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'set code' })).toHaveValue(
        'PRC',
      )
    })
    expect(screen.getByRole('textbox', { name: 'title' })).toBe(titel)
    expect(titel).toHaveValue('Saat zwei')
    expect(document.activeElement).toBe(titel)
  })

  it('sagt, wenn dasselbe Feld fremd geändert wurde, und schreibt nicht still darüber', async () => {
    const bestaende = { '1': { title: 'Saat', 'set code': 'PRB' } }
    const gesetzt = mitBestaenden(bestaende)

    const { rerender } = render(lasche('1', 0))

    const titel = await screen.findByRole('textbox', { name: 'title' })
    await waitFor(() => {
      expect(titel).toHaveValue('Saat')
    })
    await userEvent.type(titel, ' zwei')

    bestaende['1'].title = 'Ernte'
    rerender(lasche('1', 1))

    expect(
      await screen.findByText(
        einsetzen(texte.fremdGeaendert, { wert: 'Ernte' }),
      ),
    ).toBeInTheDocument()

    await act(async () => {
      titel.blur()
      await Promise.resolve()
    })
    expect(gesetzt).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: texte.behalten }))

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledWith(
        '1',
        'title',
        'Saat zwei',
        expect.any(Array),
      )
    })
  })

  it('räumt den Entwurf beim Wechsel des Sets', async () => {
    const bestaende = {
      '1': { title: 'Saat', 'set code': 'PRB' },
      '2': { title: 'Ernte', 'set code': 'ERN' },
    }
    const gesetzt = mitBestaenden(bestaende)

    const { rerender } = render(lasche('1', 0))

    const titel = await screen.findByRole('textbox', { name: 'title' })
    await waitFor(() => {
      expect(titel).toHaveValue('Saat')
    })
    await userEvent.type(titel, ' zwei')

    rerender(lasche('2', 0))

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'title' })).toHaveValue(
        'Ernte',
      )
    })
    expect(
      screen.queryByRole('button', { name: texte.behalten }),
    ).not.toBeInTheDocument()
    expect(gesetzt).not.toHaveBeenCalled()
  })
})
