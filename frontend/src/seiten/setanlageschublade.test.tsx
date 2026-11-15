// **Ein Set anlegen — ab dem ersten in einer Schublade** — AP-10
// (#2367),
// Befund B-19, Entwurf `Wege09`: „In einer Schublade, ab dem ersten Set; ⌘N
// öffnet sie und setzt den Fokus in den Namen.“ Ohne Set bleibt das Anlegen
// offen auf der Seite — der leere Einstieg erklärt sich selbst.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { PaketDTO, SetDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { Sets } from './Sets'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets

const satz: SetDTO = {
  id: '1',
  name: 'Tischprobe',
  spiel: 'magic',
  stil: 'magic-m15',
  karten: 3,
  herkunft: '',
  herkunftFehlt: false,
}

function paket(art: string, name: string): PaketDTO {
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
    rang: 0,
    artrang: 0,
    herkunftsart: 'mitgeliefert',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
    hauptdatei: '',
  }
}

function zeige(sets: SetDTO[], neuanstoss = false) {
  const SetAnlegen = vi.fn().mockResolvedValue({ id: '9', fehler: null })
  belegeApi({
    Sets: () => Promise.resolve({ sets, fehler: null }),
    Pakete: () =>
      Promise.resolve({
        pakete: [
          paket('mse-game', 'probe'),
          paket('mse-style', 'probe-schlicht'),
        ],
        befunde: [],
        fehler: null,
      }),
    SetAnlegen,
  })
  const waehlen = vi.fn()
  const neuAngestossen = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={waehlen}
        speichern={stillerSpeicherlauf()}
        neuanstoss={neuanstoss}
        neuAngestossen={neuAngestossen}
      />
    </Sprachraum>,
  )

  return { waehlen, SetAnlegen }
}

const schublade = () => screen.getByRole('dialog', { name: worte.neuTitel })

describe('das Anlegen eines Sets', () => {
  it('steht ab dem ersten Set in einer Schublade, die „Neues Set“ öffnet', async () => {
    zeige([satz])
    expect(await screen.findByText('Tischprobe')).toBeInTheDocument()
    expect(screen.queryByLabelText(worte.neuName)).toBeNull()

    await userEvent.click(screen.getByRole('button', { name: /Neues Set/ }))

    // **Gewartet wird auf den Fokus** — Mantines Falle setzt ihn nach dem
    // Öffnen (`data-autofocus`).
    await waitFor(() => {
      expect(within(schublade()).getByLabelText(worte.neuName)).toHaveFocus()
    })
  })

  // **Ein Name, nicht zwei** — der Titel der Schublade nennt sie; die
  // Überschrift des Anlegens darin steht nur für das Vorleseprogramm da
  // (gesehen am Fenster: „Ein neues Set“ zweimal untereinander).
  it('nennt sich in der Schublade nur einmal sichtbar', async () => {
    zeige([satz], true)

    const alle = await waitFor(() =>
      within(schublade()).getAllByRole('heading', { name: worte.neuTitel }),
    )
    expect(alle).toHaveLength(2)
    expect(
      alle.filter(
        (ueberschrift) => !ueberschrift.classList.contains('sr-only'),
      ),
    ).toHaveLength(1)
  })

  it('öffnet die Schublade mit ⌘N und setzt den Fokus in den Namen', async () => {
    zeige([satz], true)

    await waitFor(() => {
      expect(within(schublade()).getByLabelText(worte.neuName)).toHaveFocus()
    })
  })

  it('schließt nach dem Anlegen und schlägt das neue Set auf', async () => {
    const { waehlen, SetAnlegen } = zeige([satz], true)
    const name = await waitFor(() =>
      within(schublade()).getByLabelText(worte.neuName),
    )

    await userEvent.type(name, 'Mein Set')
    await userEvent.click(
      within(schublade()).getByRole('button', { name: worte.neuAnlegen }),
    )

    await waitFor(() => {
      expect(waehlen).toHaveBeenCalledWith('9')
    })
    expect(SetAnlegen).toHaveBeenCalled()
    expect(screen.queryByRole('dialog', { name: worte.neuTitel })).toBeNull()
  })

  it('behält eine halb getippte Eingabe über das Schließen hinweg', async () => {
    zeige([satz])
    await userEvent.click(
      await screen.findByRole('button', { name: /Neues Set/ }),
    )
    await userEvent.type(
      within(schublade()).getByLabelText(worte.neuName),
      'Halb',
    )
    await userEvent.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByRole('dialog', { name: worte.neuTitel })).toBeNull()
    })

    await userEvent.click(screen.getByRole('button', { name: /Neues Set/ }))

    expect(within(schublade()).getByLabelText(worte.neuName)).toHaveValue(
      'Halb',
    )
  })

  // **Der leere Einstieg** — ohne Set steht das Anlegen offen da und
  // erklärt den ersten Schritt.
  it('steht ohne Set offen auf der Seite', async () => {
    zeige([])
    expect(await screen.findByText(worte.neuTitel)).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).toBeNull()
  })
})
