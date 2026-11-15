// **Die Spielauswahl nur, wenn es etwas zu wählen gibt** — UX-05.AK-04
// (#1677).
//
// *„Existiert nur ein nutzbares Spielpaket, entfällt die inhaltsleere
// Spielauswahl. Bei mehreren tatsächlich nutzbaren Paketen erscheint sie
// wieder."* **Nutzbar heißt: es gibt einen Stil dazu** — ohne Stil legt MSE
// kein Set an (`Set::validate`, `src/data/set.cpp`, Fassung f9d9356 vom
// 2021-07-06), und ein Spiel, aus dem sich kein Set machen lässt, ist keine
// Wahl.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sets } from './Sets'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets

function paket(art: string, name: string) {
  return {
    art,
    name,
    anzeigename: name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
  }
}

function zeige(pakete: ReturnType<typeof paket>[]) {
  const SetAnlegen = vi.fn().mockResolvedValue({ id: '9', fehler: null })

  belegeApi({
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    Pakete: vi.fn().mockResolvedValue({ pakete, fehler: null }),
    SetAnlegen,
  })

  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={vi.fn()}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )

  return SetAnlegen
}

describe('Die Spielauswahl beim neuen Set', () => {
  it('entfällt, wenn nur ein Spiel einen Stil hat — und das Set bekommt dieses', async () => {
    const SetAnlegen = zeige([
      paket('mse-game', 'leer'),
      paket('mse-game', 'probe'),
      paket('mse-style', 'probe-schlicht'),
    ])

    // **Gewartet wird auf die Pakete**, nicht auf das Formular: eine
    // fehlende Auswahl vor ihrer Ankunft wäre keine Aussage. Seit dem
    // Assistenten (#1677) heißt „entfällt": der Schritt „Art" steht nicht da,
    // und aus drei Schritten werden zwei. Der Stil steht seit UXR-04 (#1912)
    // in einer Auswahl, seit UXR-F1 (#2007) in einer mit Suche — vorgewählt.
    expect(
      await screen.findByRole('combobox', { name: worte.neuStil }),
    ).toHaveValue('probe-schlicht')

    expect(screen.queryByRole('group', { name: worte.neuArtFrage })).toBeNull()
    expect(
      within(
        screen.getByRole('list', { name: worte.neuSchritte }),
      ).getAllByRole('listitem'),
    ).toHaveLength(2)

    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))
    await userEvent.type(await screen.findByLabelText(worte.neuName), 'Eins')
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    )

    await waitFor(() => {
      expect(SetAnlegen).toHaveBeenCalledWith('Eins', 'probe', 'probe-schlicht')
    })
  })

  it('erscheint wieder, sobald zwei Spiele einen Stil haben', async () => {
    zeige([
      paket('mse-game', 'probe'),
      paket('mse-style', 'probe-schlicht'),
      paket('mse-game', 'magic'),
      paket('mse-style', 'magic-m15'),
    ])

    const auswahl = await screen.findByRole('group', {
      name: worte.neuArtFrage,
    })

    expect(
      within(auswahl)
        .getAllByRole('radio')
        .map((feld) => (feld as HTMLInputElement).value),
      // **In MSE2s Ordnung**: bei gleichem Rang entscheidet der Name.
    ).toEqual(['magic', 'probe'])
  })

  it('bleibt, wenn kein Spiel einen Stil hat — dann sagt der Satz darunter, was fehlt', async () => {
    zeige([paket('mse-game', 'probe'), paket('mse-game', 'magic')])

    const auswahl = await screen.findByRole('group', {
      name: worte.neuArtFrage,
    })

    expect(within(auswahl).getAllByRole('radio')).toHaveLength(2)
  })
})
