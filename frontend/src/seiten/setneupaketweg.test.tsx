// **Fehlt ein Stil, führt der Satz zu den Paketen — und zurück** — UX-05
// S07.NAT-04 (#1677):
// *Fehlender Stil bietet passenden Ressourcenweg mit Rückkehr.*
//
// Bis hierher sagte der Satz, dass die Pakete „über die Paketseite kommen",
// und ließ den Weg dorthin offen. **Die Rückkehr ist die der Seitengeschichte**
// (⌘[, UX-03.AK-02): die Seite der Sets bleibt stehen, und ein halb getippter
// Name mit ihr.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { SetNeu } from './SetNeu'
import { Sets } from './Sets'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
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
  const zuDenPaketen = vi.fn()

  belegeApi({
    Pakete: vi.fn().mockResolvedValue({ pakete, fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <SetNeu angelegt={vi.fn()} zuDenPaketen={zuDenPaketen} />
    </Sprachraum>,
  )

  return zuDenPaketen
}

describe('der Weg zu den Paketen', () => {
  it('steht da, wo kein Spiel installiert ist', async () => {
    const zuDenPaketen = zeige([])

    await userEvent.click(
      await screen.findByRole('button', { name: worte.neuZuDenPaketen }),
    )

    expect(zuDenPaketen).toHaveBeenCalledTimes(1)
  })

  it('steht da, wo dem Spiel ein Stil fehlt', async () => {
    const zuDenPaketen = zeige([paket('mse-game', 'probe')])

    await userEvent.click(
      await screen.findByRole('button', { name: worte.neuZuDenPaketen }),
    )

    expect(zuDenPaketen).toHaveBeenCalledTimes(1)
    expect(screen.getByText(/⌘\[/)).toBeInTheDocument()
  })

  it('fehlt, wo nichts fehlt', async () => {
    zeige([paket('mse-game', 'probe'), paket('mse-style', 'probe-schlicht')])

    // **Gewartet wird auf die Pakete** — der Stil steht erst mit ihnen da,
    // seit UXR-04 (#1912) in einer Auswahl, seit UXR-F1 (#2007) in einer mit
    // Suche.
    expect(
      await screen.findByRole('combobox', { name: worte.neuStil }),
    ).toHaveValue('probe-schlicht')

    expect(
      screen.queryByRole('button', { name: worte.neuZuDenPaketen }),
    ).toBeNull()
  })
})

// **Und die Seite der Sets reicht den Weg durch** — zu den Paketen und nicht
// irgendwohin.
it('führt von der Seite der Sets auf die Paketseite', async () => {
  const wechseln = vi.fn()

  belegeApi({
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    Pakete: vi.fn().mockResolvedValue({ pakete: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Sets
        waehlen={vi.fn()}
        speichern={stillerSpeicherlauf()}
        wechseln={wechseln}
      />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: worte.neuZuDenPaketen }),
  )

  expect(wechseln).toHaveBeenCalledWith('pakete')
})
