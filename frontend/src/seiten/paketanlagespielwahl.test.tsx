// **Das Spiel beim Anlegen eines Pakets ist eine Auswahl mit Suche** —
// UXR-F1 (#2007),
// R7: die Spiele wachsen mit dem, was jemand bezieht. **Die erste Zeile ist
// „kein Spiel"**, und sie gilt, solange niemand ein Spiel wählt.
//
// Neben `paketanlage.test.tsx`, damit keine der beiden Dateien die 400 Zeilen
// aus §2 reißt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { Paketanlage } from './Paketanlage'

import type { PaketDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const leer = kataloge.de.seiten.pakete.anlegenSpielLeer

const spiel: PaketDTO = {
  art: 'mse-game',
  name: 'eigenes',
  anzeigename: 'eigenes',
  version: '2026-09-04',
  pfad: '/p/eigenes.mse-game',
  herkunft: '',
  pruefsumme: '',
  gruppe: '',
  rang: 0,
  braucht: [],
  artrang: 30,
  herkunftsart: 'mitgeliefert' as const,
  hauptdatei: 'style',
  geaendert: false,
  gebrauchtVon: [],
  benutzt: 0,
}

describe('Die Spielwahl der Paketanlage', () => {
  it('bietet kein Spiel und die installierten an, und sucht darin', async () => {
    const anlegen = vi.fn().mockResolvedValue({
      paket: { ...spiel, art: 'mse-style', name: 'x' },
      abgebrochen: false,
      fehler: null,
    })

    belegeApi({
      PaketArten: () =>
        Promise.resolve([
          { art: 'mse-game', brauchtSpiel: false },
          { art: 'mse-style', brauchtSpiel: true },
        ]),
      PaketAnlegen: anlegen,
    })

    render(
      <Sprachraum sprache="de">
        <Paketanlage
          spiele={[spiel, { ...spiel, name: 'anderes' }]}
          angelegt={vi.fn()}
        />
      </Sprachraum>,
    )
    await effekteNachholen()
    await screen.findByRole('option', { name: /Stil/ })

    await userEvent.selectOptions(
      screen.getByLabelText(/Art des neuen Pakets/),
      'mse-style',
    )

    const wahl = await screen.findByRole('combobox', {
      name: /Spiel des neuen Pakets/,
    })

    // **Ohne Wahl gilt „kein Spiel"** — mit seinem Namen im Feld.
    expect(wahl).toHaveValue(leer)

    await userEvent.click(wahl)

    // **Nur die Zeilen der Liste** — die Auswahl der Art daneben ist ein
    // natives Feld und hat ihre eigenen `option`.
    const zeilen = () =>
      within(screen.getByRole('listbox'))
        .getAllByRole('option')
        .map((zeile) => zeile.textContent)

    expect(zeilen()).toEqual([leer, 'eigenes', 'anderes'])

    await userEvent.keyboard('ander')

    expect(zeilen()).toEqual(['anderes'])

    await userEvent.keyboard('{Enter}')
    await userEvent.type(screen.getByLabelText(/Kurzname/), 'x')
    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    await screen.findByText(/Angelegt: x/)

    expect(anlegen).toHaveBeenCalledWith('mse-style', 'x', '', 'anderes')
  })
})
